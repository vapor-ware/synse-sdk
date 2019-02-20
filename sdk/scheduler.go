// Synse SDK
// Copyright (c) 2019 Vapor IO
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package sdk

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/vapor-ware/synse-server-grpc/go"

	"github.com/vapor-ware/synse-sdk/sdk/errors"

	log "github.com/Sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk/config"
	"golang.org/x/time/rate"
)

const (
	modeSerial   = "serial"
	modeParallel = "parallel"
)

// ListenerCtx is the context needed for a listener function to be called
// and retried at a later time if it errors out after the listener goroutine
// is initially dispatched.
type ListenerCtx struct {
	// handler is the DeviceHandler that defines the handler function.
	handler *DeviceHandler

	// device is the Device that is being listened to via the listener.
	device *Device

	// restarts is the number of times the listener has been restarted.
	restarts int
}

// NewListenerCtx creates a new ListenerCtx for the given handler and device.
func NewListenerCtx(handler *DeviceHandler, device *Device) *ListenerCtx {
	return &ListenerCtx{
		handler:  handler,
		device:   device,
		restarts: 0,
	}
}

// Scheduler is the plugin component which runs the read, write, and
// listen jobs to get data from devices and write data to devices.
type Scheduler struct {
	// Plugin component references.
	deviceManager *deviceManager
	stateManager  *StateManager

	// config is the configuration that is used by the scheduler.
	config *config.PluginSettings

	// serialLock is a lock that is used around reads/writes when
	// the scheduler is run in serial mode.
	serialLock *sync.Mutex

	// limiter is a rate limiter for making requests.
	limiter *rate.Limiter

	// writeChan is the channel that is used to queue write actions for
	// devices.
	writeChan chan *WriteContext

	// stop is a channel used to signal that the scheduler should stop.
	// This is generally used for graceful shutdown.
	stop chan struct{}
}

// NewScheduler creates a new instance of the plugin's scheduler component.
func NewScheduler(conf *config.PluginSettings, dm *deviceManager, sm *StateManager) *Scheduler {
	var limiter *rate.Limiter
	// fixme: test if this is actually checking the same thing
	if conf.Limiter != nil && conf.Limiter != (&config.LimiterSettings{}) {
		// todo: check if zero-valued, set defaults if so.
		limiter = rate.NewLimiter(
			rate.Limit(conf.Limiter.Rate),
			conf.Limiter.Burst,
		)
	}

	return &Scheduler{
		deviceManager: dm,
		stateManager:  sm,
		config:        conf,
		limiter:       limiter,
		serialLock:    &sync.Mutex{},
		writeChan:     make(chan *WriteContext, conf.Write.QueueSize),
		stop:          make(chan struct{}),
	}
}

// registerActions registers pre-run (setup) and post-run (teardown) actions
// for the scheduler.
func (scheduler *Scheduler) registerActions(plugin *Plugin) {
	// Register post-run actions.
	plugin.RegisterPostRunActions(
		&PluginAction{
			Name:   "Stop scheduler",
			Action: func(plugin *Plugin) error { return scheduler.Stop() },
		},
	)
}

// Start starts the scheduler.
func (scheduler *Scheduler) Start() {
	log.Info("[scheduler] starting...")

	go scheduler.scheduleReads()
	go scheduler.scheduleWrites()
	go scheduler.scheduleListen()
}

// Stop the scheduler.
func (scheduler *Scheduler) Stop() error {
	log.Info("[scheduler] stopping...")

	close(scheduler.stop)
	return nil
}

// Write queues up a write request into the scheduler's write queue.
// fixme: instead of taking the payload, just take the device?
func (scheduler *Scheduler) Write(payload *synse.V3WritePayload) (map[string]*synse.V3WriteData, error) {
	devices := scheduler.deviceManager.GetDevices(DeviceSelectorToTags(payload.Selector)...)

	if len(devices) > 1 {
		// fixme: better err handling
		return nil, fmt.Errorf("cannot write to more than one device at a time")
	}

	if len(devices) == 0 {
		// fixme: better err handling
		return nil, fmt.Errorf("no such device")
	}

	device := devices[0]
	if !device.IsWritable() {
		// fixme: better err handling
		return nil, fmt.Errorf("writing not enabled for device")
	}

	var response = make(map[string]*synse.V3WriteData)
	for _, data := range payload.Data {
		t := newTransaction()
		t.setStatusPending()

		// Map the transaction ID to the write context for the response.
		response[t.id] = data

		// Queue up the write.
		scheduler.writeChan <- &WriteContext{
			transaction: t,
			device:      device.id,
			data:        data,
		}
	}

	return response, nil
}

// scheduleReads schedules device reads based on the plugin configuration.
//
// This will do nothing if:
// - Reading is globally disabled for the plugin.
// - No registered device handlers implement a read function.
func (scheduler *Scheduler) scheduleReads() {
	if scheduler.config.Read.Disable {
		log.Info("[scheduler] reading will not be scheduled (reads globally disabled)")
		return
	}

	if !scheduler.deviceManager.HasReadHandlers() {
		log.Info("[scheduler] reading will not be scheduled (no read handlers registered)")
		return
	}

	interval := scheduler.config.Read.Interval
	delay := scheduler.config.Read.Delay
	mode := scheduler.config.Mode

	rlog := log.WithFields(log.Fields{
		"interval": interval,
		"delay":    delay,
		"mode":     mode,
	})

	rlog.Info("[scheduler] starting read scheduling")
	for {
		// If the stop channel is closed, stop the read loop.
		select {
		case <-scheduler.stop:
			break
		}

		var waitGroup sync.WaitGroup

		// Run all single device reads.
		for _, device := range scheduler.deviceManager.devices {
			// Increment the WaitGroup counter for each device.
			waitGroup.Add(1)

			// Launch the device read.
			go func(wg *sync.WaitGroup, device *Device) {
				scheduler.read(device)
				wg.Done()
			}(&waitGroup, device)
		}

		// Run all batch device reads.
		for _, handler := range scheduler.deviceManager.handlers {
			// Increment the WaitGroup for each bulk read action.
			waitGroup.Add(1)

			// Launch the bulk read.
			go func(wg *sync.WaitGroup, handler *DeviceHandler) {
				scheduler.bulkRead(handler)
				wg.Done()
			}(&waitGroup, handler)
		}

		// Wait for all device reads to complete.
		waitGroup.Wait()

		if interval != 0 {
			rlog.Debug("[scheduler] sleeping for read interval")
			time.Sleep(interval)
			rlog.Debug("[scheduler] waking up for read interval")
		}
	}
}

// scheduleWrites schedules device writes based on the plugin configuration.
//
// This will do nothing if:
// - Writing is globally disabled for the plugin.
// - No registered device handlers implement a write function.
func (scheduler *Scheduler) scheduleWrites() {
	if scheduler.config.Write.Disable {
		log.Info("[scheduler] writing will not be scheduled (writes globally disabled)")
		return
	}

	if !scheduler.deviceManager.HasWriteHandlers() {
		log.Info("[scheduler] writing will not be scheduled (no write handlers registered)")
		return
	}

	interval := scheduler.config.Write.Interval
	delay := scheduler.config.Write.Delay
	mode := scheduler.config.Mode

	wlog := log.WithFields(log.Fields{
		"interval": interval,
		"delay":    delay,
		"mode":     mode,
	})

	wlog.Info("[scheduler] starting write scheduling")
	for {
		// If the stop channel is closed, stop the write loop.
		select {
		case <-scheduler.stop:
			break
		}

		var waitGroup sync.WaitGroup

		// Check for any pending writes. If any exist, attempt to fulfill
		// the writes and update their transaction state accordingly.
		for i := 0; i < scheduler.config.Write.BatchSize; i++ {
			select {
			case w := <-scheduler.writeChan:
				// Increment the WaitGroup counter for all writes being executed
				// in this batch.
				waitGroup.Add(1)

				// Launch the device write.
				go func(wg *sync.WaitGroup, writeContext *WriteContext) {
					scheduler.write(writeContext)
				}(&waitGroup, w)

			default:
				// If there is nothing to write, do nothing.
			}
		}

		// Wait for all device writes to complete.
		waitGroup.Wait()

		if interval != 0 {
			wlog.Debug("[scheduler] sleeping for write interval")
			time.Sleep(interval)
			wlog.Debug("[scheduler] waking up for write interval")
		}
	}
}

// scheduleListen schedulers device listeners based on the plugin configuration.
//
// This will do nothing if:
// - Listening is globally disabled for the plugin.
// - No registered device handlers implement a listener function.
func (scheduler *Scheduler) scheduleListen() {
	if scheduler.config.Listen.Disable {
		log.Info("[scheduler] listeners will not be scheduled (listening globally disabled)")
		return
	}

	if !scheduler.deviceManager.HasListenerHandlers() {
		log.Info("[scheduler] listeners will not be scheduled (no listener handlers registered)")
		return
	}

	// For each handler which has a listener function defined, get the devices for
	// the handler and start the listener for those devices.
	for _, handler := range scheduler.deviceManager.handlers {
		hlog := log.WithField("handler", handler.Name)

		if handler.Listen != nil {
			hlog.Info("[scheduler] starting listener scheduling")

			// Get the devices for the handler.
			devices := scheduler.deviceManager.GetDevicesForHandler(handler.Name)
			if len(devices) == 0 {
				hlog.Debug("[scheduler] handler has no devices to listen")
				continue
			}

			// For each device, run the listener goroutine.
			for _, device := range devices {
				ctx := NewListenerCtx(handler, device)
				go scheduler.listen(ctx)
			}
		}
	}
}

// read reads from a single device using a handler's Read function.
func (scheduler *Scheduler) read(device *Device) {
	delay := scheduler.config.Read.Delay
	mode := scheduler.config.Mode

	rlog := log.WithFields(log.Fields{
		"delay":  delay,
		"mode":   mode,
		"device": device.id,
	})

	// Rate limiting, if configured. We want to do this before potentially
	// acquiring the serial lock so something isn't holding on to the lock
	// and just waiting.
	if scheduler.limiter != nil {
		if err := scheduler.limiter.Wait(context.Background()); err != nil {
			rlog.WithField("error", err).Error("[scheduler] error with rate limiter")
		}
	}

	// If the device does not get its readings from a bulk read operation, then
	// it is read individually. If a device is read in bulk, it will not be read
	// here; it will be read later via the bulkRead function.
	if !device.handler.supportsBulkRead() {

		// If we are running in serial mode, acquire the serial lock.
		if mode == modeSerial {
			scheduler.serialLock.Lock()
			defer scheduler.serialLock.Unlock()
		}

		// Read from the device.
		response, err := device.Read()
		if err != nil {
			// Check to see if the error is that of unsupported error. If it is, we
			// do not want to log out here (low-interval read polling would cause this
			// to pollute the logs for something that we should already know).
			_, unsupported := err.(*errors.UnsupportedCommandError)
			if !unsupported {
				rlog.Error("[scheduler] failed device read")
			}
		} else {
			scheduler.stateManager.readChan <- response
		}

		// If a delay is configured, wait for the delay before continuing
		// (and relinquishing the lock, if in serial mode).
		if delay != 0 {
			rlog.Debug("[scheduler] sleeping for read delay")
			time.Sleep(delay)
			rlog.Debug("[scheduler] waking up for read delay")
		}
	}

}

// bulkRead reads from multiple devices using a handler's BulkRead function.
func (scheduler *Scheduler) bulkRead(handler *DeviceHandler) {
	delay := scheduler.config.Read.Delay
	mode := scheduler.config.Mode

	rlog := log.WithFields(log.Fields{
		"delay":   delay,
		"mode":    mode,
		"handler": handler.Name,
	})

	// Rate limiting, if configured. We want to do this before potentially
	// acquiring the serial lock so something isn't holding on to the lock
	// and just waiting.
	if scheduler.limiter != nil {
		if err := scheduler.limiter.Wait(context.Background()); err != nil {
			rlog.WithField("error", err).Error("[scheduler] error with rate limiter")
		}
	}

	// If the handler supports bulk reading, execute bulk reads. Devices using the
	// handler will not have been read individually yet.
	if handler.supportsBulkRead() {
		devices := scheduler.deviceManager.GetDevicesForHandler(handler.Name)
		if len(devices) == 0 {
			rlog.Debug("[scheduler] handler has no devices to read")
			return
		}

		// If we are running in serial mode, acquire the serial lock.
		if mode == modeSerial {
			scheduler.serialLock.Lock()
			defer scheduler.serialLock.Unlock()
		}

		response, err := handler.BulkRead(devices)
		if err != nil {
			rlog.WithField("error", err).Error("[scheduler] handler failed bulk read")
		} else {
			for _, readCtx := range response {
				scheduler.stateManager.readChan <- readCtx
			}
		}

		// If a delay is configured, wait for the delay before continuing
		// (and relinquishing the lock, if in serial mode).
		if delay != 0 {
			rlog.Debug("[scheduler] sleeping for bulk read delay")
			time.Sleep(delay)
			rlog.Debug("[scheduler] waking up for bulk read delay")
		}
	}
}

// write writes to devices using a handler's Write function.
func (scheduler *Scheduler) write(writeCtx *WriteContext) {
	delay := scheduler.config.Write.Delay
	mode := scheduler.config.Mode

	wlog := log.WithFields(log.Fields{
		"delay":       delay,
		"mode":        mode,
		"transaction": writeCtx.transaction.id,
		"device":      writeCtx.device,
	})

	// Rate limiting, if configured. We want to do this before potentially
	// acquiring the serial lock so something isn't holding on to the lock
	// and just waiting.
	if scheduler.limiter != nil {
		if err := scheduler.limiter.Wait(context.Background()); err != nil {
			wlog.WithField("error", err).Error("[scheduler] error with rate limiter")
		}
	}

	// If we are running in serial mode, acquire the serial lock.
	if mode == modeSerial {
		scheduler.serialLock.Lock()
		defer scheduler.serialLock.Unlock()
	}

	wlog.Debug("[scheduler] starting device write")

	// Get the device.
	device := scheduler.deviceManager.GetDevice(writeCtx.device)
	if device == nil {
		writeCtx.transaction.setStatusError()
		writeCtx.transaction.message = "no device found with ID: " + writeCtx.device
		wlog.Error("[scheduler] " + writeCtx.transaction.message)
		return
	}

	if !device.IsWritable() {
		writeCtx.transaction.setStatusError()
		writeCtx.transaction.message = "device is not writable: " + writeCtx.device
		wlog.Error("[scheduler] " + writeCtx.transaction.message)
		return
	}

	writeCtx.transaction.setStatusWriting()

	// Write to the device.
	data := decodeWriteData(writeCtx.data)
	err := device.Write(data)
	if err != nil {
		wlog.WithField("error", err).Error("[scheduler] failed to write to device")
		writeCtx.transaction.setStatusError()
		writeCtx.transaction.message = err.Error()
		return
	}
	wlog.Debug("[scheduler] successfully wrote to device")
	writeCtx.transaction.setStatusDone()
}

// listen listens to devices to collect readings using a device's Listen function.
func (scheduler *Scheduler) listen(listenerCtx *ListenerCtx) {
	llog := log.WithFields(log.Fields{
		"handler": listenerCtx.handler.Name,
		"device":  listenerCtx.device.id,
	})

	llog.Info("[scheduler] starting listener for device")

	for {
		// Run the listener fore the device. Pass in the state manager's read channel,
		// as the listener is really just collecting readings.
		err := listenerCtx.handler.Listen(
			listenerCtx.device,
			scheduler.stateManager.readChan,
		)
		if err != nil {
			// Increment the number of restarts.
			listenerCtx.restarts++

			// If a listener function results in error, we want to restart it to try and
			// keep listening. Log the error and re-try listening.
			llog.WithFields(log.Fields{
				"restarts": listenerCtx.restarts,
				"error":    err,
			}).Error("[scheduler] listener failed, will restart and try again")
			continue

		} else {
			// If the listener ended without any error, we take this to mean
			// that it terminated in a way that is considered ok, so we do not
			// want to try and restart. Instead, just stop listening.
			llog.Info("[scheduler] listener completed without error, ending device listen")
			return
		}
	}
}

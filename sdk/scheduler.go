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

	log "github.com/sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/health"
	synse "github.com/vapor-ware/synse-server-grpc/go"
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

// scheduler is the plugin component which runs the read, write, and
// listen jobs to get data from devices and write data to devices.
type scheduler struct {
	// Plugin component references.
	deviceManager *deviceManager
	stateManager  *stateManager

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

// newScheduler creates a new instance of the plugin's scheduler component.
//func NewScheduler(conf *config.PluginSettings, dm *deviceManager, sm *StateManager) *Scheduler {
func newScheduler(plugin *Plugin) *scheduler {
	conf := plugin.config.Settings

	var limiter *rate.Limiter

	// If the limiter is configured and non-0 values (which signify unlimited),
	// set up the limiter.
	if conf.Limiter != nil {
		if conf.Limiter.Rate != 0 || conf.Limiter.Burst != 0 {
			limiter = rate.NewLimiter(
				rate.Limit(conf.Limiter.Rate),
				conf.Limiter.Burst,
			)
		}
	}

	return &scheduler{
		deviceManager: plugin.device,
		stateManager:  plugin.state,
		config:        conf,
		limiter:       limiter,
		serialLock:    &sync.Mutex{},
		writeChan:     make(chan *WriteContext, conf.Write.QueueSize),
		stop:          make(chan struct{}),
	}
}

// registerActions registers pre-run (setup) and post-run (teardown) actions
// for the scheduler.
func (scheduler *scheduler) registerActions(plugin *Plugin) {
	// Register pre-run actions.
	plugin.RegisterPreRunActions(
		&PluginAction{
			Name:   "Register default scheduler health checks",
			Action: scheduler.healthChecks,
		},
	)

	// Register post-run actions.
	plugin.RegisterPostRunActions(
		&PluginAction{
			Name:   "Stop scheduler",
			Action: func(p *Plugin) error { return scheduler.Stop() },
		},
	)
}

// healthChecks defines and registers the scheduler's default health checks with
// the plugin.
func (scheduler *scheduler) healthChecks(plugin *Plugin) error {
	wqh := health.NewPeriodicHealthCheck("write queue health", 30*time.Second, func() error {
		// Determine the percent usage of the write queue.
		pctUsage := (float64(len(scheduler.writeChan)) / float64(cap(scheduler.writeChan))) * 100

		// If the write queue is at 95% usage, we consider it unhealthy; the write
		// queue should be configured to be larger.
		if pctUsage > 95 {
			return fmt.Errorf("write queue usage >95%%, consider increasing size in configuration")
		}
		return nil
	})
	plugin.health.RegisterDefault(wqh)
	return nil
}

// Start starts the scheduler.
func (scheduler *scheduler) Start() {
	log.Info("[scheduler] starting...")

	go scheduler.scheduleReads()
	go scheduler.scheduleWrites()
	go scheduler.scheduleListen()
}

// Stop the scheduler.
func (scheduler *scheduler) Stop() error {
	log.Info("[scheduler] stopping...")

	close(scheduler.stop)
	return nil
}

// Write queues up a write request into the scheduler's write queue.
func (scheduler *scheduler) Write(device *Device, data []*synse.V3WriteData) ([]*synse.V3WriteTransaction, error) {
	if device == nil {
		// fixme: better err handling
		return nil, fmt.Errorf("cannot write to nil device")
	}

	if data == nil {
		return nil, fmt.Errorf("cannot write nil data")
	}

	if !device.IsWritable() {
		// fixme: better err handling
		return nil, fmt.Errorf("writing not enabled for device")
	}

	var response []*synse.V3WriteTransaction
	for _, data := range data {
		t := scheduler.stateManager.newTransaction(device.WriteTimeout)
		t.setStatusPending()

		// Map the transaction ID to the write context for the response.
		response = append(response, &synse.V3WriteTransaction{
			Id:      t.id,
			Device:  device.GetID(),
			Context: data,
			Timeout: device.WriteTimeout.String(),
		})

		// Queue up the write.
		scheduler.writeChan <- &WriteContext{
			transaction: t,
			device:      device.id,
			data:        data,
		}
	}
	return response, nil
}

func (scheduler *scheduler) WriteAndWait(device *Device, data []*synse.V3WriteData) ([]*synse.V3TransactionStatus, error) {
	if device == nil {
		// fixme: better err handling
		return nil, fmt.Errorf("cannot write to nil device")
	}

	if data == nil {
		return nil, fmt.Errorf("cannot write nil data")
	}

	if !device.IsWritable() {
		// fixme: better err handling
		return nil, fmt.Errorf("writing not enabled for device")
	}
	var response []*synse.V3TransactionStatus
	var txns []*transaction
	var waitGroup sync.WaitGroup

	for _, data := range data {
		t := scheduler.stateManager.newTransaction(device.WriteTimeout)
		t.setStatusPending()

		txns = append(txns, t)

		// Queue up the write.
		scheduler.writeChan <- &WriteContext{
			transaction: t,
			device:      device.id,
			data:        data,
		}

		waitGroup.Add(1)
		go func(t *transaction, wg *sync.WaitGroup) {
			t.wait()
			wg.Done()
		}(t, &waitGroup)
	}

	waitGroup.Wait()

	for _, t := range txns {
		response = append(response, t.encode())
	}
	return response, nil
}

// scheduleReads schedules device reads based on the plugin configuration.
//
// This will do nothing if:
// - Reading is globally disabled for the plugin.
// - No registered device handlers implement a read function.
func (scheduler *scheduler) scheduleReads() {
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
		default:
			// no stop signal
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
			time.Sleep(interval)
		}
	}
}

// scheduleWrites schedules device writes based on the plugin configuration.
//
// This will do nothing if:
// - Writing is globally disabled for the plugin.
// - No registered device handlers implement a write function.
func (scheduler *scheduler) scheduleWrites() {
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
		log.WithFields(log.Fields{
			"queue": len(scheduler.writeChan),
			"cap":   cap(scheduler.writeChan),
		}).Debug("[scheduler] writing")
		// If the stop channel is closed, stop the write loop.
		select {
		case <-scheduler.stop:
			break
		default:
			// no stop signal
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
					wg.Done()
				}(&waitGroup, w)

			default:
				// If there is nothing to write, do nothing.
			}
		}

		// Wait for all device writes to complete.
		waitGroup.Wait()

		if interval != 0 {
			//wlog.Debug("[scheduler] sleeping for write interval")
			time.Sleep(interval)
			//wlog.Debug("[scheduler] waking up for write interval")
		}
	}
}

// scheduleListen schedulers device listeners based on the plugin configuration.
//
// This will do nothing if:
// - Listening is globally disabled for the plugin.
// - No registered device handlers implement a listener function.
func (scheduler *scheduler) scheduleListen() {
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
func (scheduler *scheduler) read(device *Device) {
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
	if !device.handler.CanBulkRead() {

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
			//rlog.Debug("[scheduler] sleeping for read delay")
			time.Sleep(delay)
			//rlog.Debug("[scheduler] waking up for read delay")
		}
	}

}

// bulkRead reads from multiple devices using a handler's BulkRead function.
func (scheduler *scheduler) bulkRead(handler *DeviceHandler) {
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
	if handler.CanBulkRead() {
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
			//rlog.Debug("[scheduler] sleeping for bulk read delay")
			time.Sleep(delay)
			//rlog.Debug("[scheduler] waking up for bulk read delay")
		}
	}
}

// write writes to devices using a handler's Write function.
func (scheduler *scheduler) write(writeCtx *WriteContext) {
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

	// Write to the device. If the device write does not complete within
	// the set time bounds, error out with timeout.
	// See: https://gobyexample.com/timeouts
	writer := make(chan error, 1)
	go func() {
		data := decodeWriteData(writeCtx.data)
		writer <- device.Write(data)
	}()

	// Wait for the write to complete, or timeout.
	// FIXME (etd): this does technically give us a timeout, where we will stop
	//  waiting for the write after the given timeout and just mark the transaction
	//  dead, but thats kinda all this does.
	//
	//  from an upstream perspective, this is fine.. we want to be able to say
	//  "we expect this write to complete after N time, afterwards, timeout and consider
	//  it failed".
	//
	//  from the backend, it does present this to the frontend consumer, but it doesn't
	//  actually stop the write from happening, so while we have "timed out", the write
	//  is still going on in the background and could eventually resolve, which is
	//  not at all what we want to do.
	//
	//  this could be resolved by passing in a context and cancelling, but uhh.. its
	//  complicated and the internet has many mixed feeling-ed blog posts about it.
	//
	//  this relates to #365. I'll need to spend more time thinking this through.
	//  essentially the issue lies in the fact that context cancellation relies on the
	//  fact that you can terminate the work at some point (e.g. if iterating in an
	//  infinite loop.., emitting from a channel, ...), then you can wait on the context
	//  done signal, wait for the fn to finish up, and then you're done. to me, it seems
	//  like the write will need to be waited on either way, so i'm not sure there is
	//  a benefit to using a cancellation context unless writing is handled in a different
	//  manner..
	//
	//  it seems like having a cancelation context could be useful if there is some
	//  retry logic on the write, but thats mostly it..

	log.WithFields(log.Fields{
		"device":  device.GetID(),
		"action":  writeCtx.data.Action,
		"data":    string(writeCtx.data.Data),
		"timeout": device.WriteTimeout,
	}).Debug("[scheduler] writing")

	var err error
	select {
	case writeErr := <-writer:
		err = writeErr
	case <-time.After(device.WriteTimeout):
		// fixme: improve err message
		err = fmt.Errorf("device write timeout")
	}

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
func (scheduler *scheduler) listen(listenerCtx *ListenerCtx) {
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

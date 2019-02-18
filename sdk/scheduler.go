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
	"sync"
	"time"

	"github.com/vapor-ware/synse-sdk/sdk/errors"

	log "github.com/Sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk/config"
	"golang.org/x/time/rate"
)

const (
	modeSerial   = "serial"
	modeParallel = "parallel"
)

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

	// stop is a channel used to signal that the scheduler should stop.
	// This is generally used for graceful shutdown.
	stop chan struct{}
}

// NewScheduler creates a new instance of the plugin's scheduler component.
func NewScheduler(conf *config.PluginSettings, deviceManager *deviceManager, stateManager *StateManager) *Scheduler {
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
		deviceManager: deviceManager,
		stateManager:  stateManager,
		config:        conf,
		limiter:       limiter,
		serialLock:    &sync.Mutex{},
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
	go scheduler.scheduleReads()
	go scheduler.scheduleWrites()
	go scheduler.scheduleListen()
}

// Stop the scheduler.
func (scheduler *Scheduler) Stop() error {
	close(scheduler.stop)

	return nil
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

	// todo: figure out better way to handle this.. exiter channel??
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
				// todo: bulk read
				wg.Done()
			}(&waitGroup, handler)
		}

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
}

func (scheduler *Scheduler) read(device *Device) {
	delay := scheduler.config.Read.Delay
	mode := scheduler.config.Mode

	rlog := log.WithFields(log.Fields{
		"delay":  delay,
		"mode":   mode,
		"device": device.ID(),
	})

	// Rate limiting, if configured. We want to do this before potentially
	// acquiring the serial lock so something isn't holding on to the lock
	// and just waiting.
	if scheduler.limiter != nil {
		if err := scheduler.limiter.Wait(context.Background()); err != nil {
			rlog.WithField("error", err).Error("[scheduler] error with rate limiter")
		}
	}

	// If we are running in serial mode, acquire the serial lock.
	if mode == modeSerial {
		scheduler.serialLock.Lock()
		defer scheduler.serialLock.Unlock()
	}
	// fixme: think about where this goes w.r.t the bulk read check...

	// If the device does not get its readings from a bulk read operation, then
	// it is read individually. If a device is read in bulk, it will not be read
	// here; it will be read later via the bulkRead function.
	if !device.handler.supportsBulkRead() {
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

func (scheduler *Scheduler) bulkRead(handler *DeviceHandler) {

}

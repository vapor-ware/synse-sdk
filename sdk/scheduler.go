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
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk/config"
	"golang.org/x/time/rate"
)

// Scheduler is the plugin component which runs the read, write, and
// listen jobs to get data from devices and write data to devices.
type Scheduler struct {
	// Plugin component references.
	deviceManager *deviceManager

	// config is the configuration that is used by the scheduler.
	config *config.PluginSettings

	// serialLock is a lock that is used around reads/writes when
	// the scheduler is run in serial mode.
	serialLock *sync.Mutex

	// limiter is a rate limiter for making requests.
	limiter *rate.Limiter
}

// NewScheduler creates a new instance of the plugin's scheduler component.
func NewScheduler(conf *config.PluginSettings, deviceManager *deviceManager) *Scheduler {
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
		config:        conf,
		limiter:       limiter,
		serialLock:    &sync.Mutex{},
	}
}

// Start starts the scheduler.
func (scheduler *Scheduler) Start() {
	go scheduler.scheduleReads()
	go scheduler.scheduleWrites()
	go scheduler.scheduleListen()
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
		var waitGroup sync.WaitGroup

		// Run all single device reads.
		for _, device := range scheduler.deviceManager.devices {
			// Increment the WaitGroup counter for each device.
			waitGroup.Add(1)

			// Launch the device read.
			go func(wg *sync.WaitGroup, device *Device) {
				scheduler.read()
				wg.Done()
			}(&waitGroup, device)
		}

		// Run all batch device reads.
		// TODO

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

func (scheduler *Scheduler) read() {
	delay := scheduler.config.Read.Delay
	mode := scheduler.config.Mode

	rlog := log.WithFields(log.Fields{
		"delay": delay,
		"mode":  mode,
	})

}

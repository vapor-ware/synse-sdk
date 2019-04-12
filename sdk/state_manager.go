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
	"fmt"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/health"
	"github.com/vapor-ware/synse-sdk/sdk/output"
	"github.com/vapor-ware/synse-sdk/sdk/utils"
)

// stateManager manages the read and write (transaction) state for plugin devices.
type stateManager struct {
	config        *config.PluginSettings
	readChan      chan *ReadContext
	readings      map[string][]*output.Reading
	readingsCache *cache.Cache
	readingsLock  *sync.RWMutex
	transactions  *cache.Cache
}

// newStateManager creates a new instance of the stateManager.
func newStateManager(conf *config.PluginSettings) *stateManager {
	if conf == nil {
		panic("state manager requires a non-nil config")
	}

	var readingsCache *cache.Cache
	if conf.Cache.Enabled {
		log.WithField("ttl", conf.Cache.TTL).Debug("[state manager] readings cache enabled")
		readingsCache = cache.New(conf.Cache.TTL, conf.Cache.TTL*2)
	}

	return &stateManager{
		config:   conf,
		readChan: make(chan *ReadContext, conf.Read.QueueSize),
		readings: make(map[string][]*output.Reading),
		transactions: cache.New(
			conf.Transaction.TTL,
			conf.Transaction.TTL*2,
		),
		readingsCache: readingsCache,
		readingsLock:  &sync.RWMutex{},
	}
}

// Start starts the StateManager.
func (manager *stateManager) Start() {
	log.Info("[state manager] starting")
	go manager.updateReadings()
}

// registerActions registers pre-run (setup) and post-run (teardown) actions
// for the state manager.
func (manager *stateManager) registerActions(plugin *Plugin) {
	// Register pre-run actions.
	plugin.RegisterPreRunActions(
		&PluginAction{
			Name:   "Register default state manager health checks",
			Action: manager.healthChecks,
		},
	)
}

// healthChecks defines and registers the state manager's default health checks with
// the plugin.
func (manager *stateManager) healthChecks(plugin *Plugin) error {
	rqh := health.NewPeriodicHealthCheck("read queue health", 30*time.Second, func() error {
		// Determine the percent usage of the read queue.
		pctUsage := (float64(len(manager.readChan)) / float64(cap(manager.readChan))) * 100

		// If the read queue is at 95% usage, we consider it unhealthy; the read
		// queue should be configured to be larger.
		if pctUsage > 95 {
			return fmt.Errorf("read queue usage >95%%, consider increasing size in configuration")
		}
		return nil
	})
	plugin.health.RegisterDefault(rqh)
	return nil
}

func (manager *stateManager) updateReadings() {
	// todo: figure out how to test this...
	for {
		// Read from the read channel for incoming readings.
		reading := <-manager.readChan
		id := reading.Device
		readings := reading.Reading

		// Update the reading state.
		manager.readingsLock.Lock()
		manager.readings[id] = readings
		manager.readingsLock.Unlock()

		// Update the local readings cache, if enabled.
		manager.addReadingToCache(reading)
	}
}

// addReadingToCache adds the given reading to the readingsCache, if the plugin
// is configured to enable read caching.
func (manager *stateManager) addReadingToCache(ctx *ReadContext) {
	if manager.config.Cache.Enabled {
		now := utils.GetCurrentTime()
		item, exists := manager.readingsCache.Get(now)
		if !exists {
			newCtxs := []*ReadContext{ctx}
			manager.readingsCache.Set(now, &newCtxs, cache.DefaultExpiration)
		} else {
			cached := item.(*[]*ReadContext)
			*cached = append(*cached, ctx)
		}
	}
}

// GetReadingsForDevice gets the current reading(s) for the specified device from
// the StateManager.
func (manager *stateManager) GetReadingsForDevice(device string) []*output.Reading {
	manager.readingsLock.RLock()
	defer manager.readingsLock.RUnlock()

	return manager.readings[device]
}

// GetOutputsForDevice gets the outputs for a device, based on the outputs associated
// with any readings collected for the device.
func (manager *stateManager) GetOutputsForDevice(device string) []*output.Output {
	readings := manager.GetReadingsForDevice(device)

	var outputs []*output.Output
	for _, r := range readings {
		o := r.GetOutput()
		if o != nil {
			outputs = append(outputs, o)
		}
	}
	return outputs
}

// GetCachedReadings gets the readings in the StateManager's readingsCache. If the plugin
// is not configured to maintain a readings cache, this will just return a dump of the
// current reading state. Once the data has been passed through the given channel, this
// function will close the channel prior to returning.
func (manager *stateManager) GetCachedReadings(start, end string, readings chan *ReadContext) {
	// Whether we exit the function normally or by error, we want to close the channel
	// when we complete to signal to the reader that we are done here.
	defer close(readings)

	// Parse the timestamps for the start/end bounds of the data window.
	startTime, err := utils.ParseRFC3339(start)
	if err != nil {
		// If we can't parse the time, we don't have any business returning data.
		log.WithFields(log.Fields{
			"timestamp": start,
		}).Warn("[state manager] unable to get data: failed to parse timestamp")
		return
	}
	endTime, err := utils.ParseRFC3339(end)
	if err != nil {
		// If we can't parse the time, we don't have any business returning data.
		log.WithFields(log.Fields{
			"timestamp": end,
		}).Warn("[state manager] unable to get data: failed to parse timestamp")
		return
	}

	// If read caching is disable, dump the current state; otherwise, dump the
	// reading cache contents.
	if manager.config.Cache.Enabled {
		manager.dumpCachedReadings(startTime, endTime, readings)
	} else {
		manager.dumpCurrentReadings(readings)
	}

}

// dumpCachedReadings dumps the cached reading
func (manager *stateManager) dumpCachedReadings(start, end time.Time, readings chan *ReadContext) {
	for timestamp, item := range manager.readingsCache.Items() {
		ts, err := utils.ParseRFC3339(timestamp)
		if err != nil {
			// If we can't parse the timestamp from the cache, an error is logged
			// and we move on. We should always be using RFC3339 formatted timestamps
			// as keys when things get inserted, so if we find something in there that
			// does not conform, it means something is wrong and we should not use it
			// (data corruption, something added incorrectly, ...)
			log.Error("[cache] failed to parse RFC3339 timestamp from cache - ignoring")
			continue
		}

		// If we have a start bound, check that the cached items are
		// within that bound. If not, ignore them.
		if !start.IsZero() && ts.Before(start) {
			continue
		}

		// If we have an end bound, check that the cached items are
		// within that bound. If not, ignore them.
		if !end.IsZero() && ts.After(end) {
			continue
		}

		// Pass the read contexts to the channel
		ctxs := item.Object.(*[]*ReadContext)
		for _, ctx := range *ctxs {
			readings <- ctx
		}
	}
}

// dumpCurrentReadings dumps the current readings out to the provided channel.
func (manager *stateManager) dumpCurrentReadings(readings chan *ReadContext) {
	for id, data := range manager.GetReadings() {
		readings <- &ReadContext{
			Device:  id,
			Reading: data,
		}
	}
}

// GetReadings gets a copy of the entire current readings state in the StateManager.
func (manager *stateManager) GetReadings() map[string][]*output.Reading {
	readings := make(map[string][]*output.Reading)
	manager.readingsLock.RLock()
	defer manager.readingsLock.RUnlock()

	// Iterate over the StateManager's readings map to make a copy of it.
	// We want a copy since the underlying data should only be accessed
	// in a locked context.
	for k, v := range manager.readings {
		readings[k] = v
	}
	return readings
}

// newTransaction creates a new transaction and adds it to the transaction cache.
func (manager *stateManager) newTransaction(timeout time.Duration, customID string) (*transaction, error) {
	t := newTransaction(timeout, customID)
	_, exists := manager.transactions.Get(t.id)
	if exists {
		return nil, fmt.Errorf("transaction with ID %s already exists", t.id)
	}
	manager.transactions.Set(t.id, t, cache.DefaultExpiration)
	return t, nil
}

// getTransaction gets the transaction with the specified ID from the transaction cache.
// If the specified transaction was not found, nil is returned.
func (manager *stateManager) getTransaction(id string) *transaction {
	t, found := manager.transactions.Get(id)
	if found {
		return t.(*transaction)
	}
	log.WithFields(log.Fields{
		"id": id,
	}).Warn("[state manager] transaction not found")
	return nil
}

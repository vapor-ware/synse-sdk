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

	log "github.com/Sirupsen/logrus"
	"github.com/patrickmn/go-cache"
	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/health"
	"github.com/vapor-ware/synse-sdk/sdk/output"
	"github.com/vapor-ware/synse-sdk/sdk/utils"
)

// todo: for readings, check if reading is enabled on the device

// fixme: better way of defining this?
// cacheContexts is how ReadContexts are stored in the readings cache. Since
// we may want to filter readings based on the timestamp they were added, we
// want to store the ReadContexts against a timestamp key. In order to support
// multiple contexts at a given time, we store them as a slice.
type cacheContexts []*ReadContext

// StateManager manages the read and write (transaction) state for plugin devices.
type StateManager struct {
	readChan chan *ReadContext

	readings      map[string][]*output.Reading
	transactions  *cache.Cache
	readingsCache *cache.Cache

	readingsLock *sync.RWMutex

	// TODO; figure out which bits of config this needs
	config *config.PluginSettings
}

func NewStateManager(conf *config.PluginSettings) *StateManager {
	var readingsCache *cache.Cache
	if conf.Cache.Enabled {
		// todo: logging
		readingsCache = cache.New(conf.Cache.TTL, conf.Cache.TTL*2)
	}

	return &StateManager{
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
func (manager *StateManager) Start() {
	go manager.updateReadings()
}

// registerActions registers pre-run (setup) and post-run (teardown) actions
// for the state manager.
func (manager *StateManager) registerActions(plugin *Plugin) {
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
func (manager *StateManager) healthChecks(plugin *Plugin) error {
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
	plugin.healthManager.RegisterDefault(rqh)
	return nil
}

func (manager *StateManager) updateReadings() {
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
func (manager *StateManager) addReadingToCache(ctx *ReadContext) {
	if manager.config.Cache.Enabled {
		now := utils.GetCurrentTime()
		item, exists := manager.readingsCache.Get(now)
		if !exists {
			newCtxs := cacheContexts([]*ReadContext{ctx})
			manager.readingsCache.Set(now, &newCtxs, cache.DefaultExpiration)
		} else {
			cached := item.(*cacheContexts)
			*cached = append(*cached, ctx)
		}
	}
}

// GetReadingsForDevice gets the current reading(s) for the specified device from
// the StateManager.
func (manager *StateManager) GetReadingsForDevice(device string) []*output.Reading {
	manager.readingsLock.RLock()
	defer manager.readingsLock.RUnlock()

	return manager.readings[device]
}

// GetCachedReadings gets the readings in the StateManager's readingsCache. If the plugin
// is not configured to maintain a readings cache, this will just return a dump of the
// current reading state. Once the data has been passed through the given channel, this
// function will close the channel prior to returning.
func (manager *StateManager) GetCachedReadings(start, end string, readings chan *ReadContext) {
	// Whether we exit the function normally or by error, we want to close the channel
	// when we complete to signal to the reader that we are done here.
	defer close(readings)

	// Parse the timestamps for the start/end bounds of the data window.
	startTime, err := utils.ParseRFC3339(start)
	if err != nil {
		// todo: logging

		// if we can't parse the time... we don't really have any business returning data..
		return
	}
	endTime, err := utils.ParseRFC3339(end)
	if err != nil {
		// todo: logging
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
func (manager *StateManager) dumpCachedReadings(start, end time.Time, readings chan *ReadContext) {
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
		ctxs := item.Object.(*cacheContexts)
		for _, ctx := range *ctxs {
			readings <- ctx
		}
	}
}

// dumpCurrentReadings dumps the current readings out to the provided channel.
func (manager *StateManager) dumpCurrentReadings(readings chan *ReadContext) {
	for id, data := range manager.GetReadings() {
		// TODO: make sure this is all we need.. this may change a bit with the move
		//  to tags
		readings <- &ReadContext{
			Device:  id,
			Reading: data,
		}
	}
}

// GetReadings gets a copy of the entire current readings state in the StateManager.
func (manager *StateManager) GetReadings() map[string][]*output.Reading {
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

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
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/health"
	"github.com/vapor-ware/synse-sdk/sdk/output"
)

func Test_newStateManager_nilConfig(t *testing.T) {
	assert.Panics(t, func() {
		newStateManager(nil)
	})
}

func Test_newStateManager(t *testing.T) {
	conf := config.PluginSettings{
		Mode: modeParallel,
		Cache: &config.CacheSettings{
			Enabled: true,
			TTL:     5 * time.Minute,
		},
		Read: &config.ReadSettings{
			Disable:   false,
			Interval:  0,
			Delay:     0,
			QueueSize: 100,
		},
		Transaction: &config.TransactionSettings{
			TTL: 5 * time.Second,
		},
	}

	sm := newStateManager(&conf)

	assert.Equal(t, &conf, sm.config)
	assert.Equal(t, 100, cap(sm.readChan))
	assert.Empty(t, sm.readings)
	assert.NotNil(t, sm.transactions)
	assert.NotNil(t, sm.readingsCache)
}

func TestStateManager_registerActions(t *testing.T) {
	plugin := Plugin{}
	sm := stateManager{}

	assert.Empty(t, plugin.preRun)

	sm.registerActions(&plugin)

	assert.Len(t, plugin.preRun, 1)
}

func TestStateManager_healthChecks(t *testing.T) {
	plugin := Plugin{
		health: health.NewManager(&config.HealthSettings{}),
	}
	sm := stateManager{}

	assert.Equal(t, plugin.health.Count(), 0)

	err := sm.healthChecks(&plugin)
	assert.NoError(t, err)

	assert.Equal(t, plugin.health.Count(), 1)
}

func TestStateManager_addReadingToCache_cacheDisabled(t *testing.T) {
	sm := stateManager{
		config: &config.PluginSettings{
			Cache: &config.CacheSettings{
				Enabled: false,
			},
		},
		readingsCache: cache.New(1*time.Minute, 2*time.Minute),
	}

	assert.Equal(t, 0, sm.readingsCache.ItemCount())

	sm.addReadingToCache(&ReadContext{
		Device:  "test-1",
		Reading: []*output.Reading{{Value: 1}},
	})

	assert.Equal(t, 0, sm.readingsCache.ItemCount())

	_, exists := sm.readingsCache.Get("test-1")
	assert.False(t, exists)
}

func TestStateManager_addReadingToCache_new(t *testing.T) {
	sm := stateManager{
		config: &config.PluginSettings{
			Cache: &config.CacheSettings{
				Enabled: true,
			},
		},
		readingsCache: cache.New(1*time.Minute, 2*time.Minute),
	}

	assert.Equal(t, 0, sm.readingsCache.ItemCount())

	sm.addReadingToCache(&ReadContext{
		Device:  "test-1",
		Reading: []*output.Reading{{Value: 1}},
	})

	assert.Equal(t, 1, sm.readingsCache.ItemCount())
}

func TestStateManager_addReadingToCache_twoReadings(t *testing.T) {
	sm := stateManager{
		config: &config.PluginSettings{
			Cache: &config.CacheSettings{
				Enabled: true,
			},
		},
		readingsCache: cache.New(1*time.Minute, 2*time.Minute),
	}

	assert.Equal(t, 0, sm.readingsCache.ItemCount())

	// Add first reading
	sm.addReadingToCache(&ReadContext{
		Device:  "test-1",
		Reading: []*output.Reading{{Value: 1}},
	})

	assert.Equal(t, 1, sm.readingsCache.ItemCount())

	// Add second reading for the device
	sm.addReadingToCache(&ReadContext{
		Device:  "test-1",
		Reading: []*output.Reading{{Value: 1}},
	})

	assert.Equal(t, 1, sm.readingsCache.ItemCount())
}

func TestStateManager_GetReadingsForDevice_noDevice(t *testing.T) {
	sm := stateManager{
		readingsLock: &sync.RWMutex{},
		readings:     map[string][]*output.Reading{},
	}

	res := sm.GetReadingsForDevice("foo")
	assert.Nil(t, res)
}

func TestStateManager_GetReadingsForDevice_deviceExists(t *testing.T) {
	sm := stateManager{
		readingsLock: &sync.RWMutex{},
		readings: map[string][]*output.Reading{
			"foo": {{Value: 1}},
		},
	}

	res := sm.GetReadingsForDevice("foo")
	assert.NotNil(t, res)
	assert.Len(t, res, 1)
}

func TestStateManager_GetCachedReadings_invalidStart(t *testing.T) {
	sm := stateManager{
		config: &config.PluginSettings{
			Cache: &config.CacheSettings{
				Enabled: true,
			},
		},
	}

	readings := make(chan *ReadContext, 1)

	sm.GetCachedReadings("foobar", "2019-03-22T09:44:33Z", readings)
	assert.Empty(t, readings)

	// Verify the channel was closed
	_, isOpen := <-readings
	assert.False(t, isOpen)
}

func TestStateManager_GetCachedReadings_invalidEnd(t *testing.T) {
	sm := stateManager{
		config: &config.PluginSettings{
			Cache: &config.CacheSettings{
				Enabled: true,
			},
		},
	}

	readings := make(chan *ReadContext, 5)

	sm.GetCachedReadings("2019-03-22T09:44:33Z", "foobar", readings)
	assert.Empty(t, readings)

	// Verify the channel was closed
	_, isOpen := <-readings
	assert.False(t, isOpen)
}

func TestStateManager_GetCachedReadings_cacheEnabled(t *testing.T) {
	sm := stateManager{
		config: &config.PluginSettings{
			Cache: &config.CacheSettings{
				Enabled: true,
			},
		},
		readingsCache: cache.New(1*time.Minute, 2*time.Minute),
	}

	newCtxs := cacheContexts([]*ReadContext{{Device: "123", Reading: []*output.Reading{{Value: 3}}}})
	err := sm.readingsCache.Add("2019-03-22T09:48:00Z", &newCtxs, cache.DefaultExpiration)
	assert.NoError(t, err)

	readings := make(chan *ReadContext, 5)

	sm.GetCachedReadings("2019-03-22T09:45:00Z", "2019-03-22T09:50:00Z", readings)
	assert.Len(t, readings, 1)

	rctx := <-readings
	assert.Equal(t, "123", rctx.Device)

	// Verify the channel was closed
	_, isOpen := <-readings
	assert.False(t, isOpen)
}

func TestStateManager_GetCachedReadings_cacheDisabled(t *testing.T) {
	sm := stateManager{
		config: &config.PluginSettings{
			Cache: &config.CacheSettings{
				Enabled: false,
			},
		},
		readingsLock: &sync.RWMutex{},
		readings: map[string][]*output.Reading{
			"1": {{Value: 3}},
		},
	}

	readings := make(chan *ReadContext, 5)

	sm.GetCachedReadings("2019-03-22T09:45:00Z", "2019-03-22T09:50:00Z", readings)
	assert.Len(t, readings, 1)

	rctx := <-readings
	assert.Equal(t, "1", rctx.Device)

	// Verify the channel was closed
	_, isOpen := <-readings
	assert.False(t, isOpen)
}

func TestStateManager_dumpCachedReadings_noReadings(t *testing.T) {
	sm := stateManager{
		config: &config.PluginSettings{
			Cache: &config.CacheSettings{
				Enabled: true,
			},
		},
		readingsCache: cache.New(1*time.Minute, 2*time.Minute),
	}

	readings := make(chan *ReadContext, 5)
	defer close(readings)

	start, err := time.Parse(time.RFC3339, "2019-03-22T09:45:00Z")
	assert.NoError(t, err)

	end, err := time.Parse(time.RFC3339, "2019-03-22T09:50:00Z")
	assert.NoError(t, err)

	sm.dumpCachedReadings(start, end, readings)
	assert.Empty(t, readings)
}

func TestStateManager_dumpCachedReadings_cachedReadingBeforeStart(t *testing.T) {
	sm := stateManager{
		config: &config.PluginSettings{
			Cache: &config.CacheSettings{
				Enabled: true,
			},
		},
		readingsCache: cache.New(1*time.Minute, 2*time.Minute),
	}

	// Test data setup
	newCtxs := cacheContexts([]*ReadContext{{Device: "123", Reading: []*output.Reading{{Value: 3}}}})
	err := sm.readingsCache.Add("2019-03-22T09:40:00Z", &newCtxs, cache.DefaultExpiration)
	assert.NoError(t, err)

	readings := make(chan *ReadContext, 5)
	defer close(readings)

	start, err := time.Parse(time.RFC3339, "2019-03-22T09:45:00Z")
	assert.NoError(t, err)

	end, err := time.Parse(time.RFC3339, "2019-03-22T09:50:00Z")
	assert.NoError(t, err)

	sm.dumpCachedReadings(start, end, readings)
	assert.Empty(t, readings)
}

func TestStateManager_dumpCachedReadings_cachedReadingAfterEnd(t *testing.T) {
	sm := stateManager{
		config: &config.PluginSettings{
			Cache: &config.CacheSettings{
				Enabled: true,
			},
		},
		readingsCache: cache.New(1*time.Minute, 2*time.Minute),
	}

	// Test data setup
	newCtxs := cacheContexts([]*ReadContext{{Device: "123", Reading: []*output.Reading{{Value: 3}}}})
	err := sm.readingsCache.Add("2019-03-22T09:55:00Z", &newCtxs, cache.DefaultExpiration)
	assert.NoError(t, err)

	readings := make(chan *ReadContext, 5)
	defer close(readings)

	start, err := time.Parse(time.RFC3339, "2019-03-22T09:45:00Z")
	assert.NoError(t, err)

	end, err := time.Parse(time.RFC3339, "2019-03-22T09:50:00Z")
	assert.NoError(t, err)

	sm.dumpCachedReadings(start, end, readings)
	assert.Empty(t, readings)
}

func TestStateManager_dumpCachedReadings_cachedReadingOk(t *testing.T) {
	sm := stateManager{
		config: &config.PluginSettings{
			Cache: &config.CacheSettings{
				Enabled: true,
			},
		},
		readingsCache: cache.New(1*time.Minute, 2*time.Minute),
	}

	// Test data setup
	newCtxs := cacheContexts([]*ReadContext{{Device: "123", Reading: []*output.Reading{{Value: 3}}}})
	err := sm.readingsCache.Add("2019-03-22T09:48:00Z", &newCtxs, cache.DefaultExpiration)
	assert.NoError(t, err)

	readings := make(chan *ReadContext, 5)
	defer close(readings)

	start, err := time.Parse(time.RFC3339, "2019-03-22T09:45:00Z")
	assert.NoError(t, err)

	end, err := time.Parse(time.RFC3339, "2019-03-22T09:50:00Z")
	assert.NoError(t, err)

	sm.dumpCachedReadings(start, end, readings)
	assert.Len(t, readings, 1)

	rctx := <-readings
	assert.Equal(t, "123", rctx.Device)
}

func TestStateManager_dumpCachedReadings_cachedReadingBadTS(t *testing.T) {
	sm := stateManager{
		config: &config.PluginSettings{
			Cache: &config.CacheSettings{
				Enabled: true,
			},
		},
		readingsCache: cache.New(1*time.Minute, 2*time.Minute),
	}

	// Test data setup
	newCtxs := cacheContexts([]*ReadContext{{Device: "123", Reading: []*output.Reading{{Value: 3}}}})
	err := sm.readingsCache.Add("foobar", &newCtxs, cache.DefaultExpiration)
	assert.NoError(t, err)

	readings := make(chan *ReadContext, 5)
	defer close(readings)

	start, err := time.Parse(time.RFC3339, "2019-03-22T09:45:00Z")
	assert.NoError(t, err)

	end, err := time.Parse(time.RFC3339, "2019-03-22T09:50:00Z")
	assert.NoError(t, err)

	sm.dumpCachedReadings(start, end, readings)
	assert.Empty(t, readings)
}

func TestStateManager_dumpCurrentReadings_noReadings(t *testing.T) {
	sm := stateManager{
		readingsLock: &sync.RWMutex{},
		readings:     map[string][]*output.Reading{},
	}

	readings := make(chan *ReadContext, 5)
	defer close(readings)

	sm.dumpCurrentReadings(readings)
	assert.Empty(t, readings)
}

func TestStateManager_dumpCurrentReadings_hasReadings(t *testing.T) {
	sm := stateManager{
		readingsLock: &sync.RWMutex{},
		readings: map[string][]*output.Reading{
			"1": {{Value: 3}},
		},
	}

	readings := make(chan *ReadContext, 5)
	defer close(readings)

	sm.dumpCurrentReadings(readings)
	assert.Len(t, readings, 1)

	rctx := <-readings
	assert.Equal(t, "1", rctx.Device)
}

func TestStateManager_GetReadings_noReadings(t *testing.T) {
	sm := stateManager{
		readingsLock: &sync.RWMutex{},
		readings:     map[string][]*output.Reading{},
	}

	readings := sm.GetReadings()
	assert.Empty(t, readings)
}

func TestStateManager_GetReadings_oneReading(t *testing.T) {
	sm := stateManager{
		readingsLock: &sync.RWMutex{},
		readings: map[string][]*output.Reading{
			"1": {{Value: 3}},
		},
	}

	readings := sm.GetReadings()
	assert.Len(t, readings, 1)
	assert.Contains(t, readings, "1")
}

func TestStateManager_GetReadings_multipleReadings(t *testing.T) {
	sm := stateManager{
		readingsLock: &sync.RWMutex{},
		readings: map[string][]*output.Reading{
			"1": {{Value: 1}},
			"2": {{Value: 2}},
			"3": {{Value: 3}},
		},
	}

	readings := sm.GetReadings()
	assert.Len(t, readings, 3)
	assert.Contains(t, readings, "1")
	assert.Contains(t, readings, "2")
	assert.Contains(t, readings, "3")
}

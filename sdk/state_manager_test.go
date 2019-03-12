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

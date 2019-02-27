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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/sdk/config"
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

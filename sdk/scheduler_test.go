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

	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/health"
	synse "github.com/vapor-ware/synse-server-grpc/go"
)

func TestNewListenerCtx(t *testing.T) {
	h := &DeviceHandler{Name: "test"}
	d := &Device{Type: "test"}

	ctx := NewListenerCtx(h, d)

	assert.Equal(t, h, ctx.handler)
	assert.Equal(t, d, ctx.device)
	assert.Equal(t, 0, ctx.restarts)
}

func TestNewScheduler(t *testing.T) {
	p := &Plugin{
		config: &config.Plugin{
			Settings: &config.PluginSettings{
				Write: &config.WriteSettings{
					QueueSize: 10,
				},
			},
		},
		device: &deviceManager{},
		state:  &stateManager{},
	}

	sched := newScheduler(p)

	assert.NotNil(t, sched.deviceManager)
	assert.NotNil(t, sched.stateManager)
	assert.NotNil(t, sched.config)
	assert.NotNil(t, sched.writeChan)
	assert.NotNil(t, sched.stop)
	assert.Nil(t, sched.limiter)
}

func TestNewScheduler_withLimiter(t *testing.T) {
	p := &Plugin{
		config: &config.Plugin{
			Settings: &config.PluginSettings{
				Limiter: &config.LimiterSettings{
					Rate: 10,
				},
				Write: &config.WriteSettings{
					QueueSize: 10,
				},
			},
		},
		device: &deviceManager{},
		state:  &stateManager{},
	}

	sched := newScheduler(p)

	assert.NotNil(t, sched.deviceManager)
	assert.NotNil(t, sched.stateManager)
	assert.NotNil(t, sched.config)
	assert.NotNil(t, sched.writeChan)
	assert.NotNil(t, sched.stop)
	assert.NotNil(t, sched.limiter)
}

func TestScheduler_registerActions(t *testing.T) {
	plugin := Plugin{}
	s := scheduler{}

	assert.Empty(t, plugin.preRun)
	assert.Empty(t, plugin.postRun)

	s.registerActions(&plugin)
	assert.Len(t, plugin.preRun, 1)
	assert.Len(t, plugin.postRun, 1)
}

func TestScheduler_healthChecks(t *testing.T) {
	plugin := Plugin{
		health: health.NewManager(&config.HealthSettings{}),
	}
	sched := scheduler{}

	assert.Equal(t, plugin.health.Count(), 0)

	err := sched.healthChecks(&plugin)
	assert.NoError(t, err)

	assert.Equal(t, plugin.health.Count(), 1)
}

func TestScheduler_Stop(t *testing.T) {
	s := scheduler{
		stop: make(chan struct{}),
	}

	err := s.Stop()
	assert.NoError(t, err)

	_, isOpen := <-s.stop
	assert.False(t, isOpen)
}

func TestScheduler_Write_nilDevice(t *testing.T) {
	s := &scheduler{}

	resp, err := s.Write(nil, []*synse.V3WriteData{{Action: "test"}})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestScheduler_Write_nilData(t *testing.T) {
	s := &scheduler{}
	dev := &Device{
		handler: &DeviceHandler{
			Write: func(device *Device, data *WriteData) error {
				return nil
			},
		},
	}

	resp, err := s.Write(dev, nil)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestScheduler_Write_deviceNotWritable(t *testing.T) {
	s := &scheduler{}
	dev := &Device{
		handler: &DeviceHandler{},
	}

	resp, err := s.Write(dev, []*synse.V3WriteData{{Action: "test"}})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestScheduler_Write(t *testing.T) {
	s := &scheduler{
		stateManager: &stateManager{
			transactions: cache.New(1*time.Minute, 2*time.Minute),
		},
		writeChan: make(chan *WriteContext, 1),
	}

	dev := &Device{
		WriteTimeout: 1 * time.Minute,
		id:           "test-1",
		handler: &DeviceHandler{
			Write: func(device *Device, data *WriteData) error {
				return nil
			},
		},
	}

	resp, err := s.Write(dev, []*synse.V3WriteData{{Action: "test"}})
	assert.NoError(t, err)
	assert.Len(t, resp, 1)

	// Verify that the transaction was put in the cache.
	assert.Equal(t, 1, s.stateManager.transactions.ItemCount())

	// Verify that the write was put in the write queue.
	w, isOpen := <-s.writeChan
	assert.True(t, isOpen)
	assert.Equal(t, "test-1", w.device)
}

func TestScheduler_WriteAndWait_nilDevice(t *testing.T) {
	s := &scheduler{}

	resp, err := s.WriteAndWait(nil, []*synse.V3WriteData{{Action: "test"}})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestScheduler_WriteAndWait_nilData(t *testing.T) {
	s := &scheduler{}
	dev := &Device{
		handler: &DeviceHandler{
			Write: func(device *Device, data *WriteData) error {
				return nil
			},
		},
	}

	resp, err := s.WriteAndWait(dev, nil)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestScheduler_WriteAndWait_deviceNotWritable(t *testing.T) {
	s := &scheduler{}
	dev := &Device{
		handler: &DeviceHandler{},
	}

	resp, err := s.WriteAndWait(dev, []*synse.V3WriteData{{Action: "test"}})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestScheduler_WriteAndWait(t *testing.T) {
	s := &scheduler{
		stateManager: &stateManager{
			transactions: cache.New(1*time.Minute, 2*time.Minute),
		},
		writeChan: make(chan *WriteContext, 1),
	}

	dev := &Device{
		WriteTimeout: 1 * time.Minute,
		id:           "test-1",
		handler: &DeviceHandler{
			Write: func(device *Device, data *WriteData) error {
				return nil
			},
		},
	}

	go func() {
		// Verify that the write was put in the write queue.
		w, isOpen := <-s.writeChan
		assert.True(t, isOpen)
		assert.Equal(t, "test-1", w.device)

		// Close the transaction to unblock
		close(w.transaction.done)
	}()

	resp, err := s.WriteAndWait(dev, []*synse.V3WriteData{{Action: "test"}})
	assert.NoError(t, err)
	assert.Len(t, resp, 1)

	// Verify that the transaction was put in the cache.
	assert.Equal(t, 1, s.stateManager.transactions.ItemCount())
}

func TestScheduler_scheduleReads_read(t *testing.T) {

}

func TestScheduler_scheduleReads_readBulk(t *testing.T) {

}

func TestScheduler_scheduleWrites(t *testing.T) {

}

func TestScheduler_scheduleListen(t *testing.T) {

}

func TestScheduler_read(t *testing.T) {

}

func TestScheduler_bulkRead(t *testing.T) {

}

func TestScheduler_write(t *testing.T) {

}

func TestScheduler_listen(t *testing.T) {

}

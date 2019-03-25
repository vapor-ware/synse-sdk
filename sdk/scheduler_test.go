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
	"github.com/vapor-ware/synse-sdk/sdk/output"
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

func TestScheduler_scheduleReads_readDisabled(t *testing.T) {
	s := scheduler{
		config: &config.PluginSettings{
			Read: &config.ReadSettings{
				Disable: true,
			},
		},
	}

	assert.False(t, s.isReading)
	s.scheduleReads()
	assert.False(t, s.isReading)
}

func TestScheduler_scheduleReads_noHandlers(t *testing.T) {
	s := scheduler{
		config: &config.PluginSettings{
			Read: &config.ReadSettings{
				Disable: false,
			},
		},
		deviceManager: &deviceManager{
			handlers: map[string]*DeviceHandler{
				"test": {Name: "test"},
			},
		},
	}

	assert.False(t, s.isReading)
	s.scheduleReads()
	assert.False(t, s.isReading)
}

func TestScheduler_scheduleReads(t *testing.T) {
	handler := &DeviceHandler{
		Name: "test",
		Read: func(device *Device) (readings []*output.Reading, e error) {
			return []*output.Reading{{Value: 1}}, nil
		},
	}

	s := scheduler{
		config: &config.PluginSettings{
			Mode: "parallel",
			Read: &config.ReadSettings{
				Disable:  false,
				Interval: 10 * time.Millisecond,
				Delay:    0 * time.Second,
			},
		},
		deviceManager: &deviceManager{
			handlers: map[string]*DeviceHandler{
				"test": handler,
			},
			devices: map[string]*Device{
				"123": {
					id:      "123",
					handler: handler,
				},
			},
		},
		stateManager: &stateManager{
			readChan: make(chan *ReadContext),
		},
		stop: make(chan struct{}),
	}

	go s.scheduleReads()
	defer close(s.stop)

	reading, isOpen := <-s.stateManager.readChan
	assert.True(t, isOpen)
	assert.Equal(t, "123", reading.Device)
}

func TestScheduler_scheduleWrites_writeDisabled(t *testing.T) {
	s := scheduler{
		config: &config.PluginSettings{
			Write: &config.WriteSettings{
				Disable: true,
			},
		},
	}

	assert.False(t, s.isWriting)
	s.scheduleWrites()
	assert.False(t, s.isWriting)
}

func TestScheduler_scheduleWrites_noHandlers(t *testing.T) {
	s := scheduler{
		config: &config.PluginSettings{
			Write: &config.WriteSettings{
				Disable: false,
			},
		},
		deviceManager: &deviceManager{
			handlers: map[string]*DeviceHandler{
				"test": {Name: "test"},
			},
		},
	}

	assert.False(t, s.isWriting)
	s.scheduleWrites()
	assert.False(t, s.isWriting)
}

func TestScheduler_scheduleWrites(t *testing.T) {
	handler := &DeviceHandler{
		Name: "test",
		Write: func(device *Device, data *WriteData) error {
			return nil
		},
	}

	s := scheduler{
		config: &config.PluginSettings{
			Mode: "parallel",
			Write: &config.WriteSettings{
				Disable:   false,
				Interval:  10 * time.Millisecond,
				Delay:     0 * time.Second,
				BatchSize: 10,
			},
		},
		deviceManager: &deviceManager{
			handlers: map[string]*DeviceHandler{
				"test": handler,
			},
			devices: map[string]*Device{
				"123": {
					id:      "123",
					handler: handler,
				},
			},
		},
		stateManager: &stateManager{
			readChan:     make(chan *ReadContext),
			transactions: cache.New(1*time.Minute, 2*time.Minute),
		},
		writeChan: make(chan *WriteContext),
		stop:      make(chan struct{}),
	}

	go s.scheduleWrites()
	defer close(s.stop)

	txn := s.stateManager.newTransaction(10 * time.Minute)
	wctx := &WriteContext{
		txn,
		"123",
		&synse.V3WriteData{Action: "test"},
	}
	s.writeChan <- wctx

	txn.wait()
	assert.Equal(t, synse.WriteStatus_DONE, txn.status)
}

func TestScheduler_scheduleListen_listenDisabled(t *testing.T) {
	s := scheduler{
		config: &config.PluginSettings{
			Listen: &config.ListenSettings{
				Disable: true,
			},
		},
	}

	assert.False(t, s.isListening)
	s.scheduleListen()
	assert.False(t, s.isListening)
}

func TestScheduler_scheduleListen_noHandlers(t *testing.T) {
	s := scheduler{
		config: &config.PluginSettings{
			Listen: &config.ListenSettings{
				Disable: false,
			},
		},
		deviceManager: &deviceManager{
			handlers: map[string]*DeviceHandler{
				"test": {Name: "test"},
			},
		},
	}

	assert.False(t, s.isListening)
	s.scheduleListen()
	assert.False(t, s.isListening)
}

func TestScheduler_scheduleListen(t *testing.T) {
	handler := &DeviceHandler{
		Name: "test",
		Listen: func(device *Device, contexts chan *ReadContext) error {
			contexts <- &ReadContext{Device: device.id}
			return nil
		},
	}

	s := scheduler{
		config: &config.PluginSettings{
			Mode: "parallel",
			Listen: &config.ListenSettings{
				Disable: false,
			},
		},
		deviceManager: &deviceManager{
			handlers: map[string]*DeviceHandler{
				"test": handler,
			},
			devices: map[string]*Device{
				"123": {
					id:      "123",
					handler: handler,
					Handler: "test",
				},
			},
		},
		stateManager: &stateManager{
			readChan: make(chan *ReadContext),
		},
		stop: make(chan struct{}),
	}

	go s.scheduleListen()
	defer close(s.stop)

	reading, isOpen := <-s.stateManager.readChan
	assert.True(t, isOpen)
	assert.Equal(t, "123", reading.Device)
}

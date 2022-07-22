// Synse SDK
// Copyright (c) 2017-2022 Vapor IO
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
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/v2/sdk/config"
	"github.com/vapor-ware/synse-sdk/v2/sdk/funcs"
	"github.com/vapor-ware/synse-sdk/v2/sdk/health"
	"github.com/vapor-ware/synse-sdk/v2/sdk/output"
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
	assert.Equal(t, ErrNilDevice, err)
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
	assert.Equal(t, ErrNilData, err)
	assert.Nil(t, resp)
}

func TestScheduler_Write_deviceNotWritable(t *testing.T) {
	s := &scheduler{}
	dev := &Device{
		handler: &DeviceHandler{},
	}

	resp, err := s.Write(dev, []*synse.V3WriteData{{Action: "test"}})
	assert.Error(t, err)
	assert.Equal(t, ErrDeviceNotWritable, err)
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
	assert.Equal(t, dev, w.device)
}

func TestScheduler_WriteAndWait_nilDevice(t *testing.T) {
	s := &scheduler{}

	resp, err := s.WriteAndWait(nil, []*synse.V3WriteData{{Action: "test"}})
	assert.Error(t, err)
	assert.Equal(t, ErrNilDevice, err)
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
	assert.Equal(t, ErrNilData, err)
	assert.Nil(t, resp)
}

func TestScheduler_WriteAndWait_deviceNotWritable(t *testing.T) {
	s := &scheduler{}
	dev := &Device{
		handler: &DeviceHandler{},
	}

	resp, err := s.WriteAndWait(dev, []*synse.V3WriteData{{Action: "test"}})
	assert.Error(t, err)
	assert.Equal(t, ErrDeviceNotWritable, err)
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
		assert.Equal(t, dev, w.device)

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
	assert.Equal(t, s.deviceManager.GetDevice("123"), reading.Device)
}

// When configured in serial mode, scheduleReads should execute all reads serially, even
// if there is a mix of single-read and batch-read handlers.
//
// In order to check if things were run serially, each handler sleeps for a short period
// of time. We time the execution of the scheduler run to see if the timing profile fits
// that of a serial run.
func TestScheduler_scheduleReadsSerial(t *testing.T) {
	// Define a set of DeviceHandlers which will be used for the test case.
	singleReadHandler1 := &DeviceHandler{
		Name: "handler1",
		Read: func(device *Device) ([]*output.Reading, error) {
			time.Sleep(1 * 500 * time.Millisecond)
			return []*output.Reading{{Value: device.Info}}, nil
		},
	}
	singleReadHandler2 := &DeviceHandler{
		Name: "handler2",
		Read: func(device *Device) ([]*output.Reading, error) {
			time.Sleep(1 * 500 * time.Millisecond)
			return []*output.Reading{{Value: device.Info}}, nil
		},
	}
	batchReadHandler3 := &DeviceHandler{
		Name: "handler3",
		BulkRead: func(devices []*Device) ([]*ReadContext, error) {
			time.Sleep(1 * 500 * time.Millisecond)
			var readings []*ReadContext
			for _, d := range devices {
				readings = append(readings, &ReadContext{
					Device:  d,
					Reading: []*output.Reading{{Value: d.Info}},
				})
			}
			return readings, nil
		},
	}
	batchReadHandler4 := &DeviceHandler{
		Name: "handler4",
		BulkRead: func(devices []*Device) ([]*ReadContext, error) {
			time.Sleep(1 * 500 * time.Millisecond)
			var readings []*ReadContext
			for _, d := range devices {
				readings = append(readings, &ReadContext{
					Device:  d,
					Reading: []*output.Reading{{Value: d.Info}},
				})
			}
			return readings, nil
		},
	}

	// Create an instance of the scheduler to test.
	s := scheduler{
		config: &config.PluginSettings{
			Mode: modeSerial,
			Read: &config.ReadSettings{
				Disable:  false,
				Interval: 100 * time.Millisecond,
				Delay:    0 * time.Second,
			},
		},
		deviceManager: &deviceManager{
			handlers: map[string]*DeviceHandler{
				"handler1": singleReadHandler1,
				"handler2": singleReadHandler2,
				"handler3": batchReadHandler3,
				"handler4": batchReadHandler4,
			},
			devices: map[string]*Device{
				// We define two devices per handler here.
				"1": {id: "1", Info: "1", Handler: "handler1", handler: singleReadHandler1},
				"2": {id: "2", Info: "2", Handler: "handler1", handler: singleReadHandler1},

				"3": {id: "3", Info: "3", Handler: "handler2", handler: singleReadHandler2},
				"4": {id: "4", Info: "4", Handler: "handler2", handler: singleReadHandler2},

				"5": {id: "5", Info: "5", Handler: "handler3", handler: batchReadHandler3},
				"6": {id: "6", Info: "6", Handler: "handler3", handler: batchReadHandler3},

				"7": {id: "7", Info: "7", Handler: "handler4", handler: batchReadHandler4},
				"8": {id: "8", Info: "8", Handler: "handler4", handler: batchReadHandler4},
			},
		},
		stateManager: &stateManager{
			readChan: make(chan *ReadContext, 100),
		},
		stop:       make(chan struct{}, 1),
		serialLock: &sync.Mutex{},
	}

	// Wait a short period of time for scheduleReads to run before closing the
	// channel, stopping the scheduleReads loop. This only needs to be short, since
	// it is checked at the top of the loop. We just need enough time to ensure we
	// get into the loop.
	go func() {
		time.Sleep(100 * time.Millisecond)
		close(s.stop)
	}()

	start := time.Now()
	s.scheduleReads()
	stop := time.Now()

	// Verify that the execution time matches the expected run time for serial reads.
	// Individual reads @ 500ms (2 handlers, 2 devices each) = 2s
	// Bulk reads @ 500ms (2 handlers) = 1s
	// Total = 2s + 1s + 100ms of overhead for an expected 3100ms
	assert.InDelta(t, 3100*time.Millisecond, stop.Sub(start), float64(150*time.Millisecond))

	// Close the read channel so we can iterate over it without blocking.
	close(s.stateManager.readChan)

	// Collect all the readings
	var readings []*ReadContext
	for ctx := range s.stateManager.readChan {
		readings = append(readings, ctx)
	}

	// We can't guarantee the number of times that the scheduler loop ran during the
	// test, but we know how many readings we should get per run based on the number
	// of devices and the handler responses. There should be one ReadContext per device,
	// with 8 devices, so 8 ReadContexts.
	assert.Greater(t, len(readings), 0)
	assert.Equal(t, 0, len(readings)%8)

	// We'll only check the first 8, since that constitutes a single run. Despite being
	// run serially, we can't actually guarantee the order of the returned readings because
	// it depends on which read goroutine acquires the lock first. While values may be out
	// of sequential order, it doesn't mean they weren't run serially.
	var values []string
	for i := 0; i < 8; i++ {
		values = append(values, readings[i].Reading[0].Value.(string))
	}

	assert.Contains(t, values, "1")
	assert.Contains(t, values, "2")
	assert.Contains(t, values, "3")
	assert.Contains(t, values, "4")
	assert.Contains(t, values, "5")
	assert.Contains(t, values, "6")
	assert.Contains(t, values, "7")
	assert.Contains(t, values, "8")
}

// When configured in parallel mode, scheduleReads should execute all reads in parallel,
// even if there is a mix of single-read and batch-read handlers.
//
// In order to check if things were run in parallel, each handler sleeps for a short period
// of time. We time the execution of the scheduler run to see if the timing profile fits
// that of a parallel run.
func TestScheduler_scheduleReadsParallel(t *testing.T) {
	// Define a set of DeviceHandlers which will be used for the test case.
	singleReadHandler1 := &DeviceHandler{
		Name: "handler1",
		Read: func(device *Device) ([]*output.Reading, error) {
			time.Sleep(1 * 500 * time.Millisecond)
			return []*output.Reading{{Value: device.Info}}, nil
		},
	}
	singleReadHandler2 := &DeviceHandler{
		Name: "handler2",
		Read: func(device *Device) ([]*output.Reading, error) {
			time.Sleep(1 * 500 * time.Millisecond)
			return []*output.Reading{{Value: device.Info}}, nil
		},
	}
	batchReadHandler3 := &DeviceHandler{
		Name: "handler3",
		BulkRead: func(devices []*Device) ([]*ReadContext, error) {
			time.Sleep(1 * 500 * time.Millisecond)
			var readings []*ReadContext
			for _, d := range devices {
				readings = append(readings, &ReadContext{
					Device:  d,
					Reading: []*output.Reading{{Value: d.Info}},
				})
			}
			return readings, nil
		},
	}
	batchReadHandler4 := &DeviceHandler{
		Name: "handler4",
		BulkRead: func(devices []*Device) ([]*ReadContext, error) {
			time.Sleep(1 * 500 * time.Millisecond)
			var readings []*ReadContext
			for _, d := range devices {
				readings = append(readings, &ReadContext{
					Device:  d,
					Reading: []*output.Reading{{Value: d.Info}},
				})
			}
			return readings, nil
		},
	}

	// Create an instance of the scheduler to test.
	s := scheduler{
		config: &config.PluginSettings{
			Mode: modeParallel,
			Read: &config.ReadSettings{
				Disable:  false,
				Interval: 100 * time.Millisecond,
				Delay:    0 * time.Second,
			},
		},
		deviceManager: &deviceManager{
			handlers: map[string]*DeviceHandler{
				"handler1": singleReadHandler1,
				"handler2": singleReadHandler2,
				"handler3": batchReadHandler3,
				"handler4": batchReadHandler4,
			},
			devices: map[string]*Device{
				// We define two devices per handler here.
				"1": {id: "1", Info: "1", Handler: "handler1", handler: singleReadHandler1},
				"2": {id: "2", Info: "2", Handler: "handler1", handler: singleReadHandler1},

				"3": {id: "3", Info: "3", Handler: "handler2", handler: singleReadHandler2},
				"4": {id: "4", Info: "4", Handler: "handler2", handler: singleReadHandler2},

				"5": {id: "5", Info: "5", Handler: "handler3", handler: batchReadHandler3},
				"6": {id: "6", Info: "6", Handler: "handler3", handler: batchReadHandler3},

				"7": {id: "7", Info: "7", Handler: "handler4", handler: batchReadHandler4},
				"8": {id: "8", Info: "8", Handler: "handler4", handler: batchReadHandler4},
			},
		},
		stateManager: &stateManager{
			readChan: make(chan *ReadContext, 100),
		},
		stop:       make(chan struct{}, 1),
		serialLock: &sync.Mutex{},
	}

	// Wait a short period of time for scheduleReads to run before closing the
	// channel, stopping the scheduleReads loop. This only needs to be short, since
	// it is checked at the top of the loop. We just need enough time to ensure we
	// get into the loop.
	go func() {
		time.Sleep(100 * time.Millisecond)
		close(s.stop)
	}()

	start := time.Now()
	s.scheduleReads()
	stop := time.Now()

	// Verify that the execution time matches the expected run time for parallel reads.
	// Individual reads @ 500ms (2 handlers, 2 devices each) ~= 500ms
	// Bulk reads @ 500ms (2 handlers) ~= 500ms
	// Total = 500ms + 100ms of overhead for an expected 600ms
	assert.InDelta(t, 600*time.Millisecond, stop.Sub(start), float64(100*time.Millisecond))

	// Close the read channel so we can iterate over it without blocking.
	close(s.stateManager.readChan)

	// Collect all the readings
	var readings []*ReadContext
	for ctx := range s.stateManager.readChan {
		readings = append(readings, ctx)
	}

	// We can't guarantee the number of times that the scheduler loop ran during the
	// test, but we know how many readings we should get per run based on the number
	// of devices and the handler responses. There should be one ReadContext per device,
	// with 8 devices, so 8 ReadContexts.
	assert.Greater(t, len(readings), 0)
	assert.Equal(t, 0, len(readings)%8)

	// We'll only check the first 8, since that constitutes a single run. Being run in
	// parallel, we can't guarantee any order, but we do know which values we should get.
	var values []string
	for i := 0; i < 8; i++ {
		values = append(values, readings[i].Reading[0].Value.(string))
	}

	assert.Contains(t, values, "1")
	assert.Contains(t, values, "2")
	assert.Contains(t, values, "3")
	assert.Contains(t, values, "4")
	assert.Contains(t, values, "5")
	assert.Contains(t, values, "6")
	assert.Contains(t, values, "7")
	assert.Contains(t, values, "8")
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
					id:           "123",
					handler:      handler,
					WriteTimeout: 1 * time.Second,
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

	txn, err := s.stateManager.newTransaction(10*time.Minute, "")
	assert.NoError(t, err)

	wctx := &WriteContext{
		txn,
		s.deviceManager.GetDevice("123"),
		&synse.V3WriteData{Action: "test"},
	}
	s.writeChan <- wctx

	txn.wait()
	assert.Equal(t, synse.WriteStatus_DONE, txn.status, txn.message)
}

func TestScheduler_scheduleWrites_serial_withDelay(t *testing.T) {
	handler := &DeviceHandler{
		Name: "test",
		Write: func(device *Device, data *WriteData) error {
			return nil
		},
	}

	s := scheduler{
		config: &config.PluginSettings{
			Mode: modeSerial,
			Write: &config.WriteSettings{
				Disable:   false,
				Interval:  10 * time.Millisecond,
				Delay:     100 * time.Millisecond,
				BatchSize: 10,
			},
		},
		deviceManager: &deviceManager{
			handlers: map[string]*DeviceHandler{
				"test": handler,
			},
			devices: map[string]*Device{
				"123": {
					id:           "123",
					handler:      handler,
					WriteTimeout: 1 * time.Second,
				},
			},
		},
		stateManager: &stateManager{
			readChan:     make(chan *ReadContext),
			transactions: cache.New(1*time.Minute, 2*time.Minute),
		},
		writeChan:  make(chan *WriteContext),
		stop:       make(chan struct{}),
		serialLock: &sync.Mutex{},
	}

	go s.scheduleWrites()
	defer close(s.stop)

	txn1, err := s.stateManager.newTransaction(10*time.Minute, "")
	assert.NoError(t, err)

	txn2, err := s.stateManager.newTransaction(10*time.Minute, "")
	assert.NoError(t, err)

	wctx1 := &WriteContext{
		txn1,
		s.deviceManager.GetDevice("123"),
		&synse.V3WriteData{Action: "test"},
	}

	wctx2 := &WriteContext{
		txn2,
		s.deviceManager.GetDevice("123"),
		&synse.V3WriteData{Action: "test"},
	}

	start := time.Now()
	s.writeChan <- wctx1
	s.writeChan <- wctx2

	txn1.wait()
	txn2.wait()

	end := time.Now()

	// The delay when writing occurs after the transaction status is set to DONE, so we
	// should only detect the delay between txn1 and tnx2, but not the delay after txn2.
	// This means that this test only sees a single delay interval.
	//
	// Add a threshold of +/- 18ms to account for leeway in different testing environments.
	// In CI, tests run a bit slower.
	assert.WithinDuration(t, start.Add(100*time.Millisecond), end, 18*time.Millisecond)

	assert.Equal(t, synse.WriteStatus_DONE, txn1.status, txn1.message)
	assert.Equal(t, synse.WriteStatus_DONE, txn2.status, txn2.message)
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
			contexts <- &ReadContext{Device: device}
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
	assert.Equal(t, s.deviceManager.GetDevice("123"), reading.Device)
}

func TestScheduler_applyTransformations_NoTransformers(t *testing.T) {
	device := &Device{
		Transforms: []Transformer{},
	}
	rctx := &ReadContext{
		Reading: []*output.Reading{
			{Value: 2},
		},
	}

	err := finalizeReadings(device, rctx)
	assert.NoError(t, err)

	// Verify that the reading value did not change.
	assert.Equal(t, 2, rctx.Reading[0].Value.(int))

	// Verify that no additional context was set.
	assert.Empty(t, rctx.Reading[0].Context)
}

func TestScheduler_applyTransformations_ScaleTransformerOk(t *testing.T) {
	device := &Device{
		Transforms: []Transformer{
			&ScaleTransformer{Factor: 2},
		},
	}
	rctx := &ReadContext{
		Reading: []*output.Reading{
			{Value: 2},
		},
	}

	err := finalizeReadings(device, rctx)
	assert.NoError(t, err)

	// Verify that the reading value changed
	assert.Equal(t, float64(4), rctx.Reading[0].Value.(float64))

	// Verify that no additional context was set.
	assert.Empty(t, rctx.Reading[0].Context)
}

func TestScheduler_applyTransformations_ApplyTransformerOk(t *testing.T) {
	device := &Device{
		Transforms: []Transformer{
			&ApplyTransformer{
				Func: &funcs.Func{
					Name: "test-fn-1",
					Fn: func(value interface{}) (interface{}, error) {
						return (value.(int)) * 2, nil
					},
				},
			},
		},
	}
	rctx := &ReadContext{
		Reading: []*output.Reading{
			{Value: 2},
		},
	}

	err := finalizeReadings(device, rctx)
	assert.NoError(t, err)

	// Verify that the reading value changed.
	assert.Equal(t, 4, rctx.Reading[0].Value.(int))

	// Verify that no additional context was set.
	assert.Empty(t, rctx.Reading[0].Context)
}

func TestScheduler_applyTransformations_ApplyAndScaleTransformerOK(t *testing.T) {
	device := &Device{
		Transforms: []Transformer{
			// First scale
			&ScaleTransformer{
				Factor: 2,
			},
			// Then apply
			&ApplyTransformer{
				Func: &funcs.Func{
					Name: "test-fn-1",
					Fn: func(value interface{}) (interface{}, error) {
						return ((value.(float64)) / 4.0) + 1, nil
					},
				},
			},
		},
	}
	rctx := &ReadContext{
		Reading: []*output.Reading{
			{Value: 2},
		},
	}

	err := finalizeReadings(device, rctx)
	assert.NoError(t, err)

	// Verify that the reading value changed. We should expect the value to first follow
	// the scale transform then the apply transform.
	assert.Equal(t, float64(2), rctx.Reading[0].Value.(float64))

	// Verify that no additional context was set.
	assert.Empty(t, rctx.Reading[0].Context)
}

func TestScheduler_applyTransformations_ApplyAndScaleTransformerOK_OrderChanged(t *testing.T) {
	device := &Device{
		Transforms: []Transformer{
			// First apply
			&ApplyTransformer{
				Func: &funcs.Func{
					Name: "test-fn-1",
					Fn: func(value interface{}) (interface{}, error) {
						return (float64(value.(int)) / 4.0) + 1, nil
					},
				},
			},
			// Then scale
			&ScaleTransformer{
				Factor: 2,
			},
		},
	}
	rctx := &ReadContext{
		Reading: []*output.Reading{
			{Value: 2},
		},
	}

	fmt.Println("==============================================")

	err := finalizeReadings(device, rctx)
	assert.NoError(t, err)

	// Verify that the reading value changed. We should expect the value to first follow
	// the apply transform then the scale transform.
	assert.Equal(t, float64(3), rctx.Reading[0].Value.(float64))

	// Verify that no additional context was set.
	assert.Empty(t, rctx.Reading[0].Context)
}

func TestScheduler_applyTransformations_multipleFnsOk(t *testing.T) {
	device := &Device{
		Transforms: []Transformer{
			&ApplyTransformer{
				Func: &funcs.Func{
					Name: "test-fn-1",
					Fn: func(value interface{}) (interface{}, error) {
						return (value.(int)) * 2, nil
					},
				},
			},
			&ApplyTransformer{
				Func: &funcs.Func{
					Name: "test-fn-2",
					Fn: func(value interface{}) (interface{}, error) {
						return (value.(int)) + 3, nil
					},
				},
			},
		},
	}
	rctx := &ReadContext{
		Reading: []*output.Reading{
			{Value: 2},
		},
	}

	err := finalizeReadings(device, rctx)
	assert.NoError(t, err)

	// Verify that the reading value changed.
	assert.Equal(t, 7, rctx.Reading[0].Value.(int))

	// Verify that no additional context was set.
	assert.Empty(t, rctx.Reading[0].Context)
}

func TestScheduler_applyTransformations_multipleFnsOk_withScale(t *testing.T) {
	device := &Device{
		Transforms: []Transformer{
			&ApplyTransformer{
				Func: &funcs.Func{
					Name: "test-fn-1",
					Fn: func(value interface{}) (interface{}, error) {
						return (value.(int)) * 2, nil
					},
				},
			},
			&ApplyTransformer{
				Func: &funcs.Func{
					Name: "test-fn-2",
					Fn: func(value interface{}) (interface{}, error) {
						return (value.(int)) + 3, nil
					},
				},
			},
			&ScaleTransformer{
				Factor: 2,
			},
		},
	}
	rctx := &ReadContext{
		Reading: []*output.Reading{
			{Value: 2},
		},
	}

	err := finalizeReadings(device, rctx)
	assert.NoError(t, err)

	// Verify that the reading value changed.
	assert.Equal(t, float64(14), rctx.Reading[0].Value.(float64))

	// Verify that no additional context was set.
	assert.Empty(t, rctx.Reading[0].Context)
}

func TestScheduler_applyTransformations_ApplyTransformerErr(t *testing.T) {
	device := &Device{
		Transforms: []Transformer{
			&ApplyTransformer{
				Func: &funcs.Func{
					Name: "test-fn-1",
					Fn: func(value interface{}) (interface{}, error) {
						return nil, fmt.Errorf("test error")
					},
				},
			},
		},
	}
	rctx := &ReadContext{
		Reading: []*output.Reading{
			{Value: 2},
		},
	}

	err := finalizeReadings(device, rctx)
	assert.Error(t, err)

	// Verify that the reading value did not change.
	assert.Equal(t, 2, rctx.Reading[0].Value.(int))

	// Verify that no additional context was set.
	assert.Empty(t, rctx.Reading[0].Context)
}

func TestScheduler_applyTransformations_ApplyTransformerErr_WithScale(t *testing.T) {
	device := &Device{
		Transforms: []Transformer{
			&ApplyTransformer{
				Func: &funcs.Func{
					Name: "test-fn-1",
					Fn: func(value interface{}) (interface{}, error) {
						return nil, fmt.Errorf("test error")
					},
				},
			},
			&ScaleTransformer{
				Factor: 2,
			},
		},
	}
	rctx := &ReadContext{
		Reading: []*output.Reading{
			{Value: 2},
		},
	}

	err := finalizeReadings(device, rctx)
	assert.Error(t, err)

	// Verify that the reading value did not change.
	assert.Equal(t, 2, rctx.Reading[0].Value.(int))

	// Verify that no additional context was set.
	assert.Empty(t, rctx.Reading[0].Context)
}

func TestScheduler_applyTransformations_ApplyTransformerErr_WithScale_Reordered(t *testing.T) {
	// Same test as above, but transforms in different order.
	device := &Device{
		Transforms: []Transformer{
			&ScaleTransformer{
				Factor: 2,
			},
			&ApplyTransformer{
				Func: &funcs.Func{
					Name: "test-fn-1",
					Fn: func(value interface{}) (interface{}, error) {
						return nil, fmt.Errorf("test error")
					},
				},
			},
		},
	}
	rctx := &ReadContext{
		Reading: []*output.Reading{
			{Value: 2},
		},
	}

	err := finalizeReadings(device, rctx)
	assert.Error(t, err)

	// Verify that the reading value changed. It changed from the first scale transform.
	// This is okay since the error on the second transform would cause an error on read,
	// so we wouldn't propagate this partially transformed value.
	assert.Equal(t, float64(4), rctx.Reading[0].Value.(float64))

	// Verify that no additional context was set.
	assert.Empty(t, rctx.Reading[0].Context)
}

func TestScheduler_applyTransformations_MultipleApplyTransformerError(t *testing.T) {
	device := &Device{
		Transforms: []Transformer{
			&ApplyTransformer{
				Func: &funcs.Func{
					Name: "test-fn-1",
					Fn: func(value interface{}) (interface{}, error) {
						return (value.(int)) * 2, nil
					},
				},
			},
			&ApplyTransformer{
				Func: &funcs.Func{
					Name: "test-fn-2",
					Fn: func(value interface{}) (interface{}, error) {
						return nil, fmt.Errorf("test err")
					},
				},
			},
		},
	}
	rctx := &ReadContext{
		Reading: []*output.Reading{
			{Value: 2},
		},
	}

	err := finalizeReadings(device, rctx)
	assert.Error(t, err)

	// Verify that the reading value changed. It should change because the first
	// fn was applied successfully. It is up to the upstream caller to check the
	// error and make sure all transforms succeed before using the value.
	assert.Equal(t, 4, rctx.Reading[0].Value.(int))

	// Verify that no additional context was set.
	assert.Empty(t, rctx.Reading[0].Context)
}

func TestScheduler_applyTransformations_MultipleApplyErr_WithScale(t *testing.T) {
	device := &Device{
		Transforms: []Transformer{
			&ApplyTransformer{
				Func: &funcs.Func{
					Name: "test-fn-1",
					Fn: func(value interface{}) (interface{}, error) {
						return (value.(int)) * 2, nil
					},
				},
			},
			&ApplyTransformer{
				Func: &funcs.Func{
					Name: "test-fn-2",
					Fn: func(value interface{}) (interface{}, error) {
						return nil, fmt.Errorf("test err")
					},
				},
			},
			&ScaleTransformer{
				Factor: 3,
			},
		},
	}
	rctx := &ReadContext{
		Reading: []*output.Reading{
			{Value: 2},
		},
	}

	err := finalizeReadings(device, rctx)
	assert.Error(t, err)

	// Verify that the reading value changed. It should change because the first
	// fn was applied successfully. It is up to the upstream caller to check the
	// error and make sure all transforms succeed before using the value.
	assert.Equal(t, 4, rctx.Reading[0].Value.(int))

	// Verify that no additional context was set.
	assert.Empty(t, rctx.Reading[0].Context)
}

func TestScheduler_applyTransformations_oneFnOk_withScaleErr(t *testing.T) {
	device := &Device{
		Transforms: []Transformer{
			&ApplyTransformer{
				Func: &funcs.Func{
					Name: "test-fn-1",
					Fn: func(value interface{}) (interface{}, error) {
						return (value.(int)) * 2, nil
					},
				},
			},
			&ScaleTransformer{
				Factor: 0,
			},
		},
	}
	rctx := &ReadContext{
		Reading: []*output.Reading{
			{Value: 2},
		},
	}

	err := finalizeReadings(device, rctx)
	assert.Error(t, err)

	// Verify that the reading value changed. It should change because the
	// transform fn ran, but the scaling fn should not have run.
	assert.Equal(t, 4, rctx.Reading[0].Value.(int))

	// Verify that no additional context was set.
	assert.Empty(t, rctx.Reading[0].Context)
}

func TestScheduler_finalizeReadings_withContext(t *testing.T) {
	device := &Device{
		Context: map[string]string{"foo": "bar"},
	}
	rctx := &ReadContext{
		Reading: []*output.Reading{
			{Value: 2},
		},
	}

	err := finalizeReadings(device, rctx)
	assert.NoError(t, err)

	// Verify that the reading value did not change.
	assert.Equal(t, 2, rctx.Reading[0].Value.(int))

	// Verify that the device context was set.
	assert.Equal(t, map[string]string{"foo": "bar"}, rctx.Reading[0].Context)
}

func TestScheduler_finalizeReadings_withContextAugment(t *testing.T) {
	device := &Device{
		Context: map[string]string{"foo": "bar"},
	}
	rctx := &ReadContext{
		Reading: []*output.Reading{
			{
				Value:   2,
				Context: map[string]string{"abc": "def"},
			},
		},
	}

	err := finalizeReadings(device, rctx)
	assert.NoError(t, err)

	// Verify that the reading value did not change.
	assert.Equal(t, 2, rctx.Reading[0].Value.(int))

	// Verify that the device context was set.
	assert.Equal(t, map[string]string{"foo": "bar", "abc": "def"}, rctx.Reading[0].Context)
}

func TestScheduler_finalizeReadings_withContextOverride(t *testing.T) {
	device := &Device{
		Context: map[string]string{"foo": "bar"},
	}
	rctx := &ReadContext{
		Reading: []*output.Reading{
			{
				Value:   2,
				Context: map[string]string{"foo": "123"},
			},
		},
	}

	err := finalizeReadings(device, rctx)
	assert.NoError(t, err)

	// Verify that the reading value did not change.
	assert.Equal(t, 2, rctx.Reading[0].Value.(int))

	// Verify that the device context was set.
	assert.Equal(t, map[string]string{"foo": "bar"}, rctx.Reading[0].Context)
}

func TestScheduler_finalizeReadings_withContextAndTransform(t *testing.T) {
	device := &Device{
		Transforms: []Transformer{
			&ScaleTransformer{Factor: 2},
		},
		Context: map[string]string{"foo": "bar"},
	}
	rctx := &ReadContext{
		Reading: []*output.Reading{
			{Value: 2},
		},
	}

	err := finalizeReadings(device, rctx)
	assert.NoError(t, err)

	// Verify that the reading value changed.
	assert.Equal(t, float64(4), rctx.Reading[0].Value.(float64))

	// Verify that the device context was set.
	assert.Equal(t, map[string]string{"foo": "bar"}, rctx.Reading[0].Context)
}

func TestScheduler_finalizeReadings_nilReading(t *testing.T) {
	device := &Device{
		Transforms: []Transformer{
			&ScaleTransformer{Factor: 2},
		},
		Context: map[string]string{"foo": "bar"},
	}
	rctx := &ReadContext{
		Reading: []*output.Reading{
			{Value: nil},
		},
	}

	err := finalizeReadings(device, rctx)
	assert.NoError(t, err)

	// Verify that the reading value did not change.
	assert.Equal(t, nil, rctx.Reading[0].Value)

	// Verify that the device context was set.
	assert.Equal(t, map[string]string{"foo": "bar"}, rctx.Reading[0].Context)
}

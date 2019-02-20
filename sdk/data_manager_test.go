package sdk

//
//import (
//	"fmt"
//	"testing"
//	"time"
//
//	"github.com/stretchr/testify/assert"
//	"github.com/vapor-ware/synse-server-grpc/go"
//)
//
//// TestNewDataManager tests creating a new dataManager instance successfully.
//func TestNewDataManager(t *testing.T) {
//	d := newDataManager()
//	assert.Nil(t, d.readChannel)
//	assert.Nil(t, d.writeChannel)
//	assert.Nil(t, d.limiter)
//	assert.NotNil(t, d.dataLock)
//	assert.NotNil(t, d.rwLock)
//	assert.Empty(t, d.readings)
//}
//
//// TestDeviceManager_setupError tests calling setup on the DataManager when there
//// is no global plugin config. This should result in error.
//func TestDataManager_setupError(t *testing.T) {
//	defer Config.reset()
//
//	Config.Plugin = nil
//	err := DataManager.setup()
//	assert.Error(t, err)
//}
//
//// TestDataManager_WritesEnabled tests that writes are enabled in the dataManager
//// when they are enabled in the config.
//func TestDataManager_WritesEnabled(t *testing.T) {
//	defer Config.reset()
//
//	Config.Plugin = &PluginConfig{
//		Version: 1,
//		Network: &NetworkSettings{
//			Type:    "tcp",
//			Address: "test",
//		},
//		Settings: &PluginSettings{
//			Read: &ReadSettings{
//				Buffer: 200,
//			},
//			Write: &WriteSettings{
//				Enabled: true,
//				Buffer:  200,
//			},
//			Transaction: &TransactionSettings{
//				TTL: "2s",
//			},
//		},
//	}
//
//	assert.True(t, DataManager.writesEnabled())
//}
//
//// TestDataManager_readNoLimiter tests reading a device when a limiter is
//// not configured.
//func TestDataManager_readOneOkNoLimiter(t *testing.T) {
//	defer Config.reset()
//
//	Config.Plugin = &PluginConfig{
//		Version: 1,
//		Network: &NetworkSettings{
//			Type:    "tcp",
//			Address: "test",
//		},
//		Settings: &PluginSettings{
//			Read:        &ReadSettings{Buffer: 200},
//			Write:       &WriteSettings{Buffer: 200},
//			Listen:      &ListenSettings{Buffer: 100},
//			Transaction: &TransactionSettings{TTL: "2s"},
//		},
//	}
//
//	// Create the device to read
//	device := &Device{
//		Kind:     "test.state",
//		Location: &Location{Rack: "rack", Board: "board"},
//		Outputs: []*Output{
//			{
//				OutputType: OutputType{
//					Name: "foo",
//				},
//			},
//		},
//		Handler: &DeviceHandler{
//			Read: func(d *Device) ([]*Reading, error) {
//				output := d.GetOutput("foo")
//				reading, err := output.MakeReading("ok")
//				if err != nil {
//					return nil, err
//				}
//				return []*Reading{reading}, nil
//			},
//		},
//	}
//
//	d := newDataManager()
//	err := d.setup()
//	assert.NoError(t, err)
//
//	// Pass a reading in
//	assert.Equal(t, 0, len(d.readChannel))
//	d.readOne(device)
//	assert.Equal(t, 1, len(d.readChannel))
//
//	// Get the reading out
//	reading := <-d.readChannel
//	assert.Equal(t, 1, len(reading.Reading))
//	assert.Equal(t, "foo", reading.Reading[0].Type)
//	assert.Equal(t, "ok", reading.Reading[0].Value)
//}
//
//// TestDataManager_readOkWithLimiter tests reading a device when a limiter is
//// configured.
//func TestDataManager_readOneOkWithLimiter(t *testing.T) {
//	defer Config.reset()
//
//	Config.Plugin = &PluginConfig{
//		Version: 1,
//		Network: &NetworkSettings{
//			Type:    "tcp",
//			Address: "test",
//		},
//		Settings: &PluginSettings{
//			Read:        &ReadSettings{Buffer: 200},
//			Write:       &WriteSettings{Buffer: 200},
//			Listen:      &ListenSettings{Buffer: 100},
//			Transaction: &TransactionSettings{TTL: "2s"},
//		},
//		Limiter: &LimiterSettings{Rate: 200, Burst: 200},
//	}
//
//	// Create the device to read
//	device := &Device{
//		Kind:     "test.state",
//		Location: &Location{Rack: "rack", Board: "board"},
//		Outputs: []*Output{
//			{
//				OutputType: OutputType{
//					Name: "foo",
//				},
//			},
//		},
//		Handler: &DeviceHandler{
//			Read: func(d *Device) ([]*Reading, error) {
//				output := d.GetOutput("foo")
//				reading, err := output.MakeReading("ok")
//				if err != nil {
//					return nil, err
//				}
//				return []*Reading{reading}, nil
//			},
//		},
//	}
//
//	d := newDataManager()
//	err := d.setup()
//	assert.NoError(t, err)
//
//	// Pass a reading in
//	assert.Equal(t, 0, len(d.readChannel))
//	d.readOne(device)
//	assert.Equal(t, 1, len(d.readChannel))
//
//	// Get the reading out
//	reading := <-d.readChannel
//	assert.Equal(t, 1, len(reading.Reading))
//	assert.Equal(t, "foo", reading.Reading[0].Type)
//	assert.Equal(t, "ok", reading.Reading[0].Value)
//}
//
//// TestDataManager_readErr tests reading a device that results in error.
//func TestDataManager_readOneErr(t *testing.T) {
//	defer Config.reset()
//
//	Config.Plugin = &PluginConfig{
//		Version: 1,
//		Network: &NetworkSettings{
//			Type:    "tcp",
//			Address: "test",
//		},
//		Settings: &PluginSettings{
//			Read:        &ReadSettings{Buffer: 200},
//			Write:       &WriteSettings{Buffer: 200},
//			Listen:      &ListenSettings{Buffer: 100},
//			Transaction: &TransactionSettings{TTL: "2s"},
//		},
//	}
//
//	// Create the device to read
//	device := &Device{
//		Kind:     "test.state",
//		Location: &Location{Rack: "rack", Board: "board"},
//		Outputs: []*Output{
//			{
//				OutputType: OutputType{
//					Name: "foo",
//				},
//			},
//		},
//		Handler: &DeviceHandler{
//			Read: func(d *Device) ([]*Reading, error) {
//				return nil, fmt.Errorf("test read error")
//			},
//		},
//	}
//
//	d := newDataManager()
//	err := d.setup()
//	assert.NoError(t, err)
//
//	assert.Equal(t, 0, len(d.readChannel))
//	d.readOne(device)
//	assert.Equal(t, 0, len(d.readChannel))
//}
//
//// TestDataManager_readBulkOkNoLimiter tests bulk reading a device when a limiter is
//// not configured.
//func TestDataManager_readBulkOkNoLimiter(t *testing.T) {
//	defer func() {
//		Config.reset()
//		resetContext()
//	}()
//
//	Config.Plugin = &PluginConfig{
//		Version: 1,
//		Network: &NetworkSettings{
//			Type:    "tcp",
//			Address: "test",
//		},
//		Settings: &PluginSettings{
//			Read:        &ReadSettings{Buffer: 200},
//			Write:       &WriteSettings{Buffer: 200},
//			Listen:      &ListenSettings{Buffer: 100},
//			Transaction: &TransactionSettings{TTL: "2s"},
//		},
//	}
//
//	handler := &DeviceHandler{
//		BulkRead: func(devices []*Device) ([]*ReadContext, error) {
//			var ctxs []*ReadContext
//			for _, d := range devices {
//				output := d.GetOutput("foo")
//				reading, err := output.MakeReading("ok")
//				if err != nil {
//					return nil, err
//				}
//				ctx := &ReadContext{
//					Rack:    "rack",
//					Board:   "board",
//					Device:  "device",
//					Reading: []*Reading{reading},
//				}
//				ctxs = append(ctxs, ctx)
//			}
//			return ctxs, nil
//		},
//	}
//	ctx.deviceHandlers = []*DeviceHandler{handler}
//
//	// Create the device to read
//	device := &Device{
//		id:       "device",
//		Kind:     "test.state",
//		Location: &Location{Rack: "rack", Board: "board"},
//		Handler:  handler,
//		Outputs: []*Output{
//			{
//				OutputType: OutputType{
//					Name: "foo",
//				},
//			},
//		},
//	}
//
//	ctx.devices["test-device-1"] = device
//
//	d := newDataManager()
//	err := d.setup()
//	assert.NoError(t, err)
//
//	// Pass a reading in
//	assert.Equal(t, 0, len(d.readChannel))
//	d.readBulk(handler)
//	assert.Equal(t, 1, len(d.readChannel))
//
//	// Get the reading out
//	reading := <-d.readChannel
//	assert.Equal(t, 1, len(reading.Reading))
//	assert.Equal(t, "foo", reading.Reading[0].Type)
//	assert.Equal(t, "ok", reading.Reading[0].Value)
//}
//
//// TestDataManager_readBulkOkWithLimiter tests bulk reading a device when a limiter is
//// configured.
//func TestDataManager_readBulkOkWithLimiter(t *testing.T) {
//	defer func() {
//		Config.reset()
//		resetContext()
//	}()
//
//	Config.Plugin = &PluginConfig{
//		Version: 1,
//		Network: &NetworkSettings{
//			Type:    "tcp",
//			Address: "test",
//		},
//		Settings: &PluginSettings{
//			Read:        &ReadSettings{Buffer: 200},
//			Write:       &WriteSettings{Buffer: 200},
//			Listen:      &ListenSettings{Buffer: 100},
//			Transaction: &TransactionSettings{TTL: "2s"},
//		},
//		Limiter: &LimiterSettings{Rate: 200, Burst: 200},
//	}
//
//	handler := &DeviceHandler{
//		BulkRead: func(devices []*Device) ([]*ReadContext, error) {
//			var ctxs []*ReadContext
//			for _, d := range devices {
//				output := d.GetOutput("foo")
//				reading, err := output.MakeReading("ok")
//				if err != nil {
//					return nil, err
//				}
//				ctx := &ReadContext{
//					Rack:    "rack",
//					Board:   "board",
//					Device:  "device",
//					Reading: []*Reading{reading},
//				}
//				ctxs = append(ctxs, ctx)
//			}
//			return ctxs, nil
//		},
//	}
//	ctx.deviceHandlers = []*DeviceHandler{handler}
//
//	// Create the device to read
//	device := &Device{
//		id:       "device",
//		Kind:     "test.state",
//		Location: &Location{Rack: "rack", Board: "board"},
//		Handler:  handler,
//		Outputs: []*Output{
//			{
//				OutputType: OutputType{
//					Name: "foo",
//				},
//			},
//		},
//	}
//
//	ctx.devices["test-device-1"] = device
//
//	d := newDataManager()
//	err := d.setup()
//	assert.NoError(t, err)
//
//	// Pass a reading in
//	assert.Equal(t, 0, len(d.readChannel))
//	d.readBulk(handler)
//	assert.Equal(t, 1, len(d.readChannel))
//
//	// Get the reading out
//	reading := <-d.readChannel
//	assert.Equal(t, 1, len(reading.Reading))
//	assert.Equal(t, "foo", reading.Reading[0].Type)
//	assert.Equal(t, "ok", reading.Reading[0].Value)
//}
//
//// TestDataManager_readBulkError tests bulk reading a device when reading returns an error.
//func TestDataManager_readBulkError(t *testing.T) {
//	defer func() {
//		Config.reset()
//		resetContext()
//	}()
//
//	Config.Plugin = &PluginConfig{
//		Version: 1,
//		Network: &NetworkSettings{
//			Type:    "tcp",
//			Address: "test",
//		},
//		Settings: &PluginSettings{
//			Read:        &ReadSettings{Buffer: 200},
//			Write:       &WriteSettings{Buffer: 200},
//			Listen:      &ListenSettings{Buffer: 100},
//			Transaction: &TransactionSettings{TTL: "2s"},
//		},
//	}
//
//	handler := &DeviceHandler{
//		BulkRead: func(devices []*Device) ([]*ReadContext, error) {
//			return nil, fmt.Errorf("test error")
//		},
//	}
//	ctx.deviceHandlers = []*DeviceHandler{handler}
//
//	// Create the device to read
//	device := &Device{
//		id:       "device",
//		Kind:     "test.state",
//		Location: &Location{Rack: "rack", Board: "board"},
//		Handler:  handler,
//		Outputs: []*Output{
//			{
//				OutputType: OutputType{
//					Name: "foo",
//				},
//			},
//		},
//	}
//
//	ctx.devices["test-device-1"] = device
//
//	d := newDataManager()
//	err := d.setup()
//	assert.NoError(t, err)
//
//	// Pass a reading in
//	assert.Equal(t, 0, len(d.readChannel))
//	d.readBulk(handler)
//	assert.Equal(t, 0, len(d.readChannel))
//}
//
//// TestDataManager_serialReadSingle tests reading a single device serially.
//func TestDataManager_serialReadSingle(t *testing.T) {
//	defer func() {
//		Config.reset()
//		resetContext()
//	}()
//
//	Config.Plugin = &PluginConfig{
//		Version: 1,
//		Network: &NetworkSettings{
//			Type:    "tcp",
//			Address: "test",
//		},
//		Settings: &PluginSettings{
//			Read:        &ReadSettings{Buffer: 200},
//			Write:       &WriteSettings{Buffer: 200},
//			Listen:      &ListenSettings{Buffer: 100},
//			Transaction: &TransactionSettings{TTL: "2s"},
//		},
//	}
//
//	// Create the device to read
//	device := &Device{
//		Kind:     "test.state",
//		Location: &Location{Rack: "rack", Board: "board"},
//		Outputs: []*Output{
//			{
//				OutputType: OutputType{
//					Name: "foo",
//				},
//			},
//		},
//		Handler: &DeviceHandler{
//			Read: func(d *Device) ([]*Reading, error) {
//				output := d.GetOutput("foo")
//				reading, err := output.MakeReading("ok")
//				if err != nil {
//					return nil, err
//				}
//				return []*Reading{reading}, nil
//			},
//		},
//	}
//
//	ctx.devices["test-id-1"] = device
//
//	d := newDataManager()
//	err := d.setup()
//	assert.NoError(t, err)
//
//	// Pass a reading in
//	assert.Equal(t, 0, len(d.readChannel))
//	d.serialRead(time.Nanosecond)
//	assert.Equal(t, 1, len(d.readChannel))
//
//	// Get the reading out
//	reading := <-d.readChannel
//	assert.Equal(t, 1, len(reading.Reading))
//	assert.Equal(t, "foo", reading.Reading[0].Type)
//	assert.Equal(t, "ok", reading.Reading[0].Value)
//}
//
//// TestDataManager_serialReadSingleBulk tests reading a single device in bulk serially.
//func TestDataManager_serialReadSingleBulk(t *testing.T) {
//	defer func() {
//		Config.reset()
//		resetContext()
//	}()
//
//	Config.Plugin = &PluginConfig{
//		Version: 1,
//		Network: &NetworkSettings{
//			Type:    "tcp",
//			Address: "test",
//		},
//		Settings: &PluginSettings{
//			Read:        &ReadSettings{Buffer: 200},
//			Write:       &WriteSettings{Buffer: 200},
//			Listen:      &ListenSettings{Buffer: 100},
//			Transaction: &TransactionSettings{TTL: "2s"},
//		},
//	}
//
//	handler := &DeviceHandler{
//		BulkRead: func(devices []*Device) ([]*ReadContext, error) {
//			var ctxs []*ReadContext
//			for _, d := range devices {
//				output := d.GetOutput("foo")
//				reading, err := output.MakeReading("ok")
//				if err != nil {
//					return nil, err
//				}
//				ctx := &ReadContext{
//					Rack:    "rack",
//					Board:   "board",
//					Device:  "device",
//					Reading: []*Reading{reading},
//				}
//				ctxs = append(ctxs, ctx)
//			}
//			return ctxs, nil
//		},
//	}
//	ctx.deviceHandlers = []*DeviceHandler{handler}
//
//	// Create the device to read
//	device := &Device{
//		Kind:     "test.state",
//		Location: &Location{Rack: "rack", Board: "board"},
//		Outputs: []*Output{
//			{
//				OutputType: OutputType{
//					Name: "foo",
//				},
//			},
//		},
//		Handler:  handler,
//		bulkRead: true,
//	}
//
//	ctx.devices["test-id-1"] = device
//
//	d := newDataManager()
//	err := d.setup()
//	assert.NoError(t, err)
//
//	// Pass a reading in
//	assert.Equal(t, 0, len(d.readChannel))
//	d.serialRead(time.Nanosecond)
//	assert.Equal(t, 1, len(d.readChannel))
//
//	// Get the reading out
//	reading := <-d.readChannel
//	assert.Equal(t, 1, len(reading.Reading))
//	assert.Equal(t, "foo", reading.Reading[0].Type)
//	assert.Equal(t, "ok", reading.Reading[0].Value)
//}
//
//// TestDataManager_parallelReadSingle tests reading a single device in parallel.
//func TestDataManager_parallelReadSingle(t *testing.T) {
//	defer func() {
//		Config.reset()
//		resetContext()
//	}()
//
//	Config.Plugin = &PluginConfig{
//		Version: 1,
//		Network: &NetworkSettings{
//			Type:    "tcp",
//			Address: "test",
//		},
//		Settings: &PluginSettings{
//			Read:        &ReadSettings{Buffer: 200},
//			Write:       &WriteSettings{Buffer: 200},
//			Listen:      &ListenSettings{Buffer: 100},
//			Transaction: &TransactionSettings{TTL: "2s"},
//		},
//	}
//
//	// Create the device to read
//	device := &Device{
//		Kind:     "test.state",
//		Location: &Location{Rack: "rack", Board: "board"},
//		Outputs: []*Output{
//			{
//				OutputType: OutputType{
//					Name: "foo",
//				},
//			},
//		},
//		Handler: &DeviceHandler{
//			Read: func(d *Device) ([]*Reading, error) {
//				output := d.GetOutput("foo")
//				reading, err := output.MakeReading("ok")
//				if err != nil {
//					return nil, err
//				}
//				return []*Reading{reading}, nil
//			},
//		},
//	}
//
//	// Clear the global device map then add the device to it
//	ctx.devices["test-id-1"] = device
//
//	d := newDataManager()
//	err := d.setup()
//	assert.NoError(t, err)
//
//	// Pass a reading in
//	assert.Equal(t, 0, len(d.readChannel))
//	d.parallelRead()
//	assert.Equal(t, 1, len(d.readChannel))
//
//	// Get the reading out
//	reading := <-d.readChannel
//	assert.Equal(t, 1, len(reading.Reading))
//	assert.Equal(t, "foo", reading.Reading[0].Type)
//	assert.Equal(t, "ok", reading.Reading[0].Value)
//}
//
//// TestDataManager_parallelReadSingleBulk tests reading a single device in bulk in parallel.
//func TestDataManager_parallelReadSingleBulk(t *testing.T) {
//	defer func() {
//		Config.reset()
//		resetContext()
//	}()
//
//	Config.Plugin = &PluginConfig{
//		Version: 1,
//		Network: &NetworkSettings{
//			Type:    "tcp",
//			Address: "test",
//		},
//		Settings: &PluginSettings{
//			Read:        &ReadSettings{Buffer: 200},
//			Write:       &WriteSettings{Buffer: 200},
//			Listen:      &ListenSettings{Buffer: 100},
//			Transaction: &TransactionSettings{TTL: "2s"},
//		},
//	}
//
//	handler := &DeviceHandler{
//		BulkRead: func(devices []*Device) ([]*ReadContext, error) {
//			var ctxs []*ReadContext
//			for _, d := range devices {
//				reading, err := d.GetOutput("foo").MakeReading("ok")
//				if err != nil {
//					return nil, err
//				}
//				ctx := &ReadContext{
//					Rack:    "rack",
//					Board:   "board",
//					Device:  "device",
//					Reading: []*Reading{reading},
//				}
//				ctxs = append(ctxs, ctx)
//			}
//			return ctxs, nil
//		},
//	}
//	ctx.deviceHandlers = []*DeviceHandler{handler}
//
//	// Create the device to read
//	device := &Device{
//		Kind:     "test.state",
//		Location: &Location{Rack: "rack", Board: "board"},
//		Outputs: []*Output{
//			{
//				OutputType: OutputType{
//					Name: "foo",
//				},
//			},
//		},
//		Handler:  handler,
//		bulkRead: true,
//	}
//
//	// Clear the global device map then add the device to it
//	ctx.devices["test-id-1"] = device
//
//	d := newDataManager()
//	err := d.setup()
//	assert.NoError(t, err)
//
//	// Pass a reading in
//	assert.Equal(t, 0, len(d.readChannel))
//	d.parallelRead()
//	assert.Equal(t, 1, len(d.readChannel))
//
//	// Get the reading out
//	reading := <-d.readChannel
//	assert.Equal(t, 1, len(reading.Reading))
//	assert.Equal(t, "foo", reading.Reading[0].Type)
//	assert.Equal(t, "ok", reading.Reading[0].Value)
//}
//
//// TestDataManager_serialReadMultiple tests reading multiple devices serially.
//func TestDataManager_serialReadMultiple(t *testing.T) {
//	defer func() {
//		Config.reset()
//		resetContext()
//	}()
//
//	Config.Plugin = &PluginConfig{
//		Version: 1,
//		Network: &NetworkSettings{
//			Type:    "tcp",
//			Address: "test",
//		},
//		Settings: &PluginSettings{
//			Read:        &ReadSettings{Buffer: 200},
//			Write:       &WriteSettings{Buffer: 200},
//			Listen:      &ListenSettings{Buffer: 100},
//			Transaction: &TransactionSettings{TTL: "2s"},
//		},
//	}
//
//	// Create the devices to read
//	device1 := &Device{
//		Kind:     "test.1.state",
//		Location: &Location{Rack: "rack", Board: "board"},
//		Outputs: []*Output{
//			{
//				OutputType: OutputType{
//					Name: "foo",
//				},
//			},
//		},
//		Handler: &DeviceHandler{
//			Read: func(d *Device) ([]*Reading, error) {
//				reading, err := d.GetOutput("foo").MakeReading("ok")
//				if err != nil {
//					return nil, err
//				}
//				return []*Reading{reading}, nil
//			},
//		},
//	}
//	device2 := &Device{
//		Kind:     "test.2.state",
//		Location: &Location{Rack: "rack", Board: "board"},
//		Outputs: []*Output{
//			{
//				OutputType: OutputType{
//					Name: "foo",
//				},
//			},
//		},
//		Handler: &DeviceHandler{
//			Read: func(d *Device) ([]*Reading, error) {
//				reading, err := d.GetOutput("foo").MakeReading("ok")
//				if err != nil {
//					return nil, err
//				}
//				return []*Reading{reading}, nil
//			},
//		},
//	}
//	device3 := &Device{
//		Kind:     "test.3.state",
//		Location: &Location{Rack: "rack", Board: "board"},
//		Outputs: []*Output{
//			{
//				OutputType: OutputType{
//					Name: "foo",
//				},
//			},
//		},
//		Handler: &DeviceHandler{
//			Read: func(d *Device) ([]*Reading, error) {
//				reading, err := d.GetOutput("foo").MakeReading("ok")
//				if err != nil {
//					return nil, err
//				}
//				return []*Reading{reading}, nil
//			},
//		},
//	}
//
//	ctx.devices["test-id-1"] = device1
//	ctx.devices["test-id-2"] = device2
//	ctx.devices["test-id-3"] = device3
//
//	d := newDataManager()
//	err := d.setup()
//	assert.NoError(t, err)
//
//	assert.Equal(t, 0, len(d.readChannel))
//	d.serialRead(time.Nanosecond)
//	assert.Equal(t, 3, len(d.readChannel))
//
//	for i := 0; i < 3; i++ {
//		reading := <-d.readChannel
//		assert.Equal(t, 1, len(reading.Reading))
//		assert.Equal(t, "foo", reading.Reading[0].Type)
//		assert.Equal(t, "ok", reading.Reading[0].Value)
//	}
//}
//
//// TestDataManager_parallelReadSingle tests reading multiple devices in parallel.
//func TestDataManager_parallelReadMultiple(t *testing.T) {
//	defer func() {
//		Config.reset()
//		resetContext()
//	}()
//
//	Config.Plugin = &PluginConfig{
//		Version: 1,
//		Network: &NetworkSettings{
//			Type:    "tcp",
//			Address: "test",
//		},
//		Settings: &PluginSettings{
//			Read:        &ReadSettings{Buffer: 200},
//			Write:       &WriteSettings{Buffer: 200},
//			Listen:      &ListenSettings{Buffer: 100},
//			Transaction: &TransactionSettings{TTL: "2s"},
//		},
//	}
//
//	// Create the devices to read
//	device1 := &Device{
//		Kind:     "test.1.state",
//		Location: &Location{Rack: "rack", Board: "board"},
//		Outputs: []*Output{
//			{
//				OutputType: OutputType{
//					Name: "foo",
//				},
//			},
//		},
//		Handler: &DeviceHandler{
//			Read: func(d *Device) ([]*Reading, error) {
//				reading, err := d.GetOutput("foo").MakeReading("ok")
//				if err != nil {
//					return nil, err
//				}
//				return []*Reading{reading}, nil
//			},
//		},
//	}
//	device2 := &Device{
//		Kind:     "test.2.state",
//		Location: &Location{Rack: "rack", Board: "board"},
//		Outputs: []*Output{
//			{
//				OutputType: OutputType{
//					Name: "foo",
//				},
//			},
//		},
//		Handler: &DeviceHandler{
//			Read: func(d *Device) ([]*Reading, error) {
//				reading, err := d.GetOutput("foo").MakeReading("ok")
//				if err != nil {
//					return nil, err
//				}
//				return []*Reading{reading}, nil
//			},
//		},
//	}
//	device3 := &Device{
//		Kind:     "test.3.state",
//		Location: &Location{Rack: "rack", Board: "board"},
//		Outputs: []*Output{
//			{
//				OutputType: OutputType{
//					Name: "foo",
//				},
//			},
//		},
//		Handler: &DeviceHandler{
//			Read: func(d *Device) ([]*Reading, error) {
//				reading, err := d.GetOutput("foo").MakeReading("ok")
//				if err != nil {
//					return nil, err
//				}
//				return []*Reading{reading}, nil
//			},
//		},
//	}
//
//	ctx.devices["test-id-1"] = device1
//	ctx.devices["test-id-2"] = device2
//	ctx.devices["test-id-3"] = device3
//
//	d := newDataManager()
//	err := d.setup()
//	assert.NoError(t, err)
//
//	assert.Equal(t, 0, len(d.readChannel))
//	d.parallelRead()
//	assert.Equal(t, 3, len(d.readChannel))
//
//	for i := 0; i < 3; i++ {
//		reading := <-d.readChannel
//		assert.Equal(t, 1, len(reading.Reading))
//		assert.Equal(t, "foo", reading.Reading[0].Type)
//		assert.Equal(t, "ok", reading.Reading[0].Value)
//	}
//}
//
//// TestDataManager_writeOkNoLimiter tests writing to a device with no limiter configured.
//func TestDataManager_writeOkNoLimiter(t *testing.T) {
//	defer func() {
//		Config.reset()
//		resetContext()
//	}()
//	setupTransactionCache(time.Duration(600) * time.Second)
//
//	Config.Plugin = &PluginConfig{
//		Version: 1,
//		Network: &NetworkSettings{
//			Type:    "tcp",
//			Address: "test",
//		},
//		Settings: &PluginSettings{
//			Read:        &ReadSettings{Buffer: 200},
//			Write:       &WriteSettings{Buffer: 200},
//			Listen:      &ListenSettings{Buffer: 100},
//			Transaction: &TransactionSettings{TTL: "2s"},
//		},
//	}
//
//	// Create the device to read
//	device := &Device{
//		Kind:     "test.state",
//		Location: &Location{Rack: "rack", Board: "board"},
//		Outputs: []*Output{
//			{
//				OutputType: OutputType{
//					Name: "foo",
//				},
//			},
//		},
//		Handler: &DeviceHandler{
//			Write: func(device *Device, data *WriteData) error {
//				return nil
//			},
//		},
//	}
//	ctx.devices["rack-board-device"] = device
//
//	d := newDataManager()
//	err := d.setup()
//	assert.NoError(t, err)
//
//	ctx := &WriteContext{
//		transaction: newTransaction(),
//		device:      "device",
//		board:       "board",
//		rack:        "rack",
//		data: &synse.WriteData{
//			Action: "test",
//		},
//	}
//
//	d.write(ctx)
//
//	assert.Equal(t, stateOk, ctx.transaction.state)
//	assert.Equal(t, statusDone, ctx.transaction.status)
//	assert.Equal(t, "", ctx.transaction.message)
//}
//
//// TestDataManager_writeOkWithLimiter tests writing to a device with a limiter configured.
//func TestDataManager_writeOkWithLimiter(t *testing.T) {
//	defer func() {
//		Config.reset()
//		resetContext()
//	}()
//	setupTransactionCache(time.Duration(600) * time.Second)
//
//	Config.Plugin = &PluginConfig{
//		Version: 1,
//		Network: &NetworkSettings{
//			Type:    "tcp",
//			Address: "test",
//		},
//		Settings: &PluginSettings{
//			Read:        &ReadSettings{Buffer: 200},
//			Write:       &WriteSettings{Buffer: 200},
//			Listen:      &ListenSettings{Buffer: 100},
//			Transaction: &TransactionSettings{TTL: "2s"},
//		},
//		Limiter: &LimiterSettings{Rate: 200, Burst: 200},
//	}
//
//	// Create the device to read
//	device := &Device{
//		Kind:     "test.state",
//		Location: &Location{Rack: "rack", Board: "board"},
//		Outputs: []*Output{
//			{
//				OutputType: OutputType{
//					Name: "foo",
//				},
//			},
//		},
//		Handler: &DeviceHandler{
//			Write: func(device *Device, data *WriteData) error {
//				return nil
//			},
//		},
//	}
//	ctx.devices["rack-board-device"] = device
//
//	d := newDataManager()
//	err := d.setup()
//	assert.NoError(t, err)
//
//	ctx := &WriteContext{
//		transaction: newTransaction(),
//		device:      "device",
//		board:       "board",
//		rack:        "rack",
//		data: &synse.WriteData{
//			Action: "test",
//		},
//	}
//
//	d.write(ctx)
//
//	assert.Equal(t, stateOk, ctx.transaction.state)
//	assert.Equal(t, statusDone, ctx.transaction.status)
//	assert.Equal(t, "", ctx.transaction.message)
//}
//
//// TestDataManager_writeNoDevice tests writing to a device when that device cannot be found.
//func TestDataManager_writeNoDevice(t *testing.T) {
//	defer func() {
//		Config.reset()
//		resetContext()
//	}()
//	setupTransactionCache(time.Duration(600) * time.Second)
//
//	Config.Plugin = &PluginConfig{
//		Version: 1,
//		Network: &NetworkSettings{
//			Type:    "tcp",
//			Address: "test",
//		},
//		Settings: &PluginSettings{
//			Read:        &ReadSettings{Buffer: 200},
//			Write:       &WriteSettings{Buffer: 200},
//			Listen:      &ListenSettings{Buffer: 100},
//			Transaction: &TransactionSettings{TTL: "2s"},
//		},
//	}
//
//	d := newDataManager()
//	err := d.setup()
//	assert.NoError(t, err)
//
//	ctx := &WriteContext{
//		transaction: newTransaction(),
//		device:      "device",
//		board:       "board",
//		rack:        "rack",
//		data: &synse.WriteData{
//			Action: "test",
//		},
//	}
//
//	d.write(ctx)
//
//	assert.Equal(t, stateError, ctx.transaction.state)
//	assert.Equal(t, statusDone, ctx.transaction.status)
//	assert.NotEmpty(t, ctx.transaction.message)
//}
//
//// TestDataManager_writeError tests writing to a device when the write errors.
//func TestDataManager_writeError(t *testing.T) {
//	defer func() {
//		Config.reset()
//		resetContext()
//	}()
//	setupTransactionCache(time.Duration(600) * time.Second)
//
//	Config.Plugin = &PluginConfig{
//		Version: 1,
//		Network: &NetworkSettings{
//			Type:    "tcp",
//			Address: "test",
//		},
//		Settings: &PluginSettings{
//			Read:        &ReadSettings{Buffer: 200},
//			Write:       &WriteSettings{Buffer: 200},
//			Listen:      &ListenSettings{Buffer: 100},
//			Transaction: &TransactionSettings{TTL: "2s"},
//		},
//	}
//
//	// Create the device to read
//	device := &Device{
//		Kind:     "test.state",
//		Location: &Location{Rack: "rack", Board: "board"},
//		Outputs: []*Output{
//			{
//				OutputType: OutputType{
//					Name: "foo",
//				},
//			},
//		},
//		Handler: &DeviceHandler{
//			Write: func(device *Device, data *WriteData) error {
//				return fmt.Errorf("test error")
//			},
//		},
//	}
//	ctx.devices["rack-board-device"] = device
//
//	d := newDataManager()
//	err := d.setup()
//	assert.NoError(t, err)
//
//	ctx := &WriteContext{
//		transaction: newTransaction(),
//		device:      "device",
//		board:       "board",
//		rack:        "rack",
//		data: &synse.WriteData{
//			Action: "test",
//		},
//	}
//
//	d.write(ctx)
//
//	assert.Equal(t, stateError, ctx.transaction.state)
//	assert.Equal(t, statusDone, ctx.transaction.status)
//	assert.NotEmpty(t, ctx.transaction.message)
//}
//
//// TestDataManager_serialWriteSingle tests writing to a single device in serial.
//func TestDataManager_serialWriteSingle(t *testing.T) {
//	defer func() {
//		Config.reset()
//		resetContext()
//	}()
//	setupTransactionCache(time.Duration(600) * time.Second)
//
//	Config.Plugin = &PluginConfig{
//		Version: 1,
//		Network: &NetworkSettings{
//			Type:    "tcp",
//			Address: "test",
//		},
//		Settings: &PluginSettings{
//			Read:        &ReadSettings{Buffer: 200},
//			Write:       &WriteSettings{Buffer: 200, Max: 200},
//			Listen:      &ListenSettings{Buffer: 100},
//			Transaction: &TransactionSettings{TTL: "2s"},
//		},
//		Limiter: &LimiterSettings{Rate: 200, Burst: 200},
//	}
//
//	// Create the device to read
//	device := &Device{
//		Kind:     "test.state",
//		Location: &Location{Rack: "rack", Board: "board"},
//		Outputs: []*Output{
//			{
//				OutputType: OutputType{
//					Name: "foo",
//				},
//			},
//		},
//		Handler: &DeviceHandler{
//			Write: func(device *Device, data *WriteData) error {
//				return nil
//			},
//		},
//	}
//	ctx.devices["rack-board-device"] = device
//
//	d := newDataManager()
//	err := d.setup()
//	assert.NoError(t, err)
//
//	ctx := &WriteContext{
//		transaction: newTransaction(),
//		device:      "device",
//		board:       "board",
//		rack:        "rack",
//		data: &synse.WriteData{
//			Action: "test",
//		},
//	}
//
//	d.writeChannel <- ctx
//
//	d.serialWrite()
//
//	assert.Equal(t, stateOk, ctx.transaction.state)
//	assert.Equal(t, statusDone, ctx.transaction.status)
//	assert.Equal(t, "", ctx.transaction.message)
//}
//
//// TestDataManager_serialWriteMultiple tests writing to multiple devices in serial.
//func TestDataManager_serialWriteMultiple(t *testing.T) {
//	defer func() {
//		Config.reset()
//		resetContext()
//	}()
//	setupTransactionCache(time.Duration(600) * time.Second)
//
//	Config.Plugin = &PluginConfig{
//		Version: 1,
//		Network: &NetworkSettings{
//			Type:    "tcp",
//			Address: "test",
//		},
//		Settings: &PluginSettings{
//			Read:        &ReadSettings{Buffer: 200},
//			Write:       &WriteSettings{Buffer: 200, Max: 200},
//			Listen:      &ListenSettings{Buffer: 100},
//			Transaction: &TransactionSettings{TTL: "2s"},
//		},
//		Limiter: &LimiterSettings{Rate: 200, Burst: 200},
//	}
//
//	// Create the device to read
//	device := &Device{
//		Kind:     "test.state",
//		Location: &Location{Rack: "rack", Board: "board"},
//		Outputs: []*Output{
//			{
//				OutputType: OutputType{
//					Name: "foo",
//				},
//			},
//		},
//		Handler: &DeviceHandler{
//			Write: func(device *Device, data *WriteData) error {
//				return nil
//			},
//		},
//	}
//	ctx.devices["rack-board-device"] = device
//
//	d := newDataManager()
//	err := d.setup()
//	assert.NoError(t, err)
//
//	ctx1 := &WriteContext{
//		transaction: newTransaction(),
//		device:      "device",
//		board:       "board",
//		rack:        "rack",
//		data: &synse.WriteData{
//			Action: "test",
//		},
//	}
//	ctx2 := &WriteContext{
//		transaction: newTransaction(),
//		device:      "device",
//		board:       "board",
//		rack:        "rack",
//		data: &synse.WriteData{
//			Action: "test",
//		},
//	}
//	ctx3 := &WriteContext{
//		transaction: newTransaction(),
//		device:      "device",
//		board:       "board",
//		rack:        "rack",
//		data: &synse.WriteData{
//			Action: "test",
//		},
//	}
//
//	d.writeChannel <- ctx1
//	d.writeChannel <- ctx2
//	d.writeChannel <- ctx3
//
//	d.serialWrite()
//
//	for _, ctx := range []*WriteContext{ctx1, ctx2, ctx3} {
//		assert.Equal(t, stateOk, ctx.transaction.state)
//		assert.Equal(t, statusDone, ctx.transaction.status)
//		assert.Equal(t, "", ctx.transaction.message)
//	}
//}
//
//// TestDataManager_parallelWriteSingle tests writing to a single device in parallel.
//func TestDataManager_parallelWriteSingle(t *testing.T) {
//	defer func() {
//		Config.reset()
//		resetContext()
//	}()
//	setupTransactionCache(time.Duration(600) * time.Second)
//
//	Config.Plugin = &PluginConfig{
//		Version: 1,
//		Network: &NetworkSettings{
//			Type:    "tcp",
//			Address: "test",
//		},
//		Settings: &PluginSettings{
//			Read:        &ReadSettings{Buffer: 200},
//			Write:       &WriteSettings{Buffer: 200, Max: 200},
//			Listen:      &ListenSettings{Buffer: 100},
//			Transaction: &TransactionSettings{TTL: "2s"},
//		},
//		Limiter: &LimiterSettings{Rate: 200, Burst: 200},
//	}
//
//	// Create the device to read
//	device := &Device{
//		Kind:     "test.state",
//		Location: &Location{Rack: "rack", Board: "board"},
//		Outputs: []*Output{
//			{
//				OutputType: OutputType{
//					Name: "foo",
//				},
//			},
//		},
//		Handler: &DeviceHandler{
//			Write: func(device *Device, data *WriteData) error {
//				return nil
//			},
//		},
//	}
//	ctx.devices["rack-board-device"] = device
//
//	d := newDataManager()
//	err := d.setup()
//	assert.NoError(t, err)
//
//	ctx := &WriteContext{
//		transaction: newTransaction(),
//		device:      "device",
//		board:       "board",
//		rack:        "rack",
//		data: &synse.WriteData{
//			Action: "test",
//		},
//	}
//
//	d.writeChannel <- ctx
//
//	d.parallelWrite()
//
//	assert.Equal(t, stateOk, ctx.transaction.state)
//	assert.Equal(t, statusDone, ctx.transaction.status)
//	assert.Equal(t, "", ctx.transaction.message)
//}
//
//// TestDataManager_parallelWriteMultiple tests writing to multiple devices in parallel.
//func TestDataManager_parallelWriteMultiple(t *testing.T) {
//	defer func() {
//		Config.reset()
//		resetContext()
//	}()
//	setupTransactionCache(time.Duration(600) * time.Second)
//
//	Config.Plugin = &PluginConfig{
//		Version: 1,
//		Network: &NetworkSettings{
//			Type:    "tcp",
//			Address: "test",
//		},
//		Settings: &PluginSettings{
//			Read:        &ReadSettings{Buffer: 200},
//			Write:       &WriteSettings{Buffer: 200, Max: 200},
//			Listen:      &ListenSettings{Buffer: 100},
//			Transaction: &TransactionSettings{TTL: "2s"},
//		},
//		Limiter: &LimiterSettings{Rate: 200, Burst: 200},
//	}
//
//	// Create the device to read
//	device := &Device{
//		Kind:     "test.state",
//		Location: &Location{Rack: "rack", Board: "board"},
//		Outputs: []*Output{
//			{
//				OutputType: OutputType{
//					Name: "foo",
//				},
//			},
//		},
//		Handler: &DeviceHandler{
//			Write: func(device *Device, data *WriteData) error {
//				return nil
//			},
//		},
//	}
//	ctx.devices["rack-board-device"] = device
//
//	d := newDataManager()
//	err := d.setup()
//	assert.NoError(t, err)
//
//	ctx1 := &WriteContext{
//		transaction: newTransaction(),
//		device:      "device",
//		board:       "board",
//		rack:        "rack",
//		data: &synse.WriteData{
//			Action: "test",
//		},
//	}
//	ctx2 := &WriteContext{
//		transaction: newTransaction(),
//		device:      "device",
//		board:       "board",
//		rack:        "rack",
//		data: &synse.WriteData{
//			Action: "test",
//		},
//	}
//	ctx3 := &WriteContext{
//		transaction: newTransaction(),
//		device:      "device",
//		board:       "board",
//		rack:        "rack",
//		data: &synse.WriteData{
//			Action: "test",
//		},
//	}
//
//	d.writeChannel <- ctx1
//	d.writeChannel <- ctx2
//	d.writeChannel <- ctx3
//
//	d.parallelWrite()
//
//	for _, ctx := range []*WriteContext{ctx1, ctx2, ctx3} {
//		assert.Equal(t, stateOk, ctx.transaction.state)
//		assert.Equal(t, statusDone, ctx.transaction.status)
//		assert.Equal(t, "", ctx.transaction.message)
//	}
//}
//
//// Test creating a new instance of a listener context.
//func TestNewListenerCtx(t *testing.T) {
//	handler := &DeviceHandler{}
//	device := &Device{}
//
//	ctx := NewListenerCtx(handler, device)
//	assert.Equal(t, handler, ctx.handler)
//	assert.Equal(t, device, ctx.device)
//	assert.Equal(t, 0, ctx.restarts)
//}

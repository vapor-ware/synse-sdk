package sdk

//"testing"
//
//"github.com/stretchr/testify/assert"
//
//"fmt"
//
//"github.com/vapor-ware/synse-sdk/sdk/config"
//"golang.org/x/time/rate"

// DeviceID gets the unique identifiers out of the plugin-specific
// configuration to be used in UID generation. Required to construct Handlers.
func DeviceID(data map[string]string) string {
	return data["id"]
}

//
//// TestNewDataManager tests creating a new dataManager instance successfully.
//func TestNewDataManager(t *testing.T) {
//	// Create handlers.
//	h, err := NewHandlers(DeviceID, nil)
//	assert.NoError(t, err)
//
//	c := config.PluginConfig{
//		Version: "test",
//		Network: config.NetworkSettings{
//			Type:    "tcp",
//			Address: "test",
//		},
//		Settings: config.Settings{
//			Read:        config.ReadSettings{Buffer: 200},
//			Write:       config.WriteSettings{Buffer: 200},
//			Transaction: config.TransactionSettings{TTL: "2s"},
//		},
//	}
//	p := Plugin{handlers: h}
//	err = p.SetConfig(&c)
//	assert.NoError(t, err)
//
//	d, err := newDataManager(&p)
//	assert.NoError(t, err)
//
//	assert.Equal(t, 200, cap(d.writeChannel))
//	assert.Equal(t, 200, cap(d.readChannel))
//	assert.Equal(t, h, d.handlers)
//}
//
//// TestNewDataManager2 tests creating a new dataManager instance successfully with
//// a different configuration.
//func TestNewDataManager2(t *testing.T) {
//	// Create handlers.
//	h, err := NewHandlers(DeviceID, nil)
//	assert.NoError(t, err)
//
//	c := &config.PluginConfig{
//		Version: "test",
//		Network: config.NetworkSettings{
//			Type:    "tcp",
//			Address: "test",
//		},
//		Settings: config.Settings{
//			Read:        config.ReadSettings{Buffer: 500},
//			Write:       config.WriteSettings{Buffer: 500},
//			Transaction: config.TransactionSettings{TTL: "2s"},
//		},
//	}
//	p := Plugin{handlers: h}
//	err = p.SetConfig(c)
//	assert.NoError(t, err)
//
//	// Create the dataManager
//	d, err := newDataManager(&p)
//	assert.NoError(t, err)
//
//	assert.Equal(t, 500, cap(d.writeChannel))
//	assert.Equal(t, 500, cap(d.readChannel))
//	assert.Equal(t, h, d.handlers)
//}
//
//// TestNewDataManager_NilPlugin tests creating a new dataManager instance, passing
//// in a nil Plugin to the constructor.
//func TestNewDataManager_NilPlugin(t *testing.T) {
//	d, err := newDataManager(nil)
//	assert.Nil(t, d)
//	assert.Error(t, err)
//}
//
//// TestNewDataManager_NilPluginHandlers tests creating a new dataManager instance,
//// passing in a Plugin with nil handlers to the constructor.
//func TestNewDataManager_NilPluginHandlers(t *testing.T) {
//	p := Plugin{
//		handlers: nil,
//	}
//
//	d, err := newDataManager(&p)
//	assert.Nil(t, d)
//	assert.Error(t, err)
//}
//
//// TestNewDataManager_NilPluginConfig tests creating a new dataManager instance,
//// passing in a Plugin with nil config to the constructor.
//func TestNewDataManager_NilPluginConfig(t *testing.T) {
//	p := Plugin{
//		handlers: &Handlers{DeviceID, nil},
//	}
//
//	d, err := newDataManager(&p)
//	assert.Nil(t, d)
//	assert.Error(t, err)
//}
//
//// TestDataManager_WritesEnabled tests that writes are enabled in the dataManager
//// when they are enabled in the config.
//func TestDataManager_WritesEnabled(t *testing.T) {
//	// Create the plugin
//	c := config.PluginConfig{
//		Version: "test",
//		Network: config.NetworkSettings{
//			Type:    "tcp",
//			Address: "test",
//		},
//		Settings: config.Settings{
//			Read: config.ReadSettings{Buffer: 200},
//			Write: config.WriteSettings{
//				Enabled: true,
//				Buffer:  200,
//			},
//			Transaction: config.TransactionSettings{TTL: "2s"},
//		},
//	}
//	p := Plugin{
//		Config:   &c,
//		handlers: &Handlers{DeviceID, nil},
//	}
//
//	// Create the dataManager
//	d, err := newDataManager(&p)
//	assert.NoError(t, err)
//
//	assert.True(t, d.writesEnabled())
//}
//
//// TestDataManager_readNoLimiter tests reading a device when a limiter is
//// not configured.
//func TestDataManager_readOneOkNoLimiter(t *testing.T) {
//	// Create the plugin
//	p := Plugin{
//		Config: &config.PluginConfig{
//			Version: "test",
//			Network: config.NetworkSettings{
//				Type:    "tcp",
//				Address: "test",
//			},
//			Settings: config.Settings{
//				Read:        config.ReadSettings{Buffer: 200},
//				Write:       config.WriteSettings{Buffer: 200},
//				Transaction: config.TransactionSettings{TTL: "2s"},
//			},
//		},
//		handlers: &Handlers{DeviceID, nil},
//	}
//
//	// Create the device to read
//	device := makeTestDevice()
//	device.Handler = &DeviceHandler{
//		Read: func(d *Device) ([]*Reading, error) {
//			return []*Reading{NewReading("test", "ok")}, nil
//		},
//	}
//
//	// Create the dataManager
//	d, err := newDataManager(&p)
//	assert.NoError(t, err)
//
//	assert.Equal(t, 0, len(d.readChannel))
//	d.readOne(device)
//	assert.Equal(t, 1, len(d.readChannel))
//
//	reading := <-d.readChannel
//	assert.Equal(t, 1, len(reading.Reading))
//	assert.Equal(t, "test", reading.Reading[0].Type)
//	assert.Equal(t, "ok", reading.Reading[0].Value)
//}
//
//// TestDataManager_readOkWithLimiter tests reading a device when a limiter is
//// configured.
//func TestDataManager_readOneOkWithLimiter(t *testing.T) {
//	// Create the plugin
//	p := Plugin{
//		Config: &config.PluginConfig{
//			Version: "test",
//			Network: config.NetworkSettings{
//				Type:    "tcp",
//				Address: "test",
//			},
//			Settings: config.Settings{
//				Read:        config.ReadSettings{Buffer: 200},
//				Write:       config.WriteSettings{Buffer: 200},
//				Transaction: config.TransactionSettings{TTL: "2s"},
//			},
//			Limiter: rate.NewLimiter(200, 200),
//		},
//		handlers: &Handlers{DeviceID, nil},
//	}
//
//	// Create the device to read
//	device := makeTestDevice()
//	device.Handler = &DeviceHandler{
//		Read: func(d *Device) ([]*Reading, error) {
//			return []*Reading{NewReading("test", "ok")}, nil
//		},
//	}
//
//	// Create the dataManager
//	d, err := newDataManager(&p)
//	assert.NoError(t, err)
//
//	assert.Equal(t, 0, len(d.readChannel))
//	d.readOne(device)
//	assert.Equal(t, 1, len(d.readChannel))
//
//	reading := <-d.readChannel
//	assert.Equal(t, 1, len(reading.Reading))
//	assert.Equal(t, "test", reading.Reading[0].Type)
//	assert.Equal(t, "ok", reading.Reading[0].Value)
//}
//
//// TestDataManager_readErr tests reading a device that results in error.
//func TestDataManager_readOneErr(t *testing.T) {
//	// Create the plugin
//	p := Plugin{
//		Config: &config.PluginConfig{
//			Version: "test",
//			Network: config.NetworkSettings{
//				Type:    "tcp",
//				Address: "test",
//			},
//			Settings: config.Settings{
//				Read:        config.ReadSettings{Buffer: 200},
//				Write:       config.WriteSettings{Buffer: 200},
//				Transaction: config.TransactionSettings{TTL: "2s"},
//			},
//		},
//		handlers: &Handlers{DeviceID, nil},
//	}
//
//	// Create the device to read
//	device := makeTestDevice()
//	device.Handler = &DeviceHandler{
//		Read: func(d *Device) ([]*Reading, error) {
//			return nil, fmt.Errorf("test read error")
//		},
//	}
//
//	// Create the dataManager
//	d, err := newDataManager(&p)
//	assert.NoError(t, err)
//
//	assert.Equal(t, 0, len(d.readChannel))
//	d.readOne(device)
//	assert.Equal(t, 0, len(d.readChannel))
//}
//
//// TestDataManager_serialReadSingle tests reading a single device serially.
//func TestDataManager_serialReadSingle(t *testing.T) {
//	// Create the plugin
//	p := Plugin{
//		Config: &config.PluginConfig{
//			Version: "test",
//			Network: config.NetworkSettings{
//				Type:    "tcp",
//				Address: "test",
//			},
//			Settings: config.Settings{
//				Read:        config.ReadSettings{Buffer: 200},
//				Write:       config.WriteSettings{Buffer: 200},
//				Transaction: config.TransactionSettings{TTL: "2s"},
//			},
//		},
//		handlers: &Handlers{DeviceID, nil},
//	}
//
//	// Create the device to read
//	device := makeTestDevice()
//	device.Handler = &DeviceHandler{
//		Read: func(d *Device) ([]*Reading, error) {
//			return []*Reading{NewReading("test", "ok")}, nil
//		},
//	}
//
//	// Clear the global device map then add the device to it
//	deviceMap = make(map[string]*Device)
//	deviceMap["test-id-1"] = device
//
//	// Create the dataManager
//	d, err := newDataManager(&p)
//	assert.NoError(t, err)
//
//	assert.Equal(t, 0, len(d.readChannel))
//	d.serialRead()
//	assert.Equal(t, 1, len(d.readChannel))
//
//	reading := <-d.readChannel
//	assert.Equal(t, 1, len(reading.Reading))
//	assert.Equal(t, "test", reading.Reading[0].Type)
//	assert.Equal(t, "ok", reading.Reading[0].Value)
//}

// FIXME - race condition, likely because the deviceMap is global and used by other tests..
//// TestDataManager_serialReadMultiple tests reading multiple devices serially.
//func TestDataManager_serialReadMultiple(t *testing.T) {
//	// Create the plugin
//	p := Plugin{
//		Config:   &config.PluginConfig{
//			Version: "test",
//			Network: config.NetworkSettings{
//				Type:    "tcp",
//				Address: "test",
//			},
//			Settings: config.Settings{
//				Read: config.ReadSettings{Buffer: 200},
//				Write: config.WriteSettings{Buffer:  200},
//				Transaction: config.TransactionSettings{TTL: "2s"},
//			},
//		},
//		handlers: &Handlers{DeviceID, nil},
//	}
//
//	// Create the device to read
//	device1 := makeTestDevice()
//	device1.Type = "abc"
//	device1.Handler = &DeviceHandler{
//		Read: func(d *Device) ([]*Reading, error) {
//			return []*Reading{NewReading("test", "ok")}, nil
//		},
//	}
//	fmt.Printf("device1: %v\n", device1.ID())
//
//	device2 := makeTestDevice()
//	device2.Type = "def"
//	device2.Handler = &DeviceHandler{
//		Read: func(d *Device) ([]*Reading, error) {
//			return []*Reading{NewReading("test", "ok")}, nil
//		},
//	}
//	fmt.Printf("device2: %v\n", device2.ID())
//
//	device3 := makeTestDevice()
//	device3.Type = "ghi"
//	device3.Handler = &DeviceHandler{
//		Read: func(d *Device) ([]*Reading, error) {
//			return []*Reading{NewReading("test", "ok")}, nil
//		},
//	}
//	fmt.Printf("device3: %v\n", device3.ID())
//
//	// Clear the global device map then add devices to it
//	deviceMap = make(map[string]*Device)
//	deviceMap["test-id-1"] = device1
//	deviceMap["test-id-2"] = device2
//	deviceMap["test-id-3"] = device3
//
//	// Create the dataManager
//	d, err := newDataManager(&p)
//	assert.NoError(t, err)
//
//	assert.Equal(t, 0, len(d.readChannel))
//	d.serialRead()
//	assert.Equal(t, 3, len(d.readChannel))
//
//	reading := <-d.readChannel
//	assert.Equal(t, 3, len(reading.Reading))
//	for _, r := range reading.Reading {
//		assert.Equal(t, "test", r.Type)
//		assert.Equal(t, "ok", r.Value)
//	}
//}
//
//// TestDataManager_parallelReadSingle tests reading a single device in parallel.
//func TestDataManager_parallelReadSingle(t *testing.T) {
//	// Create the plugin
//	p := Plugin{
//		Config: &config.PluginConfig{
//			Version: "test",
//			Network: config.NetworkSettings{
//				Type:    "tcp",
//				Address: "test",
//			},
//			Settings: config.Settings{
//				Read:        config.ReadSettings{Buffer: 200},
//				Write:       config.WriteSettings{Buffer: 200},
//				Transaction: config.TransactionSettings{TTL: "2s"},
//			},
//		},
//		handlers: &Handlers{DeviceID, nil},
//	}
//
//	// Create the device to read
//	device := makeTestDevice()
//	device.Handler = &DeviceHandler{
//		Read: func(d *Device) ([]*Reading, error) {
//			return []*Reading{NewReading("test", "ok")}, nil
//		},
//	}
//
//	// Clear the global device map then add the device to it
//	deviceMap = make(map[string]*Device)
//	deviceMap["test-id-1"] = device
//
//	// Create the dataManager
//	d, err := newDataManager(&p)
//	assert.NoError(t, err)
//
//	assert.Equal(t, 0, len(d.readChannel))
//	d.parallelRead()
//	assert.Equal(t, 1, len(d.readChannel))
//
//	reading := <-d.readChannel
//	assert.Equal(t, 1, len(reading.Reading))
//	assert.Equal(t, "test", reading.Reading[0].Type)
//	assert.Equal(t, "ok", reading.Reading[0].Value)
//}

// FIXME - race condition, likely because the deviceMap is global and used by other tests..
//// TestDataManager_parallelReadSingle tests reading multiple devices in parallel.
//func TestDataManager_parallelReadMultiple(t *testing.T) {
//	// Create the plugin
//	p := Plugin{
//		Config:   &config.PluginConfig{
//			Version: "test",
//			Network: config.NetworkSettings{
//				Type:    "tcp",
//				Address: "test",
//			},
//			Settings: config.Settings{
//				Read: config.ReadSettings{Buffer: 200},
//				Write: config.WriteSettings{Buffer:  200},
//				Transaction: config.TransactionSettings{TTL: "2s"},
//			},
//		},
//		handlers: &Handlers{DeviceID, nil},
//	}
//
//	// Create the device to read
//	device := makeTestDevice()
//	device.Handler = &DeviceHandler{
//		Read: func(d *Device) ([]*Reading, error) {
//			return []*Reading{NewReading("test", "ok")}, nil
//		},
//	}
//
//	// Clear the global device map then add devices to it
//	deviceMap = make(map[string]*Device)
//	deviceMap["test-id-1"] = device
//	deviceMap["test-id-2"] = device
//	deviceMap["test-id-3"] = device
//
//	// Create the dataManager
//	d, err := newDataManager(&p)
//	assert.NoError(t, err)
//
//	assert.Equal(t, 0, len(d.readChannel))
//	d.parallelRead()
//	assert.Equal(t, 3, len(d.readChannel))
//
//	reading := <-d.readChannel
//	assert.Equal(t, 3, len(reading.Reading))
//	for _, r := range reading.Reading {
//		assert.Equal(t, "test", r.Type)
//		assert.Equal(t, "ok", r.Value)
//	}
//}

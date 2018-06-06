package sdk

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/sdk/config"
)

// DeviceID gets the unique identifiers out of the plugin-specific
// configuration to be used in UID generation. Required to construct Handlers.
func DeviceID(data map[string]string) string {
	return data["id"]
}

// TestNewDataManager tests creating a new dataManager instance successfully.
func TestNewDataManager(t *testing.T) {
	d := newDataManager()
	assert.Nil(t, d.readChannel)
	assert.Nil(t, d.writeChannel)
	assert.Nil(t, d.limiter)
	assert.NotNil(t, d.dataLock)
	assert.NotNil(t, d.rwLock)
	assert.Empty(t, d.readings)
}

// TestDataManager_WritesEnabled tests that writes are enabled in the dataManager
// when they are enabled in the config.
func TestDataManager_WritesEnabled(t *testing.T) {
	PluginConfig = &config.PluginConfig{
		ConfigVersion: config.ConfigVersion{Version: "test"},
		Network: &config.NetworkSettings{
			Type:    "tcp",
			Address: "test",
		},
		Settings: &config.PluginSettings{
			Read: &config.ReadSettings{
				Buffer: 200,
			},
			Write: &config.WriteSettings{
				Enabled: true,
				Buffer:  200,
			},
			Transaction: &config.TransactionSettings{
				TTL: "2s",
			},
		},
	}
	defer func() {
		// reset the plugin config
		PluginConfig = &config.PluginConfig{}
	}()

	assert.True(t, DataManager.writesEnabled())
}

// TestDataManager_readNoLimiter tests reading a device when a limiter is
// not configured.
func TestDataManager_readOneOkNoLimiter(t *testing.T) {
	PluginConfig = &config.PluginConfig{
		ConfigVersion: config.ConfigVersion{Version: "test"},
		Network: &config.NetworkSettings{
			Type:    "tcp",
			Address: "test",
		},
		Settings: &config.PluginSettings{
			Read:        &config.ReadSettings{Buffer: 200},
			Write:       &config.WriteSettings{Buffer: 200},
			Transaction: &config.TransactionSettings{TTL: "2s"},
		},
	}
	defer func() {
		// reset the plugin config
		PluginConfig = &config.PluginConfig{}
	}()

	// Create the device to read
	device := &Device{
		Kind:     "test.state",
		Location: &Location{Rack: "rack", Board: "board"},
		Outputs: []*Output{
			{
				OutputType: config.OutputType{
					Name: "foo",
				},
			},
		},
		Handler: &DeviceHandler{
			Read: func(d *Device) ([]*Reading, error) {
				return []*Reading{d.GetOutput("foo").MakeReading("ok")}, nil
			},
		},
	}

	d := newDataManager()
	d.setup()

	// Pass a reading in
	assert.Equal(t, 0, len(d.readChannel))
	d.readOne(device)
	assert.Equal(t, 1, len(d.readChannel))

	// Get the reading out
	reading := <-d.readChannel
	assert.Equal(t, 1, len(reading.Reading))
	assert.Equal(t, "foo", reading.Reading[0].Type)
	assert.Equal(t, "ok", reading.Reading[0].Value)
}

// TestDataManager_readOkWithLimiter tests reading a device when a limiter is
// configured.
func TestDataManager_readOneOkWithLimiter(t *testing.T) {
	PluginConfig = &config.PluginConfig{
		ConfigVersion: config.ConfigVersion{Version: "test"},
		Network: &config.NetworkSettings{
			Type:    "tcp",
			Address: "test",
		},
		Settings: &config.PluginSettings{
			Read:        &config.ReadSettings{Buffer: 200},
			Write:       &config.WriteSettings{Buffer: 200},
			Transaction: &config.TransactionSettings{TTL: "2s"},
		},
		Limiter: &config.LimiterSettings{Rate: 200, Burst: 200},
	}
	defer func() {
		// reset the plugin config
		PluginConfig = &config.PluginConfig{}
	}()

	// Create the device to read
	device := &Device{
		Kind:     "test.state",
		Location: &Location{Rack: "rack", Board: "board"},
		Outputs: []*Output{
			{
				OutputType: config.OutputType{
					Name: "foo",
				},
			},
		},
		Handler: &DeviceHandler{
			Read: func(d *Device) ([]*Reading, error) {
				return []*Reading{d.GetOutput("foo").MakeReading("ok")}, nil
			},
		},
	}

	d := newDataManager()
	d.setup()

	// Pass a reading in
	assert.Equal(t, 0, len(d.readChannel))
	d.readOne(device)
	assert.Equal(t, 1, len(d.readChannel))

	// Get the reading out
	reading := <-d.readChannel
	assert.Equal(t, 1, len(reading.Reading))
	assert.Equal(t, "foo", reading.Reading[0].Type)
	assert.Equal(t, "ok", reading.Reading[0].Value)
}

// TestDataManager_readErr tests reading a device that results in error.
func TestDataManager_readOneErr(t *testing.T) {
	PluginConfig = &config.PluginConfig{
		ConfigVersion: config.ConfigVersion{Version: "test"},
		Network: &config.NetworkSettings{
			Type:    "tcp",
			Address: "test",
		},
		Settings: &config.PluginSettings{
			Read:        &config.ReadSettings{Buffer: 200},
			Write:       &config.WriteSettings{Buffer: 200},
			Transaction: &config.TransactionSettings{TTL: "2s"},
		},
	}
	defer func() {
		// reset the plugin config
		PluginConfig = &config.PluginConfig{}
	}()

	// Create the device to read
	device := &Device{
		Kind:     "test.state",
		Location: &Location{Rack: "rack", Board: "board"},
		Outputs: []*Output{
			{
				OutputType: config.OutputType{
					Name: "foo",
				},
			},
		},
		Handler: &DeviceHandler{
			Read: func(d *Device) ([]*Reading, error) {
				return nil, fmt.Errorf("test read error")
			},
		},
	}

	d := newDataManager()
	d.setup()

	assert.Equal(t, 0, len(d.readChannel))
	d.readOne(device)
	assert.Equal(t, 0, len(d.readChannel))
}

// TestDataManager_serialReadSingle tests reading a single device serially.
func TestDataManager_serialReadSingle(t *testing.T) {
	PluginConfig = &config.PluginConfig{
		ConfigVersion: config.ConfigVersion{Version: "test"},
		Network: &config.NetworkSettings{
			Type:    "tcp",
			Address: "test",
		},
		Settings: &config.PluginSettings{
			Read:        &config.ReadSettings{Buffer: 200},
			Write:       &config.WriteSettings{Buffer: 200},
			Transaction: &config.TransactionSettings{TTL: "2s"},
		},
	}
	defer func() {
		// reset the plugin config
		PluginConfig = &config.PluginConfig{}
	}()

	// Create the device to read
	device := &Device{
		Kind:     "test.state",
		Location: &Location{Rack: "rack", Board: "board"},
		Outputs: []*Output{
			{
				OutputType: config.OutputType{
					Name: "foo",
				},
			},
		},
		Handler: &DeviceHandler{
			Read: func(d *Device) ([]*Reading, error) {
				return []*Reading{d.GetOutput("foo").MakeReading("ok")}, nil
			},
		},
	}

	// Clear the global device map then add the device to it
	deviceMap = make(map[string]*Device)
	deviceMap["test-id-1"] = device

	d := newDataManager()
	d.setup()

	// Pass a reading in
	assert.Equal(t, 0, len(d.readChannel))
	d.serialRead()
	assert.Equal(t, 1, len(d.readChannel))

	// Get the reading out
	reading := <-d.readChannel
	assert.Equal(t, 1, len(reading.Reading))
	assert.Equal(t, "foo", reading.Reading[0].Type)
	assert.Equal(t, "ok", reading.Reading[0].Value)
}

// TestDataManager_parallelReadSingle tests reading a single device in parallel.
func TestDataManager_parallelReadSingle(t *testing.T) {
	PluginConfig = &config.PluginConfig{
		ConfigVersion: config.ConfigVersion{Version: "test"},
		Network: &config.NetworkSettings{
			Type:    "tcp",
			Address: "test",
		},
		Settings: &config.PluginSettings{
			Read:        &config.ReadSettings{Buffer: 200},
			Write:       &config.WriteSettings{Buffer: 200},
			Transaction: &config.TransactionSettings{TTL: "2s"},
		},
	}
	defer func() {
		// reset the plugin config
		PluginConfig = &config.PluginConfig{}
	}()

	// Create the device to read
	device := &Device{
		Kind:     "test.state",
		Location: &Location{Rack: "rack", Board: "board"},
		Outputs: []*Output{
			{
				OutputType: config.OutputType{
					Name: "foo",
				},
			},
		},
		Handler: &DeviceHandler{
			Read: func(d *Device) ([]*Reading, error) {
				return []*Reading{d.GetOutput("foo").MakeReading("ok")}, nil
			},
		},
	}

	// Clear the global device map then add the device to it
	deviceMap = make(map[string]*Device)
	deviceMap["test-id-1"] = device

	d := newDataManager()
	d.setup()

	// Pass a reading in
	assert.Equal(t, 0, len(d.readChannel))
	d.parallelRead()
	assert.Equal(t, 1, len(d.readChannel))

	// Get the reading out
	reading := <-d.readChannel
	assert.Equal(t, 1, len(reading.Reading))
	assert.Equal(t, "foo", reading.Reading[0].Type)
	assert.Equal(t, "ok", reading.Reading[0].Value)
}

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

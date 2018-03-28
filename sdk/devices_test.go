package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/sdk/config"
)

// testDeviceFields is a test helper function to check the Device
// fields against the specified prototype and device configs.
func testDeviceFields(t *testing.T, dev *Device, proto *config.PrototypeConfig, deviceConfig *config.DeviceConfig) {
	assert.Equal(t, proto.Type, dev.Type)
	assert.Equal(t, proto.Model, dev.Model)
	assert.Equal(t, proto.Manufacturer, dev.Manufacturer)
	assert.Equal(t, proto.Protocol, dev.Protocol)

	assert.Equal(t, deviceConfig.Location, dev.Location)

	assert.Equal(t, len(proto.Output), len(dev.Output))
	for i := 0; i < len(dev.Output); i++ {
		assert.Equal(t, proto.Output[i], dev.Output[i])
	}

	assert.Equal(t, deviceConfig.Data, dev.Data)
	for k, v := range dev.Data {
		assert.Equal(t, deviceConfig.Data[k], v)
	}
}

// TestDeviceFields tests that the Device instance fields match up with
// the expected configuration fields from which they should originate.
func TestDeviceFields(t *testing.T) {
	testDevice := makeTestDevice()

	testDeviceFields(t, testDevice, testDevice.pconfig, testDevice.dconfig)
	assert.Equal(t, "664f6cfa51c9bef163682bd2a766613b", testDevice.ID())
	assert.Equal(t, "TestRack-TestBoard-664f6cfa51c9bef163682bd2a766613b", testDevice.GUID())
}

// TestEncodeDevice tests encoding the SDK Device to the gRPC MetainfoResponse model.
func TestEncodeDevice(t *testing.T) {
	testDevice := makeTestDevice()
	encoded := testDevice.encode()

	assert.Equal(t, testDevice.ID(), encoded.Uid)
	assert.Equal(t, testDevice.Type, encoded.Type)
	assert.Equal(t, testDevice.Model, encoded.Model)
	assert.Equal(t, testDevice.Manufacturer, encoded.Manufacturer)
	assert.Equal(t, testDevice.Protocol, encoded.Protocol)
	assert.Equal(t, "", encoded.Info)
	assert.Equal(t, "", encoded.Comment)
	assert.Equal(t, testDevice.Location.Rack, encoded.Location.Rack)
	assert.Equal(t, testDevice.Location.Board, encoded.Location.Board)
}

// TestNewDevice tests creating a new device and validating its fields.
func TestNewDevice(t *testing.T) {
	// Create Handlers.
	handlers, err := NewHandlers(testDeviceIdentifier, nil)
	assert.NoError(t, err)

	// Initialize Plugin with handlers.
	p := Plugin{
		handlers: handlers,
	}

	protoConfig := makePrototypeConfig()
	deviceConfig := makeDeviceConfig()

	d, err := NewDevice(protoConfig, deviceConfig, &testDeviceHandler, &p)
	assert.NoError(t, err)

	testDeviceFields(t, d, protoConfig, deviceConfig)
	assert.Equal(t, &testDeviceHandler, d.Handler)
	assert.Equal(t, deviceConfig, d.dconfig)
	assert.Equal(t, protoConfig, d.pconfig)
}

// TestNewDevice2 tests creating a new device and getting a error on validation
// of the device (invalid handlers).
func TestNewDevice2(t *testing.T) {
	p := Plugin{
		handlers: &Handlers{},
	}

	protoConfig := makePrototypeConfig()
	deviceConfig := makeDeviceConfig()

	_, err := NewDevice(protoConfig, deviceConfig, &testDeviceHandler, &p)
	assert.Error(t, err)
}

// TestNewDevice3 tests creating a new device and getting an error on validation
// of the device (instance-protocol mismatch).
func TestNewDevice3(t *testing.T) {
	// Create handlers.
	handlers, err := NewHandlers(testDeviceIdentifier, nil)
	assert.NoError(t, err)

	// Initialize plugin with handlers.
	p := Plugin{
		handlers: handlers,
	}

	protoConfig := makePrototypeConfig()
	deviceConfig := makeDeviceConfig()
	deviceConfig.Type = "foo"

	_, err = NewDevice(protoConfig, deviceConfig, &testDeviceHandler, &p)
	assert.Error(t, err)
}

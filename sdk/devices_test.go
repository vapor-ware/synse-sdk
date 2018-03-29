package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"fmt"

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
// of the device (instance-protocol mismatch on type).
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

// TestNewDevice3 tests creating a new device and getting an error on validation
// of the device (instance-protocol mismatch on model).
func TestNewDevice4(t *testing.T) {
	// Create handlers.
	handlers, err := NewHandlers(testDeviceIdentifier, nil)
	assert.NoError(t, err)

	// Initialize plugin with handlers.
	p := Plugin{
		handlers: handlers,
	}

	protoConfig := makePrototypeConfig()
	deviceConfig := makeDeviceConfig()
	deviceConfig.Model = "foo"

	_, err = NewDevice(protoConfig, deviceConfig, &testDeviceHandler, &p)
	assert.Error(t, err)
}

// TestDeviceIsReadable tests whether a device is readable in the case
// when it is readable.
func TestDeviceIsReadable(t *testing.T) {
	device := Device{
		Handler: &DeviceHandler{
			Read: func(device *Device) ([]*Reading, error) {
				return []*Reading{}, nil
			},
		},
	}

	readable := device.IsReadable()
	assert.True(t, readable)
}

// TestDeviceIsNotReadable tests whether a device is readable in the case
// when it is not readable.
func TestDeviceIsNotReadable(t *testing.T) {
	device := Device{
		Handler: &DeviceHandler{},
	}

	readable := device.IsReadable()
	assert.False(t, readable)
}

// TestDeviceIsWritable tests whether a device is writable in the case
// when it is writable.
func TestDeviceIsWritable(t *testing.T) {
	device := Device{
		Handler: &DeviceHandler{
			Write: func(device *Device, data *WriteData) error {
				return nil
			},
		},
	}

	writable := device.IsWritable()
	assert.True(t, writable)
}

// TestDeviceIsNotWritable tests whether a device is writable in the case
// when it is not writable.
func TestDeviceIsNotWritable(t *testing.T) {
	device := Device{
		Handler: &DeviceHandler{},
	}

	writable := device.IsWritable()
	assert.False(t, writable)
}

// TestDeviceReadNotReadable tests reading from a device when it is not
// a readable device.
func TestDeviceReadNotReadable(t *testing.T) {
	device := makeTestDevice()

	ctx, err := device.Read()
	assert.Nil(t, ctx)
	assert.Error(t, err)
	assert.IsType(t, &UnsupportedCommandError{}, err)
}

// TestDeviceReadErr tests reading from a device when the device is readable,
// but reading from it will return an error.
func TestDeviceReadErr(t *testing.T) {
	device := makeTestDevice()
	device.Handler.Read = func(device *Device) ([]*Reading, error) {
		return nil, fmt.Errorf("test error")
	}

	ctx, err := device.Read()
	assert.Nil(t, ctx)
	assert.Error(t, err)
	assert.Equal(t, "test error", err.Error())
}

// TestDeviceReadErr2 tests reading from a device when the device is readable,
// but the device rack specification is invalid.
func TestDeviceReadErr2(t *testing.T) {
	device := makeTestDevice()
	device.Location = config.Location{
		Rack:  map[string]string{"invalid-key": "invalid-value"},
		Board: "TestBoard",
	}
	device.Handler.Read = func(device *Device) ([]*Reading, error) {
		return []*Reading{NewReading("test", "value")}, nil
	}

	ctx, err := device.Read()
	assert.Nil(t, ctx)
	assert.Error(t, err)
}

// TestDeviceReadOk tests reading from a device when the device is readable,
// and the device config is correct, so we get back a good reading.
func TestDeviceReadOk(t *testing.T) {
	device := makeTestDevice()
	device.Handler.Read = func(device *Device) ([]*Reading, error) {
		return []*Reading{NewReading("test", "value")}, nil
	}

	ctx, err := device.Read()

	assert.NotNil(t, ctx)
	assert.NoError(t, err)
	assert.Equal(t, device.Location.Rack, ctx.Rack)
	assert.Equal(t, device.Location.Board, ctx.Board)
	assert.Equal(t, device.ID(), ctx.Device)
	assert.Equal(t, 1, len(ctx.Reading))
	assert.Equal(t, "test", ctx.Reading[0].Type)
	assert.Equal(t, "value", ctx.Reading[0].Value)
}

// TestDeviceWriteNotWritable tests writing to a device when it is not
// a writable device.
func TestDeviceWriteNotWritable(t *testing.T) {
	device := makeTestDevice()

	err := device.Write(&WriteData{Action: "test"})
	assert.Error(t, err)
	assert.IsType(t, &UnsupportedCommandError{}, err)
}

// TestDeviceWriteErr tests writing to a device when it is
// a writable device, but the write returns an error.
func TestDeviceWriteErr(t *testing.T) {
	device := makeTestDevice()
	device.Handler.Write = func(device *Device, data *WriteData) error {
		return fmt.Errorf("test error")
	}

	err := device.Write(&WriteData{Action: "test"})
	assert.Error(t, err)
	assert.Equal(t, "test error", err.Error())
}

// TestDeviceWriteOk tests writing to a device when it is
// a writable device.
func TestDeviceWriteOk(t *testing.T) {
	device := makeTestDevice()
	device.Handler.Write = func(device *Device, data *WriteData) error {
		return nil
	}

	err := device.Write(&WriteData{Action: "test"})
	assert.NoError(t, err)
}

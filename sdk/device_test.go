package sdk

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
)

// TestDevicesInit tests that the devices data structures were properly initialized
func TestDevicesInit(t *testing.T) {
	assert.NotNil(t, deviceMap)
	assert.NotNil(t, deviceHandlers)
}

// TestDeviceHandler_supportsBulkRead tests whether a DeviceHandler supports bulk reads
func TestDeviceHandler_supportsBulkRead(t *testing.T) {
	var testTable = []struct {
		desc         string
		supportsBulk bool
		handler      DeviceHandler
	}{
		{
			desc:         "empty handler, does not support reads",
			supportsBulk: false,
			handler:      DeviceHandler{},
		},
		{
			desc:         "supports only writes",
			supportsBulk: false,
			handler: DeviceHandler{
				Write: func(device *Device, data *WriteData) error {
					return nil
				},
			},
		},
		{
			desc:         "supports individual reads",
			supportsBulk: false,
			handler: DeviceHandler{
				Read: func(device *Device) ([]*Reading, error) {
					return nil, nil
				},
			},
		},
		{
			desc:         "supports individual read and bulk read",
			supportsBulk: false,
			handler: DeviceHandler{
				Read: func(device *Device) ([]*Reading, error) {
					return nil, nil
				},
				BulkRead: func(devices []*Device) ([]*ReadContext, error) {
					return nil, nil
				},
			},
		},
		{
			desc:         "supports bulk reads",
			supportsBulk: true,
			handler: DeviceHandler{
				BulkRead: func(devices []*Device) ([]*ReadContext, error) {
					return nil, nil
				},
			},
		},
	}

	for _, testCase := range testTable {
		actual := testCase.handler.supportsBulkRead()
		assert.Equal(t, testCase.supportsBulk, actual, testCase.desc)
	}
}

// TestDeviceHandler_getDevicesForHandler tests getting devices for the handler,
// when none exist.
func TestDeviceHandler_getDevicesForHandler(t *testing.T) {
	handler := DeviceHandler{Name: "test"}
	devices := handler.getDevicesForHandler()
	assert.Equal(t, 0, len(devices))
}

// TestDeviceHandler_getDevicesForHandler2 tests getting devices for the handler,
// when one exists.
func TestDeviceHandler_getDevicesForHandler2(t *testing.T) {
	handler := DeviceHandler{Name: "test"}
	deviceMap["123"] = &Device{
		Handler: &handler,
	}
	defer delete(deviceMap, "123")

	devices := handler.getDevicesForHandler()
	assert.Equal(t, 1, len(devices))
}

// Test_getHandlerForDevice tests getting the handler for a Device when a handler
// with the given name doesn't exist.
func Test_getHandlerForDevice(t *testing.T) {
	handler, err := getHandlerForDevice("bar")
	assert.Error(t, err)
	assert.Nil(t, handler)
}

// Test_getHandlerForDevice2 tests getting the handler for a Device when a handler
// with the given name exists.
func Test_getHandlerForDevice2(t *testing.T) {
	deviceHandlers = []*DeviceHandler{
		{Name: "foo"},
	}
	defer func() {
		deviceHandlers = []*DeviceHandler{}
	}()

	handler, err := getHandlerForDevice("foo")
	assert.NoError(t, err)
	assert.Equal(t, "foo", handler.Name)
}

// TestDevice_GetType tests getting the type of the device.
func TestDevice_GetType(t *testing.T) {
	var testTable = []struct {
		desc     string
		expected string
		device   Device
	}{
		{
			desc:     "empty kind",
			expected: "",
			device:   Device{},
		},
		{
			desc:     "single name for kind",
			expected: "foo",
			device: Device{
				Kind: "foo",
			},
		},
		{
			desc:     "namespaced name for kind",
			expected: "foo",
			device: Device{
				Kind: "test.foo",
			},
		},
		{
			desc:     "deep namespaced name for kind",
			expected: "foo",
			device: Device{
				Kind: "test.example.sample.something.foo",
			},
		},
	}

	for _, testCase := range testTable {
		actual := testCase.device.GetType()
		assert.Equal(t, testCase.expected, actual, testCase.desc)
	}
}

// TestDevice_GetOutputNil tests getting an output when it doesn't exist.
func TestDevice_GetOutputNil(t *testing.T) {
	device := &Device{
		Outputs: []*Output{},
	}
	output := device.GetOutput("foo")
	assert.Nil(t, output)
}

// TestDevice_GetOutput tests getting an output when it does exist.
func TestDevice_GetOutput(t *testing.T) {
	device := &Device{
		Outputs: []*Output{
			{
				OutputType: config.OutputType{Name: "foo"},
			},
		},
	}
	output := device.GetOutput("foo")
	assert.NotNil(t, output)
	assert.Equal(t, "foo", output.Name)
}

// TestMakeDevices tests making a single device.
func TestMakeDevices(t *testing.T) {
	// Add an output to the output map
	outputTypeMap["something"] = &config.OutputType{
		Name: "something",
	}
	defer delete(outputTypeMap, "something")

	// Add a handler to the handler list
	deviceHandlers = []*DeviceHandler{
		{Name: "test"},
	}
	defer func() {
		deviceHandlers = []*DeviceHandler{}
	}()

	// Create the device config from which Devices will be made
	cfg := &config.DeviceConfig{
		Locations: []*config.Location{
			{
				Name:  "foo",
				Rack:  &config.LocationData{Name: "rack"},
				Board: &config.LocationData{Name: "board"},
			},
		},
		Devices: []*config.DeviceKind{
			{
				Name: "test",
				Outputs: []*config.DeviceOutput{
					{
						Type: "something",
					},
				},
				Instances: []*config.DeviceInstance{
					{
						Info:     "test info",
						Location: "foo",
					},
				},
			},
		},
	}

	devices, err := makeDevices(cfg)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(devices))
}

// TestMakeDevices2 tests making devices when no device kinds are specified
func TestMakeDevices2(t *testing.T) {
	cfg := &config.DeviceConfig{
		Locations: []*config.Location{
			{
				Name:  "foo",
				Rack:  &config.LocationData{Name: "rack"},
				Board: &config.LocationData{Name: "board"},
			},
		},
		Devices: []*config.DeviceKind{},
	}

	devices, err := makeDevices(cfg)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(devices))
}

// TestMakeDevices3 tests making devices when no device instances are specified
func TestMakeDevices3(t *testing.T) {
	cfg := &config.DeviceConfig{
		Locations: []*config.Location{
			{
				Name:  "foo",
				Rack:  &config.LocationData{Name: "rack"},
				Board: &config.LocationData{Name: "board"},
			},
		},
		Devices: []*config.DeviceKind{
			{
				Name: "test",
				Outputs: []*config.DeviceOutput{
					{
						Type: "something",
					},
				},
				Instances: []*config.DeviceInstance{},
			},
		},
	}

	devices, err := makeDevices(cfg)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(devices))
}

// TestMakeDevices4 tests making a single device, when the instance output is invalid.
func TestMakeDevices4(t *testing.T) {
	// Create the device config from which Devices will be made
	cfg := &config.DeviceConfig{
		Locations: []*config.Location{
			{
				Name:  "foo",
				Rack:  &config.LocationData{Name: "rack"},
				Board: &config.LocationData{Name: "board"},
			},
		},
		Devices: []*config.DeviceKind{
			{
				Name: "test",
				Instances: []*config.DeviceInstance{
					{
						Info:     "test info",
						Location: "foo",
						Outputs: []*config.DeviceOutput{
							{
								Type: "something",
							},
						},
					},
				},
			},
		},
	}

	devices, err := makeDevices(cfg)
	assert.Error(t, err)
	assert.Nil(t, devices)
}

// TestMakeDevices5 tests making a single device, when there is no associated handler
// defined.
func TestMakeDevices5(t *testing.T) {
	// Add an output to the output map
	outputTypeMap["something"] = &config.OutputType{
		Name: "something",
	}
	defer delete(outputTypeMap, "something")

	// Create the device config from which Devices will be made
	cfg := &config.DeviceConfig{
		Locations: []*config.Location{
			{
				Name:  "foo",
				Rack:  &config.LocationData{Name: "rack"},
				Board: &config.LocationData{Name: "board"},
			},
		},
		Devices: []*config.DeviceKind{
			{
				Name: "test",
				Outputs: []*config.DeviceOutput{
					{
						Type: "something",
					},
				},
				Instances: []*config.DeviceInstance{
					{
						Info:     "test info",
						Location: "foo",
					},
				},
			},
		},
	}

	devices, err := makeDevices(cfg)
	assert.Error(t, err)
	assert.Nil(t, devices)
}

// TestMakeDevices6 tests making a single device, when the locations do not
// match up.
func TestMakeDevices6(t *testing.T) {
	// Add an output to the output map
	outputTypeMap["something"] = &config.OutputType{
		Name: "something",
	}
	defer delete(outputTypeMap, "something")

	// Create the device config from which Devices will be made
	cfg := &config.DeviceConfig{
		Locations: []*config.Location{
			{
				Name:  "foo",
				Rack:  &config.LocationData{Name: "rack"},
				Board: &config.LocationData{Name: "board"},
			},
		},
		Devices: []*config.DeviceKind{
			{
				Name: "test",
				Outputs: []*config.DeviceOutput{
					{
						Type: "something",
					},
				},
				Instances: []*config.DeviceInstance{
					{
						Info:     "test info",
						Location: "baz",
					},
				},
			},
		},
	}

	devices, err := makeDevices(cfg)
	assert.Error(t, err)
	assert.Nil(t, devices)
}

// TestMakeDevices7 tests making a single device when the kind specifies
// an override handler.
func TestMakeDevices7(t *testing.T) {
	// Add an output to the output map
	outputTypeMap["something"] = &config.OutputType{
		Name: "something",
	}
	defer delete(outputTypeMap, "something")

	// Add a handler to the handler list
	deviceHandlers = []*DeviceHandler{
		{Name: "override"},
	}
	defer func() {
		deviceHandlers = []*DeviceHandler{}
	}()

	// Create the device config from which Devices will be made
	cfg := &config.DeviceConfig{
		Locations: []*config.Location{
			{
				Name:  "foo",
				Rack:  &config.LocationData{Name: "rack"},
				Board: &config.LocationData{Name: "board"},
			},
		},
		Devices: []*config.DeviceKind{
			{
				Name:        "test",
				HandlerName: "override",
				Outputs: []*config.DeviceOutput{
					{
						Type: "something",
					},
				},
				Instances: []*config.DeviceInstance{
					{
						Info:     "test info",
						Location: "foo",
					},
				},
			},
		},
	}

	devices, err := makeDevices(cfg)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(devices))
}

// TestMakeDevices8 tests making a single device when the instances specifies
// an override handler.
func TestMakeDevices8(t *testing.T) {
	// Add an output to the output map
	outputTypeMap["something"] = &config.OutputType{
		Name: "something",
	}
	defer delete(outputTypeMap, "something")

	// Add a handler to the handler list
	deviceHandlers = []*DeviceHandler{
		{Name: "override"},
	}
	defer func() {
		deviceHandlers = []*DeviceHandler{}
	}()

	// Create the device config from which Devices will be made
	cfg := &config.DeviceConfig{
		Locations: []*config.Location{
			{
				Name:  "foo",
				Rack:  &config.LocationData{Name: "rack"},
				Board: &config.LocationData{Name: "board"},
			},
		},
		Devices: []*config.DeviceKind{
			{
				Name: "test",
				Outputs: []*config.DeviceOutput{
					{
						Type: "something",
					},
				},
				Instances: []*config.DeviceInstance{
					{
						Info:        "test info",
						Location:    "foo",
						HandlerName: "override",
					},
				},
			},
		},
	}

	devices, err := makeDevices(cfg)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(devices))
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
	device := Device{
		Handler: &DeviceHandler{},
	}

	ctx, err := device.Read()
	assert.Nil(t, ctx)
	assert.Error(t, err)
	assert.IsType(t, &errors.UnsupportedCommandError{}, err)
}

// TestDeviceReadErr tests reading from a device when the device is readable,
// but reading from it will return an error.
func TestDeviceReadErr(t *testing.T) {
	device := Device{
		Handler: &DeviceHandler{
			Read: func(device *Device) ([]*Reading, error) {
				return nil, fmt.Errorf("test error")
			},
		},
	}

	ctx, err := device.Read()
	assert.Nil(t, ctx)
	assert.Error(t, err)
	assert.Equal(t, "test error", err.Error())
}

// TestDeviceReadOk tests reading from a device when the device is readable,
// and the device config is correct, so we get back a good reading.
func TestDeviceReadOk(t *testing.T) {
	device := Device{
		Outputs: []*Output{
			{
				OutputType: config.OutputType{Name: "foo"},
			},
		},
		Location: &Location{
			Rack:  "rack",
			Board: "board",
		},
		Handler: &DeviceHandler{
			Read: func(device *Device) ([]*Reading, error) {
				return []*Reading{
					device.GetOutput("foo").MakeReading("value"),
				}, nil
			},
		},
	}

	ctx, err := device.Read()

	assert.NotNil(t, ctx)
	assert.NoError(t, err)
	assert.Equal(t, device.Location.Rack, ctx.Rack)
	assert.Equal(t, device.Location.Board, ctx.Board)
	assert.Equal(t, device.ID(), ctx.Device)
	assert.Equal(t, 1, len(ctx.Reading))
	assert.Equal(t, "foo", ctx.Reading[0].Type)
	assert.Equal(t, "value", ctx.Reading[0].Value)
}

// TestDeviceWriteNotWritable tests writing to a device when it is not
// a writable device.
func TestDeviceWriteNotWritable(t *testing.T) {
	device := Device{
		Handler: &DeviceHandler{},
	}

	err := device.Write(&WriteData{Action: "test"})
	assert.Error(t, err)
	assert.IsType(t, &errors.UnsupportedCommandError{}, err)
}

// TestDeviceWriteErr tests writing to a device when it is
// a writable device, but the write returns an error.
func TestDeviceWriteErr(t *testing.T) {
	device := Device{
		Handler: &DeviceHandler{
			Write: func(device *Device, data *WriteData) error {
				return fmt.Errorf("test error")
			},
		},
	}

	err := device.Write(&WriteData{Action: "test"})
	assert.Error(t, err)
	assert.Equal(t, "test error", err.Error())
}

// TestDeviceWriteOk tests writing to a device when it is
// a writable device.
func TestDeviceWriteOk(t *testing.T) {
	device := Device{
		Handler: &DeviceHandler{
			Write: func(device *Device, data *WriteData) error {
				return nil
			},
		},
	}

	err := device.Write(&WriteData{Action: "test"})
	assert.NoError(t, err)
}

// TestLocation_encode tests encoding a Location to the grpc message
func TestLocation_encode(t *testing.T) {
	location := Location{
		Rack:  "foo",
		Board: "bar",
	}
	out := location.encode()
	assert.Equal(t, "foo", out.GetRack())
	assert.Equal(t, "bar", out.GetBoard())
}

// TestNewLocationFromConfig tests successfully getting a Location from config
func TestNewLocationFromConfig(t *testing.T) {
	cfg := &config.Location{
		Name:  "test",
		Rack:  &config.LocationData{Name: "rack"},
		Board: &config.LocationData{Name: "board"},
	}

	loc, err := NewLocationFromConfig(cfg)
	assert.NoError(t, err)
	assert.Equal(t, "rack", loc.Rack)
	assert.Equal(t, "board", loc.Board)
}

// TestNewLocationFromConfig2 tests getting a Location from config with bad rack
func TestNewLocationFromConfig2(t *testing.T) {
	cfg := &config.Location{
		Name:  "test",
		Rack:  &config.LocationData{FromEnv: "NOT_AN_ENV"},
		Board: &config.LocationData{Name: "board"},
	}

	loc, err := NewLocationFromConfig(cfg)
	assert.Error(t, err)
	assert.Nil(t, loc)
}

// TestNewLocationFromConfig3 tests getting a Location from config with bad board
func TestNewLocationFromConfig3(t *testing.T) {
	cfg := &config.Location{
		Name:  "test",
		Rack:  &config.LocationData{Name: "rack"},
		Board: &config.LocationData{FromEnv: "NOT_AN_ENV"},
	}

	loc, err := NewLocationFromConfig(cfg)
	assert.Error(t, err)
	assert.Nil(t, loc)
}

// TestOutput_encode tests successfully converting an output to its grpc message.
func TestOutput_encode(t *testing.T) {
	output := Output{
		OutputType: config.OutputType{
			Name:      "foo",
			Precision: 2,
		},
	}
	out := output.encode()
	assert.Equal(t, "foo", out.Name)
	assert.Equal(t, int32(2), out.Precision)
}

// TestOutput_encode2 tests successfully converting an output to its grpc message
// with bad scaling factor.
func TestOutput_encode2(t *testing.T) {
	output := Output{
		OutputType: config.OutputType{
			Name:          "foo",
			Precision:     2,
			ScalingFactor: "abc",
		},
	}
	out := output.encode()
	assert.Equal(t, "foo", out.Name)
	assert.Equal(t, int32(2), out.Precision)
	assert.Equal(t, float64(0), out.ScalingFactor)
}

// TestDevice_encode tests encoding a Device to its grpc message.
func TestDevice_encode(t *testing.T) {
	device := &Device{
		Kind:   "test",
		Plugin: "foo",
		Info:   "bar",
		Location: &Location{
			Rack:  "rack",
			Board: "board",
		},
		Outputs: []*Output{
			{
				OutputType: config.OutputType{
					Name: "output",
					Unit: config.Unit{
						Name:   "x",
						Symbol: "X",
					},
				},
				Info: "out info",
			},
		},
	}
	out := device.encode()
	assert.Equal(t, "test", out.GetKind())
	assert.Equal(t, "foo", out.GetPlugin())
	assert.Equal(t, "bar", out.GetInfo())
	assert.Equal(t, "rack", out.GetLocation().GetRack())
	assert.Equal(t, "board", out.GetLocation().GetBoard())
	assert.Equal(t, "output", out.GetOutput()[0].GetName())
	assert.Equal(t, "x", out.GetOutput()[0].GetUnit().GetName())
	assert.Equal(t, "X", out.GetOutput()[0].GetUnit().GetSymbol())
}

// Test_updateDeviceMap tests updating the device map.
func Test_updateDeviceMap(t *testing.T) {
	device := &Device{
		Kind: "test",
		Location: &Location{
			Rack:  "rack",
			Board: "board",
		},
	}

	assert.Equal(t, 0, len(deviceMap))
	updateDeviceMap([]*Device{device})
	assert.Equal(t, 1, len(deviceMap))

	delete(deviceMap, device.GUID())
}

// Test_getInstanceOutputs tests getting instance output when none are defined.
func Test_getInstanceOutputs(t *testing.T) {
	kind := &config.DeviceKind{}
	instance := &config.DeviceInstance{}

	outputs, err := getInstanceOutputs(kind, instance)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(outputs))
}

// Test_getInstanceOutputs2 tests getting instance output when the Kind defines them.
func Test_getInstanceOutputs2(t *testing.T) {
	outputTypeMap["test"] = &config.OutputType{Name: "test"}
	defer delete(outputTypeMap, "test")

	kind := &config.DeviceKind{
		Outputs: []*config.DeviceOutput{
			{
				Type: "test",
			},
		},
	}
	instance := &config.DeviceInstance{}

	outputs, err := getInstanceOutputs(kind, instance)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(outputs))
}

// Test_getInstanceOutputs3 tests getting instance output when the Instance defines them.
func Test_getInstanceOutputs3(t *testing.T) {
	outputTypeMap["test"] = &config.OutputType{Name: "test"}
	defer delete(outputTypeMap, "test")

	kind := &config.DeviceKind{}
	instance := &config.DeviceInstance{
		Outputs: []*config.DeviceOutput{
			{
				Type: "test",
			},
		},
	}

	outputs, err := getInstanceOutputs(kind, instance)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(outputs))
}

// Test_getInstanceOutputs4 tests getting instance output when the Kind and Instance defines them.
func Test_getInstanceOutputs4(t *testing.T) {
	outputTypeMap["foo"] = &config.OutputType{Name: "foo"}
	outputTypeMap["bar"] = &config.OutputType{Name: "bar"}
	defer delete(outputTypeMap, "foo")
	defer delete(outputTypeMap, "bar")

	kind := &config.DeviceKind{
		Outputs: []*config.DeviceOutput{
			{
				Type: "foo",
			},
		},
	}
	instance := &config.DeviceInstance{
		Outputs: []*config.DeviceOutput{
			{
				Type: "bar",
			},
		},
	}

	outputs, err := getInstanceOutputs(kind, instance)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(outputs))
}

// Test_getInstanceOutputs5 tests getting instance output when duplicates are defined.
func Test_getInstanceOutputs5(t *testing.T) {
	outputTypeMap["test"] = &config.OutputType{Name: "test"}
	defer delete(outputTypeMap, "test")

	kind := &config.DeviceKind{
		Outputs: []*config.DeviceOutput{
			{
				Type: "test",
			},
		},
	}
	instance := &config.DeviceInstance{
		Outputs: []*config.DeviceOutput{
			{
				Type: "test",
			},
		},
	}

	outputs, err := getInstanceOutputs(kind, instance)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(outputs))
}

// Test_getInstanceOutputs6 tests getting instance output when the Kind and Instance defines them,
// but the instance should not inherit from the kind..
func Test_getInstanceOutputs6(t *testing.T) {
	outputTypeMap["foo"] = &config.OutputType{Name: "foo"}
	outputTypeMap["bar"] = &config.OutputType{Name: "bar"}
	defer delete(outputTypeMap, "foo")
	defer delete(outputTypeMap, "bar")

	kind := &config.DeviceKind{
		Outputs: []*config.DeviceOutput{
			{
				Type: "foo",
			},
		},
	}
	instance := &config.DeviceInstance{
		DisableOutputInheritance: true,
		Outputs: []*config.DeviceOutput{
			{
				Type: "bar",
			},
		},
	}

	outputs, err := getInstanceOutputs(kind, instance)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(outputs))
}

// Test_getInstanceOutputs7 tests getting instance output when the Kind defines them,
// but there is no corresponding type.
func Test_getInstanceOutputs7(t *testing.T) {
	kind := &config.DeviceKind{
		Outputs: []*config.DeviceOutput{
			{
				Type: "test",
			},
		},
	}
	instance := &config.DeviceInstance{}

	outputs, err := getInstanceOutputs(kind, instance)
	assert.Error(t, err)
	assert.Nil(t, outputs)
}

// Test_getInstanceOutputs8 tests getting instance output when the Instance defines them,
// but there is no corresponding type.
func Test_getInstanceOutputs8(t *testing.T) {
	kind := &config.DeviceKind{}
	instance := &config.DeviceInstance{
		Outputs: []*config.DeviceOutput{
			{
				Type: "test",
			},
		},
	}

	outputs, err := getInstanceOutputs(kind, instance)
	assert.Error(t, err)
	assert.Nil(t, outputs)
}

// ----------
// Examples
// ----------

// A device with a Read function defined in its DeviceHandler should
// be readable.
func ExampleDevice_IsReadable_true() {
	device := Device{
		Handler: &DeviceHandler{
			Read: func(device *Device) ([]*Reading, error) {
				return []*Reading{}, nil
			},
		},
	}

	readable := device.IsReadable()
	fmt.Println(readable)
	// Output: true
}

// A device without a Read function defined in its DeviceHandler should
// not be readable.
func ExampleDevice_IsReadable_false() {
	device := Device{
		Handler: &DeviceHandler{},
	}

	readable := device.IsReadable()
	fmt.Println(readable)
	// Output: false
}

// A device with a Write function defined in its DeviceHandler should
// be writable.
func ExampleDevice_IsWritable_true() {
	device := Device{
		Handler: &DeviceHandler{
			Write: func(device *Device, data *WriteData) error {
				return nil
			},
		},
	}

	writable := device.IsWritable()
	fmt.Println(writable)
	// Output: true
}

// A device without a Write function defined in its DeviceHandler should
// not be writable.
func ExampleDevice_IsWritable_false() {
	device := Device{
		Handler: &DeviceHandler{},
	}

	writable := device.IsWritable()
	fmt.Println(writable)
	// Output: false
}

// Get the GUID of the device.
func ExampleDevice_GUID() {
	device := Device{
		id: "baz",
		Location: &Location{
			Rack:  "foo",
			Board: "bar",
		},
	}

	guid := device.GUID()
	fmt.Println(guid)
	// Output: foo-bar-baz
}

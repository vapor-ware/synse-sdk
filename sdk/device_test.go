package sdk

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
)

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
	defer resetContext()

	handler := DeviceHandler{Name: "test"}
	ctx.devices["123"] = &Device{
		Handler: &handler,
	}

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
	defer resetContext()

	ctx.deviceHandlers = []*DeviceHandler{
		{Name: "foo"},
	}

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
				OutputType: OutputType{Name: "foo"},
			},
		},
	}
	output := device.GetOutput("foo")
	assert.NotNil(t, output)
	assert.Equal(t, "foo", output.Name)
}

// TestDevice_JSON_1 tests dumping an empty Device to a JSON string.
func TestDevice_JSON_1(t *testing.T) {
	d := Device{}
	out, err := d.JSON()
	assert.NoError(t, err)
	assert.Equal(
		t,
		`{"Kind":"","Metadata":null,"Plugin":"","Info":"","Location":null,"Data":null,"Outputs":null,"SortOrdinal":0}`,
		out,
	)
}

// TestDevice_JSON_2 tests dumping a Device to a JSON string.
func TestDevice_JSON_2(t *testing.T) {
	d := Device{
		Kind:        "foo",
		Metadata:    map[string]string{"test": "data"},
		Info:        "info",
		Handler:     &DeviceHandler{},
		SortOrdinal: 1,
		Location: &Location{
			Rack:  "rack",
			Board: "board",
		},
	}

	out, err := d.JSON()
	assert.NoError(t, err)
	assert.Equal(
		t,
		`{"Kind":"foo","Metadata":{"test":"data"},"Plugin":"","Info":"info","Location":{"Rack":"rack","Board":"board"},"Data":null,"Outputs":null,"SortOrdinal":1}`,
		out,
	)
}

// TestMakeDevices tests making a single device.
func TestMakeDevices(t *testing.T) {
	defer resetContext()

	// Add an output to the output map
	ctx.outputTypes["something"] = &OutputType{
		Name: "something",
	}

	// Add a handler to the handler list
	ctx.deviceHandlers = []*DeviceHandler{
		{Name: "test"},
	}

	// Create the device config from which Devices will be made
	cfg := &DeviceConfig{
		Locations: []*LocationConfig{
			{
				Name:  "foo",
				Rack:  &LocationData{Name: "rack"},
				Board: &LocationData{Name: "board"},
			},
		},
		Devices: []*DeviceKind{
			{
				Name: "test",
				Outputs: []*DeviceOutput{
					{
						Type: "something",
					},
				},
				Instances: []*DeviceInstance{
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
	cfg := &DeviceConfig{
		Locations: []*LocationConfig{
			{
				Name:  "foo",
				Rack:  &LocationData{Name: "rack"},
				Board: &LocationData{Name: "board"},
			},
		},
		Devices: []*DeviceKind{},
	}

	devices, err := makeDevices(cfg)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(devices))
}

// TestMakeDevices3 tests making devices when no device instances are specified
func TestMakeDevices3(t *testing.T) {
	cfg := &DeviceConfig{
		Locations: []*LocationConfig{
			{
				Name:  "foo",
				Rack:  &LocationData{Name: "rack"},
				Board: &LocationData{Name: "board"},
			},
		},
		Devices: []*DeviceKind{
			{
				Name: "test",
				Outputs: []*DeviceOutput{
					{
						Type: "something",
					},
				},
				Instances: []*DeviceInstance{},
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
	cfg := &DeviceConfig{
		Locations: []*LocationConfig{
			{
				Name:  "foo",
				Rack:  &LocationData{Name: "rack"},
				Board: &LocationData{Name: "board"},
			},
		},
		Devices: []*DeviceKind{
			{
				Name: "test",
				Instances: []*DeviceInstance{
					{
						Info:     "test info",
						Location: "foo",
						Outputs: []*DeviceOutput{
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
	defer resetContext()
	// Add an output to the output map
	ctx.outputTypes["something"] = &OutputType{
		Name: "something",
	}

	// Create the device config from which Devices will be made
	cfg := &DeviceConfig{
		Locations: []*LocationConfig{
			{
				Name:  "foo",
				Rack:  &LocationData{Name: "rack"},
				Board: &LocationData{Name: "board"},
			},
		},
		Devices: []*DeviceKind{
			{
				Name: "test",
				Outputs: []*DeviceOutput{
					{
						Type: "something",
					},
				},
				Instances: []*DeviceInstance{
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
	defer resetContext()

	// Add an output to the output map
	ctx.outputTypes["something"] = &OutputType{
		Name: "something",
	}

	// Create the device config from which Devices will be made
	cfg := &DeviceConfig{
		Locations: []*LocationConfig{
			{
				Name:  "foo",
				Rack:  &LocationData{Name: "rack"},
				Board: &LocationData{Name: "board"},
			},
		},
		Devices: []*DeviceKind{
			{
				Name: "test",
				Outputs: []*DeviceOutput{
					{
						Type: "something",
					},
				},
				Instances: []*DeviceInstance{
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
	defer resetContext()

	// Add an output to the output map
	ctx.outputTypes["something"] = &OutputType{
		Name: "something",
	}

	// Add a handler to the handler list
	ctx.deviceHandlers = []*DeviceHandler{
		{Name: "override"},
	}

	// Create the device config from which Devices will be made
	cfg := &DeviceConfig{
		Locations: []*LocationConfig{
			{
				Name:  "foo",
				Rack:  &LocationData{Name: "rack"},
				Board: &LocationData{Name: "board"},
			},
		},
		Devices: []*DeviceKind{
			{
				Name:        "test",
				HandlerName: "override",
				Outputs: []*DeviceOutput{
					{
						Type: "something",
					},
				},
				Instances: []*DeviceInstance{
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
	defer resetContext()

	// Add an output to the output map
	ctx.outputTypes["something"] = &OutputType{
		Name: "something",
	}

	// Add a handler to the handler list
	ctx.deviceHandlers = []*DeviceHandler{
		{Name: "override"},
	}

	// Create the device config from which Devices will be made
	cfg := &DeviceConfig{
		Locations: []*LocationConfig{
			{
				Name:  "foo",
				Rack:  &LocationData{Name: "rack"},
				Board: &LocationData{Name: "board"},
			},
		},
		Devices: []*DeviceKind{
			{
				Name: "test",
				Outputs: []*DeviceOutput{
					{
						Type: "something",
					},
				},
				Instances: []*DeviceInstance{
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
				OutputType: OutputType{Name: "foo"},
			},
		},
		Location: &Location{
			Rack:  "rack",
			Board: "board",
		},
		Handler: &DeviceHandler{
			Read: func(device *Device) ([]*Reading, error) {
				reading, err := device.GetOutput("foo").MakeReading("value")
				if err != nil {
					return nil, err
				}
				return []*Reading{reading}, nil
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

// TestLocationConfig_Resolve tests successfully getting a Location from config
func TestLocationConfig_Resolve(t *testing.T) {
	cfg := &LocationConfig{
		Name:  "test",
		Rack:  &LocationData{Name: "rack"},
		Board: &LocationData{Name: "board"},
	}

	loc, err := cfg.Resolve()
	assert.NoError(t, err)
	assert.Equal(t, "rack", loc.Rack)
	assert.Equal(t, "board", loc.Board)
}

// TestLocationConfig_Resolve2 tests getting a Location from config with bad rack
func TestLocationConfig_Resolve2(t *testing.T) {
	cfg := &LocationConfig{
		Name:  "test",
		Rack:  &LocationData{FromEnv: "NOT_AN_ENV"},
		Board: &LocationData{Name: "board"},
	}

	loc, err := cfg.Resolve()
	assert.Error(t, err)
	assert.Nil(t, loc)
}

// TestLocationConfig_Resolve3 tests getting a Location from config with bad board
func TestLocationConfig_Resolve3(t *testing.T) {
	cfg := &LocationConfig{
		Name:  "test",
		Rack:  &LocationData{Name: "rack"},
		Board: &LocationData{FromEnv: "NOT_AN_ENV"},
	}

	loc, err := cfg.Resolve()
	assert.Error(t, err)
	assert.Nil(t, loc)
}

// TestOutput_encode tests successfully converting an output to its grpc message.
func TestOutput_encode(t *testing.T) {
	output := Output{
		OutputType: OutputType{
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
		OutputType: OutputType{
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
				OutputType: OutputType{
					Name: "output",
					Unit: Unit{
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
	defer resetContext()

	device := &Device{
		Kind: "test",
		Location: &Location{
			Rack:  "rack",
			Board: "board",
		},
	}

	assert.Equal(t, 0, len(ctx.devices))
	updateDeviceMap([]*Device{device})
	//r := recover()
	//assert.NotNil(t, r)
	assert.Equal(t, 1, len(ctx.devices))
}

// Test_updateDeviceMap2 tests updating the device map when a device
// with that id already exists.
func Test_updateDeviceMap2(t *testing.T) {
	defer resetContext()

	// this will be run after we panic
	defer func() {
		r := recover()
		assert.NotNil(t, r)
		assert.Equal(t, 1, len(ctx.devices))
	}()

	device := &Device{
		Kind: "test",
		Location: &Location{
			Rack:  "rack",
			Board: "board",
		},
	}
	// manually add the device to the device map
	ctx.devices[device.GUID()] = device
	assert.Equal(t, 1, len(ctx.devices))

	// now try updating the map - this will be a duplicate since
	// we already have this device in the map.
	updateDeviceMap([]*Device{device})
}

// Test_getInstanceOutputs tests getting instance output when none are defined.
func Test_getInstanceOutputs(t *testing.T) {
	kind := &DeviceKind{}
	instance := &DeviceInstance{}

	outputs, err := getInstanceOutputs(kind, instance)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(outputs))
}

// Test_getInstanceOutputs2 tests getting instance output when the Kind defines them.
func Test_getInstanceOutputs2(t *testing.T) {
	defer resetContext()

	ctx.outputTypes["test"] = &OutputType{Name: "test"}

	kind := &DeviceKind{
		Outputs: []*DeviceOutput{
			{
				Type: "test",
			},
		},
	}
	instance := &DeviceInstance{}

	outputs, err := getInstanceOutputs(kind, instance)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(outputs))
}

// Test_getInstanceOutputs3 tests getting instance output when the Instance defines them.
func Test_getInstanceOutputs3(t *testing.T) {
	defer resetContext()

	ctx.outputTypes["test"] = &OutputType{Name: "test"}

	kind := &DeviceKind{}
	instance := &DeviceInstance{
		Outputs: []*DeviceOutput{
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
	defer resetContext()

	ctx.outputTypes["foo"] = &OutputType{Name: "foo"}
	ctx.outputTypes["bar"] = &OutputType{Name: "bar"}

	kind := &DeviceKind{
		Outputs: []*DeviceOutput{
			{
				Type: "foo",
			},
		},
	}
	instance := &DeviceInstance{
		Outputs: []*DeviceOutput{
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
	defer resetContext()

	ctx.outputTypes["test"] = &OutputType{Name: "test"}

	kind := &DeviceKind{
		Outputs: []*DeviceOutput{
			{
				Type: "test",
			},
		},
	}
	instance := &DeviceInstance{
		Outputs: []*DeviceOutput{
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
	defer resetContext()

	ctx.outputTypes["foo"] = &OutputType{Name: "foo"}
	ctx.outputTypes["bar"] = &OutputType{Name: "bar"}

	kind := &DeviceKind{
		Outputs: []*DeviceOutput{
			{
				Type: "foo",
			},
		},
	}
	instance := &DeviceInstance{
		DisableOutputInheritance: true,
		Outputs: []*DeviceOutput{
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
	kind := &DeviceKind{
		Outputs: []*DeviceOutput{
			{
				Type: "test",
			},
		},
	}
	instance := &DeviceInstance{}

	outputs, err := getInstanceOutputs(kind, instance)
	assert.Error(t, err)
	assert.Nil(t, outputs)
}

// Test_getInstanceOutputs8 tests getting instance output when the Instance defines them,
// but there is no corresponding type.
func Test_getInstanceOutputs8(t *testing.T) {
	kind := &DeviceKind{}
	instance := &DeviceInstance{
		Outputs: []*DeviceOutput{
			{
				Type: "test",
			},
		},
	}

	outputs, err := getInstanceOutputs(kind, instance)
	assert.Error(t, err)
	assert.Nil(t, outputs)
}

// TestNewDeviceConfig tests initializing a new DeviceConfig.
func TestNewDeviceConfig(t *testing.T) {
	cfg := NewDeviceConfig()

	assert.IsType(t, &DeviceConfig{}, cfg)
	assert.Equal(t, currentDeviceSchemeVersion, cfg.Version)
	assert.Equal(t, 0, len(cfg.Locations))
	assert.Equal(t, 0, len(cfg.Devices))
}

// TestDeviceConfig_JSON_1 tests dumping an empty DeviceConfig to a JSON string.
func TestDeviceConfig_JSON_1(t *testing.T) {
	d := DeviceConfig{}
	out, err := d.JSON()
	assert.NoError(t, err)
	assert.Equal(
		t,
		`{"Version":"","Locations":null,"Devices":null}`,
		out,
	)
}

// TestDeviceConfig_JSON_2 tests dumping a DeviceConfig to a JSON string.
func TestDeviceConfig_JSON_2(t *testing.T) {
	d := DeviceConfig{
		SchemeVersion: SchemeVersion{Version: "1.0"},
		Locations:     []*LocationConfig{{Name: "test", Rack: &LocationData{Name: "test"}, Board: &LocationData{Name: "test"}}},
		Devices:       []*DeviceKind{{Name: "test"}},
	}

	out, err := d.JSON()
	assert.NoError(t, err)
	assert.Equal(
		t,
		`{"Version":"1.0","Locations":[{"Name":"test","Rack":{"Name":"test","FromEnv":""},"Board":{"Name":"test","FromEnv":""}}],"Devices":[{"Name":"test","Metadata":null,"Instances":null,"Outputs":null,"HandlerName":""}]}`,
		out,
	)
}

// TestDeviceConfig_GetLocation_Ok tests getting locations from a DeviceConfig successfully.
func TestDeviceConfig_GetLocation_Ok(t *testing.T) {
	var testTable = []struct {
		desc     string
		location string
		config   DeviceConfig
	}{
		{
			desc:     "DeviceConfig has single location",
			location: "test",
			config: DeviceConfig{
				Locations: []*LocationConfig{
					{Name: "test", Rack: &LocationData{Name: "test"}, Board: &LocationData{Name: "test"}},
				},
			},
		},
		{
			desc:     "DeviceConfig has multiple locations",
			location: "test",
			config: DeviceConfig{
				Locations: []*LocationConfig{
					{Name: "test", Rack: &LocationData{Name: "test"}, Board: &LocationData{Name: "test"}},
					{Name: "foo", Rack: &LocationData{Name: "foo"}, Board: &LocationData{Name: "foo"}},
					{Name: "bar", Rack: &LocationData{Name: "bar"}, Board: &LocationData{Name: "bar"}},
				},
			},
		},
		{
			desc:     "DeviceConfig has multiple locations",
			location: "bar",
			config: DeviceConfig{
				Locations: []*LocationConfig{
					{Name: "test", Rack: &LocationData{Name: "test"}, Board: &LocationData{Name: "test"}},
					{Name: "foo", Rack: &LocationData{Name: "foo"}, Board: &LocationData{Name: "foo"}},
					{Name: "bar", Rack: &LocationData{Name: "bar"}, Board: &LocationData{Name: "bar"}},
				},
			},
		},
	}

	for _, testCase := range testTable {
		l, err := testCase.config.GetLocation(testCase.location)
		assert.NoError(t, err, testCase.desc)
		assert.NotNil(t, l, testCase.desc)
		assert.Equal(t, testCase.location, l.Name, testCase.desc)
	}
}

// TestDeviceConfig_GetLocation_Err tests getting locations from a DeviceConfig unsuccessfully.
func TestDeviceConfig_GetLocation_Err(t *testing.T) {
	var testTable = []struct {
		desc     string
		location string
		config   DeviceConfig
	}{
		{
			desc:     "DeviceConfig has no locations defined",
			location: "test",
			config: DeviceConfig{
				Locations: []*LocationConfig{},
			},
		},
		{
			desc:     "Specified name does not match any location",
			location: "baz",
			config: DeviceConfig{
				Locations: []*LocationConfig{
					{Name: "test", Rack: &LocationData{Name: "test"}, Board: &LocationData{Name: "test"}},
					{Name: "foo", Rack: &LocationData{Name: "foo"}, Board: &LocationData{Name: "foo"}},
					{Name: "bar", Rack: &LocationData{Name: "bar"}, Board: &LocationData{Name: "bar"}},
				},
			},
		},
	}

	for _, testCase := range testTable {
		l, err := testCase.config.GetLocation(testCase.location)
		assert.Error(t, err, testCase.desc)
		assert.Nil(t, l, testCase.desc)
	}
}

// TestDeviceConfig_Validate_Ok tests validating a DeviceConfig with no errors.
func TestDeviceConfig_Validate_Ok(t *testing.T) {
	var testTable = []struct {
		desc   string
		config DeviceConfig
	}{
		{
			desc: "DeviceConfig has valid version",
			config: DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
			},
		},
		{
			desc: "DeviceConfig has valid version and location",
			config: DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations:     []*LocationConfig{{Name: "test", Rack: &LocationData{Name: "test"}, Board: &LocationData{Name: "test"}}},
			},
		},
		{
			desc: "DeviceConfig has valid version, location, and DeviceKind",
			config: DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations:     []*LocationConfig{{Name: "test", Rack: &LocationData{Name: "test"}, Board: &LocationData{Name: "test"}}},
				Devices:       []*DeviceKind{{Name: "test"}},
			},
		},
		{
			desc: "DeviceConfig has valid version, invalid Locations (Locations not validated here)",
			config: DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations:     []*LocationConfig{{Name: ""}},
			},
		},
		{
			desc: "DeviceConfig has valid version and locations, invalid DeviceKinds (DeviceKinds not validated here)",
			config: DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations:     []*LocationConfig{{Name: "test", Rack: &LocationData{Name: "test"}, Board: &LocationData{Name: "test"}}},
				Devices:       []*DeviceKind{{Name: ""}},
			},
		},
	}

	for _, testCase := range testTable {
		merr := errors.NewMultiError("test")

		testCase.config.Validate(merr)
		assert.NoError(t, merr.Err(), testCase.desc)
	}
}

// TestDeviceConfig_Validate_Error tests validating a DeviceConfig with errors.
func TestDeviceConfig_Validate_Error(t *testing.T) {
	var testTable = []struct {
		desc     string
		errCount int
		config   DeviceConfig
	}{
		{
			desc:     "DeviceConfig has invalid version",
			errCount: 1,
			config: DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "abc"},
			},
		},
	}

	for _, testCase := range testTable {
		merr := errors.NewMultiError("test")

		testCase.config.Validate(merr)
		assert.Error(t, merr.Err(), testCase.desc)
		assert.Equal(t, testCase.errCount, len(merr.Errors), merr.Error())
	}
}

// TestLocation_Equals tests whether a location equals another one.
func TestLocation_Equals(t *testing.T) {
	testLoc := LocationConfig{
		Name:  "test",
		Rack:  &LocationData{Name: "rack"},
		Board: &LocationData{Name: "board"},
	}

	var testTable = []struct {
		desc      string
		isEqual   bool
		location1 *LocationConfig
		location2 *LocationConfig
	}{
		{
			desc:      "pointer to the same Location",
			isEqual:   true,
			location1: &testLoc,
			location2: &testLoc,
		},
		{
			desc:      "different location instances, same data (empty)",
			isEqual:   true,
			location1: &LocationConfig{},
			location2: &LocationConfig{},
		},
		{
			desc:      "different location instances, same data",
			isEqual:   true,
			location1: &testLoc,
			location2: &LocationConfig{
				Name:  "test",
				Rack:  &LocationData{Name: "rack"},
				Board: &LocationData{Name: "board"},
			},
		},
		{
			desc:      "different locations, different name",
			isEqual:   false,
			location1: &testLoc,
			location2: &LocationConfig{
				Name:  "foo",
				Rack:  &LocationData{Name: "rack"},
				Board: &LocationData{Name: "board"},
			},
		},
		{
			desc:      "different locations, different rack",
			isEqual:   false,
			location1: &testLoc,
			location2: &LocationConfig{
				Name:  "test",
				Rack:  &LocationData{Name: "foo"},
				Board: &LocationData{Name: "board"},
			},
		},
		{
			desc:      "different locations, different board",
			isEqual:   false,
			location1: &testLoc,
			location2: &LocationConfig{
				Name:  "test",
				Rack:  &LocationData{Name: "rack"},
				Board: &LocationData{Name: "foo"},
			},
		},
	}

	for _, testCase := range testTable {
		equals := testCase.location1.Equals(testCase.location2)
		assert.Equal(t, testCase.isEqual, equals, testCase.desc)
	}
}

// TestLocationData_Equals tests whether two LocationData instances are equal.
func TestLocationData_Equals(t *testing.T) {
	testLoc := LocationData{
		Name: "foo",
	}

	var testTable = []struct {
		desc    string
		isEqual bool
		loc1    *LocationData
		loc2    *LocationData
	}{
		{
			desc:    "pointer to the same LocationData",
			isEqual: true,
			loc1:    &testLoc,
			loc2:    &testLoc,
		},
		{
			desc:    "different LocationData, same data (empty)",
			isEqual: true,
			loc1:    &LocationData{},
			loc2:    &LocationData{},
		},
		{
			desc:    "different LocationData, same data (name)",
			isEqual: true,
			loc1: &LocationData{
				Name: "test",
			},
			loc2: &LocationData{
				Name: "test",
			},
		},
		{
			desc:    "different LocationData, same data (from env)",
			isEqual: true,
			loc1: &LocationData{
				FromEnv: "HOSTNAME",
			},
			loc2: &LocationData{
				FromEnv: "HOSTNAME",
			},
		},
		{
			desc:    "different LocationData, different data (name)",
			isEqual: false,
			loc1: &LocationData{
				Name: "foo",
			},
			loc2: &LocationData{
				Name: "bar",
			},
		},
		{
			desc:    "different LocationData, different data (from env)",
			isEqual: false,
			loc1: &LocationData{
				FromEnv: "NODENAME",
			},
			loc2: &LocationData{
				FromEnv: "HOSTNAME",
			},
		},
		{
			desc:    "different LocationData, different data (mixed)",
			isEqual: false,
			loc1: &LocationData{
				Name: "test",
			},
			loc2: &LocationData{
				FromEnv: "HOSTNAME",
			},
		},
		{
			desc:    "different LocationData, different data (empty)",
			isEqual: false,
			loc1:    &LocationData{},
			loc2: &LocationData{
				FromEnv: "HOSTNAME",
			},
		},
	}

	for _, testCase := range testTable {
		equals := testCase.loc1.Equals(testCase.loc2)
		assert.Equal(t, testCase.isEqual, equals, testCase.desc)
	}
}

// TestLocation_Validate_Ok tests validating a Location with no errors.
func TestLocation_Validate_Ok(t *testing.T) {
	var testTable = []struct {
		desc     string
		location LocationConfig
	}{
		{
			desc: "Valid Location instance",
			location: LocationConfig{
				Name:  "test",
				Rack:  &LocationData{Name: "test"},
				Board: &LocationData{Name: "test"},
			},
		},
	}

	for _, testCase := range testTable {
		merr := errors.NewMultiError("test")

		testCase.location.Validate(merr)
		assert.NoError(t, merr.Err(), testCase.desc)
	}
}

// TestLocation_Validate_Error tests validating a Location with errors.
func TestLocation_Validate_Error(t *testing.T) {
	var testTable = []struct {
		desc     string
		errCount int
		location LocationConfig
	}{
		{
			desc:     "Location requires a name, but has none",
			errCount: 1,
			location: LocationConfig{
				Rack:  &LocationData{Name: "test"},
				Board: &LocationData{Name: "test"},
			},
		},
		{
			desc:     "Location requires a rack, but has none",
			errCount: 1,
			location: LocationConfig{
				Name:  "test",
				Board: &LocationData{Name: "test"},
			},
		},
		{
			desc:     "Location requires a board, but has none",
			errCount: 1,
			location: LocationConfig{
				Name: "test",
				Rack: &LocationData{Name: "test"},
			},
		},
		{
			desc:     "Location both rack and board, has neither",
			errCount: 2,
			location: LocationConfig{
				Name: "test",
			},
		},
		{
			desc:     "Location has an invalid rack",
			errCount: 1,
			location: LocationConfig{
				Name:  "test",
				Rack:  &LocationData{Name: ""},
				Board: &LocationData{Name: "test"},
			},
		},
		{
			desc:     "Location has an invalid board",
			errCount: 1,
			location: LocationConfig{
				Name:  "test",
				Rack:  &LocationData{Name: "test"},
				Board: &LocationData{Name: ""},
			},
		},
		{
			desc:     "Location missing all fields",
			errCount: 3,
			location: LocationConfig{},
		},
	}

	for _, testCase := range testTable {
		merr := errors.NewMultiError("test")

		testCase.location.Validate(merr)
		assert.Error(t, merr.Err(), testCase.desc)
		assert.Equal(t, testCase.errCount, len(merr.Errors), merr.Error())
	}
}

// TestLocationData_Validate_Ok tests validating a LocationData with no errors.
func TestLocationData_Validate_Ok(t *testing.T) {
	var testTable = []struct {
		desc         string
		locationData LocationData
	}{
		{
			desc:         "LocationData has a valid name",
			locationData: LocationData{Name: "test"},
		},
		{
			desc:         "LocationData has a valid fromEnv",
			locationData: LocationData{FromEnv: "TEST_ENV"},
		},
		{
			desc:         "LocationData has both name and fromEnv",
			locationData: LocationData{Name: "foo", FromEnv: "TEST_ENV"},
		},
	}

	err := os.Setenv("TEST_ENV", "test")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err := os.Unsetenv("TEST_ENV")
		if err != nil {
			t.Error(err)
		}
	}()

	for _, testCase := range testTable {
		merr := errors.NewMultiError("test")

		testCase.locationData.Validate(merr)
		assert.NoError(t, merr.Err(), testCase.desc)
	}
}

// TestLocationData_Validate_Error tests validating a LocationData with errors.
func TestLocationData_Validate_Error(t *testing.T) {
	var testTable = []struct {
		desc         string
		errCount     int
		locationData LocationData
	}{
		{
			desc:         "LocationData has no fromEnv or name",
			errCount:     2,
			locationData: LocationData{},
		},
		{
			desc:         "LocationData fromEnv does not resolve",
			errCount:     2,
			locationData: LocationData{FromEnv: "FOO_BAR_BAZ"},
		},
	}

	for _, testCase := range testTable {
		merr := errors.NewMultiError("test")

		testCase.locationData.Validate(merr)
		assert.Error(t, merr.Err(), testCase.desc)
		assert.Equal(t, testCase.errCount, len(merr.Errors), merr.Error())
	}
}

// TestLocationData_Get_Ok tests getting the locational data with no errors.
func TestLocationData_Get_Ok(t *testing.T) {
	var testTable = []struct {
		desc         string
		locationData LocationData
		expected     string
	}{
		{
			desc:         "LocationData has a valid name",
			locationData: LocationData{Name: "test"},
			expected:     "test",
		},
		{
			desc:         "LocationData has a valid fromEnv",
			locationData: LocationData{FromEnv: "TEST_ENV"},
			expected:     "foo",
		},
		{
			desc:         "LocationData has both name and fromEnv (should take name)",
			locationData: LocationData{Name: "test", FromEnv: "TEST_ENV"},
			expected:     "test",
		},
		{
			desc:         "LocationData has no fromEnv or name",
			locationData: LocationData{},
			expected:     "",
		},
	}

	err := os.Setenv("TEST_ENV", "foo")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err := os.Unsetenv("TEST_ENV")
		if err != nil {
			t.Error(err)
		}
	}()

	for _, testCase := range testTable {
		actual, err := testCase.locationData.Get()
		assert.NoError(t, err, testCase.desc)
		assert.Equal(t, testCase.expected, actual, testCase.desc)
	}
}

// TestLocationData_Get_Error tests getting the location data with errors.
func TestLocationData_Get_Error(t *testing.T) {
	var testTable = []struct {
		desc         string
		locationData LocationData
	}{
		{
			desc:         "LocationData fromEnv does not resolve",
			locationData: LocationData{FromEnv: "FOO_BAR_BAZ"},
		},
	}

	for _, testCase := range testTable {
		actual, err := testCase.locationData.Get()
		assert.Error(t, err, testCase.desc)
		assert.Equal(t, "", actual, testCase.desc)
	}
}

// TestDeviceKind_Validate_Ok tests validating a DeviceKind with no errors.
func TestDeviceKind_Validate_Ok(t *testing.T) {
	var testTable = []struct {
		desc string
		kind DeviceKind
	}{
		{
			desc: "DeviceKind has a valid name",
			kind: DeviceKind{
				Name: "test",
			},
		},
		{
			desc: "DeviceKind has a valid name and instances",
			kind: DeviceKind{
				Name:      "test",
				Instances: []*DeviceInstance{{Location: "test"}},
			},
		},
		{
			desc: "DeviceKind has a valid name, instances, and outputs",
			kind: DeviceKind{
				Name:      "test",
				Instances: []*DeviceInstance{{Location: "test"}},
				Outputs:   []*DeviceOutput{{Type: "test"}},
			},
		},
		{
			desc: "DeviceKind has valid name, invalid instances (DeviceInstance not validated here)",
			kind: DeviceKind{
				Name:      "test",
				Instances: []*DeviceInstance{{Location: ""}},
			},
		},
		{
			desc: "DeviceKind has valid name, invalid outputs (DeviceOutputs not validated here)",
			kind: DeviceKind{
				Name:    "test",
				Outputs: []*DeviceOutput{{Type: ""}},
			},
		},
	}

	for _, testCase := range testTable {
		merr := errors.NewMultiError("test")

		testCase.kind.Validate(merr)
		assert.NoError(t, merr.Err(), testCase.desc)
	}
}

// TestDeviceKind_Validate_Error tests validating a DeviceKind with errors.
func TestDeviceKind_Validate_Error(t *testing.T) {
	var testTable = []struct {
		desc     string
		errCount int
		kind     DeviceKind
	}{
		{
			desc:     "DeviceKind has no name specified",
			errCount: 1,
			kind:     DeviceKind{},
		},
	}

	for _, testCase := range testTable {
		merr := errors.NewMultiError("test")

		testCase.kind.Validate(merr)
		assert.Error(t, merr.Err(), testCase.desc)
		assert.Equal(t, testCase.errCount, len(merr.Errors), merr.Error())
	}
}

// TestDeviceInstance_Validate_Ok tests validating a DeviceInstance with no errors.
func TestDeviceInstance_Validate_Ok(t *testing.T) {
	var testTable = []struct {
		desc     string
		instance DeviceInstance
	}{
		{
			desc: "DeviceInstance has a valid location",
			instance: DeviceInstance{
				Location: "test",
			},
		},
		{
			desc: "DeviceInstance has a valid location and outputs",
			instance: DeviceInstance{
				Location: "test",
				Outputs:  []*DeviceOutput{{Type: "test"}},
			},
		},
		{
			desc: "DeviceInstance has valid location, invalid outputs (DeviceOutputs not validated here)",
			instance: DeviceInstance{
				Location: "test",
				Outputs:  []*DeviceOutput{{Type: ""}},
			},
		},
	}

	for _, testCase := range testTable {
		merr := errors.NewMultiError("test")

		testCase.instance.Validate(merr)
		assert.NoError(t, merr.Err(), testCase.desc)
	}
}

// TestDeviceInstance_Validate_Error tests validating a DeviceInstance with errors.
func TestDeviceInstance_Validate_Error(t *testing.T) {
	var testTable = []struct {
		desc     string
		errCount int
		instance DeviceInstance
	}{
		{
			desc:     "DeviceInstance has a no location",
			errCount: 1,
			instance: DeviceInstance{
				Location: "",
			},
		},
	}

	for _, testCase := range testTable {
		merr := errors.NewMultiError("test")

		testCase.instance.Validate(merr)
		assert.Error(t, merr.Err(), testCase.desc)
		assert.Equal(t, testCase.errCount, len(merr.Errors), merr.Error())
	}
}

// TestDeviceOutput_Validate_Ok tests validating a DeviceOutput with no errors.
func TestDeviceOutput_Validate_Ok(t *testing.T) {
	var testTable = []struct {
		desc   string
		output DeviceOutput
	}{
		{
			desc: "DeviceOutput has valid type",
			output: DeviceOutput{
				Type: "test",
			},
		},
	}

	for _, testCase := range testTable {
		merr := errors.NewMultiError("test")

		testCase.output.Validate(merr)
		assert.NoError(t, merr.Err(), testCase.desc)
	}
}

// TestDeviceOutput_Validate_Error tests validating a DeviceOutput with errors.
func TestDeviceOutput_Validate_Error(t *testing.T) {
	var testTable = []struct {
		desc     string
		errCount int
		output   DeviceOutput
	}{
		{
			desc:     "DeviceOutput has no type",
			errCount: 1,
			output: DeviceOutput{
				Type: "",
			},
		},
	}

	for _, testCase := range testTable {
		merr := errors.NewMultiError("test")

		testCase.output.Validate(merr)
		assert.Error(t, merr.Err(), testCase.desc)
		assert.Equal(t, testCase.errCount, len(merr.Errors), merr.Error())
	}
}

// TestDeviceConfig_ValidateDeviceConfigDataOk tests validating config data when there
// are no errors.
func TestDeviceConfig_ValidateDeviceConfigDataOk(t *testing.T) {
	var testTable = []struct {
		desc   string
		config DeviceConfig
	}{
		{
			desc: "no data field in the device config",
			config: DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations:     []*LocationConfig{},
				Devices:       []*DeviceKind{},
			},
		},
		{
			desc: "data in the device kind output",
			config: DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations:     []*LocationConfig{},
				Devices: []*DeviceKind{
					{
						Outputs: []*DeviceOutput{
							{
								Data: map[string]interface{}{
									"foo": "bar",
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "data in the device instance",
			config: DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations:     []*LocationConfig{},
				Devices: []*DeviceKind{
					{
						Instances: []*DeviceInstance{
							{
								Data: map[string]interface{}{
									"foo": "bar",
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "data in the device instance output",
			config: DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations:     []*LocationConfig{},
				Devices: []*DeviceKind{
					{
						Instances: []*DeviceInstance{
							{
								Outputs: []*DeviceOutput{
									{
										Data: map[string]interface{}{
											"foo": "bar",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// define the validator function. it does nothing and returns nil,
	// so we should never get an error.
	var validator = func(_ map[string]interface{}) error {
		return nil
	}

	for _, testCase := range testTable {
		err := testCase.config.ValidateDeviceConfigData(validator)
		assert.NoError(t, err.Err(), testCase.desc)
	}
}

// TestDeviceConfig_ValidateDeviceConfigDataError tests validating config data when there
// are errors.
func TestDeviceConfig_ValidateDeviceConfigDataError(t *testing.T) {
	var testTable = []struct {
		desc   string
		config DeviceConfig
	}{
		{
			desc: "data in the device kind output",
			config: DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations:     []*LocationConfig{},
				Devices: []*DeviceKind{
					{
						Outputs: []*DeviceOutput{
							{
								Data: map[string]interface{}{
									"foo": "bar",
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "data in the device instance",
			config: DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations:     []*LocationConfig{},
				Devices: []*DeviceKind{
					{
						Instances: []*DeviceInstance{
							{
								Data: map[string]interface{}{
									"foo": "bar",
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "data in the device instance output",
			config: DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations:     []*LocationConfig{},
				Devices: []*DeviceKind{
					{
						Instances: []*DeviceInstance{
							{
								Outputs: []*DeviceOutput{
									{
										Data: map[string]interface{}{
											"foo": "bar",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// define the validator function. it does nothing and returns an error, so
	// all things should fail validation.
	var validator = func(_ map[string]interface{}) error {
		return fmt.Errorf("test error")
	}

	for _, testCase := range testTable {
		err := testCase.config.ValidateDeviceConfigData(validator)
		assert.Error(t, err.Err(), testCase.desc)
	}
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

package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
)

// TestVerificationInit tests that the init function initialized things correctly.
func TestVerificationInit(t *testing.T) {
	assert.NotNil(t, deviceConfigLocations)
	assert.Empty(t, deviceConfigLocations)
	assert.NotNil(t, deviceConfigKinds)
	assert.Empty(t, deviceConfigKinds)
}

// TestVerifyConfigs tests verifying the unified device config.
func TestVerifyConfigs(t *testing.T) {
	cfg := &DeviceConfig{
		SchemeVersion: SchemeVersion{Version: "1.0"},
		Locations:     []*LocationConfig{},
		Devices:       []*DeviceKind{},
	}

	err := verifyConfigs(cfg)
	assert.NoError(t, err.Err())
}

// Test_verifyDeviceConfigLocations_Ok tests that there are no conflicting locations
func Test_verifyDeviceConfigLocations_Ok(t *testing.T) {
	defer func() {
		deviceConfigLocations = map[string]*LocationConfig{}
	}()

	cfg := &DeviceConfig{
		SchemeVersion: SchemeVersion{Version: "1.0"},
		Locations: []*LocationConfig{
			{
				Name:  "foo",
				Rack:  &LocationData{Name: "rack"},
				Board: &LocationData{Name: "board"},
			},
			{
				Name:  "foo",
				Rack:  &LocationData{Name: "rack"},
				Board: &LocationData{Name: "board"},
			}, {
				Name:  "bar",
				Rack:  &LocationData{Name: "test"},
				Board: &LocationData{Name: "test"},
			},
			{
				Name:  "baz",
				Rack:  &LocationData{Name: "1"},
				Board: &LocationData{Name: "2"},
			},
		},
		Devices: []*DeviceKind{},
	}

	err := errors.NewMultiError("test")
	verifyDeviceConfigLocations(cfg, err)
	assert.NoError(t, err.Err())
}

// Test_verifyDeviceConfigLocations_Error tests that there are conflicting locations
func Test_verifyDeviceConfigLocations_Error(t *testing.T) {
	defer func() {
		deviceConfigLocations = map[string]*LocationConfig{}
	}()

	cfg := &DeviceConfig{
		SchemeVersion: SchemeVersion{Version: "1.0"},
		Locations: []*LocationConfig{
			{
				Name:  "foo",
				Rack:  &LocationData{Name: "rack"},
				Board: &LocationData{Name: "board"},
			},
			{
				Name:  "foo",
				Rack:  &LocationData{Name: "different"},
				Board: &LocationData{Name: "different"},
			},
			{
				Name:  "bar",
				Rack:  &LocationData{Name: "rack"},
				Board: &LocationData{Name: "board"},
			},
		},
		Devices: []*DeviceKind{},
	}

	err := errors.NewMultiError("test")
	verifyDeviceConfigLocations(cfg, err)
	assert.Error(t, err.Err())
	assert.Equal(t, 1, len(err.Errors), err.Error())
}

// Test_verifyDeviceConfigDeviceKinds_Ok tests that there are no duplicate device kinds defined.
func Test_verifyDeviceConfigDeviceKinds_Ok(t *testing.T) {
	defer func() {
		deviceConfigKinds = map[string]*DeviceKind{}
	}()

	cfg := &DeviceConfig{
		SchemeVersion: SchemeVersion{Version: "1.0"},
		Locations:     []*LocationConfig{},
		Devices: []*DeviceKind{
			{Name: "test"},
			{Name: "foo"},
			{Name: "bar"},
		},
	}

	err := errors.NewMultiError("test")
	verifyDeviceConfigDeviceKinds(cfg, err)
	assert.NoError(t, err.Err())
}

// Test_verifyDeviceConfigDeviceKinds_Error tests that there are duplicate device kinds defined.
func Test_verifyDeviceConfigDeviceKinds_Error(t *testing.T) {
	defer func() {
		deviceConfigKinds = map[string]*DeviceKind{}
	}()

	cfg := &DeviceConfig{
		SchemeVersion: SchemeVersion{Version: "1.0"},
		Locations:     []*LocationConfig{},
		Devices: []*DeviceKind{
			{Name: "test"},
			{Name: "foo"},
			{Name: "bar"},
			{Name: "foo"},
			{Name: "test"},
		},
	}

	err := errors.NewMultiError("test")
	verifyDeviceConfigDeviceKinds(cfg, err)
	assert.Error(t, err.Err())
	assert.Equal(t, 2, len(err.Errors), err.Error())
}

// Test_verifyDeviceConfigInstances_Ok tests that the device instances are all correct.
func Test_verifyDeviceConfigInstances_Ok(t *testing.T) {
	defer delete(deviceConfigLocations, "foo")
	defer delete(deviceConfigLocations, "bar")

	// add the expected locations to the location map
	deviceConfigLocations["foo"] = &LocationConfig{
		Name:  "foo",
		Rack:  &LocationData{Name: "rack"},
		Board: &LocationData{Name: "board"},
	}
	deviceConfigLocations["bar"] = &LocationConfig{
		Name:  "bar",
		Rack:  &LocationData{Name: "test"},
		Board: &LocationData{Name: "test"},
	}

	cfg := &DeviceConfig{
		SchemeVersion: SchemeVersion{Version: "1.0"},
		Locations:     []*LocationConfig{},
		Devices: []*DeviceKind{
			{
				Name: "test",
				Instances: []*DeviceInstance{
					{Location: "foo"},
					{Location: "foo"},
					{Location: "bar"},
				},
			},
			{
				Name: "foo",
				Instances: []*DeviceInstance{
					{Location: "bar"},
					{Location: "foo"},
					{Location: "bar"},
				},
			},
		},
	}

	err := errors.NewMultiError("test")
	verifyDeviceConfigInstances(cfg, err)
	assert.NoError(t, err.Err())
}

// Test_verifyDeviceConfigInstances_Error tests errors being detected in the instance
// verification process.
func Test_verifyDeviceConfigInstances_Error(t *testing.T) {
	defer delete(deviceConfigLocations, "foo")

	// add some expected locations to the location map
	deviceConfigLocations["foo"] = &LocationConfig{
		Name:  "foo",
		Rack:  &LocationData{Name: "rack"},
		Board: &LocationData{Name: "board"},
	}

	cfg := &DeviceConfig{
		SchemeVersion: SchemeVersion{Version: "1.0"},
		Locations:     []*LocationConfig{},
		Devices: []*DeviceKind{
			{
				Name: "test",
				Instances: []*DeviceInstance{
					{Location: "foo"},
					{Location: ""},    // err: empty definition
					{Location: "bar"}, // err: doesn't exist
				},
			},
			{
				Name: "foo",
				Instances: []*DeviceInstance{
					{Location: "bar"}, // err: doesn't exist
					{Location: "foo"},
					{Location: "foo"},
				},
			},
		},
	}

	err := errors.NewMultiError("test")
	verifyDeviceConfigInstances(cfg, err)
	assert.Error(t, err.Err())
	assert.Equal(t, 3, len(err.Errors), err.Error())
}

// Test_verifyDeviceConfigOutputs_Ok tests verifying no issues with device outputs.
func Test_verifyDeviceConfigOutputs_Ok(t *testing.T) {
	defer resetContext()

	// add some expected outputs
	ctx.outputTypes["foo"] = &OutputType{Name: "foo"}
	ctx.outputTypes["bar"] = &OutputType{Name: "bar"}

	cfg := &DeviceConfig{
		SchemeVersion: SchemeVersion{Version: "1.0"},
		Locations:     []*LocationConfig{},
		Devices: []*DeviceKind{
			{
				Name: "test",
				Outputs: []*DeviceOutput{
					{Type: "foo"},
				},
				Instances: []*DeviceInstance{
					{
						Location: "foo",
						Outputs: []*DeviceOutput{
							{Type: "foo"},
							{Type: "bar"},
						},
					},
					{
						Location: "foo",
						Outputs: []*DeviceOutput{
							{Type: "foo"},
							{Type: "bar"},
						},
					},
				},
			},
			{
				Name: "foo",
				Outputs: []*DeviceOutput{
					{Type: "foo"},
					{Type: "bar"},
				},
				Instances: []*DeviceInstance{
					{
						Location: "bar",
						Outputs: []*DeviceOutput{
							{Type: "foo"},
						},
					},
				},
			},
		},
	}

	err := errors.NewMultiError("test")
	verifyDeviceConfigOutputs(cfg, err)
	assert.NoError(t, err.Err())
}

// Test_verifyDeviceConfigOutputs_Error tests verification errors with device outputs.
func Test_verifyDeviceConfigOutputs_Error(t *testing.T) {
	defer resetContext()

	// add some expected outputs
	ctx.outputTypes["foo"] = &OutputType{Name: "foo"}

	cfg := &DeviceConfig{
		SchemeVersion: SchemeVersion{Version: "1.0"},
		Locations:     []*LocationConfig{},
		Devices: []*DeviceKind{
			{
				Name: "test",
				Outputs: []*DeviceOutput{
					{Type: "foo"},
				},
				Instances: []*DeviceInstance{
					{
						Location: "foo",
						Outputs: []*DeviceOutput{
							{Type: "foo"},
							{Type: "bar"}, // err: doesn't exist
						},
					},
					{
						Location: "foo",
						Outputs: []*DeviceOutput{
							{Type: "foo"},
							{Type: "bar"}, // err: doesn't exist
						},
					},
				},
			},
			{
				Name: "foo",
				Outputs: []*DeviceOutput{
					{Type: "foo"},
					{Type: "bar"}, // err: doesn't exist
				},
				Instances: []*DeviceInstance{
					{
						Location: "bar",
						Outputs: []*DeviceOutput{
							{Type: "foo"},
						},
					},
				},
			},
		},
	}

	err := errors.NewMultiError("test")
	verifyDeviceConfigOutputs(cfg, err)
	assert.Error(t, err.Err())
	assert.Equal(t, 3, len(err.Errors), err.Error())
}

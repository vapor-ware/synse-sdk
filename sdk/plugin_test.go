package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/policies"
)

// TestNewPlugin tests creating a new plugin.
func TestNewPlugin(t *testing.T) {
	var testTable = []struct {
		desc    string
		options []PluginOption
	}{
		{
			desc:    "no plugin options",
			options: []PluginOption{},
		},
		{
			desc: "one options",
			options: []PluginOption{
				CustomDeviceIdentifier(func(i map[string]interface{}) string {
					return "test"
				}),
			},
		},
		{
			desc: "multiple options",
			options: []PluginOption{
				CustomDeviceIdentifier(func(i map[string]interface{}) string {
					return "test"
				}),
				CustomDynamicDeviceRegistration(func(i map[string]interface{}) ([]*Device, error) {
					return nil, nil
				}),
			},
		},
		{
			desc: "duplicate options",
			options: []PluginOption{
				CustomDeviceIdentifier(func(i map[string]interface{}) string {
					return "foo"
				}),
				CustomDeviceIdentifier(func(i map[string]interface{}) string {
					return "bar"
				}),
			},
		},
	}

	for _, testCase := range testTable {
		plugin := NewPlugin(testCase.options...)
		assert.NotNil(t, plugin)
	}
}

// TestPlugin_SetConfigPolicies tests setting the config policies for the plugin.
func TestPlugin_SetConfigPolicies(t *testing.T) {
	plugin := NewPlugin()
	plugin.SetConfigPolicies(
		policies.DeviceConfigOptional,
		policies.PluginConfigRequired,
	)
	assert.Equal(t, 2, len(plugin.policies))
}

// TestPlugin_RegisterOutputTypes tests registering the output types for the plugin.
func TestPlugin_RegisterOutputTypes(t *testing.T) {
	defer func() {
		outputTypeMap = map[string]*config.OutputType{}
	}()

	types := []*config.OutputType{
		{Name: "foo"},
		{Name: "bar"},
		{Name: "baz"},
	}
	plugin := NewPlugin()
	assert.Equal(t, 0, len(outputTypeMap))

	err := plugin.RegisterOutputTypes(types...)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(outputTypeMap))
}

// TestPlugin_RegisterOutputTypesError tests registering the output types for the
// plugin when duplicate types are specified.
func TestPlugin_RegisterOutputTypesError(t *testing.T) {
	defer func() {
		outputTypeMap = map[string]*config.OutputType{}
	}()

	types := []*config.OutputType{
		{Name: "foo"},
		{Name: "bar"},
		{Name: "foo"},
	}
	plugin := NewPlugin()
	assert.Equal(t, 0, len(outputTypeMap))

	err := plugin.RegisterOutputTypes(types...)
	assert.Error(t, err)
	assert.True(t, len(outputTypeMap) > 0)
}

// TestPlugin_RegisterPreRunActions tests registering pre-run actions.
func TestPlugin_RegisterPreRunActions(t *testing.T) {
	defer func() {
		preRunActions = []pluginAction{}
	}()

	actions := []pluginAction{
		func(_ *Plugin) error { return nil },
		func(_ *Plugin) error { return nil },
		func(_ *Plugin) error { return nil },
	}

	plugin := NewPlugin()

	assert.Equal(t, 0, len(preRunActions))
	plugin.RegisterPreRunActions(actions...)
	assert.Equal(t, 3, len(preRunActions))
}

// TestPlugin_RegisterPostRunActions tests registering post-run actions.
func TestPlugin_RegisterPostRunActions(t *testing.T) {
	defer func() {
		postRunActions = []pluginAction{}
	}()

	actions := []pluginAction{
		func(_ *Plugin) error { return nil },
		func(_ *Plugin) error { return nil },
		func(_ *Plugin) error { return nil },
	}

	plugin := NewPlugin()

	assert.Equal(t, 0, len(postRunActions))
	plugin.RegisterPostRunActions(actions...)
	assert.Equal(t, 3, len(postRunActions))
}

// TestPlugin_RegisterDeviceSetupActions tests registering device setup actions.
func TestPlugin_RegisterDeviceSetupActions(t *testing.T) {
	defer func() {
		deviceSetupActions = map[string][]deviceAction{}
	}()

	action := func(_ *Plugin, _ *Device) error { return nil }

	plugin := NewPlugin()

	assert.Equal(t, 0, len(deviceSetupActions))
	plugin.RegisterDeviceSetupActions("kind=test", action, action, action)
	assert.Equal(t, 1, len(deviceSetupActions))
	assert.Equal(t, 3, len(deviceSetupActions["kind=test"]))
}

// TestPlugin_RegisterDeviceSetupActions2 tests registering device setup actions when
// some already exist.
func TestPlugin_RegisterDeviceSetupActions2(t *testing.T) {
	defer func() {
		deviceSetupActions = map[string][]deviceAction{}
	}()

	action := func(_ *Plugin, _ *Device) error { return nil }

	// add something to the device setup actions to start with
	deviceSetupActions["kind=test"] = []deviceAction{action}

	plugin := NewPlugin()

	assert.Equal(t, 1, len(deviceSetupActions))
	plugin.RegisterDeviceSetupActions("kind=test", action)
	assert.Equal(t, 1, len(deviceSetupActions))
	assert.Equal(t, 2, len(deviceSetupActions["kind=test"]))
}

// TestPlugin_RegisterDeviceHandlers tests registering DeviceHandlers with the plugin.
func TestPlugin_RegisterDeviceHandlers(t *testing.T) {
	defer func() {
		deviceHandlers = []*DeviceHandler{}
	}()

	fooHandler := &DeviceHandler{Name: "foo"}
	barHandler := &DeviceHandler{Name: "bar"}

	plugin := NewPlugin()

	assert.Equal(t, 0, len(deviceHandlers))
	plugin.RegisterDeviceHandlers(fooHandler, barHandler)
	assert.Equal(t, 2, len(deviceHandlers))
}

// TestPlugin_logStartupInfo tests logging out info which is done on startup.
// There isn't much to check here other than it runs and completes without
// any issues.
func TestPlugin_logStartupInfo(t *testing.T) {
	plugin := NewPlugin()
	plugin.logStartupInfo()
}

// TestPlugin_resolveFlags tests resolving flags. In this case, no flags are
// set, so it should ultimately do nothing.
func TestPlugin_resolveFlags(t *testing.T) {
	plugin := NewPlugin()
	plugin.resolveFlags()
}

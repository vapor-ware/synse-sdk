package sdk

// DeviceIdentifier is a function that produces a string that can be used to
// identify a device deterministically. The returned string should be a composite
// from the Device's config data.
type DeviceIdentifier func(map[string]interface{}) string

// DynamicDeviceRegistrar is a function that takes a Plugin config's "dynamic
// registration" data and generates Device instances from it. How this is done
// is specific to the plugin/protocol.
type DynamicDeviceRegistrar func(map[string]interface{}) ([]*Device, error)

// DynamicDeviceConfigRegistrar is a function that takes a Plugin config's "dynamic
// registration" data and generates DeviceConfig instances from it. How this is done
// is specific to the plugin/protocol.
type DynamicDeviceConfigRegistrar func(map[string]interface{}) ([]*DeviceConfig, error)

// DeviceDataValidator is a function that takes the `Data` field of a device config
// and performs some validation on it. This allows users to provide validation on the
// plugin-specific config fields.
type DeviceDataValidator func(map[string]interface{}) error

// Context is the global context for the plugin. It stores various plugin settings,
// including handler functions for customizable plugin functionality.
var Context = newPluginContext()

// PluginContext holds context information for the plugin. Having the context
// global allows simpler access, without having to pass references to the plugin
// through many of our functions.
type PluginContext struct {
	deviceIdentifier             DeviceIdentifier
	dynamicDeviceRegistrar       DynamicDeviceRegistrar
	dynamicDeviceConfigRegistrar DynamicDeviceConfigRegistrar
	deviceDataValidator          DeviceDataValidator
}

// newPluginContext creates a new instance of the plugin context, supplying the default
// values for any context fields that have defaults.
func newPluginContext() *PluginContext {
	return &PluginContext{
		deviceIdentifier:             defaultDeviceIdentifier,
		dynamicDeviceRegistrar:       defaultDynamicDeviceRegistration,
		dynamicDeviceConfigRegistrar: defaultDynamicDeviceConfigRegistration,
		deviceDataValidator:          defaultDeviceDataValidator,
	}
}

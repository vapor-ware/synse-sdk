package sdk

// Context is the global context for the plugin. It stores various plugin settings,
// including handler functions for customizable plugin functionality.
var Context = newPluginContext()

// pluginContext holds context information for the plugin. Having the context
// global allows simpler access, without having to pass references to the plugin
// through many of our functions.
type PluginContext struct {
	deviceIdentifier             DeviceIdentifier
	dynamicDeviceRegistrar       DynamicDeviceRegistrar
	dynamicDeviceConfigRegistrar DynamicDeviceConfigRegistrar
}

// newPluginContext creates a new instance of the plugin context, supplying the default
// values for any context fields that have defaults.
func newPluginContext() *PluginContext {
	return &PluginContext{
		deviceIdentifier:             defaultDeviceIdentifier,
		dynamicDeviceRegistrar:       defaultDynamicDeviceRegistration,
		dynamicDeviceConfigRegistrar: defaultDynamicDeviceConfigRegistration,
	}
}

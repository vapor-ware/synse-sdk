package sdk

// A PluginOption sets optional configurations or functional capabilities for
// a plugin. This includes things like device identification and device
// registration behaviors.
type PluginOption func(*Plugin)

// defaultOptions defines the default plugin options. These are applied to a
// new Plugin (via `NewPlugin`) if there are no corresponding custom options
// specified.
var defaultOptions = []PluginOption{
	defaultDeviceIdentifierOption,
	defaultDynamicDeviceRegistrationOption,
	defaultDynamicDeviceConfigRegistrationOption,
}

// CustomDeviceIdentifier lets you set a custom function for creating a deterministic
// identifier for a device using the config data for the device.
func CustomDeviceIdentifier(identifier DeviceIdentifier) PluginOption {
	return func(plugin *Plugin) {
		plugin.deviceIdentifier = identifier
	}
}

// defaultDeviceIdentifierOption applies the default behavior for creating a deterministic
// identifier to the plugin, if it does not already have one set.
func defaultDeviceIdentifierOption(plugin *Plugin) {
	if plugin.deviceIdentifier == nil {
		plugin.deviceIdentifier = defaultDeviceIdentifier
	}
}

// CustomDynamicDeviceRegistration lets you set a custom function for dynamically registering
// Device instances using the data from the "dynamic registration" field in the Plugin config.
func CustomDynamicDeviceRegistration(registrar DynamicDeviceRegistrar) PluginOption {
	return func(plugin *Plugin) {
		plugin.dynamicDeviceRegistrar = registrar
	}
}

// defaultDynamicDeviceRegistrationOption applies the default behavior for dynamic device
// registration to the plugin, if it does not already have one set.
func defaultDynamicDeviceRegistrationOption(plugin *Plugin) {
	if plugin.dynamicDeviceRegistrar == nil {
		plugin.dynamicDeviceRegistrar = defaultDynamicDeviceRegistration
	}
}

// CustomDynamicDeviceConfigRegistration lets you set a custom function for dynamically
// registering DeviceConfig instances using the data from the "dynamic registration" field
// in the Plugin config.
func CustomDynamicDeviceConfigRegistration(registrar DynamicDeviceConfigRegistrar) PluginOption {
	return func(plugin *Plugin) {
		plugin.dynamicDeviceConfigRegistrar = registrar
	}
}

// defaultDynamicDeviceConfigRegistrationOption applies the default behavior for dynamic
// device config registration to the plugin, if it does not already have one set.
func defaultDynamicDeviceConfigRegistrationOption(plugin *Plugin) {
	if plugin.dynamicDeviceConfigRegistrar == nil {
		plugin.dynamicDeviceConfigRegistrar = defaultDynamicDeviceConfigRegistration
	}
}

package sdk

// A PluginOption sets optional configurations or functional capabilities for
// a plugin. This includes things like device identification and device
// registration behaviors.
type PluginOption func(*Plugin)

// CustomDeviceIdentifier lets you set a custom function for creating a deterministic
// identifier for a device using the config data for the device.
func CustomDeviceIdentifier(identifier DeviceIdentifier) PluginOption {
	return func(plugin *Plugin) {
		plugin.deviceIdentifier = identifier
	}
}

// CustomDynamicDeviceRegistration lets you set a custom function for dynamically registering
// Device instances using the data from the "dynamic registration" field in the Plugin config.
func CustomDynamicDeviceRegistration(registrar DynamicDeviceRegistrar) PluginOption {
	return func(plugin *Plugin) {
		plugin.dynamicRegistrar = registrar
	}
}

// CustomDynamicDeviceConfigRegistration lets you set a custom function for dynamically
// registering DeviceConfig instances using the data from the "dynamic registration" field
// in the Plugin config.
func CustomDynamicDeviceConfigRegistration(registrar DynamicDeviceConfigRegistrar) PluginOption {
	return func(plugin *Plugin) {
		plugin.dynamicConfigRegistrar = registrar
	}
}

// CustomDeviceDataValidator lets you set a custom function for validating the Data field
// of a device's config. By default, this data is not validated by the SDK, since it is
// plugin-specific.
func CustomDeviceDataValidator(validator DeviceDataValidator) PluginOption {
	return func(plugin *Plugin) {
		plugin.deviceDataValidator = validator
	}
}

// PluginConfigRequired is a PluginOption which designates that a Plugin should require
// a plugin config and will fail if it does not detect one. By default, a Plugin considers
// them optional and will use a set of default configurations if no config is found.
func PluginConfigRequired() PluginOption {
	return func(plugin *Plugin) {
		plugin.pluginCfgRequired = true
	}
}

// DeviceConfigOptional is a PluginOption which designates that a Plugin should not require
// device configurations to be required, as they are by default. This can be the case when
// a plugin may support dynamic device configuration, so pre-defined device configs may not
// be specified.
func DeviceConfigOptional() PluginOption {
	return func(plugin *Plugin) {
		plugin.deviceCfgRequired = false
	}
}

// DynamicConfigRequired is a PluginOption which designates that a Plugin requires dynamic
// device configuration. By default, dynamic device configuration is optional. This can
// be set if a plugin is designed to only load devices in a dynamic fashion, and not through
// pre-defined config files.
func DynamicConfigRequired() PluginOption {
	return func(plugin *Plugin) {
		plugin.dynamicCfgRequired = true
	}
}
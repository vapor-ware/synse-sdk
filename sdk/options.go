package sdk

// A PluginOption sets optional configurations or functional capabilities for
// a plugin. This includes things like device identification and device
// registration behaviors.
type PluginOption func(*PluginContext)

// CustomDeviceIdentifier lets you set a custom function for creating a deterministic
// identifier for a device using the config data for the device.
func CustomDeviceIdentifier(identifier DeviceIdentifier) PluginOption {
	return func(ctx *PluginContext) {
		ctx.deviceIdentifier = identifier
	}
}

// CustomDynamicDeviceRegistration lets you set a custom function for dynamically registering
// Device instances using the data from the "dynamic registration" field in the Plugin config.
func CustomDynamicDeviceRegistration(registrar DynamicDeviceRegistrar) PluginOption {
	return func(ctx *PluginContext) {
		ctx.dynamicDeviceRegistrar = registrar
	}
}

// CustomDynamicDeviceConfigRegistration lets you set a custom function for dynamically
// registering DeviceConfig instances using the data from the "dynamic registration" field
// in the Plugin config.
func CustomDynamicDeviceConfigRegistration(registrar DynamicDeviceConfigRegistrar) PluginOption {
	return func(ctx *PluginContext) {
		ctx.dynamicDeviceConfigRegistrar = registrar
	}
}

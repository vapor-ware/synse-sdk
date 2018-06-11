package sdk

// ctx is the global context for the plugin. It stores various plugin settings,
// data, and handler functions for customizable plugin functionality.
var ctx = newPluginContext()

// PluginContext holds context information for the plugin. Having the context
// global allows simpler access, without having to pass references to the plugin
// through many of our functions.
type PluginContext struct {
	// The handler functions that can extend/modify a plugin's behavior.
	// These can be set via PluginOptions, or can use a default handler.
	deviceIdentifier             DeviceIdentifier
	dynamicDeviceRegistrar       DynamicDeviceRegistrar
	dynamicDeviceConfigRegistrar DynamicDeviceConfigRegistrar
	deviceDataValidator          DeviceDataValidator

	// outputTypes is a map where the the key is the name of the output type
	// and the value is the corresponding OutputType.
	outputTypes map[string]*OutputType

	// devices holds all of the known devices configured for the plugin.
	devices map[string]*Device

	// deviceHandlers holds all of the DeviceHandlers that are registered with the plugin.
	deviceHandlers []*DeviceHandler

	/// preRunActions holds all of the known plugin actions to run prior to starting
	// up the plugin server and data manager.
	preRunActions []pluginAction

	// postRunActions holds all of the known plugin actions to run after terminating
	// the plugin server and data manager.
	postRunActions []pluginAction

	// deviceSetupActions holds all of the known device device setup actions to run
	// prior to starting up the plugin server and data manager. The map key is the
	// filter used to apply the deviceAction value to a Device instance.
	deviceSetupActions map[string][]deviceAction
}

// newPluginContext creates a new instance of the plugin context, supplying the default
// values for any context fields that have defaults.
func newPluginContext() *PluginContext {
	return &PluginContext{
		deviceIdentifier:             defaultDeviceIdentifier,
		dynamicDeviceRegistrar:       defaultDynamicDeviceRegistration,
		dynamicDeviceConfigRegistrar: defaultDynamicDeviceConfigRegistration,
		deviceDataValidator:          defaultDeviceDataValidator,

		outputTypes:        map[string]*OutputType{},
		devices:            map[string]*Device{},
		deviceHandlers:     []*DeviceHandler{},
		preRunActions:      []pluginAction{},
		postRunActions:     []pluginAction{},
		deviceSetupActions: map[string][]deviceAction{},
	}
}

// resetContext is a utility function that is used as a test helper to clear the plugin
// context. This should not be used outside of testing.
func resetContext() { // nolint
	ctx = newPluginContext()
}

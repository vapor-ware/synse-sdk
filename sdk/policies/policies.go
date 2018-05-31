package policies

// ConfigPolicy is a type that defines a behavior profile for the plugin
// on how it should handle configurations and configuration errors.
type ConfigPolicy uint8

const (
	// ignore first iota value since we don't want a zero-value.
	_ ConfigPolicy = iota

	// PluginConfigOptional is a policy that allows no plugin config to
	// be specified, so the plugin can just use default values. This is
	// the default policy for plugins.
	PluginConfigOptional

	// PluginConfigRequired is a policy that requires a plugin to have a
	// plugin configuration specified.
	PluginConfigRequired

	// DeviceConfigOptional is a policy that allows no device config files
	// to be specified. Some plugins may not device config files, as they
	// could use dynamic registration exclusively.
	DeviceConfigOptional

	// DeviceConfigRequired is a policy that requires device config files to
	// be present for the plugin to run. This is the default policy for plugins.
	DeviceConfigRequired
)

// policyStrings maps ConfigPolicies to their name.
var policyStrings = map[ConfigPolicy]string{
	PluginConfigOptional: "PluginConfigOptional",
	PluginConfigRequired: "PluginConfigRequired",
	DeviceConfigOptional: "DeviceConfigOptional",
	DeviceConfigRequired: "DeviceConfigRequired",
}

// String returns the name of the ConfigPolicy.
func (policy ConfigPolicy) String() string {
	if name, ok := policyStrings[policy]; ok {
		return name
	}
	return "unknown"
}

// PolicyManager is a global policyManager instance that holds the
// policies for the plugin.
var PolicyManager = policyManager{}

// policyManager is used to track the different sets of policies set for
// a plugin.
type policyManager struct {
	pluginConfigPolicy ConfigPolicy
	deviceConfigPolicy ConfigPolicy
}

// GetPluginConfigPolicy gets the plugin config policy for the plugin. If
// none was set by the plugin, this will return the default policy.
func (pm *policyManager) GetPluginConfigPolicy() ConfigPolicy {
	if PolicyManager.pluginConfigPolicy != 0 {
		return PolicyManager.pluginConfigPolicy
	}
	return PluginConfigOptional
}

// GetDeviceConfigPolicy gets the device config policy for the plugin. If
// none was set by the plugin, this will return the default policy.
func (pm *policyManager) GetDeviceConfigPolicy() ConfigPolicy {
	if PolicyManager.deviceConfigPolicy != 0 {
		return PolicyManager.deviceConfigPolicy
	}
	return DeviceConfigRequired
}

// Set sets the ConfigPolicies for the plugin.
func Set(policies []ConfigPolicy) {
	for _, policy := range policies {
		switch policy {
		case PluginConfigRequired, PluginConfigOptional:
			PolicyManager.pluginConfigPolicy = policy
		case DeviceConfigRequired, DeviceConfigOptional:
			PolicyManager.deviceConfigPolicy = policy
		}
	}
}

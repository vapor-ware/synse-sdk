package policies

import "github.com/vapor-ware/synse-sdk/sdk/logger"

// ConfigPolicy is a type that defines a behavior profile for the plugin
// on how it should handle configurations and configuration errors.
type ConfigPolicy uint8

const (
	// NoPolicy is the zero-value policy. It means that there is no policy
	// specified.
	NoPolicy ConfigPolicy = iota

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
	NoPolicy: "NoPolicy",
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

// policyManager is a global policyManager instance that holds the
// policies for the plugin.
var policyManager = manager{}

// manager is used to track the policies set for a plugin.
type manager struct {
	pluginConfigPolicy ConfigPolicy
	deviceConfigPolicy ConfigPolicy
}

// Set sets the policies for the manager to track.
func (m *manager) Set(policies []ConfigPolicy) {
	for _, policy := range policies {
		switch policy {
		case PluginConfigRequired, PluginConfigOptional:
			m.pluginConfigPolicy = policy
		case DeviceConfigRequired, DeviceConfigOptional:
			m.deviceConfigPolicy = policy
		}
	}
}

// GetPluginConfigPolicy gets the plugin config policy for the plugin. If
// none was explicitly set by the plugin, this will return the default policy.
func GetPluginConfigPolicy() ConfigPolicy {
	if policyManager.pluginConfigPolicy == 0 {
		policyManager.pluginConfigPolicy = PluginConfigOptional
	}
	return policyManager.pluginConfigPolicy
}

// GetDeviceConfigPolicy gets the device config policy for the plugin. If
// none was explicitly set by the plugin, this will return the default policy.
func GetDeviceConfigPolicy() ConfigPolicy {
	if policyManager.deviceConfigPolicy == 0 {
		policyManager.deviceConfigPolicy =  DeviceConfigRequired
	}
	return policyManager.deviceConfigPolicy
}

// Apply applies the given config policies to the SDK policy manager. Before
// policies are added to the policy manager, they are first verified OK by
// checking the policy constraints.
func Apply(policies []ConfigPolicy) error {
	err := CheckConstraints(policies)
	if err.HasErrors() {
		logger.Error("applied config policies do not pass constraint checks")
		return err
	}

	// Constraint checking passed, apply the policies to the policy manager.
	policyManager.Set(policies)
	return nil
}

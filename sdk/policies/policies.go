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
	NoPolicy:             "NoPolicy",
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

// defaultManager is a global policyManager instance that holds the
// policies for the plugin.
var defaultManager = manager{}

// manager is used to track the policies set for a plugin.
type manager struct {
	policies []ConfigPolicy

	pluginConfigPolicy ConfigPolicy
	deviceConfigPolicy ConfigPolicy
}

// Add adds a ConfigPolicy to the policies tracked by the manager.
func (m *manager) Add(policy ConfigPolicy) {
	m.policies = append(m.policies, policy)
}

// Add adds a ConfigPolicy to the SDK's policy manager.
func Add(policy ConfigPolicy) {
	defaultManager.Add(policy)
}

// Set sets multiple policies for the manager to track.
func (m *manager) Set(policies []ConfigPolicy) {
	m.policies = append(m.policies, policies...)
}

// Set sets multiple policies for the SDK's policy manager.
func Set(policies []ConfigPolicy) {
	defaultManager.Set(policies)
}

// GetPluginConfigPolicy gets the plugin config policy for the manager. If
// none was explicitly set, this will return the default policy.
func (m *manager) GetPluginConfigPolicy() ConfigPolicy {
	if m.pluginConfigPolicy == NoPolicy {
		for _, p := range m.policies {
			switch p {
			case PluginConfigRequired, PluginConfigOptional:
				m.pluginConfigPolicy = p
			}
		}
		if m.pluginConfigPolicy == NoPolicy {
			m.pluginConfigPolicy = PluginConfigOptional
		}
	}
	return m.pluginConfigPolicy
}

// GetPluginConfigPolicy gets the plugin config policy that was registered
// with the SDK's policy manager. If none was explicitly set, the default
// policy is returned.
func GetPluginConfigPolicy() ConfigPolicy {
	return defaultManager.GetPluginConfigPolicy()
}

// GetDeviceConfigPolicy gets the device config policy for the manager. If
// none was explicitly set, this will return the default policy.
func (m *manager) GetDeviceConfigPolicy() ConfigPolicy {
	if m.deviceConfigPolicy == NoPolicy {
		for _, p := range m.policies {
			switch p {
			case DeviceConfigRequired, DeviceConfigOptional:
				m.deviceConfigPolicy = p
			}
		}
		if m.deviceConfigPolicy == NoPolicy {
			m.deviceConfigPolicy = DeviceConfigRequired
		}
	}
	return m.deviceConfigPolicy
}

// GetDeviceConfigPolicy gets the device config policy that was registered
// with the SDK's policy manager. If none was explicitly set, the default
// policy is returned.
func GetDeviceConfigPolicy() ConfigPolicy {
	return defaultManager.GetDeviceConfigPolicy()
}

// Check checks the policy constraint functions against the manager's set of
// tracked policies. This should be done prior to getting any policies to ensure
// that the policy set is valid to begin with.
func (m *manager) Check() error {
	err := checkConstraints(m.policies)
	if err.HasErrors() {
		logger.Error("applied config policies do not pass constraint checks")
		return err
	}
	return nil
}

// Check checks the policy constraint functions against the policies tracked by
// the SDK's policy manager. This should be done prior to getting any policies
// to ensure that the policy set is valid to begin with.
func Check() error {
	return defaultManager.Check()
}

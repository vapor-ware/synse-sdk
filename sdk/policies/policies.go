package policies

import (
	log "github.com/sirupsen/logrus"
)

// ConfigPolicy is a type that defines a behavior profile for the plugin
// on how it should handle configurations and configuration errors.
type ConfigPolicy uint8

const (
	// NoPolicy is the zero-value policy. It means that there is no policy
	// specified.
	NoPolicy ConfigPolicy = iota

	// PluginConfigFileOptional is a policy that allows zero or more plugin
	// configurations from config file. This is the default policy for
	// plugin config.
	PluginConfigFileOptional

	// PluginConfigFileRequired is a policy that requires a plugin to have a
	// plugin configuration specified.
	PluginConfigFileRequired

	// PluginConfigFileProhibited is a policy that prevents a plugin from using
	// plugin configurations from config file. This can be used if a plugin
	// needs to restrict its configuration paths.
	PluginConfigFileProhibited

	// DeviceConfigFileOptional is a policy that allows zero or more device
	// configurations from config file.
	DeviceConfigFileOptional

	// DeviceConfigFileRequired is a policy that requires a plugin to have
	// one or more device configuration files specified. This is the default
	// policy for device config files.
	DeviceConfigFileRequired

	// DeviceConfigFileProhibited is a policy that prevents a plugin from using
	// device configurations from config file(s). This can be used if a plugin
	// needs to restrict its configuration paths.
	DeviceConfigFileProhibited

	// DeviceConfigDynamicOptional is a policy that allows zero or more device
	// configurations from dynamic device registration. This is the default policy
	// for dynamic device config.
	DeviceConfigDynamicOptional

	// DeviceConfigDynamicRequired is a policy that requires a plugin to have
	// one or more device configurations from dynamic device registration.
	DeviceConfigDynamicRequired

	// DeviceConfigDynamicProhibited is a policy that prevents a plugin from using
	// device configurations from dynamic registration function(s). This can be
	// used if a plugin needs to restrict its configuration paths. This will prohibit
	// both dynamic registration which generates DeviceConfig instance and Device
	// instances.
	DeviceConfigDynamicProhibited

	// TypeConfigFileOptional is a policy that allows zero or more output type
	// configurations from config file. This is the default policy for output
	// type file config.
	TypeConfigFileOptional

	// TypeConfigFileRequired is a policy that requires a plugin to have
	// one or more output type configurations from config file. It does not prohibit
	// the plugin from defining additional type configs directly in its code.
	TypeConfigFileRequired

	// TypeConfigFileProhibited is a policy that prevents a plugin from using
	// type configurations from config file(s). This can be used if a plugin
	// needs to restrict its configuration paths.
	TypeConfigFileProhibited
)

// policyStrings maps ConfigPolicies to their name.
var policyStrings = map[ConfigPolicy]string{
	NoPolicy: "NoPolicy",

	PluginConfigFileOptional:   "PluginConfigFileOptional",
	PluginConfigFileRequired:   "PluginConfigFileRequired",
	PluginConfigFileProhibited: "PluginConfigFileProhibited",

	DeviceConfigFileOptional:   "DeviceConfigFileOptional",
	DeviceConfigFileRequired:   "DeviceConfigFileRequired",
	DeviceConfigFileProhibited: "DeviceConfigFileProhibited",

	DeviceConfigDynamicOptional:   "DeviceConfigDynamicOptional",
	DeviceConfigDynamicRequired:   "DeviceConfigDynamicRequired",
	DeviceConfigDynamicProhibited: "DeviceConfigDynamicProhibited",

	TypeConfigFileOptional:   "TypeConfigFileOptional",
	TypeConfigFileRequired:   "TypeConfigFileRequired",
	TypeConfigFileProhibited: "TypeConfigFileProhibited",
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

	pluginConfigFilePolicy    ConfigPolicy
	deviceConfigFilePolicy    ConfigPolicy
	deviceConfigDynamicPolicy ConfigPolicy
	typeConfigFilePolicy      ConfigPolicy
}

// Add adds a ConfigPolicy to the policies tracked by the manager.
func (m *manager) Add(policy ConfigPolicy) {
	m.policies = append(m.policies, policy)
}

// Add adds a ConfigPolicy to the SDK's policy manager.
func Add(policy ConfigPolicy) {
	defaultManager.Add(policy)
}

// Clear clears the policy manager of all policies and settings.
func (m *manager) Clear() {
	m.policies = []ConfigPolicy{}
	m.pluginConfigFilePolicy = NoPolicy
	m.deviceConfigFilePolicy = NoPolicy
	m.deviceConfigDynamicPolicy = NoPolicy
	m.typeConfigFilePolicy = NoPolicy
}

// Clear clears the SDK's policy manager of all policies and settings.
func Clear() {
	defaultManager.Clear()
}

// Set sets multiple policies for the manager to track.
func (m *manager) Set(policies []ConfigPolicy) {
	m.policies = append(m.policies, policies...)
}

// Set sets multiple policies for the SDK's policy manager.
func Set(policies []ConfigPolicy) {
	defaultManager.Set(policies)
}

// GetPluginConfigFilePolicy gets the plugin config file policy for the manager. If
// no policy was explicitly set, this will return the default policy.
func (m *manager) GetPluginConfigFilePolicy() ConfigPolicy {
	if m.pluginConfigFilePolicy == NoPolicy {
		for _, p := range m.policies {
			switch p {
			case PluginConfigFileRequired, PluginConfigFileOptional, PluginConfigFileProhibited:
				m.pluginConfigFilePolicy = p
			}
		}
		if m.pluginConfigFilePolicy == NoPolicy {
			m.pluginConfigFilePolicy = PluginConfigFileOptional
		}
	}
	return m.pluginConfigFilePolicy
}

// GetPluginConfigFilePolicy gets the plugin config policy that was registered
// with the SDK's policy manager. If no policy was explicitly set, the default
// policy is returned.
func GetPluginConfigFilePolicy() ConfigPolicy {
	return defaultManager.GetPluginConfigFilePolicy()
}

// GetDeviceConfigFilePolicy gets the device config policy for the manager. If
// no policy was explicitly set, this will return the default policy.
func (m *manager) GetDeviceConfigFilePolicy() ConfigPolicy {
	if m.deviceConfigFilePolicy == NoPolicy {
		for _, p := range m.policies {
			switch p {
			case DeviceConfigFileRequired, DeviceConfigFileOptional, DeviceConfigFileProhibited:
				m.deviceConfigFilePolicy = p
			}
		}
		if m.deviceConfigFilePolicy == NoPolicy {
			m.deviceConfigFilePolicy = DeviceConfigFileRequired
		}
	}
	return m.deviceConfigFilePolicy
}

// GetDeviceConfigFilePolicy gets the device config policy that was registered
// with the SDK's policy manager. If no policy was explicitly set, the default
// policy is returned.
func GetDeviceConfigFilePolicy() ConfigPolicy {
	return defaultManager.GetDeviceConfigFilePolicy()
}

// GetDeviceConfigDynamicPolicy gets the device config policy for dynamic
// registration that was registered with the manager. If no policy was
// explicitly set, this will return the default policy.
func (m *manager) GetDeviceConfigDynamicPolicy() ConfigPolicy {
	if m.deviceConfigDynamicPolicy == NoPolicy {
		for _, p := range m.policies {
			switch p {
			case DeviceConfigDynamicRequired, DeviceConfigDynamicOptional, DeviceConfigDynamicProhibited:
				m.deviceConfigDynamicPolicy = p
			}
		}
		if m.deviceConfigDynamicPolicy == NoPolicy {
			m.deviceConfigDynamicPolicy = DeviceConfigDynamicOptional
		}
	}
	return m.deviceConfigDynamicPolicy
}

// GetDeviceConfigDynamicPolicy gets the device config policy for dynamic
// registration that was registered with the SDK's policy manager. If no
// policy was explicitly set, the default policy is returned.
func GetDeviceConfigDynamicPolicy() ConfigPolicy {
	return defaultManager.GetDeviceConfigDynamicPolicy()
}

// GetTypeConfigFilePolicy gets the output type config policy for the manager.
// If no policy was explicitly set, this will return the default policy.
func (m *manager) GetTypeConfigFilePolicy() ConfigPolicy {
	if m.typeConfigFilePolicy == NoPolicy {
		for _, p := range m.policies {
			switch p {
			case TypeConfigFileOptional, TypeConfigFileRequired, TypeConfigFileProhibited:
				m.typeConfigFilePolicy = p
			}
		}
		if m.typeConfigFilePolicy == NoPolicy {
			m.typeConfigFilePolicy = TypeConfigFileOptional
		}
	}
	return m.typeConfigFilePolicy
}

// GetTypeConfigFilePolicy gets the output type config policy that was registered
// with the SDK's policy manager. If no policy was explicitly set, the default
// policy is returned.
func GetTypeConfigFilePolicy() ConfigPolicy {
	return defaultManager.GetTypeConfigFilePolicy()
}

// Check checks the policy constraint functions against the manager's set of
// tracked policies. This should be done prior to getting any policies to ensure
// that the policy set is valid to begin with.
func (m *manager) Check() error {
	err := checkConstraints(m.policies)
	if err.HasErrors() {
		log.Error("[policies] applied config policies do not pass constraint checks")
		return err
	}
	log.Debug("[policies] plugin policies pass constraint checks")
	return nil
}

// Check checks the policy constraint functions against the policies tracked by
// the SDK's policy manager. This should be done prior to getting any policies
// to ensure that the policy set is valid to begin with.
func Check() error {
	return defaultManager.Check()
}

package policies

// ConfigPolicy is a type that defines a behavior profile for the plugin
// on how it should handle configurations and configuration errors.
type ConfigPolicy uint8

const (
	// PluginConfigOptional is a policy that allows no plugin config to
	// be specified, so the plugin can just use default values. This is
	// the default policy for plugins.
	PluginConfigOptional ConfigPolicy = iota

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

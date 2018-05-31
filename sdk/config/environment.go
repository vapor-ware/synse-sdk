package config

const (
	// EnvDeviceConfig is the environment variable that can be used to
	// specify the directory which holds the device configs.
	EnvDeviceConfig = "PLUGIN_DEVICE_CONFIG"

	// EnvPluginConfig is the environment variable that can be used to
	// specify the plugin config for any non-default location. This
	// environment variable can either specify the directory containing
	// the plugin config, or it can specify the plugin config file
	// itself.
	EnvPluginConfig = "PLUGIN_CONFIG"
)

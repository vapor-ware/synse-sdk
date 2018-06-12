package sdk

const (
	// EnvDeviceConfig is the environment variable that can be used to
	// specify the directory which holds the device configs, or a single
	// file that specifies the device configs.
	EnvDeviceConfig = "PLUGIN_DEVICE_CONFIG"

	// EnvOutputTypeConfig is the environment variable that can be used to
	// specify the directory which holds the output type configs, or a single
	// file that specifies the output type configs.
	EnvOutputTypeConfig = "PLUGIN_TYPE_CONFIG"

	// EnvPluginConfig is the environment variable that can be used to
	// specify the plugin config for any non-default location. This
	// environment variable can either specify the directory containing
	// the plugin config, or it can specify the plugin config file
	// itself.
	EnvPluginConfig = "PLUGIN_CONFIG"
)

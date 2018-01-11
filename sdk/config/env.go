package config

const (
	// EnvDevicePath is the environment variable that can be used to
	// specify a non-default directory for device configs. If the
	// prototype config and device config directories are contained
	// within the same base directory, you can use the
	// PLUGIN_DEVICE_CONFIG environment variable instead.
	EnvDevicePath = "PLUGIN_DEVICE_PATH"

	// EnvProtoPath is the environment variable that can be used to
	// specify a non-default directory for protocol configs. If the
	// prototype config and device config directories are contained
	// within the same base directory, you can use the
	// PLUGIN_DEVICE_CONFIG environment variable instead.
	EnvProtoPath = "PLUGIN_PROTO_PATH"

	// EnvDeviceConfig is the environment variable that can be used to
	// specify the directory which holds both a "proto" and "device"
	// subdirectory (corresponding to the PLUGIN_PROTO_PATH and
	// PLUGIN_DEVICE_PATH, respectively).
	EnvDeviceConfig = "PLUGIN_DEVICE_CONFIG"

	// EnvPluginConfig is the environment variable that can be used to
	// specify the config directory for any non-default location.
	EnvPluginConfig = "PLUGIN_CONFIG"
)

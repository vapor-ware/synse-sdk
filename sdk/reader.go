package sdk

const (
	pluginConfigFileName = "config"
)

var (
	// pluginConfigSearchPaths define the search paths, in order of evaluation,
	// that are used when looking for the plugin configuration file.
	pluginConfigSearchPaths = []string{".", "./config", "/etc/synse/plugin/config"}

	// deviceConfigSearchPaths define the search paths, in order of evaluation,
	// that are used when looking for device configuration files.
	deviceConfigSearchPaths = []string{"./config/device", "/etc/synse/plugin/config/device"}

	// typeConfigSearchPaths define the search paths, in order of evaluation,
	// that are used when looking for output type configuration files.
	typeConfigSearchPaths = []string{"./config/type", "/etc/synse/plugin/config/type"}
)

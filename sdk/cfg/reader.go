package cfg

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
	"gopkg.in/yaml.v2"
)

var (
	// pluginConfigSearchPaths define the search paths, in order of evaluation,
	// that are used when looking for the plugin configuration file.
	pluginConfigSearchPaths = []string{".", "./config", "/etc/synse/plugin/config"}

	// deviceConfigSearchPaths define the search paths, in order of evaluation,
	// that are used when looking for device configuration files.
	deviceConfigSearchPaths = []string{"./config/device", "/etc/synse/plugin/config/device"}

	// supportedExts are the extensions supported for configuration files.
	supportedExts = []string{".yml", ".yaml"}
)

// NewPluginConfig creates a new instance of PluginConfig, populated from
// the configuration read in by Viper. This will include config options from
// the command line and from file.
func NewPluginConfig() (*PluginConfig, error) {
	// First, we setup all the lookup info for the viper instance.
	viper.SetConfigName("config")

	// Set the environment variable lookup
	viper.SetEnvPrefix("plugin")
	viper.AutomaticEnv()

	// If the PLUGIN_CONFIG environment variable is set, we will only search for
	// the config in that specified path, as we should expect the user-specified
	// value to be there. Otherwise, we will look through a set of pre-defined
	// configuration locations (in order of search):
	//  - current working directory
	//  - local config directory
	//  - the default config location in /etc
	configPath := os.Getenv(EnvPluginConfig)
	if configPath != "" {
		viper.AddConfigPath(configPath)
	} else {
		for _, path := range pluginConfigSearchPaths {
			viper.AddConfigPath(path)
		}
	}

	// Set default values for the PluginConfig
	SetDefaults()

	// will be used for the ConfigContext
	//configFile := viper.ConfigFileUsed()

	// Read in the configuration
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	config := &PluginConfig{}
	err = mapstructure.Decode(viper.AllSettings(), config)
	if err != nil {
		return nil, err
	}

	// FIXME: we should probably return a ConfigContext from this.
	return config, nil
}

// SetDefaults sets the default values for the PluginConfig via Viper.
func SetDefaults() {
	viper.SetDefault("debug", false)
	viper.SetDefault("settings.mode", modeSerial)
	viper.SetDefault("settings.read.interval", "1s")
	viper.SetDefault("settings.read.buffer", 100)
	viper.SetDefault("settings.read.enabled", true)
	viper.SetDefault("settings.write.interval", "1s")
	viper.SetDefault("settings.write.buffer", 100)
	viper.SetDefault("settings.write.max", 100)
	viper.SetDefault("settings.write.enabled", true)
	viper.SetDefault("settings.transaction.ttl", "5m")
}

// getDeviceConfigFilePaths gets the file paths for device configuration files
// by searching various config search paths.
//
// If first attempts to resolve the environment override. If present, the
// environment override will be used and the built-in search paths will be
// ignored. A failure to resolve a user-specified override should fail the
// configuration flow, so the user knows something is wrong.
//
// This does not check that any
func getDeviceConfigFilePaths() ([]string, error) { // nolint: gocyclo
	// First, we will check to see if an environment override is set.
	//
	// The environment override can specify either:
	//  - a directory which contains multiple device configuration files
	//  - the path to a single configuration file
	override := os.Getenv(EnvDevicePath)
	if override != "" {
		logger.Debugf("getting device configs from ENV override: %s", override)
		info, err := os.Stat(override)
		if err != nil {
			return nil, err
		}

		// If the environment variable specifies a directory, get all valid
		// config files from that directory. Otherwise, consider the override
		// value to be a file.
		if info.IsDir() {
			configs, err := getConfigPathsFromDir(override)
			if err != nil {
				return nil, err
			}
			// If there were no config files in the override path, return an error.
			// We expect and user-defined overrides to be correct.
			if len(configs) == 0 {
				return nil, fmt.Errorf("no valid config files found in override path: %s", override)
			}
			return configs, nil
		}

		// If we get here, the override is not a directory, so we will consider
		// it to be a file. Check that the file is valid.
		if !isValidConfig(info) {
			return nil, fmt.Errorf("environment-specified config '%s' is not a valid config file", override)
		}
		logger.Debugf("found valid config file: %s", override)
		return []string{override}, nil
	}

	// If no override is set, look through the known search paths.
	// The first search path to contain files with the supported extensions
	// will be the source of the configs. This does not guarantee that said
	// files are actually device config files though -- that will be determined
	// when marshaling the data into the appropriate structs.
	var configs []string
	var err error

	for _, path := range deviceConfigSearchPaths {
		logger.Debugf("searching for device configs in: %s", path)

		configs, err = getConfigPathsFromDir(path)
		if err != nil {
			return nil, err
		}

		if len(configs) != 0 {
			logger.Debugf("device configs found")
			break
		}
	}

	// If there are no configs after searching all paths, return an error
	if len(configs) == 0 {
		// TODO: this should probably be a specific config error so we can
		// catch it later on. When configuration policies are implemented, this
		// error might be ignored.
		return nil, fmt.Errorf("no device configuration files found")
	}
	return configs, nil
}

// getConfigPathsFromDir gets the filepaths of all the valid configuration files
// from the specified directory. Configuration files are considered valid if they
// have a supported configuration file extension.
func getConfigPathsFromDir(dirpath string) ([]string, error) {
	var files []string

	contents, err := ioutil.ReadDir(dirpath)
	if err != nil {
		return nil, err
	}

	for _, f := range contents {
		if isValidConfig(f) {
			name := filepath.Join(dirpath, f.Name())
			logger.Debugf("found valid config file: %s", name)
			files = append(files, name)
		}
	}
	return files, nil
}

func isValidConfig(f os.FileInfo) bool {
	if !f.IsDir() {
		fileExt := filepath.Ext(f.Name())
		for _, ext := range supportedExts {
			if fileExt == ext {
				return true
			}
		}
	}
	return false
}

// GetDeviceConfigsFromFile finds the files containing device configurations and
// marshals them into a DeviceConfig struct. These DeviceConfigs are wrapped in a
// ConfigContext which provides the source file for the configuration as well.
//
// All ConfigContexts returned by this function will have their IsDeviceConfig
// function return true.
func GetDeviceConfigsFromFile() ([]*ConfigContext, error) {
	var cfgs []*ConfigContext

	files, err := getDeviceConfigFilePaths()
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		config := &DeviceConfig{}
		err := UnmarshalConfigFile(file, config)
		if err != nil {
			return nil, err
		}
		cfgs = append(cfgs, NewConfigContext(file, config))
	}

	return cfgs, nil
}

// UnmarshalConfigFile unmarshals the contents of the specified file into the
// specified struct.
func UnmarshalConfigFile(filepath string, out interface{}) error {
	// Read the file contents
	contents, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	// Unmarshal into the given struct.
	// Note: Right now, we only support YAML config files. If that changes,
	// we'll need to update this to support different encodings.
	return yaml.Unmarshal(contents, out)
}

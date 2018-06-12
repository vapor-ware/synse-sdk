package sdk

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
	"gopkg.in/yaml.v2"
)

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

	// supportedExts are the extensions supported for configuration files.
	supportedExts = []string{".yml", ".yaml"}
)

// getOutputTypeConfigsFromFile finds the files containing output type configurations
// and marshals them into an OutputType struct. These OutputTypes are wrapped in a
// ConfigContext which provides the source file for the configuration as well.
//
// All ConfigContexts returned by this function will have their IsOutputTypeConfig
// function return true.
func getOutputTypeConfigsFromFile() ([]*ConfigContext, error) {
	var cfgs []*ConfigContext

	// Search for output type config files. No name is specified as an arg here because
	// output type config files do not require any particular name.
	files, err := findConfigs(typeConfigSearchPaths, EnvOutputTypeConfig, "")
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		config := &OutputType{}
		err := unmarshalConfigFile(file, config)
		if err != nil {
			return nil, fmt.Errorf("file: %s -> %s", file, err)
		}
		cfgs = append(cfgs, NewConfigContext(file, config))
	}

	return cfgs, nil
}

// getDeviceConfigsFromFile finds the files containing device configurations and
// marshals them into a DeviceConfig struct. These DeviceConfigs are wrapped in a
// ConfigContext which provides the source file for the configuration as well.
//
// All ConfigContexts returned by this function will have their IsDeviceConfig
// function return true.
func getDeviceConfigsFromFile() ([]*ConfigContext, error) {
	var cfgs []*ConfigContext

	// Search for device config files. No name is specified as an arg here because
	// device config files do not require any particular name.
	files, err := findConfigs(deviceConfigSearchPaths, EnvDeviceConfig, "")
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		config := &DeviceConfig{}
		err := unmarshalConfigFile(file, config)
		if err != nil {
			return nil, fmt.Errorf("file: %s -> %s", file, err)
		}
		cfgs = append(cfgs, NewConfigContext(file, config))
	}

	return cfgs, nil
}

// getPluginConfigFromFile finds the plugin configuration file, resolves plugin
// config defaults, and marshals the config data into a PluginConfig struct. The
// PluginConfig is wrapped in a ConfigContext which provides the source file for
// the configuration as well.
//
// The ConfigContext returned by this function will have its IsPluginConfig
// function return true.
func getPluginConfigFromFile() (*ConfigContext, error) {
	// Search for the plugin config file. It should have the name "config".
	files, err := findConfigs(pluginConfigSearchPaths, EnvPluginConfig, pluginConfigFileName)
	if err != nil {
		return nil, err
	}
	if len(files) > 1 {
		return nil, fmt.Errorf("only one plugin config should be defined, but found: %v", files)
	}

	// Resolve the defaults for the config first
	config, err := NewDefaultPluginConfig()
	if err != nil {
		return nil, err
	}
	// Unmarshal the config data
	err = unmarshalConfigFile(files[0], config)
	if err != nil {
		return nil, fmt.Errorf("file: %s -> %s", files[0], err)
	}

	return NewConfigContext(files[0], config), nil
}

// unmarshalConfigFile unmarshals the contents of the specified file into the
// specified struct.
func unmarshalConfigFile(filepath string, out interface{}) error {
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

// findConfigs gets the paths for configuration file(s) by searching through
// any environment overrides and through the specified config search paths.
//
// If first attempts to resolve the environment override. If present, the
// environment override will be used and the built-in search paths will be
// ignored. A failure to resolve a user-specified override will fail the
// configuration flow, so the user knows something is wrong.
func findConfigs(searchPaths []string, env, name string) (configs []string, err error) {
	// First, we will check to see if an environment override is set.
	//
	// The environment override can specify either:
	//  - a directory which contains multiple device configuration files
	//  - the path to a single configuration file
	configs, err = searchEnv(env, name)
	if err != nil {
		return
	}

	// If we got any configs from searching via ENV, return those, otherwise
	// we will keep looking.
	if len(configs) > 0 {
		return
	}

	// If no override is set, look through the known search paths.
	// The first search path to contain files with the supported extensions
	// will be the source of the configs. This does not guarantee that those
	// files are actually config files though -- that will be determined
	// when marshaling the data into the appropriate structs.
	for _, path := range searchPaths {
		logger.Debugf("searching for configs in: %s", path)

		configs, err = searchDir(path, name)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return
		}

		if len(configs) != 0 {
			logger.Debugf("config(s) found")
			break
		}
	}

	// If there are no configs after searching all paths, return an error
	if len(configs) == 0 {
		return nil, errors.NewConfigsNotFoundError(searchPaths)
	}
	return
}

// searchEnv searches for configuration files based on the value set for the
// specified environment variable.
//
// The value specified by the environment variable can be either a directory
// containing configuration files, or the fully qualified path to a single
// configuration file.
//
// If a "name" is passed in as a parameter, the config file will only be valid
// if it has that name and if its extension is in the list of supported extensions.
// If it is not set, it will be considered valid if its extension is in the list
// of supported extensions.
func searchEnv(env, name string) (configs []string, err error) {
	// If there is no ENV provided, there is nothing to search for.
	if env == "" {
		return
	}

	envValue := os.Getenv(env)

	// If no value is set for the env, there is nothing to search for.
	if envValue == "" {
		return
	}

	info, err := os.Stat(envValue)
	if err != nil {
		return
	}

	// If the environment variable specifies a directory, get all the valid
	// config files from that directory.
	if info.IsDir() {
		configs, err = searchDir(envValue, name)
		if err != nil {
			return
		}

		// Since the ENV is used for user-specified overrides, if we don't
		// find anything here, we will return an error.
		if len(configs) == 0 {
			return configs, errors.NewConfigsNotFoundError([]string{envValue})
		}
		return
	}

	// Otherwise, the environment variable specifies a file. Check that
	// the file is valid.
	if !isValidConfig(info, name) {
		return configs, fmt.Errorf("environment-specified config '%s' is not a valid config file", envValue)
	}

	logger.Debugf("found valid config file: %s", envValue)
	configs = append(configs, envValue)
	return
}

// searchDir gets the filepaths of all the valid configuration files
// from the specified directory. Configuration files are considered valid if they
// have a supported configuration file extension.
func searchDir(dirpath, name string) ([]string, error) {
	var files []string

	contents, err := ioutil.ReadDir(dirpath)
	if err != nil {
		return nil, err
	}

	for _, f := range contents {
		if isValidConfig(f, name) {
			name := filepath.Join(dirpath, f.Name())
			logger.Debugf("found valid config file: %s", name)
			files = append(files, name)
		}
	}
	return files, nil
}

// isValidConfig checks if the given FileInfo corresponds to a file that could be
// a valid configuration file. It checks that it is actually a file (not a Dir)
// and checks that its extension matches the supported extensions.
func isValidConfig(f os.FileInfo, name string) bool {
	if !f.IsDir() {
		fileExt := filepath.Ext(f.Name())

		// If a file name was give, check that the file matches that name
		if name != "" {
			fileName := strings.TrimRight(f.Name(), fileExt)
			if fileName != name {
				return false
			}
		}

		// Check if the extension is supported
		for _, ext := range supportedExts {
			if fileExt == ext {
				return true
			}
		}
	}
	return false
}

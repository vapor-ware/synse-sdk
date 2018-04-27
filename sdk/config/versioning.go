package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	"github.com/vapor-ware/synse-sdk/sdk/logger"
)

// TODO - add in support for deprecated versions. this doesn't need to go
// in immediately, since we have no versions to deprecate initially.

// Versions
var (
	// version "1" or "1.0"
	v1maj0min = configVersion{Major: 1, Minor: 0}
)

// Common Configuration Versioning
// -------------------------------

// configVersion represents the version found in a configuration file.
type configVersion struct {
	Major int
	Minor int

	cfgFile string
}

// ToString converts the ConfigVersion to a version string.
func (v *configVersion) ToString() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

// getConfigVersion gets the version of the specified configuration file.
func getConfigVersion(path string, data []byte) (*configVersion, error) {
	var version cfgVersion
	version.cfgFile = path
	err := yaml.Unmarshal(data, &version)
	if err != nil {
		logger.Errorf("Failed to Unmarshal config data for %s: %v", path, err)
		return nil, err
	}
	v, err := version.toConfigVersion()
	if err != nil {
		logger.Errorf("Failed to create new configVersion for %s: %v", path, err)
		return nil, err
	}
	return v, nil
}

// isSupportedVersion checks whether the given ConfigVersion is in the slice
// of supported versions.
func isSupportedVersion(cfg *configVersion, supported []string) bool {
	isSupported := false
	for _, version := range supported {
		if cfg.ToString() == version {
			isSupported = true
			break
		}
	}
	return isSupported
}

// cfgVersion is an intermediary struct used to pull out the version
// information from a configuration file.
type cfgVersion struct {
	Version string `yaml:"version"`

	cfgFile string
}

// toConfigVersion converts the cfgVersion struct into a corresponding
// configVersion representation.
func (v *cfgVersion) toConfigVersion() (*configVersion, error) {
	var min, maj int
	var err error

	if v.Version == "" {
		return nil, fmt.Errorf("no version info found in config file (%v)", v.cfgFile)
	}

	s := strings.Split(v.Version, ".")
	if len(s) == 1 {
		maj, err = strconv.Atoi(s[0])
		if err != nil {
			return nil, err
		}
		min = 0
	} else {
		maj, err = strconv.Atoi(s[0])
		if err != nil {
			return nil, err
		}
		min, err = strconv.Atoi(s[1])
		if err != nil {
			return nil, err
		}
	}
	return &configVersion{
		Major:   maj,
		Minor:   min,
		cfgFile: v.cfgFile,
	}, nil
}

// Device Configuration Versioning
// -------------------------------

// deviceConfigVersionHandler defines an interface that all versions of the
// configuration will need to implement, which specifies how to parse the
// configuration for that given version.
type deviceConfigVersionHandler interface {
	processPrototypeConfig([]byte) ([]*PrototypeConfig, error)
	processDeviceConfig([]byte) ([]*DeviceConfig, error)
}

// deviceConfigHandler defines which device config versions are supported as
// well as the config handlers for each of those supported versions.
var deviceConfigHandlers = map[string]deviceConfigVersionHandler{
	// versions: "1", "1.0"
	v1maj0min.ToString(): &v1DeviceConfigHandler{},
}

// supportedDeviceConfigVersions defines the collection of versions which the
// current version of the SDK supports for device instance/prototype configuration
// files.
var supportedDeviceConfigVersions = func() []string {
	s := make([]string, len(deviceConfigHandlers))
	i := 0
	for k := range deviceConfigHandlers {
		s[i] = k
		i++
	}
	return s
}()

// getDeviceConfigVersionHandler gets the handler for the given device
// configuration version.
func getDeviceConfigVersionHandler(cv *configVersion) (deviceConfigVersionHandler, error) {
	if !isSupportedVersion(cv, supportedDeviceConfigVersions) {
		return nil, fmt.Errorf("config version '%s' not supported (%s)", cv.ToString(), cv.cfgFile)
	}
	h := deviceConfigHandlers[cv.ToString()]
	if h == nil {
		return nil, fmt.Errorf("no handler defined for config version '%s' (%s)", cv.ToString(), cv.cfgFile)
	}
	return h, nil
}

// Plugin Configuration Versioning
// -------------------------------

// pluginConfigVersionHandler defines an interface that all versions of the
// configuration will need to implement, which specifies how to parse the
// configuration for that given version.
type pluginConfigVersionHandler interface {
	processPluginConfig(v *viper.Viper) (*PluginConfig, error)
}

// pluginConfigHandler defines which plugin config versions are supported as
// well as the config handlers for each of those supported versions.
var pluginConfigHandlers = map[string]pluginConfigVersionHandler{
	// versions: "1", "1.0"
	v1maj0min.ToString(): &v1PluginConfigHandler{},
}

// supportedPluginConfigVersions defines the collection of versions which the
// current version of the SDK supports for plugin configuration files.
var supportedPluginConfigVersions = func() []string {
	s := make([]string, len(pluginConfigHandlers))
	i := 0
	for k := range pluginConfigHandlers {
		s[i] = k
		i++
	}
	return s
}()

// getPluginConfigVersionHandler gets the handler for the given plugin
// configuration version.
func getPluginConfigVersionHandler(cv *configVersion) (pluginConfigVersionHandler, error) {
	if !isSupportedVersion(cv, supportedPluginConfigVersions) {
		return nil, fmt.Errorf("config version '%s' not supported (%s)", cv.ToString(), cv.cfgFile)
	}
	h := pluginConfigHandlers[cv.ToString()]
	if h == nil {
		return nil, fmt.Errorf("no handler defined for config version '%s' (%s)", cv.ToString(), cv.cfgFile)
	}
	return h, nil
}

// parseVersionedPluginConfig takes a Viper instance and reads in the Plugin configuration
// with it. If successful, it will check the version field in the config and parse the
// configuration appropriately based on the version number.
func parseVersionedPluginConfig(v *viper.Viper) (*PluginConfig, error) {

	configFile := v.ConfigFileUsed()

	// Read in the configuration file
	err := v.ReadInConfig()
	if err != nil {
		logger.Errorf("Failed to read in plugin config %s: %v", configFile, err)
		return nil, err
	}

	// Get the plugin configuration version
	cv := cfgVersion{v.GetString("version"), configFile}
	version, err := cv.toConfigVersion()
	if err != nil {
		logger.Errorf("Failed to get config version from %s: %v", configFile, err)
		return nil, err
	}

	// Get the handler for the given configuration version.
	cfgHandler, err := getPluginConfigVersionHandler(version)
	if err != nil {
		logger.Errorf("Failed to get the plugin config version handler for %s: %v", configFile, err)
		return nil, err
	}

	// Parse the config with the versioned handler
	c, err := cfgHandler.processPluginConfig(v)
	if err != nil {
		logger.Errorf("Failed to parse plugin config %s: %v", configFile, err)
		return nil, err
	}
	return c, nil
}

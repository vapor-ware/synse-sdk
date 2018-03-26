package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/vapor-ware/synse-sdk/sdk/logger"
	"github.com/vapor-ware/synse-server-grpc/go"
)

const (
	defaultDeviceConfig = "/etc/synse/plugin/config/device"
)

// DeviceConfig represents a single device instance.
type DeviceConfig struct {
	Version  string
	Type     string
	Model    string
	Location Location
	Data     map[string]string
}

// Location represents a location of a device instance.
type Location struct {
	Rack  interface{} `yaml:"rack"`
	Board string      `yaml:"board"`

	// internal field to hold the resolved rack value
	rack string
}

// GetRack gets the resolved string value of the Rack field of the
// Location. If the Rack field cannot resolve, an error is returned.
func (l *Location) GetRack() (string, error) {
	if l.rack == "" {
		err := l.Validate()
		if err != nil {
			logger.Errorf("Failed to validate Location data for %v: %v", l, err)
			return "", err
		}
	}
	return l.rack, nil
}

// Validate validates the contents of the Rack interface, and if it
// is correct, it will populate the internal `rack` field with that
// value. Otherwise, it will return an error.
func (l *Location) Validate() error {
	switch l.Rack.(type) {
	// In this case, this is just the Rack defined directly.
	case string:
		logger.Debugf("Location.Rack is a string")
		l.rack = l.Rack.(string)
		return nil

	// In this case, we have a map - the only key we expect are the
	// string "from_env" with a string value.
	case map[interface{}]interface{}:
		logger.Debugf("Location.Rack is a map[interface{}]interface{}")
		stringMap := make(map[string]string)
		for k, v := range l.Rack.(map[interface{}]interface{}) {
			keyString, ok := k.(string)
			if !ok {
				return fmt.Errorf("location rack map key is not a string: %v", k)
			}
			valueString, ok := v.(string)
			if !ok {
				return fmt.Errorf("location rack map value is not a string: %v", v)
			}
			stringMap[keyString] = valueString
		}

		// Check for the "from_env" key
		fromEnv := stringMap["from_env"]
		if fromEnv == "" {
			return fmt.Errorf("location rack is a map, but no supported keys were found in it (supported keys: 'from_env')")
		}
		envValue, ok := os.LookupEnv(fromEnv)
		if !ok {
			return fmt.Errorf("location rack set to use key %v, but no env value found", fromEnv)
		}
		l.rack = envValue
		return nil

	default:
		return fmt.Errorf("failed to resolve location rack (type: %T): %v", l.Rack, l.Rack)
	}
}

// Encode translates the Location to a corresponding gRPC MetaLocation.
func (l *Location) Encode() *synse.MetaLocation {
	return &synse.MetaLocation{
		Rack:  l.rack,
		Board: l.Board,
	}
}

// ParseDeviceConfig parses the YAML files found in the device instance
// configuration directory, if any are found, into DeviceConfig structs.
func ParseDeviceConfig() ([]*DeviceConfig, error) {
	logger.Debug("ParseDeviceConfig start")
	var cfgs []*DeviceConfig

	path := os.Getenv(EnvDeviceConfig)
	if path != "" {
		path = filepath.Join(path, "device")
	} else {
		path = os.Getenv(EnvDevicePath)
		if path == "" {
			path = defaultDeviceConfig
		}
	}
	logger.Debugf("Searching %s for device configurations.", path)

	_, err := os.Stat(path)
	if err != nil {
		logger.Errorf("Device configuration path %s does not exist: %v", path, err)
		return nil, err
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		logger.Errorf("Failed to read device configuration directory %s: %v", path, err)
		return nil, err
	}

	logger.Debugf("Found configuration files: %v", files)
	for _, f := range files {
		if isValidConfig(f) {
			fpath := filepath.Join(path, f.Name())
			logger.Debugf("Reading file: %s", fpath)

			// Get the file contents
			yml, err := ioutil.ReadFile(fpath)
			if err != nil {
				logger.Errorf("Failed to read file %s: %v", fpath, err)
				return nil, err
			}

			// Get the version of the configuration file
			ver, err := getConfigVersion(fpath, yml)
			if err != nil {
				logger.Errorf("Failed to get configuration version for file %s: %v", fpath, err)
				return nil, err
			}
			logger.Debugf("Got device configuration file version: %s", ver)

			// Get the handler for the given configuration version
			cfgHandler, err := getDeviceConfigVersionHandler(ver)
			if err != nil {
				logger.Errorf("Failed to get handler for device config version %s: %v", ver.ToString(), err)
				return nil, err
			}
			logger.Debugf("Got device configuration handler: %v", cfgHandler)

			// Process the configuration files with the specific handler
			// for the version of that config file.
			c, err := cfgHandler.processDeviceConfig(yml)
			if err != nil {
				logger.Errorf("Failed to process device configuration for %s: %v", fpath, err)
				return nil, err
			}
			logger.Debugf("Successfully processed device configuration in %s", fpath)

			// Add the parsed configurations to the tracked device configs.
			cfgs = append(cfgs, c...)

		} else {
			logger.Debugf("%s is not a valid config -- skipping.", filepath.Join(path, f.Name()))
		}
	}

	logger.Debugf("Finished parsing device configurations: %v", cfgs)
	return cfgs, nil
}

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
		l.rack = l.Rack.(string)
		return nil

	// In this case, we have a map - the only key we expect are the
	// string "from_env" with a string value.
	case map[interface{}]interface{}:
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
			return fmt.Errorf("location rack is a map, but no supported keys were found in it")
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

	logger.Debug("ParseDeviceConfig 2")
	_, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	logger.Debug("ParseDeviceConfig 3")
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	logger.Debugf("ParseDeviceConfig 4. files: %v", files)
	for _, f := range files {
		logger.Debug("ParseDeviceConfig 5")
		if isValidConfig(f) {
			// Get the file contents
			fpath := filepath.Join(path, f.Name())
			logger.Debugf("ParseDeviceConfig reading file %v", fpath)
			yml, err := ioutil.ReadFile(fpath)
			if err != nil {
				return nil, err
			}

			// Get the version of the configuration file
			logger.Debug("ParseDeviceConfig 6")
			ver, err := getConfigVersion(fpath, yml)
			if err != nil {
				return nil, err
			}

			// Get the handler for the given configuration version
			logger.Debug("ParseDeviceConfig 7")
			cfgHandler, err := getDeviceConfigVersionHandler(ver)
			if err != nil {
				return nil, err
			}

			// Process the configuration files with the specific handler
			// for the version of that config file.
			logger.Debug("ParseDeviceConfig 8")
			c, err := cfgHandler.processDeviceConfig(yml)
			if err != nil {
				return nil, err
			}

			logger.Debug("ParseDeviceConfig 9")
			cfgs = append(cfgs, c...)
		}
	}

	logger.Debug("ParseDeviceConfig 10")
	return cfgs, nil
}

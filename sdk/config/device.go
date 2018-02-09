package config

import (
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
	Rack  string `yaml:"rack"`
	Board string `yaml:"board"`
}

// Encode translates the Location to a corresponding gRPC MetaLocation.
func (l *Location) Encode() *synse.MetaLocation {
	return &synse.MetaLocation{
		Rack:  l.Rack,
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

	logger.Debugf("ParseDeviceConfig 3. path: %v", path)
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

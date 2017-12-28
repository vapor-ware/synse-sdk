package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/vapor-ware/synse-server-grpc/go"
)

const (
	defaultDeviceConfig = "/etc/synse/plugin/config/device"
)

const (
	// EnvDevicePath is the environment variable that can be used to
	// specify a non-default directory for device configs.
	EnvDevicePath = "PLUGIN_DEVICE_PATH"
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
	var cfgs []*DeviceConfig

	path := os.Getenv(EnvDevicePath)
	if path == "" {
		path = defaultDeviceConfig
	}

	_, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		if isValidConfig(f) {
			// Get the file contents
			fpath := filepath.Join(path, f.Name())
			yml, err := ioutil.ReadFile(fpath)
			if err != nil {
				return nil, err
			}

			// Get the version of the configuration file
			ver, err := getConfigVersion(yml)
			if err != nil {
				return nil, err
			}

			// Get the handler for the given configuration version
			cfgHandler, err := getDeviceConfigVersionHandler(ver)
			if err != nil {
				return nil, err
			}

			// Process the configuration files with the specific handler
			// for the version of that config file.
			c, err := cfgHandler.processDeviceConfig(yml)
			if err != nil {
				return nil, err
			}

			cfgs = append(cfgs, c...)
		}
	}
	return cfgs, nil
}

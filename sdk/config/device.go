package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/vapor-ware/synse-server-grpc/go"
	"gopkg.in/yaml.v2"
)

const (
	defaultDeviceConfig = "/etc/synse/plugin/config/device"
)

const (
	// EnvDevicePath is the environment variable that can be used to
	// specify a non-default directory for device configs.
	EnvDevicePath = "PLUGIN_DEVICE_PATH"
)

type v1deviceConfig struct {
	Version   string              `yaml:"version"`
	Locations map[string]Location `yaml:"locations"`
	Devices   []v1device          `yaml:"devices"`
}

type v1device struct {
	Type      string              `yaml:"type"`
	Model     string              `yaml:"model"`
	Instances []map[string]string `yaml:"instances"`
}

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
			var scheme v1deviceConfig

			fpath := filepath.Join(path, f.Name())
			yml, err := ioutil.ReadFile(fpath)
			if err != nil {
				return nil, err
			}
			err = yaml.Unmarshal(yml, &scheme)
			if err != nil {
				return nil, err
			}

			for _, device := range scheme.Devices {
				for _, i := range device.Instances {
					locationTag := i["location"]
					if locationTag == "" {
						return nil, fmt.Errorf("no location defined for device: %#v", device)
					}
					location := scheme.Locations[locationTag]

					cfg := DeviceConfig{
						Version:  scheme.Version,
						Type:     device.Type,
						Model:    device.Model,
						Location: location,
						Data:     i,
					}
					cfgs = append(cfgs, &cfg)
				}
			}
		}
	}
	return cfgs, nil
}

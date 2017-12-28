package config

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

// V1 Device Prototype
// -------------------

type v1protoConfig struct {
	Version    string            `yaml:"version"`
	Prototypes []PrototypeConfig `yaml:"prototypes"`
}

// V1 Device Instance
// ------------------

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

// V1 Device Config Handler
// ------------------------

type v1DeviceConfigHandler struct{}

func (h *v1DeviceConfigHandler) processPrototypeConfig(yml []byte) ([]*PrototypeConfig, error) {
	var cfgs []*PrototypeConfig
	var scheme v1protoConfig

	err := yaml.Unmarshal(yml, &scheme)
	if err != nil {
		return nil, err
	}

	for _, p := range scheme.Prototypes {
		p.Version = scheme.Version
		cfgs = append(cfgs, &p)
	}
	return cfgs, nil
}

func (h *v1DeviceConfigHandler) processDeviceConfig(yml []byte) ([]*DeviceConfig, error) {
	var cfgs []*DeviceConfig
	var scheme v1deviceConfig

	err := yaml.Unmarshal(yml, &scheme)
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
	return cfgs, nil
}

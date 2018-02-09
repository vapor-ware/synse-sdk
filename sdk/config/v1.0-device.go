package config

import (
	"fmt"

	"github.com/vapor-ware/synse-sdk/sdk/logger"
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

	if scheme.Prototypes == nil {
		return nil, fmt.Errorf("no prototypes defined in proto config")
	}

	for _, p := range scheme.Prototypes {
		p.Version = scheme.Version
		cfgs = append(cfgs, &p)
	}
	return cfgs, nil
}

// This function parses out the devices from the device configuration in the yaml file.
func (h *v1DeviceConfigHandler) processDeviceConfig(yml []byte) ([]*DeviceConfig, error) {
	// TODO: Fix these traces. Get something.
	//logger.Debugf("processDeviceConfig start. yml: %+v", yml)

	// Logging the yml is a bit tricky since logrus emits a literal "\n" character
	// rather than a newline. We want something that is human readable.
	// TODO: Put this in the logger file.
	ymlString := string(yml[:])
	//ymlString = strings.Replace(ymlString, "\\n", "\n", -1)
	logger.InfoMultiline(ymlString)

	//logger.Debugf("processDeviceConfig start. yml: %+v", string(yml[:])) // had \n for newlines.
	//logger.Debugf("processDeviceConfig start. yml: %v", string(yml[:]))
	logger.Debugf("processDeviceConfig start. yml: %+v", ymlString) // had \n for newlines.
	logger.Debugf("processDeviceConfig start. yml: %v", ymlString)
	var cfgs []*DeviceConfig
	var scheme v1deviceConfig

	err := yaml.Unmarshal(yml, &scheme)
	if err != nil {
		logger.Error("Failed to unmarshal yaml. %v", err)
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
			logger.Debugf("processDeviceConfig adding cfg: %+v", cfg)
			cfgs = append(cfgs, &cfg)
		}
	}
	logger.Debugf("processDeviceConfig returning: %+v", cfgs)
	return cfgs, nil
}

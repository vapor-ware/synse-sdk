package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/vapor-ware/synse-sdk/sdk/logger"
	"github.com/vapor-ware/synse-server-grpc/go"
)

const (
	defaultProtoConfig = "/etc/synse/plugin/config/proto"
)

// PrototypeConfig represents the configuration for a device prototype.
type PrototypeConfig struct {
	Version      string
	Type         string         `yaml:"type"`
	Model        string         `yaml:"model"`
	Manufacturer string         `yaml:"manufacturer"`
	Protocol     string         `yaml:"protocol"`
	Output       []DeviceOutput `yaml:"output"`
}

// DeviceOutput represents a reading output for a device.
type DeviceOutput struct {
	Type      string `yaml:"type"`
	DataType  string `yaml:"data_type"`
	Precision int32  `yaml:"precision"`
	Unit      *Unit  `yaml:"unit"`
	Range     *Range `yaml:"range"`
}

// Encode translates the DeviceOutput to a corresponding gRPC MetaOutput.
func (o *DeviceOutput) Encode() *synse.MetaOutput {
	unit := &Unit{}
	if o.Unit != nil {
		unit = o.Unit
	}

	rang := &Range{}
	if o.Range != nil {
		rang = o.Range
	}

	return &synse.MetaOutput{
		Type:      o.Type,
		DataType:  o.DataType,
		Precision: o.Precision,
		Unit:      unit.Encode(),
		Range:     rang.Encode(),
	}
}

// Unit describes the unit of measure for a device output.
type Unit struct {
	Name   string `yaml:"name"`
	Symbol string `yaml:"symbol"`
}

// Encode translates the Unit to a corresponding gRPC MetaOutputUnit.
func (u *Unit) Encode() *synse.MetaOutputUnit {
	return &synse.MetaOutputUnit{
		Name:   u.Name,
		Symbol: u.Symbol,
	}
}

// Range describes the minimum and maximum allowable numerical values for a reading.
type Range struct {
	Min int32 `yaml:"min"`
	Max int32 `yaml:"max"`
}

// Encode translates the Range to a corresponding gRPC MetaOutputRange.
func (r *Range) Encode() *synse.MetaOutputRange {
	return &synse.MetaOutputRange{
		Min: r.Min,
		Max: r.Max,
	}
}

// ParsePrototypeConfig parses the YAML files found in the prototype configuration
// directory, if any are found, into PrototypeConfig structs.
func ParsePrototypeConfig() ([]*PrototypeConfig, error) {
	logger.Debugf("ParsePrototypeConfig start")
	var cfgs []*PrototypeConfig

	path := os.Getenv(EnvDeviceConfig)
	if path != "" {
		path = filepath.Join(path, "proto")
	} else {
		path = os.Getenv(EnvProtoPath)
		if path == "" {
			path = defaultProtoConfig
		}
	}
	logger.Debugf("Searching %s for prototype configurations.", path)

	_, err := os.Stat(path)
	if err != nil {
		logger.Errorf("Device prototype configuration path %s does not exist: %v", path, err)
		return nil, err
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		logger.Errorf("Failed to read prototype configuration directory %s: %v", path, err)
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
			logger.Debugf("Got prototype configuration file version: %s", ver)

			// Get the handler for the given configuration version
			cfgHandler, err := getDeviceConfigVersionHandler(ver)
			if err != nil {
				logger.Errorf("Failed to get handler for device config version %s: %v", ver.ToString(), err)
				return nil, err
			}
			logger.Debugf("Got prototype configuration handler: %v", cfgHandler)

			// Process the configuration files with the specific handler
			// for the version of that config file.
			c, err := cfgHandler.processPrototypeConfig(yml)
			if err != nil {
				logger.Errorf("Failed to process prototype configuration for %s: %v", fpath, err)
				return nil, err
			}
			logger.Debugf("Successfully processed prototype configuration in %s", fpath)

			// Add the parsed configurations to the tracked prototype configs.
			cfgs = append(cfgs, c...)

		} else {
			logger.Debugf("%s is not a valid config -- skipping", filepath.Join(path, f.Name()))
		}
	}

	logger.Debugf("Finised parsing prototype configurations: %v", cfgs)
	return cfgs, nil
}

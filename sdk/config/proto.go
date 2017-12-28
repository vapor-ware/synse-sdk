package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/vapor-ware/synse-server-grpc/go"
)

const (
	defaultProtoConfig = "/etc/synse/plugin/config/proto"
)

const (
	// EnvProtoPath is the environment variable that can be used to
	// specify a non-default directory for protocol configs.
	EnvProtoPath = "PLUGIN_PROTO_PATH"
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
	var cfgs []*PrototypeConfig

	path := os.Getenv(EnvProtoPath)
	if path == "" {
		path = defaultProtoConfig
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
			c, err := cfgHandler.processPrototypeConfig(yml)
			if err != nil {
				return nil, err
			}

			cfgs = append(cfgs, c...)
		}
	}
	return cfgs, nil
}

package sdk

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/vapor-ware/synse-server-grpc/go"
	"gopkg.in/yaml.v2"
)

// PrototypeConfig represents the configuration for a device prototype.
type PrototypeConfig struct {
	Version      string         `yaml:"version"`
	Type         string         `yaml:"type"`
	Model        string         `yaml:"model"`
	Manufacturer string         `yaml:"manufacturer"`
	Protocol     string         `yaml:"protocol"`
	Output       []DeviceOutput `yaml:"output"`
}

// DeviceOutput represents the reading output for a configured device.
type DeviceOutput struct {
	Type      string       `yaml:"type"`
	DataType  string       `yaml:"data_type"`
	Unit      *OutputUnit  `yaml:"unit"`
	Precision int32        `yaml:"precision"`
	Range     *OutputRange `yaml:"range"`
}

// encode translates the DeviceOutput to a corresponding gRPC MetaOutput.
func (o *DeviceOutput) encode() *synse.MetaOutput {

	unit := &OutputUnit{}
	if o.Unit != nil {
		unit = o.Unit
	}

	rang := &OutputRange{}
	if o.Range != nil {
		rang = o.Range
	}

	return &synse.MetaOutput{
		Type:      o.Type,
		DataType:  o.DataType,
		Precision: o.Precision,
		Unit:      unit.encode(),
		Range:     rang.encode(),
	}
}

// OutputUnit describes the unit of measure for a device output.
type OutputUnit struct {
	Name   string `yaml:"name"`
	Symbol string `yaml:"symbol"`
}

// encode translates the OutputUnit to a corresponding gRPC MetaOutputUnit.
func (u *OutputUnit) encode() *synse.MetaOutputUnit {
	return &synse.MetaOutputUnit{
		Name:   u.Name,
		Symbol: u.Symbol,
	}
}

// OutputRange describes the min and max valid numerical values for a reading.
type OutputRange struct {
	Min int32 `yaml:"min"`
	Max int32 `yaml:"max"`
}

// encode translates the OutputRange to a corresponding gRPC MetaOutputRange.
func (r *OutputRange) encode() *synse.MetaOutputRange {
	return &synse.MetaOutputRange{
		Min: r.Min,
		Max: r.Max,
	}
}

// parsePrototypeConfig searches the configuration directory for device
// prototype configuration files. If it finds any, it reads them and populates
// PrototypeConfig structs for each of the device prototypes.
func parsePrototypeConfig(dir string) ([]*PrototypeConfig, error) {

	var protos []*PrototypeConfig
	protoPath := filepath.Join(dir, "proto")

	_, err := os.Stat(protoPath)
	if err != nil {
		Logger.Error("Unable to find prototype config directory.")
		return protos, err
	}

	files, err := ioutil.ReadDir(protoPath)
	if err != nil {
		Logger.Error("Unable to read files in prototype config directory.")
		return protos, err
	}

	for _, f := range files {
		var protoCfg PrototypeConfig

		path := filepath.Join(protoPath, f.Name())
		yamlFile, err := ioutil.ReadFile(path)
		if err != nil {
			Logger.Errorf("Could not read config file %v.", f.Name())
			return protos, err
		}

		err = yaml.Unmarshal(yamlFile, &protoCfg)
		if err != nil {
			Logger.Errorf("Failed to parse YAML from %v.", path)
			return protos, err
		}

		protos = append(protos, &protoCfg)
	}
	return protos, nil
}

// InstanceConfig represents the configuration for a device instance.
type InstanceConfig struct {
	Version   string                    `yaml:"version"`
	Type      string                    `yaml:"type"`
	Model     string                    `yaml:"model"`
	Locations map[string]DeviceLocation `yaml:"locations"`
	Devices   []map[string]string       `yaml:"devices"`
}

// DeviceLocation represents the location of a device instance.
type DeviceLocation struct {
	Rack  string `yaml:"rack"`
	Board string `yaml:"board"`
}

// encode translates the DeviceLocation to a corresponding gRPC MetaLocation.
func (l *DeviceLocation) encode() *synse.MetaLocation {
	return &synse.MetaLocation{
		Rack:  l.Rack,
		Board: l.Board,
	}
}

// DeviceConfig represents a single device instance. It is essentially the
// same as the InstanceConfig except that it represents a single element from
// its Devices field and has its location resolved.
type DeviceConfig struct {
	Version  string
	Type     string
	Model    string
	Location DeviceLocation
	Data     map[string]string
}

// parseDeviceConfig searches the configuration directory for device
// instance configuration files. If it finds any, it reads them and populates
// DeviceConfig structs for each of the device instances.
func parseDeviceConfig(dir string) ([]*DeviceConfig, error) {

	var devices []*DeviceConfig
	devicePath := filepath.Join(dir, "device")

	_, err := os.Stat(devicePath)
	if err != nil {
		Logger.Error("Unable to find device config directory.")
		return devices, err
	}

	files, err := ioutil.ReadDir(devicePath)
	if err != nil {
		Logger.Error("Unable to read files in device config directory.")
		return devices, err
	}

	for _, f := range files {
		var instanceCfg InstanceConfig

		path := filepath.Join(devicePath, f.Name())
		yamlFile, err := ioutil.ReadFile(path)
		if err != nil {
			Logger.Errorf("Could not read config file %v.", f.Name())
			return devices, err
		}

		err = yaml.Unmarshal(yamlFile, &instanceCfg)
		if err != nil {
			Logger.Errorf("Failed to parse YAML from %v.", path)
			return devices, err
		}

		for _, data := range instanceCfg.Devices {
			loc := data["location"]
			if loc == "" {
				Logger.Errorf("No location defined for device in %v.", f.Name())
				return devices, errors.New("no location defined for device")
			}
			location := instanceCfg.Locations[loc]

			deviceCfg := DeviceConfig{
				Version:  instanceCfg.Version,
				Type:     instanceCfg.Type,
				Model:    instanceCfg.Model,
				Location: location,
				Data:     data,
			}

			devices = append(devices, &deviceCfg)

		}
	}
	return devices, nil
}

package sdk

import (
	"path/filepath"
	"os"
	"fmt"
	"io/ioutil"
	"gopkg.in/yaml.v2"

	"./pb"
)

// version: 1.0
// type: emulated-temperature
// model: emul8-temp
// manufacturer: vaporio
// protocol: emulator
// output:
//   - type: temperature
//     unit:
//       name: celsius
//       symbol: C
//     precision: 2
//     range:
//       min: 0
//       max: 100


type PrototypeConfig struct {
	Version       string          `yaml:"version"`
	Type          string          `yaml:"type"`
	Model         string          `yaml:"model"`
	Manufacturer  string          `yaml:"manufacturer"`
	Protocol      string          `yaml:"protocol"`
	Output        []DeviceOutput  `yaml:"output"`
}

type DeviceOutput struct {
	Type       string       `yaml:"type"`
	Unit       *OutputUnit   `yaml:"unit"`
	Precision  int32        `yaml:"precision"`
	Range      *OutputRange  `yaml:"range"`
}

func (o *DeviceOutput) ToMetaOutput() *pb.MetaOutput {

	unit := &OutputUnit{}
	if o.Unit != nil {
		unit = o.Unit
	}

	rang := &OutputRange{}
	if o.Range != nil {
		rang = o.Range
	}

	return &pb.MetaOutput{
		Type: o.Type,
		Precision: o.Precision,
		Unit: unit.ToMetaOutputUnit(),
		Range: rang.ToMetaOutputRange(),
	}
}

type OutputUnit struct {
	Name    string  `yaml:"name"`
	Symbol  string  `yaml:"symbol"`
}

func (u *OutputUnit) ToMetaOutputUnit() *pb.MetaOutputUnit {
	return &pb.MetaOutputUnit{
		Name: u.Name,
		Symbol: u.Symbol,
	}
}

type OutputRange struct {
	Min  int32  `yaml:"min"`
	Max  int32  `yaml:"max"`
}

func (r *OutputRange) ToMetaOutputRange() *pb.MetaOutputRange {
	return &pb.MetaOutputRange{
		Min: r.Min,
		Max: r.Max,
	}
}


func ParsePrototypeConfig(dir string) ([]PrototypeConfig, error) {

	var protos []PrototypeConfig
	protoPath := filepath.Join(dir, "proto")

	_, err := os.Stat(protoPath)
	if err != nil {
		fmt.Printf("Error: Unable to find prototype config directory.\n")
		return protos, err
	}

	files, err := ioutil.ReadDir(protoPath)
	if err != nil {
		fmt.Printf("Error: Unable to read files in prototype config directory.\n")
		return protos, err
	}

	for _, f := range files {
		var protoCfg PrototypeConfig

		yamlFile, err := ioutil.ReadFile(filepath.Join(protoPath, f.Name()))
		if err != nil {
			fmt.Printf("Error: Could not read file %v\n", f.Name())
			return protos, err
		}

		err = yaml.Unmarshal(yamlFile, &protoCfg)
		if err != nil {
			fmt.Printf("Error: Failed to parse yaml from %v\n", f.Name())
			return protos, err
		}

		protos = append(protos, protoCfg)
	}
	return protos, nil
}



// version: 1.0
// type: emulated-temperature
// model: emul8-temp
//
// locations:
//   unknown:
//     rack: unknown
//     board: unknown
//
// devices:
//   - id: 1
//     location: unknown
//     comment: first emulated temperature device
//     info: CEC temp 1
//   - id: 2
//     location: unknown
//     comment: second emulated temperature device
//     info: CEC temp 2
//   - id: 3
//     location: unknown
//     comment: third emulated temperature device
//     info: CEC temp 3

type InstanceConfig struct {
	Version string `yaml:"version"`
	Type string `yaml:"type"`
	Model string `yaml:"model"`
	Locations map[string]DeviceLocation `yaml:"locations"`
	Devices []map[string]string `yaml:"devices"`
}

type DeviceLocation struct {
	Rack string `yaml:"rack"`
	Board string `yaml:"board"`
}

func (l *DeviceLocation) ToMetalLocation() *pb.MetaLocation {
	return &pb.MetaLocation{
		Rack: l.Rack,
		Board: l.Board,
	}
}


type DeviceConfig struct {
	Version string
	Type string
	Model string
	Location DeviceLocation
	Data map[string]string
}


func ParseDeviceConfig(dir string) ([]DeviceConfig, error) {

	var devices []DeviceConfig
	devicePath := filepath.Join(dir, "device")

	_, err := os.Stat(devicePath)
	if err != nil {
		fmt.Printf("Error: Unable to find device config directory.\n")
		return devices, err
	}

	files, err := ioutil.ReadDir(devicePath)
	if err != nil {
		fmt.Printf("Error: Unable to read files in device config directory.\n")
		return devices, err
	}

	for _, f := range files {
		var instanceCfg InstanceConfig

		yamlFile, err := ioutil.ReadFile(filepath.Join(devicePath, f.Name()))
		if err != nil {
			fmt.Printf("Error: Could not read file %v\n", f.Name())
			return devices, err
		}

		err = yaml.Unmarshal(yamlFile, &instanceCfg)
		if err != nil {
			fmt.Printf("Error: Failed to parse yaml from %v\n", f.Name())
			return devices, err
		}

		for _, data := range instanceCfg.Devices {
			loc := data["location"]
			if loc == "" {
				// FIXME: figure out what to do here. error out?
				fmt.Printf("Error: No location defined for device.\n")
			}
			location := instanceCfg.Locations[loc]

			deviceCfg := DeviceConfig{
				Version: instanceCfg.Version,
				Type: instanceCfg.Type,
				Model: instanceCfg.Model,
				Location: location,
				Data: data,
			}

			devices = append(devices, deviceCfg)

		}
	}
	return devices, nil
}

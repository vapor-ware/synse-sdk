package sdk

import (
	"path/filepath"
	"os"
	"fmt"
	"io/ioutil"
	"gopkg.in/yaml.v2"

	synse "github.com/vapor-ware/synse-server-grpc/go"
)

// PluginConfig specifies the configuration options for the plugin itself.
type PluginConfig struct {

	// The name of the plugin.
	Name string `yaml:"name"`

	// The plugin version.
	Version string `yaml:"version"`

	// Log at DEBUG level.
	Debug bool `yaml:"debug"`

	// The size of the writes buffer. Since writes are processed
	// asynchronously, when a write request is received it is put
	// into a queue. Writes are processed at the beginning of every
	// iteration of the background read-write loop, but only a few
	// write transactions are processed at a time (see the
	// `WritesPerLoop` configuration option, below). This option
	// defines the size of the buffer which writes are queued in.
	//
	// Typically, the read-write loop will iterate quickly, so
	// the buffer will decumulate quickly. If writes are expected to
	// take a long time, or many writes are expected for the plugin,
	// this buffer size may need to be increased.
	WriteBufferSize int `yaml:"write_buffer_size"`

	// To prevent numerous writes requests from blocking the read block
	// of the read-write loop, we will only process a portion of the
	// queued writes at a time. This option defines the number of
	// write transactions to process per iteration of the read-write
	// loop.
	//
	// If write operations are expected to take a while for the plugin,
	// this number should be decreased so the read block can execute
	// more frequently.
	WritesPerLoop int `yaml:"writes_per_loop"`

	// A delay, in milliseconds, to wait at the end of the read-write
	// loop. This may not be needed and can be omitted (defaulting to
	// the value of 0), but it is surfaced as an option which can help
	// limit CPU/memory usage. For instance, if a plugin is written to
	// support a device which will only update its reading every 0.25
	// seconds, then it may not make sense to run the read-write loop
	// continuously. Instead `250` (milliseconds) could be specified here
	// so the loop polls the device at the same rate it updates.
	LoopDelay int `yaml:"loop_delay"`

	// When devices are read, those readings are put into a channel which
	// the ReadingManager continuously reads from to update its state.
	// ReadBufferSize defines the size of the read channel buffer.
	// Because it is being read continuously, it generally should not
	// be an issue, but if many devices are expected to be configured
	// off of a plugin (e.g. many reads occurring), increasing the read
	// buffer might become necessary.
	ReadBufferSize int `yaml:"read_buffer_size"`
}


func (c *PluginConfig) FromFile(path string) (*PluginConfig, error) {

	return &PluginConfig{}, nil
}


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

func (o *DeviceOutput) ToMetaOutput() *synse.MetaOutput {

	unit := &OutputUnit{}
	if o.Unit != nil {
		unit = o.Unit
	}

	rang := &OutputRange{}
	if o.Range != nil {
		rang = o.Range
	}

	return &synse.MetaOutput{
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

func (u *OutputUnit) ToMetaOutputUnit() *synse.MetaOutputUnit {
	return &synse.MetaOutputUnit{
		Name: u.Name,
		Symbol: u.Symbol,
	}
}

type OutputRange struct {
	Min  int32  `yaml:"min"`
	Max  int32  `yaml:"max"`
}

func (r *OutputRange) ToMetaOutputRange() *synse.MetaOutputRange {
	return &synse.MetaOutputRange{
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

func (l *DeviceLocation) ToMetalLocation() *synse.MetaLocation {
	return &synse.MetaLocation{
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

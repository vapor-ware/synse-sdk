package sdk

import (
	"errors"
	"fmt"
	"os"
	"io/ioutil"
	"path/filepath"

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

	// The time (in seconds) that transaction data should be tracked for
	// after it has completed.
	TransactionTTL int `yaml:"transaction_ttl"`
}


// FromFile reads in a YAML file and parses it into a PluginConfig struct.
func (c *PluginConfig) FromFile(path string) (*PluginConfig, error) {

	return &PluginConfig{}, nil
}


// Merge updates the fields of the PluginConfig struct with those of
// another PluginConfig. This is used primarily to combine user configurations
// with default configurations.
func (c *PluginConfig) Merge(config PluginConfig) error {

	// Since Go structs will default to the zero value for a struct field
	// that is unspecified on initialization, we need to perform some checks
	// here to see whether we should use the default configuration or not.
	// For some fields, the zero value is allowed, so we cannot differentiate
	// a configured 0 from the int zero value. In those cases, we will just
	// have the configuration default also be zero.
	// (we could technically be able to do these checks reliably by having
	// all fields be pointers. any field that isn't configured would be a
	// nil pointer. while that works for parsing, its a bit cumbersome for
	// usage.)

	// These are required fields. If they do not exist, fail.
	if config.Name == "" || config.Version == "" {
		return errors.New("Bad plugin configuration. Requires both a Name and Version.")
	}

	c.Name = config.Name
	c.Version = config.Version

	// The read buffer cannot be 0 (otherwise we would be unable to buffer
	// reads), so take a zero value here to mean "default".
	if config.ReadBufferSize != 0 {
		c.ReadBufferSize = config.ReadBufferSize
	}

	// The write buffer cannot be 0 (otherwise we would be unable to buffer
	// writes), so take a zero value here to mean "default".
	if config.WriteBufferSize != 0 {
		c.WriteBufferSize = config.WriteBufferSize
	}

	// We cannot have 0 writes per loop, otherwise no writes would ever be
	// fulfilled. Take a zero value here to mean "default".
	if config.WritesPerLoop != 0 {
		c.WritesPerLoop = config.WritesPerLoop
	}

	// We don't want the transaction TTL to be 0, otherwise it will be removed
	// almost immediately after completion, leaving no time for any subsequent
	// transaction check to finish successfully. Take a zero value here to
	// mean "default"
	if config.TransactionTTL != 0 {
		c.TransactionTTL = config.TransactionTTL
	}

	// LoopDelay can be 0 (the default), so no check is needed.
	c.LoopDelay = config.LoopDelay

	// Debug can be false (the default), so no check is needed.
	c.Debug = config.Debug

	return nil
}

// GetDefaultConfig returns a PluginConfig instance that holds the default
// values for the plugin configuration. Name and Version do not have default
// values because they are required to be specified by the plugin itself.
func GetDefaultConfig() *PluginConfig {
	return &PluginConfig{
		Debug: false,
		ReadBufferSize: 100,
		WriteBufferSize: 100,
		WritesPerLoop: 5,
		LoopDelay: 0,
		TransactionTTL: 60 * 5,  // five minutes
	}
}


// global configuration to be set via the plugin server Configure function.
var Config = GetDefaultConfig()

// FIXME - make private here and add a public call in sdk.go?
// ConfigurePlugin takes a plugin-specified configuration and sets it as
// the configuration that is used by the SDK.
func ConfigurePlugin(config PluginConfig) error {
	Config.Merge(config)
	return nil
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


// PrototypeConfig
type PrototypeConfig struct {
	Version       string          `yaml:"version"`
	Type          string          `yaml:"type"`
	Model         string          `yaml:"model"`
	Manufacturer  string          `yaml:"manufacturer"`
	Protocol      string          `yaml:"protocol"`
	Output        []DeviceOutput  `yaml:"output"`
}

// DeviceOutput
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

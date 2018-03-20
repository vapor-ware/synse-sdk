package main

// Simple Plugin Example
// ---------------------
// This file provides an example of what a simple synse background process
// plugin could look like. It contains three basic parts:
//   1.  a plugin handler  - this is where the plugin logic is defined.
//   2.  a device handler  - this is where protocol-specific helpers are defined.
//   3.  the main method   - this is where the plugin is initialized and run.

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"

	"github.com/vapor-ware/synse-sdk/sdk"
	"github.com/vapor-ware/synse-sdk/sdk/config"
)

// Build time variables for setting the version info of a Plugin.
var (
	BuildDate     string
	GitCommit     string
	GitTag        string
	GoVersion     string
	VersionString string
)

// Read defines the behavior for device reads in this example plugin.
func Read(device *sdk.Device) ([]*sdk.Reading, error) {
	return []*sdk.Reading{
		sdk.NewReading(
			device.Type,
			strconv.Itoa(rand.Int()),
		),
	}, nil
}

// Write defines the behavior for device writes in this example plugin.
func Write(device *sdk.Device, data *sdk.WriteData) error {
	fmt.Printf("[simple plugin handler]: WRITE (%v)\n", device.ID())
	fmt.Printf("Data   -> %v\n", data.Raw)
	fmt.Printf("Action -> %v\n", data.Action)
	return nil
}

// ledHandler defines the read/write behavior for the "emul8-led"
// emulated-led device.
var ledHandler = sdk.DeviceHandler{
	Type:  "emulated-led",
	Model: "emul8-led",
	Read:  Read,
	Write: Write,
}

// temperatureHandler defines the read/write behavior for the "emul8-temp"
// emulated-temperature device.
var temperatureHandler = sdk.DeviceHandler{
	Type:  "emulated-temperature",
	Model: "emul8-temp",
	Read:  Read,
	Write: Write,
}

// SimplePluginHandler fulfils the SDK's PluginHandler interface. It requires a
// Read and Write function to be defined, which specify how the plugin will read
// and write to the configured devices.
//
// Both the read and write functions operate on a single device at a time, which
// is given as a parameter. These functions will be called against all configured
// devices for the plugin, so while this example handles all reads the same, a
// more complex plugin could need to further dispatch read/write operations depending
// on the device type, model, etc.

// SimpleDeviceHandler fulfils the SDK's DeviceHandler interface.
// Each device that is generated from the configurations will be able to
// access this handler. This makes it convenient for storing helpers which
// relate to the devices themselves. For example, on of the required functions
// off of the SDK DeviceHandler interface is `GetProtocolIdentifiers`. What
// this does is allow the plugin to device which bits of protocol-specific
// data should be used when generating the device ID. For this simple plugin,
// the device configuration contains:
//
//     devices:
//       - id: 1
//         location: unknown
//         comment: first emulated temperature device
//         info: CEC temp 1
//       - id: 2
//         location: unknown
//         comment: second emulated temperature device
//         info: CEC temp 2
//
// The contents of the objects in the devices list are arbitrary and protocol-
// specific. As such, we need the plugin to define which bits of information
// here are to be used when generating the ID. In this case, we use the "id"
// field, but a concatenation of any number of fields is permissible.

// GetProtocolIdentifiers gets the unique identifiers out of the plugin-specific
// configuration to be used in UID generation.
func GetProtocolIdentifiers(data map[string]string) string {
	return data["id"]
}

// The Main Function
//   This is the entry point for the plugin. With both handlers defined,
//   all that is left to do is create a new PluginServer instance and
//   then run it.
//
//   When a PluginServer is run, it will read in the configuration, generate
//   the devices from config, start the read-write loop, and start the GRPC
//   server.
func main() {
	// Set the prototype and device instance config paths to be relative to the
	// current working directory instead of using the default location. This way
	// the plugin can be run from within this directory.
	os.Setenv("PLUGIN_DEVICE_PATH", "./config/device")
	os.Setenv("PLUGIN_PROTO_PATH", "./config/proto")

	// Create handlers for the plugin.
	handlers, err := sdk.NewHandlers(GetProtocolIdentifiers, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Configuration for the Simple Plugin.
	cfg := config.PluginConfig{
		Name:    "simple-plugin",
		Version: "1.0",
		Debug:   true,
		Network: config.NetworkSettings{
			Type:    "unix",
			Address: "simple-plugin.sock",
		},
		Settings: config.Settings{
			Read: config.ReadSettings{
				Interval: "500s",
			},
			Write: config.WriteSettings{
				Interval: "60s",
			},
			Transaction: config.TransactionSettings{
				TTL: "500m",
			},
		},
	}

	// Create a new Plugin and configure it.
	plugin, err := sdk.NewPlugin(handlers, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	plugin.RegisterDeviceHandlers(
		&temperatureHandler,
		&ledHandler,
	)

	// Set build-time version info
	plugin.SetVersion(sdk.VersionInfo{
		BuildDate:     BuildDate,
		GitCommit:     GitCommit,
		GitTag:        GitTag,
		GoVersion:     GoVersion,
		VersionString: VersionString,
	})

	// Run the plugin.
	err = plugin.Run()
	if err != nil {
		log.Fatal(err)
	}
}

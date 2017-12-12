package main

// Simple Plugin Example
// ---------------------
// This file provides an example of what a simple synse background process
// plugin could look like. It contains three basic parts:
//   1.  a plugin handler  - this is where the plugin logic is defined.
//   2.  a device handler  - this is where protocol-specific helpers are defined.
//   3.  the main method   - this is where the plugin is initialized and run.

import (
	"github.com/vapor-ware/synse-sdk/sdk"

	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"
)

// SimplePluginHandler fulfils the SDK's PluginHandler interface. It requires a
// Read and Write function to be defined, which specify how the plugin will read
// and write to the configured devices.
//
// Both the read and write functions operate on a single device at a time, which
// is given as a parameter. These functions will be called against all configured
// devices for the plugin, so while this example handles all reads the same, a
// more complex plugin could need to further dispatch read/write operations depending
// on the device type, model, etc.
type SimplePluginHandler struct{}

func (h *SimplePluginHandler) Read(device *sdk.Device) (*sdk.ReadContext, error) {

	val := rand.Int()
	strVal := strconv.Itoa(val)
	return &sdk.ReadContext{
		Device:  device.ID(),
		Board:   device.Location().Board,
		Rack:    device.Location().Rack,
		Reading: []*sdk.Reading{{time.Now().String(), device.Type(), strVal}},
	}, nil
}

func (h *SimplePluginHandler) Write(device *sdk.Device, data *sdk.WriteData) error {

	fmt.Printf("[simple plugin handler]: WRITE\n")

	fmt.Printf("Data -> %v\n", data.Raw)
	fmt.Printf("Action -> %v\n", data.Action)
	return nil
}

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
type SimpleDeviceHandler struct{}

// GetProtocolIdentifiers gets the unique identifiers out of the plugin-specific
// configuration to be used in UID generation.
func (h *SimpleDeviceHandler) GetProtocolIdentifiers(data map[string]string) string {
	return data["id"]
}

// EnumerateDevices is used to auto-enumerate device configurations for plugins
// that support it. This example plugin does not support it, so we just return
// the appropriate error.
func (h *SimpleDeviceHandler) EnumerateDevices(map[string]interface{}) ([]*sdk.DeviceConfig, error) {
	return nil, &sdk.EnumerationNotSupported{}
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
	// Configuration for the Simple Plugin.
	config := sdk.PluginConfig{
		Name:    "simple-plugin",
		Version: "1.0.0",
		Debug:   true,
		Socket: sdk.PluginConfigSocket{
			Network: "tcp",
			Address: ":50051",
		},
	}

	// Collect the Simple Plugin handlers.
	handlers := sdk.Handlers{
		Plugin: &SimplePluginHandler{},
		Device: &SimpleDeviceHandler{},
	}

	// Create a new Plugin and configure it.
	plugin := sdk.NewPlugin(&handlers)
	err := plugin.SetConfig(&config)
	if err != nil {
		log.Fatal(err)
	}

	// Register the Plugin devices - this will read the device
	// instance and prototype config to determine what it will
	// read from / write to.
	err = plugin.RegisterDevices()
	if err != nil {
		log.Fatal(err)
	}

	// Run the plugin.
	err = plugin.Run()
	if err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"

	"github.com/vapor-ware/synse-sdk/sdk"
	"github.com/vapor-ware/synse-sdk/sdk/config"
)

// The Plugin metadata. At a minimum, all plugins need a name. This information
// gets registered with the plugin (below, in the main function) and is surfaced
// to Synse Server to identify the plugin.
var (
	// FIXME: plugin names, namespacing. what is the "right" way to do this?
	pluginName       = "Simple Plugin"
	pluginMaintainer = "Vapor IO"
	pluginDesc       = "A simple example plugin"
)

// Output types are defined, either statically in the plugin code, or via YAML
// configuration files. They define the potential outputs of the plugin's devices.
// A single device could support multiple outputs, but at a minimum requires one.
var (
	// The output for temperature devices.
	temperatureOutput = config.OutputType{
		Name:      "simple.temperature",
		Precision: 2,
		Unit: config.Unit{
			Name:   "celsius",
			Symbol: "C",
		},
	}

	// The output for LED devices.
	ledOutput = config.OutputType{
		Name: "simple.led",
	}
)

// Device Handlers need to be defined to tell the plugin how to handle reads and
// writes for the different kinds of devices it supports.
var (
	// ledHandler defines the read/write behavior for the "example.led" device kind.
	ledHandler = sdk.DeviceHandler{
		Name: "example.led",

		Read: func(device *sdk.Device) ([]*sdk.Reading, error) {
			return []*sdk.Reading{
				device.GetOutput("simple.led").MakeReading(
					strconv.Itoa(rand.Int()), // nolint: gas
				),
			}, nil
		},
		Write: func(device *sdk.Device, data *sdk.WriteData) error {
			fmt.Printf("[led handler]: WRITE (%v)\n", device.ID())
			fmt.Printf("Data   -> %v\n", data.Data)
			fmt.Printf("Action -> %v\n", data.Action)
			return nil
		},
	}

	// temperatureHandler defines the read/write behavior for the "example.temperature" device kind.
	temperatureHandler = sdk.DeviceHandler{
		Name: "example.temperature",

		Read: func(device *sdk.Device) ([]*sdk.Reading, error) {
			return []*sdk.Reading{
				device.GetOutput("simple.temperature").MakeReading(
					strconv.Itoa(rand.Int()), // nolint: gas
				),
			}, nil
		},
		Write: func(device *sdk.Device, data *sdk.WriteData) error {
			fmt.Printf("[temperature handler]: WRITE (%v)\n", device.ID())
			fmt.Printf("Data   -> %v\n", data.Data)
			fmt.Printf("Action -> %v\n", data.Action)
			return nil
		},
	}
)

func main() {
	// Set the metadata for the plugin.
	sdk.SetPluginMeta(
		pluginName,
		pluginMaintainer,
		pluginDesc,
		"",
	)

	// Create a new Plugin instance.
	plugin := sdk.NewPlugin()

	// Register our output types with the Plugin.
	err := plugin.RegisterOutputTypes(
		&temperatureOutput,
		&ledOutput,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Register our device handlers with the Plugin.
	plugin.RegisterDeviceHandlers(
		&temperatureHandler,
		&ledHandler,
	)

	// Run the plugin.
	if err := plugin.Run(); err != nil {
		log.Fatal(err)
	}
}

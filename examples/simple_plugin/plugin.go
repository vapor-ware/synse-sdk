package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"

	"github.com/vapor-ware/synse-sdk/sdk"
	"github.com/vapor-ware/synse-sdk/sdk/output"
)

// The Plugin metadata. At a minimum, all plugins need a name. This information
// gets registered with the plugin (below, in the main function) and is surfaced
// to Synse Server to identify the plugin.
var (
	pluginName       = "simple plugin"
	pluginMaintainer = "vaporio"
	pluginDesc       = "A simple example plugin"
)

// Output types are defined, either statically in the plugin code, or via YAML
// configuration files. They define the potential outputs of the plugin's devices.
// A single device could support multiple outputs, but at a minimum requires one.
var (
	// The output for temperature devices.
	temperatureOutput = output.Output{
		Name:      "temperature",
		Precision: 2,
		Type:      "temperature",
		Units: map[output.SystemOfMeasure]*output.Unit{
			// fixme: use built-in once it exists
			output.NONE: {Name: "Fahrenheit", Symbol: "F", System: string(output.NONE)},
		},
		Converters: map[output.SystemOfMeasure]func(value interface{}, to output.SystemOfMeasure) (interface{}, error){
			// Define a converter that just returns the same value.
			// fixme: once we have built-in outputs, we can just use that here instead.
			output.NONE: func(value interface{}, to output.SystemOfMeasure) (i interface{}, e error) {
				return value, nil
			},
		},
	}

	// The output for on/off state devices.
	stateOutput = output.Output{
		Name: "state",
	}
)

// Device Handlers need to be defined to tell the plugin how to handle reads and
// writes for the different kinds of devices it supports.
var (
	// ledHandler defines the read/write behavior for the "example.led" device kind.
	ledHandler = sdk.DeviceHandler{
		Name: "example.led",

		Read: func(device *sdk.Device) ([]*output.Reading, error) {
			reading := stateOutput.From(strconv.Itoa(rand.Int()))

			return []*output.Reading{
				reading,
			}, nil
		},
		Write: func(device *sdk.Device, data *sdk.WriteData) error {
			fmt.Printf("[led handler]: WRITE (%v)\n", device.GetID())
			fmt.Printf("Data   -> %v\n", data.Data)
			fmt.Printf("Action -> %v\n", data.Action)
			return nil
		},
	}

	// temperatureHandler defines the read/write behavior for the "example.temperature" device kind.
	temperatureHandler = sdk.DeviceHandler{
		Name: "example.temperature",

		Read: func(device *sdk.Device) ([]*output.Reading, error) {
			reading := temperatureOutput.From(strconv.Itoa(rand.Int())) // nolint: gas, gosec

			return []*output.Reading{
				reading,
			}, nil
		},
		Write: func(device *sdk.Device, data *sdk.WriteData) error {
			fmt.Printf("[temperature handler]: WRITE (%v)\n", device.GetID())
			fmt.Printf("Data   -> %v\n", data.Data)
			fmt.Printf("Action -> %v\n", data.Action)
			return nil
		},
	}
)

func main() {
	// Create a new Plugin instance.
	plugin, err := sdk.NewPlugin()
	if err != nil {
		log.Fatal(err)
	}

	// Set plugin metadata.
	plugin.SetInfo(&sdk.PluginMetadata{
		Name:        pluginName,
		Maintainer:  pluginMaintainer,
		Description: pluginDesc,
	})

	// Register custom outputs
	// fixme: won't need to do this once there are built-ins
	err = plugin.RegisterOutputs(
		&temperatureOutput,
		&stateOutput,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Register device handlers for the plugin.
	err = plugin.RegisterDeviceHandlers(
		&temperatureHandler,
		&ledHandler,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Run the plugin.
	if err := plugin.Run(); err != nil {
		log.Fatal(err)
	}
}

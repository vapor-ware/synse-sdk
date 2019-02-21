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

// Device Handlers need to be defined to tell the plugin how to handle reads and
// writes for the different kinds of devices it supports.
var (
	// ledHandler defines the read/write behavior for the "example.led" device kind.
	ledHandler = sdk.DeviceHandler{
		Name: "example.led",

		Read: func(device *sdk.Device) ([]*output.Reading, error) {
			reading := output.State.From(strconv.Itoa(rand.Int()))

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
			reading := output.Temperature.FromImperial(strconv.Itoa(rand.Int())) // nolint: gas, gosec

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
	sdk.SetPluginInfo(
		pluginName,
		pluginMaintainer,
		pluginDesc,
		"",
	)

	// Create a new Plugin instance.
	plugin, err := sdk.NewPlugin()
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

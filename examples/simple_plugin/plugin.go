package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/vapor-ware/synse-sdk/v2/sdk"
	"github.com/vapor-ware/synse-sdk/v2/sdk/output"
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
			reading, err := output.State.MakeReading(strconv.Itoa(rand.Intn(200)))
			if err != nil {
				return nil, err
			}

			return []*output.Reading{
				reading,
			}, nil
		},
		Write: func(device *sdk.Device, data *sdk.WriteData) error {
			// simulate a bit of delay in writing
			time.Sleep(3 * time.Second)
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
			reading, err := output.Temperature.MakeReading(strconv.Itoa(rand.Intn(100))) // nolint: gas, gosec
			if err != nil {
				return nil, err
			}

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

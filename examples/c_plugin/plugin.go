package main

import (
	"log"

	"github.com/vapor-ware/synse-sdk/sdk"
)

var (
	pluginName       = "C Plugin"
	pluginMaintainer = "vaporio"
	pluginDesc       = "An example plugin that demonstrates C code integration"
)

var (
	// The output for temperature devices.
	temperatureOutput = sdk.OutputType{
		Name:      "temperature",
		Precision: 2,
		Unit: sdk.Unit{
			Name:   "celsius",
			Symbol: "C",
		},
	}
)

// temperatureHandler defines the read/write behavior for the "temp2010"
// temperature device.
var temperatureHandler = sdk.DeviceHandler{
	Name: "temperature",
	Read: func(device *sdk.Device) ([]*sdk.Reading, error) {
		id, ok := device.Data["id"].(int)
		if !ok {
			log.Fatalf("invalid device ID - should be an integer in configuration")
		}
		value := cRead(id, device.Kind)
		reading, err := device.GetOutput("temperature").MakeReading(value)
		if err != nil {
			return nil, err
		}
		return []*sdk.Reading{reading}, nil
	},
}

func main() {
	// Set the metainfo for the plugin.
	sdk.SetPluginMeta(
		pluginName,
		pluginMaintainer,
		pluginDesc,
		"",
	)

	// Create a new Plugin instance
	plugin := sdk.NewPlugin()

	// Register the output types
	err := plugin.RegisterOutputTypes(
		&temperatureOutput,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Register device handlers
	plugin.RegisterDeviceHandlers(
		&temperatureHandler,
	)

	// Run the plugin
	if err := plugin.Run(); err != nil {
		log.Fatal(err)
	}
}

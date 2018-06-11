package main

import (
	"log"

	"github.com/vapor-ware/synse-sdk/sdk"
	"github.com/vapor-ware/synse-sdk/sdk/oldconfig"
)

var (
	pluginName       = "C Plugin"
	pluginMaintainer = "Vapor IO"
	pluginDesc       = "An example plugin that demonstrates C code integration"
)

var (
	// The output for temperature devices.
	temperatureOutput = oldconfig.OutputType{
		Name:      "temperature",
		Precision: 2,
		Unit: oldconfig.Unit{
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
		return []*sdk.Reading{
			device.GetOutput("temperature").MakeReading(value),
		}, nil
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

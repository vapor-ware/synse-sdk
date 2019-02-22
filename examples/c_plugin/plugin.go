package main

import (
	"log"

	"github.com/vapor-ware/synse-sdk/sdk"
	"github.com/vapor-ware/synse-sdk/sdk/output"
)

var (
	pluginName       = "C Plugin"
	pluginMaintainer = "vaporio"
	pluginDesc       = "An example plugin that demonstrates C code integration"
)

// temperatureHandler defines the read/write behavior for the "temp2010"
// temperature device.
var temperatureHandler = sdk.DeviceHandler{
	Name: "temperature",
	Read: func(device *sdk.Device) ([]*output.Reading, error) {
		id, ok := device.Data["id"].(int)
		if !ok {
			log.Fatalf("invalid device ID - should be an integer in configuration")
		}
		value := cRead(id, device.Type)

		reading := output.Temperature.FromMetric(value)
		return []*output.Reading{reading}, nil
	},
}

func main() {
	// Set the metadata for the plugin.
	sdk.SetPluginInfo(
		pluginName,
		pluginMaintainer,
		pluginDesc,
		"",
	)

	// Create a new Plugin instance
	plugin, err := sdk.NewPlugin()
	if err != nil {
		log.Fatal(err)
	}

	// Register device handlers
	err = plugin.RegisterDeviceHandlers(
		&temperatureHandler,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Run the plugin
	if err := plugin.Run(); err != nil {
		log.Fatal(err)
	}
}

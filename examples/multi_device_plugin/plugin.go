package main

import (
	"fmt"
	"log"

	"github.com/vapor-ware/synse-sdk/examples/multi_device_plugin/devices"
	"github.com/vapor-ware/synse-sdk/examples/multi_device_plugin/outputs"
	"github.com/vapor-ware/synse-sdk/sdk"
)

var (
	pluginName       = "Multi-Device Plugin"
	pluginMaintainer = "Vapor IO"
	pluginDesc       = "An example plugin that demonstrates registering multiple devices"
)

// ProtocolIdentifier gets the unique identifiers out of the plugin-specific
// configuration to be used in UID generation.
func ProtocolIdentifier(data map[string]interface{}) string {
	return fmt.Sprint(data["id"])
}

func main() {
	// Set the metainfo for the plugin.
	sdk.SetPluginMeta(
		pluginName,
		pluginMaintainer,
		pluginDesc,
		"",
	)

	// Create a new Plugin instance with a custom device identifier.
	plugin := sdk.NewPlugin(
		sdk.CustomDeviceIdentifier(ProtocolIdentifier),
	)

	// Register our output types with the plugin
	err := plugin.RegisterOutputTypes(
		&outputs.AirflowOutput,
		&outputs.TemperatureOutput,
		&outputs.VoltageOutput,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Register device handlers
	plugin.RegisterDeviceHandlers(
		&devices.Temp2010,
		&devices.Air8884,
		&devices.Volt1103,
	)

	// Run the plugin.
	if err := plugin.Run(); err != nil {
		log.Fatal(err)
	}
}

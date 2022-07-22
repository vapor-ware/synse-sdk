package main

import (
	"fmt"
	"log"

	"github.com/vapor-ware/synse-sdk/v2/examples/multi_device_plugin/devices"
	"github.com/vapor-ware/synse-sdk/v2/examples/multi_device_plugin/outputs"
	"github.com/vapor-ware/synse-sdk/v2/sdk"
)

var (
	pluginName       = "multi device plugin"
	pluginMaintainer = "vaporio"
	pluginDesc       = "An example plugin that demonstrates registering multiple devices"
)

// ProtocolIdentifier gets the unique identifiers out of the plugin-specific
// configuration to be used in UID generation.
func ProtocolIdentifier(data map[string]interface{}) string {
	return fmt.Sprint(data["id"])
}

func main() {
	// Set the metadata for the plugin.
	sdk.SetPluginInfo(
		pluginName,
		pluginMaintainer,
		pluginDesc,
		"",
	)

	// Create a new Plugin instance with a custom device identifier.
	plugin, err := sdk.NewPlugin(
		sdk.CustomDeviceIdentifier(ProtocolIdentifier),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Register custom output types with the plugin
	err = plugin.RegisterOutputs(
		&outputs.AirflowOutput,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Register device handlers
	err = plugin.RegisterDeviceHandlers(
		&devices.Temp2010,
		&devices.Air8884,
		&devices.Volt1103,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Run the plugin.
	if err := plugin.Run(); err != nil {
		log.Fatal(err)
	}
}

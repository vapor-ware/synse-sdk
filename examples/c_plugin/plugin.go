package main

import (
	"log"
	"os"
	"strconv"

	"github.com/vapor-ware/synse-sdk/sdk"
)

var (
	pluginName       = "C Plugin"
	pluginMaintainer = "Vapor IO"
	pluginDesc       = "An example plugin that demonstrates C code integration"
)

// temperatureHandler defines the read/write behavior for the "temp2010"
// temperature device.
var temperatureHandler = sdk.DeviceHandler{
	Type:  "temperature",
	Model: "temp2010",
	Read: func(device *sdk.Device) ([]*sdk.Reading, error) {
		id, err := strconv.Atoi(device.Data["id"])
		if err != nil {
			log.Fatalf("invalid device ID - should be an integer in configuration")
		}

		value := cRead(id, device.Model)
		return []*sdk.Reading{
			sdk.NewReading(
				device.Type,
				value,
			),
		}, nil
	},
}

// ProtocolIdentifier gets the unique identifiers out of the plugin-specific
// configuration to be used in UID generation.
func ProtocolIdentifier(data map[string]string) string {
	return data["id"]
}

// checkErr is a helper used in the main function to check errors. If any errors
// are present, this will exit with log.Fatal.
func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// The main function - this is where we will configure, create, and run
// the plugin.
func main() {
	// Set the metainfo for the plugin.
	sdk.SetPluginMeta(
		pluginName,
		pluginMaintainer,
		pluginDesc,
		"",
	)

	// Set the prototype and device instance config paths to be relative to the
	// current working directory instead of using the default location. This way
	// the plugin can be run from within this directory.
	checkErr(os.Setenv("PLUGIN_DEVICE_CONFIG", "./config"))

	// Create handlers for the plugin.
	handlers, err := sdk.NewHandlers(ProtocolIdentifier, nil)
	checkErr(err)

	// The configuration comes from the files in the environment path.
	plugin, err := sdk.NewPlugin(handlers, nil)
	checkErr(err)

	plugin.RegisterDeviceHandlers(
		&temperatureHandler,
	)

	// Run the plugin.
	checkErr(plugin.Run())
}

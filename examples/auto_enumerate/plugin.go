package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"

	"github.com/vapor-ware/synse-sdk/sdk"
	"github.com/vapor-ware/synse-sdk/sdk/config"
)

// temperatureHandler defines the read/write behavior for the "temp2010"
// temperature device.
var temperatureHandler = sdk.DeviceHandler{
	Type:  "temperature",
	Model: "temp2010",
	Read: func(device *sdk.Device) ([]*sdk.Reading, error) {
		value := strconv.Itoa(rand.Int()) // nolint: gas
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

// EnumerateDevices is used to auto-enumerate device configurations for plugins
// that support it. This is plugin specific. The config parameter here is a map
// of configuration values that are taken from the list defined in the plugin
// config file under the "auto_enumerate" option. This method will be called on
// each member of that list sequentially.
//
// The example implementation here is a bit contrived - it takes a base address
// from the config and creates 3 devices off of that base. This isn't necessarily
// "auto-enumeration" by definition, but it is a valid usage. A more appropriate
// example could be taking an IP from the configuration, and using that to hit some
// endpoint which would give back all the information on the devices it manages.
func EnumerateDevices(cfg map[string]interface{}) ([]*config.DeviceConfig, error) {

	var res []*config.DeviceConfig

	baseAddr := cfg["base"]
	for i := 0; i < 3; i++ {
		devAddr := fmt.Sprintf("%v-%v", baseAddr, i)

		// validate the location
		location := config.Location{
			Rack:  "rack-1",
			Board: "board-1",
		}
		err := location.Validate()
		if err != nil {
			return nil, err
		}

		// create a new device - here, we are using the base address and appending
		// index of the loop to create the id of the device. we are hardcoding in
		// the type and model as temperature and temp2010, respectively, because
		// we need the devices to match the prototypes were support. in this example,
		// we only have the temperature device prototype. in a real case, this info
		// should be gathered from whatever the real source of auto-enumeration is,
		// e.g. for IPMI - the SDR records.
		d := config.DeviceConfig{
			Version:  "1.0",
			Type:     "temperature",
			Model:    "temp2010",
			Location: location,
			// we want to have "id" in the map because our `GetProtocolIdentifiers"
			// uses the "id" field here to create the internal device uid.
			Data: map[string]string{
				"id": devAddr,
			},
		}
		res = append(res, &d)
	}

	return res, nil
}

// checkErr is a helper used in the main function to check errors. If any errors
// are present, this will exit with log.Fatal.
func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// Set the prototype and device instance config paths to be relative to the
	// current working directory instead of using the default location. This way
	// the plugin can be run from within this directory.
	// TODO: https://github.com/vapor-ware/synse-sdk/issues/125
	checkErr(os.Setenv("PLUGIN_DEVICE_PATH", "./config/device"))
	checkErr(os.Setenv("PLUGIN_PROTO_PATH", "./config/proto"))

	// Create handlers for the plugin.
	handlers, err := sdk.NewHandlers(ProtocolIdentifier, EnumerateDevices)
	checkErr(err)

	// Create the plugin. The configuration comes from the paths set above.
	plugin, err := sdk.NewPlugin(handlers, nil)
	checkErr(err)

	// Register handlers for our devices.
	plugin.RegisterDeviceHandlers(
		&temperatureHandler,
	)

	// Run the plugin.
	checkErr(plugin.Run())
}

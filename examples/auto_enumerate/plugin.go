package main

import (
	"github.com/vapor-ware/synse-sdk/sdk"

	"log"
	"strconv"
	"time"
	"math/rand"
	"fmt"
)


// ExamplePluginHandler is a plugin-specific handler required by the
// SDK. It defines the plugin's read and write functionality.
type ExamplePluginHandler struct {}

func (h *ExamplePluginHandler) Read(in sdk.Device) (sdk.ReadResource, error) {

	val := rand.Int()
	strVal := strconv.Itoa(val)
	return sdk.ReadResource{
		Device:  in.UID(),
		Reading: []sdk.Reading{{time.Now().String(), in.Type(),strVal}},
	}, nil
}

func (h *ExamplePluginHandler) Write(in sdk.Device, data *sdk.WriteData) error {
	return nil
}


// ExampleDeviceHandler is a plugin-specific handler required by the
// SDK. It defines functions which are needed to parse/make sense of
// some of the plugin-specific configurations.
type ExampleDeviceHandler struct {}

// GetProtocolIdentifiers gets the unique identifiers out of the plugin-specific
// configuration to be used in UID generation.
func (h *ExampleDeviceHandler) GetProtocolIdentifiers(data map[string]string) string {
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
func (h *ExampleDeviceHandler) EnumerateDevices(config map[string]interface{}) ([]sdk.DeviceConfig, error) {

	var res []sdk.DeviceConfig

	baseAddr := config["base"]
	for i := 0; i < 3; i++ {
		devAddr := fmt.Sprintf("%v-%v", baseAddr, i)

		// create a new device - here, we are using the base address and appending
		// index of the loop to create the id of the device. we are hardcoding in
		// the type and model as temperature and temp2010, respectively, because
		// we need the devices to match the prototypes were support. in this example,
		// we only have the temperature device prototype. in a real case, this info
		// should be gathered from whatever the real source of auto-enumeration is,
		// e.g. for IPMI - the SDR records.
		d := sdk.DeviceConfig{
			Version: "1.0",
			Type: "temperature",
			Model: "temp2010",
			Location: sdk.DeviceLocation{
				Rack: "rack-1",
				Board: "board-1",
			},
			// we want to have "id" in the map because our `GetProtocolIdentifiers"
			// uses the "id" field here to create the internal device uid.
			Data: map[string]string{
				"id": devAddr,
			},
		}
		res = append(res, d)
	}

	return res, nil
}


// The main function - this is where we will configure, create, and run
// the plugin.
func main() {
	config := sdk.PluginConfig{}
	err := config.FromFile("plugin.yml")
	if err != nil {
		log.Fatal(err)
	}

	p, err := sdk.NewPlugin(
		config,
		&ExamplePluginHandler{},
		&ExampleDeviceHandler{},
	)
	if err != nil {
		log.Fatal(err)
	}

	p.Run()
}
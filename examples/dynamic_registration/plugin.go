package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"

	"github.com/vapor-ware/synse-sdk/v2/sdk"
	"github.com/vapor-ware/synse-sdk/v2/sdk/config"
	"github.com/vapor-ware/synse-sdk/v2/sdk/output"
)

var (
	pluginName       = "dynamic registration plugin"
	pluginMaintainer = "vaporio"
	pluginDesc       = "An example plugin that demonstrates dynamically registering devices"
)

// temperatureHandler defines the read/write behavior for the "temp2010"
// temperature device.
var temperatureHandler = sdk.DeviceHandler{
	Name: "temperature",
	Read: func(device *sdk.Device) ([]*output.Reading, error) {
		value := strconv.Itoa(rand.Intn(100)) // nolint: gas, gosec
		reading, err := output.Temperature.MakeReading(value)
		if err != nil {
			return nil, err
		}
		return []*output.Reading{reading}, nil
	},
}

// ProtocolIdentifier gets the unique identifiers out of the plugin-specific
// configuration to be used in UID generation.
func ProtocolIdentifier(data map[string]interface{}) string {
	return data["id"].(string)
}

// DynamicDeviceConfig is used to dynamically register device configurations for plugins
// that support it. This is plugin specific. The config parameter here is a map
// of configuration values that are taken from the config defined in the plugin
// config file under the "dynamicRegistration" option.
//
// The example implementation here is a bit contrived - it takes a base address
// from the config and creates 3 devices off of that base. This isn't necessarily
// "dynamic registration" by definition, but it is a valid usage. A more appropriate
// example could be taking an IP from the configuration, and using that to hit some
// endpoint which would give back all the information on the devices it manages.
func DynamicDeviceConfig(cfg map[string]interface{}) ([]*config.DeviceProto, error) {
	// create a new device - here, we are using the base address and appending
	// index of the loop to create the id of the device. we are hardcoding in
	// the name as temperature and temp2010, respectively, because we need the
	// devices to match to their device handlers. in a real case, all of this info
	// should be gathered from whatever the real source of dynamic registration is,
	// e.g. for IPMI - the SDR records.
	res := []*config.DeviceProto{
		{
			Type: "temperature",
			Context: map[string]string{
				"model": "temp2010",
			},
			Instances: []*config.DeviceInstance{
				{
					Handler: "temperature",
					Info:    "test device",
					Data: map[string]interface{}{
						"id": fmt.Sprint(cfg["base"]),
					},
				},
			},
		},
	}
	return res, nil
}

func main() {
	// Set the metadata for the plugin.
	sdk.SetPluginInfo(
		pluginName,
		pluginMaintainer,
		pluginDesc,
		"",
	)

	// Create a new Plugin instance with custom identifier and dynamic registration
	// functions supplied.
	plugin, err := sdk.NewPlugin(
		sdk.CustomDeviceIdentifier(ProtocolIdentifier),
		sdk.CustomDynamicDeviceConfigRegistration(DynamicDeviceConfig),
		sdk.DynamicConfigRequired(),
		sdk.DeviceConfigOptional(),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Register device handlers
	err = plugin.RegisterDeviceHandlers(&temperatureHandler)
	if err != nil {
		log.Fatal(err)
	}

	// Run the plugin.
	if err := plugin.Run(); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"

	"github.com/vapor-ware/synse-sdk/sdk"
	"github.com/vapor-ware/synse-sdk/sdk/config"
)

var (
	pluginName       = "Dynamic Registration Plugin"
	pluginMaintainer = "Vapor IO"
	pluginDesc       = "An example plugin that demonstrates dynamically registering devices"
)

var (
	// The output for temperature devices.
	temperatureOutput = config.OutputType{
		Name:      "temperature",
		Precision: 2,
		Unit: config.Unit{
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
		value := strconv.Itoa(rand.Int()) // nolint: gas
		return []*sdk.Reading{
			device.GetOutput("temperature").MakeReading(value),
		}, nil
	},
}

// ProtocolIdentifier gets the unique identifiers out of the plugin-specific
// configuration to be used in UID generation.
func ProtocolIdentifier(data map[string]interface{}) string {
	return data["id"].(string)
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
func DynamicDeviceConfig(cfg map[string]interface{}) ([]*config.DeviceConfig, error) {
	var res []*config.DeviceConfig

	baseAddr := cfg["base"]
	for i := 0; i < 3; i++ {
		devAddr := fmt.Sprintf("%v-%v", baseAddr, i)

		// create a new device - here, we are using the base address and appending
		// index of the loop to create the id of the device. we are hardcoding in
		// the type and model as temperature and temp2010, respectively, because
		// we need the devices to match the prototypes were support. in this example,
		// we only have the temperature device prototype. in a real case, this info
		// should be gathered from whatever the real source of auto-enumeration is,
		// e.g. for IPMI - the SDR records.
		d := config.DeviceConfig{
			ConfigVersion: config.ConfigVersion{
				Version: "1.0",
			},
			Locations: []*config.Location{
				{
					Name:  "foobar",
					Rack:  &config.LocationData{Name: "foo"},
					Board: &config.LocationData{Name: "bar"},
				},
			},
			Devices: []*config.DeviceKind{
				{
					Name: "temperature",
					Metadata: map[string]string{
						"model": "temp2010",
					},
					Instances: []*config.DeviceInstance{
						{
							Info:     fmt.Sprintf("Test Device %d", i),
							Location: "foobar",
							Data: map[string]interface{}{
								"id": devAddr,
							},
							Outputs: []*config.DeviceOutput{
								{
									Type: "temperature",
								},
							},
						},
					},
				},
			},
		}

		res = append(res, &d)
	}
	return res, nil
}

func main() {
	// Set the metainfo for the plugin.
	sdk.SetPluginMeta(
		pluginName,
		pluginMaintainer,
		pluginDesc,
		"",
	)

	// Create a new Plugin instance with custom identifier and dynamic registration
	// functions supplied.
	plugin := sdk.NewPlugin(
		sdk.CustomDeviceIdentifier(ProtocolIdentifier),
		sdk.CustomDynamicDeviceConfigRegistration(DynamicDeviceConfig),
	)

	// Register output types
	plugin.RegisterOutputTypes(&temperatureOutput)

	// Register device handlers
	plugin.RegisterDeviceHandlers(&temperatureHandler)

	// Run the plugin.
	if err := plugin.Run(); err != nil {
		log.Fatal(err)
	}
}

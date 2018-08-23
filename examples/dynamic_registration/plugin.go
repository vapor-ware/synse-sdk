package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"

	"github.com/vapor-ware/synse-sdk/sdk"
	"github.com/vapor-ware/synse-sdk/sdk/policies"
)

var (
	pluginName       = "dynamic registration plugin"
	pluginMaintainer = "vaporio"
	pluginDesc       = "An example plugin that demonstrates dynamically registering devices"
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
		value := strconv.Itoa(rand.Int()) // nolint: gas, gosec
		reading, err := device.GetOutput("temperature").MakeReading(value)
		if err != nil {
			return nil, err
		}
		return []*sdk.Reading{reading}, nil
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
func DynamicDeviceConfig(cfg map[string]interface{}) ([]*sdk.DeviceConfig, error) {
	var res []*sdk.DeviceConfig

	// create a new device - here, we are using the base address and appending
	// index of the loop to create the id of the device. we are hardcoding in
	// the name as temperature and temp2010, respectively, because we need the
	// devices to match to their device handlers. in a real case, all of this info
	// should be gathered from whatever the real source of dynamic registration is,
	// e.g. for IPMI - the SDR records.
	d := sdk.DeviceConfig{
		SchemeVersion: sdk.SchemeVersion{
			Version: "1.0",
		},
		Locations: []*sdk.LocationConfig{
			{
				Name:  "foobar",
				Rack:  &sdk.LocationData{Name: "foo"},
				Board: &sdk.LocationData{Name: "bar"},
			},
		},
		Devices: []*sdk.DeviceKind{
			{
				Name: "temperature",
				Metadata: map[string]string{
					"model": "temp2010",
				},
				Instances: []*sdk.DeviceInstance{
					{
						Info:     "test device",
						Location: "foobar",
						Data: map[string]interface{}{
							"id": fmt.Sprint(cfg["base"]),
						},
						Outputs: []*sdk.DeviceOutput{
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

	// Set the device config policy to optional - this means that we will not
	// fail if there are no device config files found.
	policies.Add(policies.DeviceConfigFileOptional)

	// Register output types
	err := plugin.RegisterOutputTypes(&temperatureOutput)
	if err != nil {
		log.Fatal(err)
	}

	// Register device handlers
	plugin.RegisterDeviceHandlers(&temperatureHandler)

	// Run the plugin.
	if err := plugin.Run(); err != nil {
		log.Fatal(err)
	}
}

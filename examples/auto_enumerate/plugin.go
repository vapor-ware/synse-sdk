package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/vapor-ware/synse-sdk/sdk"
	"github.com/vapor-ware/synse-sdk/sdk/config"
)

// Build time variables for setting the version info of a Plugin.
var (
	BuildDate     string
	GitCommit     string
	GitTag        string
	GoVersion     string
	VersionString string
)

// temperatureHandler defines the read/write behavior for the "temp2010"
// temperature device.
var temperatureHandler = sdk.DeviceHandler{
	Type:  "temperature",
	Model: "temp2010",
	Read: func(device *sdk.Device) ([]*sdk.Reading, error) {
		value := strconv.Itoa(rand.Int())
		return []*sdk.Reading{{
			Timestamp: time.Now().String(),
			Type:      device.Type,
			Value:     value,
		}}, nil
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

		// create a new device - here, we are using the base address and appending
		// index of the loop to create the id of the device. we are hardcoding in
		// the type and model as temperature and temp2010, respectively, because
		// we need the devices to match the prototypes were support. in this example,
		// we only have the temperature device prototype. in a real case, this info
		// should be gathered from whatever the real source of auto-enumeration is,
		// e.g. for IPMI - the SDR records.
		d := config.DeviceConfig{
			Version: "1.0",
			Type:    "temperature",
			Model:   "temp2010",
			Location: config.Location{
				Rack:  "rack-1",
				Board: "board-1",
			},
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

func main() {
	// Set the prototype and device instance config paths to be relative to the
	// current working directory instead of using the default location. This way
	// the plugin can be run from within this directory.
	os.Setenv("PLUGIN_DEVICE_PATH", "./config/device")
	os.Setenv("PLUGIN_PROTO_PATH", "./config/proto")

	// Create a new Plugin and configure it.
	plugin := sdk.NewPlugin()
	err := plugin.Configure()
	if err != nil {
		log.Fatal(err)
	}

	// Create handlers for the plugin and register them.
	handlers, err := sdk.NewHandlers(ProtocolIdentifier, EnumerateDevices)
	if err != nil {
		log.Fatal(err)
	}
	plugin.RegisterHandlers(handlers)

	plugin.RegisterDeviceHandlers(
		&temperatureHandler,
	)

	// Set build-time version info
	plugin.SetVersion(sdk.VersionInfo{
		BuildDate:     BuildDate,
		GitCommit:     GitCommit,
		GitTag:        GitTag,
		GoVersion:     GoVersion,
		VersionString: VersionString,
	})

	// Run the plugin.
	err = plugin.Run()
	if err != nil {
		log.Fatal(err)
	}
}

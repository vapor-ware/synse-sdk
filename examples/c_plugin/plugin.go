package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/vapor-ware/synse-sdk/sdk"
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
		id, err := strconv.Atoi(device.Data["id"])
		if err != nil {
			log.Fatalf("invalid device ID - should be an integer in configuration")
		}

		value := cRead(id, device.Model)
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

// The main function - this is where we will configure, create, and run
// the plugin.
func main() {
	// Set the prototype and device instance config paths to be relative to the
	// current working directory instead of using the default location. This way
	// the plugin can be run from within this directory.
	os.Setenv("PLUGIN_DEVICE_CONFIG", "./config")

	// Create a new Plugin and configure it.
	plugin := sdk.NewPlugin()
	err := plugin.Configure()
	if err != nil {
		log.Fatal(err)
	}

	plugin.RegisterDeviceIdentifier(ProtocolIdentifier)
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

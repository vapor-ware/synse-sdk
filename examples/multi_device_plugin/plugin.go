package main

import (
	"log"
	"os"

	"github.com/vapor-ware/synse-sdk/examples/multi_device_plugin/devices"
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
	os.Setenv("PLUGIN_DEVICE_PATH", "./config/device")
	os.Setenv("PLUGIN_PROTO_PATH", "./config/proto")

	plugin := sdk.NewPlugin()
	err := plugin.Configure()
	if err != nil {
		log.Fatal(err)
	}

	plugin.RegisterDeviceIdentifier(ProtocolIdentifier)
	plugin.RegisterDeviceHandlers(
		&devices.Temp2010,
		&devices.Air8884,
		&devices.Volt1103,
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

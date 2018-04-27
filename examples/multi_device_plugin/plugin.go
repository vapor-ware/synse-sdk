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
	// Set the prototype and device instance config paths to be relative to the
	// current working directory instead of using the default location. This way
	// the plugin can be run from within this directory.
	checkErr(os.Setenv("PLUGIN_DEVICE_PATH", "./config/device"))
	checkErr(os.Setenv("PLUGIN_PROTO_PATH", "./config/proto"))

	// Create handlers for the plugin.
	handlers, err := sdk.NewHandlers(ProtocolIdentifier, nil)
	checkErr(err)

	// Create the plugin.
	// The configuration comes from the environment set above.
	plugin, err := sdk.NewPlugin(handlers, nil)
	checkErr(err)

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
	checkErr(plugin.Run())
}

package main

import (
	"log"
	"os"

	"github.com/vapor-ware/synse-sdk/examples/pre_run_actions/devices"
	"github.com/vapor-ware/synse-sdk/sdk"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
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

// preRunAction1 defines a function we will use as a pre-run action.
func preRunAction1(p *sdk.Plugin) error {
	logger.Debug("preRunAction1 -> adding to config context")
	p.Config.Context = make(map[string]interface{})
	p.Config.Context["example_ctx"] = true
	return nil
}

// preRunAction2 defines a function we will use as a pre-run action.
func preRunAction2(p *sdk.Plugin) error {
	logger.Debug("preRunAction2 -> displaying plugin config")
	logger.Debug(p.Config)
	return nil
}

// deviceSetupAction defines a function we will use as a device setup action.
func deviceSetupAction(p *sdk.Plugin, d *sdk.Device) error {
	logger.Debug("deviceSetupAction1 -> print device info for the given filter")
	logger.Debug("device")
	logger.Debugf("  id:    %v", d.ID())
	logger.Debugf("  type:  %v", d.Type)
	logger.Debugf("  model: %v", d.Model)
	return nil
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
		&devices.Air8884,
		&devices.Temp2010,
	)

	// Set build-time version info
	plugin.SetVersion(sdk.VersionInfo{
		BuildDate:     BuildDate,
		GitCommit:     GitCommit,
		GitTag:        GitTag,
		GoVersion:     GoVersion,
		VersionString: VersionString,
	})

	// Register pre-run actions and device setup actions for the plugin.
	// This can happen at any point before plugin.Run() is called.
	plugin.RegisterPreRunActions(
		preRunAction1,
		preRunAction2,
	)

	logger.Debug("Registering action for filter 'type=airflow'")
	plugin.RegisterDeviceSetupActions(
		"type=airflow",
		deviceSetupAction,
	)

	logger.Debug("Registering action for filter 'model=*'")
	plugin.RegisterDeviceSetupActions(
		"model=*",
		deviceSetupAction,
	)

	// Run the plugin.
	err = plugin.Run()
	if err != nil {
		log.Fatal(err)
	}
}

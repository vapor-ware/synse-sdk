package main

import (
	"log"
	"os"

	"github.com/vapor-ware/synse-sdk/examples/pre_run_actions/devices"
	"github.com/vapor-ware/synse-sdk/sdk"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
)

var (
	pluginName       = "Pre-Run Action Plugin"
	pluginMaintainer = "Vapor IO"
	pluginDesc       = "An example plugin that demonstrates pre-run action capabilities"
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
	// Set the metainfo for the plugin.
	sdk.SetPluginMeta(
		pluginName,
		pluginMaintainer,
		pluginDesc,
		"",
	)

	// Set the prototype and device instance config paths to be relative to the
	// current working directory instead of using the default location. This way
	// the plugin can be run from within this directory.
	checkErr(os.Setenv("PLUGIN_DEVICE_CONFIG", "./config"))

	// Create handlers for the Plugin.
	handlers, err := sdk.NewHandlers(ProtocolIdentifier, nil)
	checkErr(err)

	// Create a new Plugin and configure it.
	// The configuration comes from the environment settings above.
	plugin, err := sdk.NewPlugin(handlers, nil)
	checkErr(err)

	plugin.RegisterDeviceHandlers(
		&devices.Air8884,
		&devices.Temp2010,
	)

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
	checkErr(plugin.Run())
}

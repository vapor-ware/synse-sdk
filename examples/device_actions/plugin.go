package main

import (
	"fmt"
	"log"

	logger "github.com/Sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/examples/device_actions/devices"
	"github.com/vapor-ware/synse-sdk/sdk"
)

var (
	pluginName       = "device action plugin"
	pluginMaintainer = "vaporio"
	pluginDesc       = "An example plugin that demonstrates pre-run action capabilities"
)

// ProtocolIdentifier gets the unique identifiers out of the plugin-specific
// configuration to be used in UID generation.
func ProtocolIdentifier(data map[string]interface{}) string {
	return fmt.Sprint(data["id"])
}

// preRunAction1 defines a function we will use as a pre-run action.
func preRunAction1(_ *sdk.Plugin) error {
	logger.Debug("preRunAction1 -> adding to config context")
	sdk.Config.Plugin.Context["example_ctx"] = true
	return nil
}

// preRunAction2 defines a function we will use as a pre-run action.
func preRunAction2(_ *sdk.Plugin) error {
	logger.Debug("preRunAction2 -> displaying plugin config")
	logger.Debug(sdk.Config.Plugin)
	return nil
}

// deviceSetupAction defines a function we will use as a device setup action.
func deviceSetupAction(_ *sdk.Plugin, d *sdk.Device) error {
	logger.Debug("deviceSetupAction1 -> print device info for the given filter")
	logger.Debug("device")
	logger.Debugf("  id:    %v", d.ID())
	logger.Debugf("  kind:  %v", d.Kind)
	logger.Debugf("  meta:  %v", d.Metadata)
	return nil
}

func main() {
	// Set the metainfo for the plugin.
	sdk.SetPluginMeta(
		pluginName,
		pluginMaintainer,
		pluginDesc,
		"",
	)

	// Create a new Plugin instance with a custom device identifier.
	plugin := sdk.NewPlugin(
		sdk.CustomDeviceIdentifier(ProtocolIdentifier),
	)

	// Register the device handlers
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

	logger.Debug("Registering action for filter 'kind=airflow'")
	plugin.RegisterDeviceSetupActions(
		"kind=airflow",
		deviceSetupAction,
	)

	// Run the plugin.
	if err := plugin.Run(); err != nil {
		log.Fatal(err)
	}
}

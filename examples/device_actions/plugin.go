package main

import (
	"fmt"
	"log"

	logger "github.com/sirupsen/logrus"
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
	logger.Debug("preRunAction1 -> doing some action")
	return nil
}

// preRunAction2 defines a function we will use as a pre-run action.
func preRunAction2(p *sdk.Plugin) error {
	logger.Debug("preRunAction2 -> displaying plugin")
	logger.Debug(p)
	return nil
}

// deviceSetupAction defines a function we will use as a device setup action.
func deviceSetupAction(_ *sdk.Plugin, d *sdk.Device) error {
	logger.Debug("deviceSetupAction1 -> print device info for the given filter")
	logger.Debug("device")
	logger.Debugf("  id:    %v", d.GetID())
	logger.Debugf("  type:  %v", d.Type)
	logger.Debugf("  meta:  %v", d.Metadata)
	return nil
}

func main() {
	// Set the metadata for the plugin.
	sdk.SetPluginInfo(
		pluginName,
		pluginMaintainer,
		pluginDesc,
		"",
	)

	// Create a new Plugin instance with a custom device identifier.
	plugin, err := sdk.NewPlugin(
		sdk.CustomDeviceIdentifier(ProtocolIdentifier),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Register the device handlers
	err = plugin.RegisterDeviceHandlers(
		&devices.Air8884,
		&devices.Temp2010,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Register pre-run actions and device setup actions for the plugin.
	// This can happen at any point before plugin.Run() is called.
	plugin.RegisterPreRunActions(
		&sdk.PluginAction{
			Name:   "action 1",
			Action: preRunAction1,
		},
		&sdk.PluginAction{
			Name:   "action 2",
			Action: preRunAction2,
		},
	)

	logger.Debug("Registering action for filter 'kind=airflow'")
	err = plugin.RegisterDeviceSetupActions(
		&sdk.DeviceAction{
			Name:   "example action",
			Filter: map[string][]string{"type": {"airflow"}},
			Action: deviceSetupAction,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Run the plugin.
	if err := plugin.Run(); err != nil {
		log.Fatal(err)
	}
}

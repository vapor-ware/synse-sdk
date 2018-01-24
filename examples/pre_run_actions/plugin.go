package main

import (
	"log"

	"os"

	"github.com/vapor-ware/synse-sdk/examples/pre_run_actions/devices"
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

// lookup is a simple lookup table that maps the known device models
// that are supported by this plugin to the handler for that model.
//
// this is not the only way to route commands to the appropriate handler
// for a given device. there may be better ways, but this is simple
// enough and keeps this example clear and understandable.
var lookup = map[string]devices.DeviceInterface{
	"air8884":  &devices.Air8884{},
	"temp2010": &devices.Temp2010{},
}

// ExamplePluginHandler is a plugin-specific handler required by the
// SDK. It defines the plugin's read and write functionality.
type ExamplePluginHandler struct{}

func (h *ExamplePluginHandler) Read(device *sdk.Device) (*sdk.ReadContext, error) {
	handler := lookup[device.Model()]
	if handler == nil {
		log.Fatalf("Unsupported device model: %+v", device)
	}
	return handler.Read(device)
}

func (h *ExamplePluginHandler) Write(device *sdk.Device, data *sdk.WriteData) error {
	handler := lookup[device.Model()]
	if handler == nil {
		log.Fatalf("Unsupported device model: %+v", device)
	}
	return handler.Write(device, data)
}

// ExampleDeviceHandler is a plugin-specific handler required by the
// SDK. It defines functions which are needed to parse/make sense of
// some of the plugin-specific configurations.
type ExampleDeviceHandler struct{}

// GetProtocolIdentifiers gets the unique identifiers out of the plugin-specific
// configuration to be used in UID generation.
func (h *ExampleDeviceHandler) GetProtocolIdentifiers(data map[string]string) string {
	return data["id"]
}

// EnumerateDevices is used to auto-enumerate device configurations for plugins
// that support it. This example plugin does not support it, so we just return
// the appropriate error.
func (h *ExampleDeviceHandler) EnumerateDevices(map[string]interface{}) ([]*config.DeviceConfig, error) {
	return nil, &sdk.EnumerationNotSupported{}
}

// preRunAction1 defines a function we will use as a pre-run action.
func preRunAction1(p *sdk.Plugin) error {
	sdk.Logger.Debug("preRunAction1 -> adding to config context")
	p.Config.Context = make(map[string]interface{})
	p.Config.Context["example_ctx"] = true
	return nil
}

// preRunAction2 defines a function we will use as a pre-run action.
func preRunAction2(p *sdk.Plugin) error {
	sdk.Logger.Debug("preRunAction2 -> displaying plugin config")
	sdk.Logger.Debug(p.Config)
	return nil
}

// deviceSetupAction defines a function we will use as a device setup action.
func deviceSetupAction(p *sdk.Plugin, d *sdk.Device) error {
	sdk.Logger.Debug("deviceSetupAction1 -> print device info for the given filter")
	sdk.Logger.Debug("device")
	sdk.Logger.Debugf("  id:    %v", d.ID())
	sdk.Logger.Debugf("  type:  %v", d.Type())
	sdk.Logger.Debugf("  model: %v", d.Model())
	return nil
}

// The main function - this is where we will configure, create, and run
// the plugin.
func main() {
	// Set the prototype and device instance config paths to be relative to the
	// current working directory instead of using the default location. This way
	// the plugin can be run from within this directory.
	os.Setenv("PLUGIN_DEVICE_CONFIG", "./config")

	// Collect the Plugin handlers.
	handlers := sdk.Handlers{
		Plugin: &ExamplePluginHandler{},
		Device: &ExampleDeviceHandler{},
	}

	// Create a new Plugin and configure it.
	plugin := sdk.NewPlugin(&handlers)
	err := plugin.Configure()
	if err != nil {
		log.Fatal(err)
	}

	// Set build-time version info
	plugin.SetVersion(sdk.VersionInfo{
		BuildDate:     BuildDate,
		GitCommit:     GitCommit,
		GitTag:        GitTag,
		GoVersion:     GoVersion,
		VersionString: VersionString,
	})

	// Register the Plugin devices.
	err = plugin.RegisterDevices()
	if err != nil {
		log.Fatal(err)
	}

	// Register pre-run actions and device setup actions for the plugin.
	// This can happen at any point before plugin.Run() is called.
	plugin.RegisterPreRunActions(
		preRunAction1,
		preRunAction2,
	)

	sdk.Logger.Debug("Registering action for filter 'type=airflow'")
	plugin.RegisterDeviceSetupActions(
		"type=airflow",
		deviceSetupAction,
	)

	sdk.Logger.Debug("Registering action for filter 'model=*'")
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

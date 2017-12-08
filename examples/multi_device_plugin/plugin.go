package main

import (
	"log"

	"github.com/vapor-ware/synse-sdk/examples/multi_device_plugin/devices"
	"github.com/vapor-ware/synse-sdk/sdk"
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
	"volt1103": &devices.Volt1103{},
}

// ExamplePluginHandler is a plugin-specific handler required by the
// SDK. It defines the plugin's read and write functionality.
type ExamplePluginHandler struct{}

func (h *ExamplePluginHandler) Read(in *sdk.Device) (*sdk.ReadResource, error) {
	handler := lookup[in.Model()]
	if handler == nil {
		log.Fatalf("Unsupported device model: %+v", in)
	}
	return handler.Read(in)
}

func (h *ExamplePluginHandler) Write(in *sdk.Device, data *sdk.WriteData) error {
	handler := lookup[in.Model()]
	if handler == nil {
		log.Fatalf("Unsupported device model: %+v", in)
	}
	return handler.Write(in, data)
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
func (h *ExampleDeviceHandler) EnumerateDevices(map[string]interface{}) ([]*sdk.DeviceConfig, error) {
	return nil, &sdk.EnumerationNotSupported{}
}

// The main function - this is where we will configure, create, and run
// the plugin.
func main() {
	config := sdk.PluginConfig{}
	err := config.FromFile("plugin.yml")
	if err != nil {
		log.Fatal(err)
	}

	p, err := sdk.NewPlugin(
		&config,
		&ExamplePluginHandler{},
		&ExampleDeviceHandler{},
	)
	if err != nil {
		log.Fatal(err)
	}

	err = p.Run()
	if err != nil {
		log.Fatal(err)
	}
}

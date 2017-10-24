package main

import (
	"github.com/vapor-ware/synse-sdk/sdk"

	"log"
	"strconv"
	"time"
)


// ExamplePluginHandler is a plugin-specific handler required by the
// SDK. It defines the plugin's read and write functionality.
type ExamplePluginHandler struct {}

func (h *ExamplePluginHandler) Read(in sdk.Device) (sdk.ReadResource, error) {
	id, err := strconv.Atoi(in.Data()["id"]); if err != nil {
		log.Fatalf("Invalid device ID - should be an integer in configuration.")
	}
	value := cRead(id, in.Model())
	return sdk.ReadResource{
		Device: in.UID(),
		Reading: []sdk.Reading{{time.Now().String(), in.Type(), value}},
	}, nil
}

func (h *ExamplePluginHandler) Write(in sdk.Device, data *sdk.WriteData) error {
	return nil
}


// ExampleDeviceHandler is a plugin-specific handler required by the
// SDK. It defines functions which are needed to parse/make sense of
// some of the plugin-specific configurations.
type ExampleDeviceHandler struct {}

// GetProtocolIdentifiers gets the unique identifiers out of the plugin-specific
// configuration to be used in UID generation.
func (h *ExampleDeviceHandler) GetProtocolIdentifiers(data map[string]string) string {
	return data["id"]
}

// EnumerateDevices is used to auto-enumerate device configurations for plugins
// that support it. This example plugin does not support it, so we just return
// the appropriate error.
func (h *ExampleDeviceHandler) EnumerateDevices(map[string]interface{}) ([]sdk.DeviceConfig, error) {
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
		config,
		&ExamplePluginHandler{},
		&ExampleDeviceHandler{},
	)
	if err != nil {
		log.Fatal(err)
	}

	p.Run()
}
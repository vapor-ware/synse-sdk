package main

import (
	"github.com/vapor-ware/synse-sdk/sdk"

	"log"
	"strconv"
	"time"
	"github.com/vapor-ware/synse-sdk/sdk/config"
	"os"
)

// ExamplePluginHandler is a plugin-specific handler required by the
// SDK. It defines the plugin's read and write functionality.
type ExamplePluginHandler struct{}

func (h *ExamplePluginHandler) Read(device *sdk.Device) (*sdk.ReadContext, error) {
	id, err := strconv.Atoi(device.Data()["id"])
	if err != nil {
		log.Fatalf("Invalid device ID - should be an integer in configuration.")
	}
	value := cRead(id, device.Model())
	return &sdk.ReadContext{
		Device:  device.ID(),
		Board:   device.Location().Board,
		Rack:    device.Location().Rack,
		Reading: []*sdk.Reading{{time.Now().String(), device.Type(), value}},
	}, nil
}

func (h *ExamplePluginHandler) Write(device *sdk.Device, data *sdk.WriteData) error {
	return &sdk.UnsupportedCommandError{}
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

// The main function - this is where we will configure, create, and run
// the plugin.
func main() {
	// Set the prototype and device instance config paths to be relative to the
	// current working directory instead of using the default location. This way
	// the plugin can be run from within this directory.
	os.Setenv("PLUGIN_DEVICE_PATH", "./config/device")
	os.Setenv("PLUGIN_PROTO_PATH", "./config/proto")

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

	// Register the Plugin devices.
	err = plugin.RegisterDevices()
	if err != nil {
		log.Fatal(err)
	}

	// Run the plugin.
	err = plugin.Run()
	if err != nil {
		log.Fatal(err)
	}
}

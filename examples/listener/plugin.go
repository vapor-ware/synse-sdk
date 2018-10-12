package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"

	"github.com/vapor-ware/synse-sdk/sdk"
)

var (
	pluginName       = "listener plugin"
	pluginMaintainer = "vaporio"
	pluginDesc       = "An example plugin with listener device"
)

// Output types are defined, either statically in the plugin code, or via YAML
// configuration files. They define the potential outputs of the plugin's devices.
// A single device could support multiple outputs, but at a minimum requires one.
var (
	// The random data coming back from the pusher is random and meaningless,
	// so we don't ascribe any precision or unit to it.
	pusherOutput = sdk.OutputType{
		Name: "push_data",
	}
)

// Device Handlers need to be defined to tell the plugin how to handle reads and
// writes for the different kinds of devices it supports.
var (
	// pusherHandler defines the listen behavior for the "pusher" device kind.
	pusherHandler = sdk.DeviceHandler{
		Name: "pusher",
		Listen: func(device *sdk.Device, data chan *sdk.ReadContext) error {
			// The device data defines the host/port to listen on.
			address := device.Data["address"].(string)

			addr, err := net.ResolveUDPAddr("udp", address)
			if err != nil {
				return err
			}
			conn, err := net.ListenUDP("udp", addr)
			if err != nil {
				return err
			}
			buffer := make([]byte, 4)
			for {
				size, err := conn.Read(buffer)
				if err != nil {
					// failed read, try again
					continue
				}
				if size != 4 {
					// Unexpected packet size, try again
					continue
				}
				value := binary.LittleEndian.Uint32(buffer)
				fmt.Printf("[listener] got data: %v\n", value)
				reading, err := device.GetOutput("push_data").MakeReading(value)
				if err != nil {
					return err
				}
				data <- sdk.NewReadContext(device, []*sdk.Reading{reading})
			}
		},
	}
)

func main() {
	// Set the metadata for the plugin.
	sdk.SetPluginMeta(
		pluginName,
		pluginMaintainer,
		pluginDesc,
		"",
	)

	// Create a new Plugin instance.
	plugin := sdk.NewPlugin()

	// Register our output types with the Plugin.
	err := plugin.RegisterOutputTypes(
		&pusherOutput,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Register our device handlers with the Plugin.
	plugin.RegisterDeviceHandlers(
		&pusherHandler,
	)

	// Run the plugin.
	if err := plugin.Run(); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"encoding/binary"
	"fmt"
	"github.com/vapor-ware/synse-sdk/sdk/config"
	"log"
	"net"

	"github.com/vapor-ware/synse-sdk/sdk"
	"github.com/vapor-ware/synse-sdk/sdk/output"
)

var (
	pluginName       = "listener plugin"
	pluginMaintainer = "vaporio"
	pluginDesc       = "An example plugin with listener device"
)

// ReadingKey is the key into a Device's Data map which holds the latest reading
// value for the device.
const ReadingKey = "reading_data"

// Output types are defined, either statically in the plugin code, or via YAML
// configuration files. They define the potential outputs of the plugin's devices.
// A single device could support multiple outputs, but at a minimum requires one.
var (
	// The random data coming back from the pusher is random and meaningless,
	// so we don't ascribe any precision or unit to it.
	serverOutput = output.Output{
		Name: "random",
		Type: "number",
	}
)


var (
	// Define a PluginAction which subscribes to the data and parses it, assigning the data
	// to the device it is associated with.
	initSubscription = sdk.PluginAction{
		Name: "initiate the metrics stream",
		Action: func(p *sdk.Plugin) error {

			// Connect to the stream. This would generally use data which comes from configuration.
			// This address corresponds with the "data server".
			addr, err := net.ResolveUDPAddr("udp4", "localhost:8553")
			if err != nil {
				return err
			}
			conn, err := net.DialUDP("udp4", nil, addr)
			if err != nil {
				return err
			}

			// Run the collection logic in a goroutine so it does not block the plugin.
			go func() {
				buffer := make([]byte, 4)
				log.Print("[metrics] listening...")
				for {
					size, _, err := conn.ReadFromUDP(buffer)
					log.Print("[metrics] read from udp!")

					if err != nil {
						log.Print("[metrics] failed read, trying again")
						continue
					}
					if size != 4 {
						log.Print("[metrics] unexpected packet size, trying again")
						continue
					}
					value := binary.LittleEndian.Uint32(buffer)
					log.Printf("[metrics] got data: %v\n", value)

					reading := serverOutput.MakeReading(value)

					device, err := p.NewDevice(
						&config.DeviceProto{
							Type: "example",
							Handler: "example-listener",
						},
						&config.DeviceInstance{
							Info: "An example device providing streamed data",
						},
					)
					if err != nil {
						log.Fatal("failed to create device")
					}

					// If the device is not yet registered with the plugin, register it now.
					deviceID := p.GenerateDeviceID(device)
					d := p.GetDevice(deviceID)
					if d == nil  {
						log.Printf("registering new device: %v", deviceID)
						if err := p.AddDevice(d); err != nil {
							log.Fatal(err)
						}
					}
					device = d

					// Add the reading to the Device Data.
					device.Data[ReadingKey] = []*output.Reading{reading}
				}
			}()

			return nil
		},
	}
)


// Device Handlers need to be defined to tell the plugin how to handle reads and
// writes for the different kinds of devices it supports.
var (
	// For our "listener device", the handler just reads from the Device data which gets
	// populated from the out-of-band Action which is doing the actual polling and data
	// parsing.
	deviceHandler = sdk.DeviceHandler{
		Name: "example-listener",
		Read: func(device *sdk.Device) (readings []*output.Reading, e error) {
			r, exists := device.Data[ReadingKey]
			if !exists {
				return nil, fmt.Errorf("no reading associated with device %v", device.GetID())
			}

			readings, ok := r.([]*output.Reading)
			if !ok {
				return  nil,  fmt.Errorf("unexpected device reading data type: %T", r)
			}

			return readings, nil
		},
	}
)


func main() {
	// Set the metadata for the plugin.
	sdk.SetPluginInfo(
		pluginName,
		pluginMaintainer,
		pluginDesc,
		"",
	)

	// Create a new Plugin instance.
	plugin, err := sdk.NewPlugin(
		sdk.DeviceConfigOptional(),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Register our output types with the Plugin.
	err = plugin.RegisterOutputs(
		&serverOutput,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Register our device handlers with the Plugin.
	err = plugin.RegisterDeviceHandlers(
		&deviceHandler,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Register actions with the Plugin.
	plugin.RegisterPreRunActions(
		&initSubscription,
	)

	// Run the plugin.
	if err := plugin.Run(); err != nil {
		log.Fatal(err)
	}
}

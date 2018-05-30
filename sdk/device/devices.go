package device

import (
	"fmt"

	"github.com/vapor-ware/synse-sdk/sdk"
	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/types"
	"github.com/vapor-ware/synse-server-grpc/go"
)

// Device is the internal model for a single device (physical or virtual) that
// a plugin can read to or write from.
type Device struct {
	Location config.Location

	Kind string

	Metadata map[string]string

	Plugin string

	Info string

	Data map[string]interface{}

	Outputs []*Output

	// id is the deterministic id of the device
	id string

	// bulkRead is a flag that determines whether or not the device should be
	// read in bulk, i.e. in a batch with other devices of the same kind.
	bulkRead bool
}

type Output struct {
	types.ReadingType

	Info string
	Data map[string]interface{}
}

func (device *Device) IsReadable() bool {
	// todo
}

func (device *Device) IsWritable() bool {
	// todo
}

func (device *Device) Read() (*sdk.ReadContext, error) {
	if device.IsReadable() {
		// todo
	}
	return nil, fmt.Errorf("reading not supported")
}

func (device *Device) Write(data *sdk.WriteData) error {
	if device.IsWritable() {
		// todo
	}
	return fmt.Errorf("writing not supported")
}

func (device *Device) GUID() string {
	rack, err := device.Location.Rack.Get()
	if err != nil {
		// error... we shouldn't get this because we passed validation though.
	}
	board, err := device.Location.Board.Get()
	if err != nil {
		// error... we shouldn't get this because we passed validation though.
	}
	return makeIDString(rack, board, device.ID())
}

func (device *Device) ID() string {
	if device.id == "" {
		// make the device ID
		// todo: get the plugin-specific identifier (DeviceIdentifier handler, e.g.)
		// todo: use some plugin specific data.. plugin name?
		// todo: use some device-specific data.. device kind (e.g. name)

		device.id = newUID()
	}
	return device.id
}

func (device *Device) encode() *synse.Device {
	var output []*synse.Output
	for _, out := range device.Outputs {
		output = append(output, out.Encode())
	}
	return &synse.Device{
		Timestamp: sdk.GetCurrentTime(),
		Uid:       device.ID(),
		Kind:      device.Kind,
		Metadata:  device.Metadata,
		Plugin:    device.Plugin,
		Info:      device.Info,
		Location:  device.Location.Encode(),
		Output:    output,
	}
}

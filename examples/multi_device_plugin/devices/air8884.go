package devices

import (
	"time"

	"github.com/vapor-ware/synse-sdk/sdk"
)

// Air8884 is the handler for the example "air8884" device model.
type Air8884 struct{}

func (d *Air8884) Read(device *sdk.Device) (*sdk.ReadContext, error) {
	return &sdk.ReadContext{
		Device:  device.UID(),
		Reading: []*sdk.Reading{{time.Now().String(), device.Type(), "100"}},
	}, nil
}

func (d *Air8884) Write(device *sdk.Device, data *sdk.WriteData) error {
	return &sdk.UnsupportedCommandError{}
}

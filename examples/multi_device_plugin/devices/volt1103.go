package devices

import (
	"time"

	"github.com/vapor-ware/synse-sdk/sdk"
)

// Volt1103 is the handler for the example "volt1103" device model.
type Volt1103 struct{}

func (d *Volt1103) Read(device *sdk.Device) (*sdk.ReadContext, error) {
	return &sdk.ReadContext{
		Device:  device.UID(),
		Reading: []*sdk.Reading{{time.Now().String(), device.Type(), "1"}},
	}, nil
}

func (d *Volt1103) Write(device *sdk.Device, data *sdk.WriteData) error {
	return &sdk.UnsupportedCommandError{}
}

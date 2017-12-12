package devices

import (
	"time"

	"github.com/vapor-ware/synse-sdk/sdk"
)

// Temp2010 is the handler for the example "temp2010" device model.
type Temp2010 struct{}

func (d *Temp2010) Read(device *sdk.Device) (*sdk.ReadContext, error) {
	return &sdk.ReadContext{
		Device:  device.UID(),
		Reading: []*sdk.Reading{{time.Now().String(), device.Type(), "10"}},
	}, nil
}

func (d *Temp2010) Write(device *sdk.Device, data *sdk.WriteData) error {
	return &sdk.UnsupportedCommandError{}
}

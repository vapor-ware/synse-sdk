package devices

import (
	"time"

	"../../../sdk"
)


// Air8884 is the handler for the example "air8884" device model.
type Air8884 struct {}

func (d *Air8884) Read(in sdk.Device) (sdk.ReadResource, error) {
	return sdk.ReadResource{
		Device: in.UID(),
		Reading: []sdk.Reading{{time.Now().String(), in.Type(), "100"}},
	}, nil
}

func (d *Air8884) Write(in sdk.Device, data *sdk.WriteData) (error) {
	return nil
}

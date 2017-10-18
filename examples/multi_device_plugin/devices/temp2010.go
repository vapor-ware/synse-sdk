package devices

import (
	"time"

	"../../../sdk"
)


// Temp2010 is the handler for the example "temp2010" device model.
type Temp2010 struct {}

func (d *Temp2010) Read(in sdk.Device) (sdk.ReadResource, error) {
	return sdk.ReadResource{
		Device: in.UID(),
		Reading: []sdk.Reading{{time.Now().String(), in.Type(), "10"}},
	}, nil
}

func (d *Temp2010) Write(in sdk.Device, data *sdk.WriteData) (error) {
	return nil
}

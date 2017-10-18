package devices

import (
	"time"

	"../../../sdk"
)


type Volt1103 struct {}

func (d *Volt1103) Read(in sdk.Device) (sdk.ReadResource, error) {
	return sdk.ReadResource{
		Device: in.Uid(),
		Reading: []sdk.Reading{{time.Now().String(), in.Type(), "1"}},
	}, nil
}

func (d *Volt1103) Write(in sdk.Device, data *sdk.WriteData) (error) {
	return nil
}

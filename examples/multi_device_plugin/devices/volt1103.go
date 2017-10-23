package devices

import (
	"time"

	"github.com/vapor-ware/synse-sdk/sdk"
)


// Volt1103 is the handler for the example "volt1103" device model.
type Volt1103 struct {}

func (d *Volt1103) Read(in sdk.Device) (sdk.ReadResource, error) {
	return sdk.ReadResource{
		Device: in.UID(),
		Reading: []sdk.Reading{{time.Now().String(), in.Type(), "1"}},
	}, nil
}

func (d *Volt1103) Write(in sdk.Device, data *sdk.WriteData) (error) {
	return nil
}

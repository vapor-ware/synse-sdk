package devices

import (
	"github.com/vapor-ware/synse-sdk/sdk"
	"github.com/vapor-ware/synse-sdk/sdk/output"
)

// Temp2010 is the handler for the example temperature device with model "temp2010".
var Temp2010 = sdk.DeviceHandler{
	Name: "temperature",

	Read: func(device *sdk.Device) ([]*output.Reading, error) {
		reading := output.Temperature.MakeReading(10)
		return []*output.Reading{reading}, nil
	},
}

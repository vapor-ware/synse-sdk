package devices

import (
	"github.com/vapor-ware/synse-sdk/sdk"
)

// Temp2010 is the handler for the example temperature device with model "temp2010".
var Temp2010 = sdk.DeviceHandler{
	Name: "temperature",

	Read: func(device *sdk.Device) ([]*sdk.Reading, error) {
		reading, err := device.GetOutput("temperature").MakeReading("10")
		if err != nil {
			return nil, err
		}
		return []*sdk.Reading{reading}, nil
	},
}

package devices

import (
	"github.com/vapor-ware/synse-sdk/v2/sdk"
	"github.com/vapor-ware/synse-sdk/v2/sdk/output"
)

// Volt1103 is the handler for the example voltage device with model "volt1103".
var Volt1103 = sdk.DeviceHandler{
	Name: "voltage",

	Read: func(device *sdk.Device) ([]*output.Reading, error) {
		reading, err := output.Voltage.MakeReading(1)
		if err != nil {
			return nil, err
		}
		return []*output.Reading{reading}, nil
	},
}

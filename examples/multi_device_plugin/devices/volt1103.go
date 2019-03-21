package devices

import (
	"github.com/vapor-ware/synse-sdk/sdk"
	"github.com/vapor-ware/synse-sdk/sdk/output"
)

// Volt1103 is the handler for the example voltage device with model "volt1103".
var Volt1103 = sdk.DeviceHandler{
	Name: "voltage",

	Read: func(device *sdk.Device) ([]*output.Reading, error) {
		reading := output.Voltage.MakeReading(1)
		return []*output.Reading{reading}, nil
	},
}

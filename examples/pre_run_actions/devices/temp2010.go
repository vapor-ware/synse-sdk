package devices

import (
	"github.com/vapor-ware/synse-sdk/sdk"
)

// Temp2010 is the handler for the example "temp2010" device model.
var Temp2010 = sdk.DeviceHandler{
	Type:  "temperature",
	Model: "temp2010",

	Read: func(device *sdk.Device) ([]*sdk.Reading, error) {
		return []*sdk.Reading{
			sdk.NewReading(
				device.Type,
				"10",
			),
		}, nil
	},
}

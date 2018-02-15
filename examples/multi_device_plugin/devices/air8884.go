package devices

import (
	"github.com/vapor-ware/synse-sdk/sdk"
)

// Air8884 is the handler for the example "air8884" device model.
var Air8884 = sdk.DeviceHandler{
	Type:  "airflow",
	Model: "air8884",

	Read: func(device *sdk.Device) ([]*sdk.Reading, error) {
		return []*sdk.Reading{
			sdk.NewReading(
				device.Type,
				"100",
			),
		}, nil
	},
}

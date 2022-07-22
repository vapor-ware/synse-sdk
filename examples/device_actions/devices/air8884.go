package devices

import (
	"github.com/vapor-ware/synse-sdk/v2/sdk"
	"github.com/vapor-ware/synse-sdk/v2/sdk/output"
)

// Air8884 is the handler for the example airflow device with model "air8884".
var Air8884 = sdk.DeviceHandler{
	Name: "airflow",

	Read: func(device *sdk.Device) ([]*output.Reading, error) {
		reading, err := output.Status.MakeReading("100")
		if err != nil {
			return nil, err
		}
		return []*output.Reading{reading}, nil
	},
}

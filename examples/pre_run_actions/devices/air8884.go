package devices

import (
	"github.com/vapor-ware/synse-sdk/sdk"
)

// Air8884 is the handler for the example airflow device with model "air8884".
var Air8884 = sdk.DeviceHandler{
	Name: "airflow",

	Read: func(device *sdk.Device) ([]*sdk.Reading, error) {
		return []*sdk.Reading{
			device.GetOutput("airflow").MakeReading("100"),
		}, nil
	},
}

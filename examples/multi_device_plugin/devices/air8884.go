package devices

import (
	"github.com/vapor-ware/synse-sdk/v2/examples/multi_device_plugin/outputs"
	"github.com/vapor-ware/synse-sdk/v2/sdk"
	"github.com/vapor-ware/synse-sdk/v2/sdk/output"
)

// Air8884 is the handler for the example airflow device with model "air8884".
var Air8884 = sdk.DeviceHandler{
	Name: "airflow",

	Read: func(device *sdk.Device) ([]*output.Reading, error) {
		reading, err := outputs.AirflowOutput.MakeReading(100)
		if err != nil {
			return nil, err
		}
		return []*output.Reading{reading}, nil
	},
	Write: func(device *sdk.Device, data *sdk.WriteData) error {
		// Defines a write function which does nothing -- this is just so
		// the example plugin has a writable device, which may be useful
		// for debugging/testing.
		return nil
	},
}

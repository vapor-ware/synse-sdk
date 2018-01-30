package devices

import (
	"time"

	"github.com/vapor-ware/synse-sdk/sdk"
)

// Volt1103 is the handler for the example "volt1103" device model.
var Volt1103 = sdk.DeviceHandler{
	Type:  "voltage",
	Model: "volt1103",

	Read: func(device *sdk.Device) ([]*sdk.Reading, error) {
		return []*sdk.Reading{{
			Timestamp: time.Now().String(),
			Type:      device.Type,
			Value:     "1",
		}}, nil
	},
}

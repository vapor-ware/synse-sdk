package devices

import (
	"github.com/vapor-ware/synse-sdk/sdk"
)

// DeviceInterface is the interface that all of the device model handlers
// should fulfil. The gRPC read/write commands get routed through these
// functions appropriately for the given device.
type DeviceInterface interface {
	Read(*sdk.Device) (*sdk.ReadContext, error)
	Write(*sdk.Device, *sdk.WriteData) error
}

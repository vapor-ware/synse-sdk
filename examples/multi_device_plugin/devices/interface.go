package devices

import (
	"../../../sdk"
)


// DeviceInterface is the interface that all of the device model handlers
// should fulfil. The gRPC read/write commands get routed through these
// functions appropriately for the given device.
type DeviceInterface interface {
	Read(sdk.Device) (sdk.ReadResource, error)
	Write(sdk.Device, *sdk.WriteData) (error)
}
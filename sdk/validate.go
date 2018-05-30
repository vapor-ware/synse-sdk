package sdk

import (
	"fmt"

	"github.com/vapor-ware/synse-server-grpc/go"
)

// validateReadRequest checks to make sure that a ReadRequest has all of the
// fields populated that we need in order to process it as a valid request.
func validateReadRequest(request *synse.DeviceFilter) error {
	device := request.GetDevice()
	if device == "" {
		return invalidArgumentErr("no device UID supplied to Read")
	}
	board := request.GetBoard()
	if board == "" {
		return invalidArgumentErr("no board supplied to Read")
	}
	rack := request.GetRack()
	if rack == "" {
		return invalidArgumentErr("no rack supplied to Read")
	}
	return nil
}

// validateWriteRequest checks to make sure that a ReadRequest has all of the
// fields populated that we need in order to process it as a valid request.
func validateWriteRequest(request *synse.WriteInfo) error {
	device := request.DeviceFilter.GetDevice()
	if device == "" {
		return invalidArgumentErr("no device UID supplied to Write")
	}
	board := request.DeviceFilter.GetBoard()
	if board == "" {
		return invalidArgumentErr("no board supplied to Write")
	}
	rack := request.DeviceFilter.GetRack()
	if rack == "" {
		return invalidArgumentErr("no rack supplied to Write")
	}
	return nil
}

// validateForRead validates that a device with the given device ID is readable.
func validateForRead(deviceID string) error {
	device := deviceMap[deviceID]
	if device == nil {
		return fmt.Errorf("no device with ID %v found", deviceID)
	}

	if !device.IsReadable() {
		return fmt.Errorf("reading not enabled for device %v (no read handler)", deviceID)
	}

	return nil
}

// validateForWrite validates that a device with the given device ID is writable.
func validateForWrite(deviceID string) error {
	device := deviceMap[deviceID]
	if device == nil {
		return fmt.Errorf("no device with ID %v found", deviceID)
	}

	if !device.IsWritable() {
		return fmt.Errorf("writing not enabled for device %v (no write handler)", deviceID)
	}

	return nil
}

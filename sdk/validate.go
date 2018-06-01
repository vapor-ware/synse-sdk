package sdk

import (
	"fmt"

	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-server-grpc/go"
)

// validateReadRequest checks to make sure that a ReadRequest has all of the
// fields populated that we need in order to process it as a valid request.
func validateReadRequest(request *synse.DeviceFilter) error {
	if request.GetDevice() == "" {
		return errors.InvalidArgumentErr("no device UID supplied to Read")
	}
	if request.GetBoard() == "" {
		return errors.InvalidArgumentErr("no board supplied to Read")
	}
	if request.GetRack() == "" {
		return errors.InvalidArgumentErr("no rack supplied to Read")
	}
	return nil
}

// validateWriteRequest checks to make sure that a ReadRequest has all of the
// fields populated that we need in order to process it as a valid request.
func validateWriteRequest(request *synse.WriteInfo) error {
	if request.DeviceFilter.GetDevice() == "" {
		return errors.InvalidArgumentErr("no device UID supplied to Write")
	}
	if request.DeviceFilter.GetBoard() == "" {
		return errors.InvalidArgumentErr("no board supplied to Write")
	}
	if request.DeviceFilter.GetRack() == "" {
		return errors.InvalidArgumentErr("no rack supplied to Write")
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

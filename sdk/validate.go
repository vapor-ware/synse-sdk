package sdk

//// validateDeviceFilter checks to make sure that a DeviceFilter has all of the
//// fields populated that we need in order to process it as a valid request.
//func validateDeviceFilter(request *synse.DeviceFilter) error {
//	if request.GetDevice() == "" {
//		return errors.InvalidArgumentErr("no device UID supplied in request")
//	}
//	if request.GetBoard() == "" {
//		return errors.InvalidArgumentErr("no board supplied in request")
//	}
//	if request.GetRack() == "" {
//		return errors.InvalidArgumentErr("no rack supplied in request")
//	}
//	return nil
//}

//// validateWriteInfo checks to make sure that a WriteInfo has all of the
//// fields populated that we need in order to process it as a valid request.
//func validateWriteInfo(request *synse.WriteInfo) error {
//	return validateDeviceFilter(request.DeviceFilter)
//}

//// validateForRead validates that a device with the given device ID is readable.
//func validateForRead(deviceID string) error {
//	device := ctx.devices[deviceID]
//	if device == nil {
//		return fmt.Errorf("no device found with ID %s", deviceID)
//	}
//
//	if !device.IsReadable() {
//		return fmt.Errorf("reading not enabled for device %s (no read handler)", deviceID)
//	}
//	return nil
//}
//
//// validateForWrite validates that a device with the given device ID is writable.
//func validateForWrite(deviceID string) error {
//	device := ctx.devices[deviceID]
//	if device == nil {
//		return fmt.Errorf("no device found with ID %s", deviceID)
//	}
//
//	if !device.IsWritable() {
//		return fmt.Errorf("writing not enabled for device %s (no write handler)", deviceID)
//	}
//	return nil
//}

// Synse SDK
// Copyright (c) 2019 Vapor IO
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package sdk

import "fmt"

// DeviceHandler specifies the read and write handlers for a Device
// based on its type and model.
type DeviceHandler struct {

	// Name is the name of the handler. This is how the handler will be referenced
	// and associated with Device instances via their configuration. This name should
	// match with the "Handler" configuration field.
	Name string

	// Write is a function that handles Write requests for the handler's devices. If
	// the devices do not support writing, this can be left unspecified.
	Write func(*Device, *WriteData) error

	// Read is a function that handles Read requests for the handler's devices. If the
	// devices do not support reading, this can be left unspecified.
	Read func(*Device) ([]*Reading, error)

	// BulkRead is a function that handles bulk read operations for the handler's devices.
	// A bulk read is where all devices of a given kind are read at once, instead of individually.
	// If a device does not support bulk read, this can be left as nil. Additionally,
	// a device can only be bulk read if there is no Read handler set.
	BulkRead func([]*Device) ([]*ReadContext, error)

	// Listen is a function that will listen for push-based data from the device.
	// This function is called one per device using the handler, even if there are
	// other handler functions (e.g. read, write) defined. The listener function
	// will run in a separate goroutine for each device. The goroutines are started
	// before the read/write loops.
	Listen func(*Device, chan *ReadContext) error
}

// GetDevices gets all of the devices that use this handler.
//
// If the DeviceManager is not initialized or contains no devices, this
// returns an empty slice.
func (handler *DeviceHandler) GetDevices() []*Device {
	return DeviceManager.GetDevicesForHandler(handler.Name)
}

// supportsBulkRead checks if the handler supports bulk reading for its Devices.
//
// If BulkRead is set for the device handler and Read is not, then the handler
// supports bulk reading. If both BulkRead and Read are defined, bulk reading
// will not be considered supported and the handler will default to individual
// reads.
func (handler *DeviceHandler) supportsBulkRead() bool {
	return handler.Read == nil && handler.BulkRead != nil
}

//// getDevicesForHandler gets a list of all the devices which use the DeviceHandler.
//func (deviceHandler *DeviceHandler) getDevicesForHandler() []*Device {
//	var devices []*Device
//
//	for _, v := range ctx.devices {
//		if v.Handler == deviceHandler {
//			devices = append(devices, v)
//		}
//	}
//	return devices
//}

// getHandlerForDevice gets the DeviceHandler for a device, based on the handler name.
func getHandlerForDevice(handlerName string) (*DeviceHandler, error) {
	for _, handler := range ctx.deviceHandlers {
		if handler.Name == handlerName {
			return handler, nil
		}
	}
	return nil, fmt.Errorf("no handler found with name: %s", handlerName)
}

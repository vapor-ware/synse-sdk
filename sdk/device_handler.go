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

import "github.com/vapor-ware/synse-sdk/sdk/output"

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
	Read func(*Device) ([]*output.Reading, error)

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

	// Actions specifies a list of the supported write actions for the handler.
	// This is optional and is just used as metadata surfaced by the SDK to the
	// client via the gRPC API.
	Actions []string
}

// CanRead returns true if the handler has a read function defined; false otherwise.
func (handler *DeviceHandler) CanRead() bool {
	if handler == nil {
		return false
	}
	return handler.Read != nil
}

// CanBulkRead returns true if the handler has a bulk read function defined and no
// regular read function defined (both cannot coexist); false otherwise.
func (handler *DeviceHandler) CanBulkRead() bool {
	if handler == nil {
		return false
	}
	// Can only bulk read if no read handler is defined.
	return handler.Read == nil && handler.BulkRead != nil
}

// CanWrite returns true if the handler has a write function defined; false otherwise.
func (handler *DeviceHandler) CanWrite() bool {
	if handler == nil {
		return false
	}
	return handler.Write != nil
}

// CanListen returns true if the handler has a listen function defined; false otherwise.
func (handler *DeviceHandler) CanListen() bool {
	if handler == nil {
		return false
	}
	return handler.Listen != nil
}

// GetCapabilitiesMode gets the capabilities mode string representation for a device
// based on its device handler. This will be one of: "r" (read-only), "w" (write-only),
// or "rw" (read-write).
//
// Note that a device is considered readable here if it can supply reading data.
// Currently, Read, BulkRead, and Listen can all supply reading data, so if any one
// of those are defined for the handler, the capabilities string will reflect that.
func (handler *DeviceHandler) GetCapabilitiesMode() string {
	var capabilities = ""
	if handler.CanRead() || handler.CanBulkRead() || handler.CanListen() {
		capabilities += "r"
	}
	if handler.CanWrite() {
		capabilities += "w"
	}
	return capabilities
}

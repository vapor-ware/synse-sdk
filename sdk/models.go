// Synse SDK
// Copyright (c) 2017-2020 Vapor IO
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

import (
	"github.com/vapor-ware/synse-sdk/v2/sdk/output"
	synse "github.com/vapor-ware/synse-server-grpc/go"
)

// ReadContext provides the context for a device reading. This context
// identifies the device being read and associates it with a set of readings
// at a given time.
//
// A single device can provide more than one reading (e.g. a humidity sensor
// could provide both a humidity and temperature reading). To accommodate, the
// ReadContext allows for multiple readings to be associated with the device.
// Note that the collection of readings in a single ReadContext would correspond
// to a single Read request.
type ReadContext struct {
	Device  *Device
	Reading []*output.Reading
}

// NewReadContext creates a new instance of a ReadContext from the given
// device and corresponding readings.
func NewReadContext(device *Device, readings []*output.Reading) *ReadContext {
	return &ReadContext{
		Device:  device,
		Reading: readings,
	}
}

// WriteContext describes a single write transaction.
type WriteContext struct {
	transaction *transaction
	device      *Device
	data        *synse.V3WriteData
}

// WriteData is an SDK alias for the Synse gRPC WriteData. This is done to
// make writing new plugins easier.
type WriteData synse.V3WriteData

// encode translates the WriteData to a corresponding gRPC WriteData.
func (w *WriteData) encode() *synse.V3WriteData {
	return &synse.V3WriteData{
		Data:   w.Data,
		Action: w.Action,
	}
}

// decodeWriteData decodes the gRPC WriteData to the SDK WriteData.
func decodeWriteData(data *synse.V3WriteData) *WriteData {
	return &WriteData{
		Data:   data.Data,
		Action: data.Action,
	}
}

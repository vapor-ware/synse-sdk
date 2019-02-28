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

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/sdk/output"
)

func TestDeviceHandler_CanRead_nil(t *testing.T) {
	var handler *DeviceHandler
	assert.False(t, handler.CanRead())
}

func TestDeviceHandler_CanRead_true(t *testing.T) {
	handler := DeviceHandler{
		Read: func(device *Device) (readings []*output.Reading, e error) {
			return nil, nil
		},
	}
	assert.True(t, handler.CanRead())
}

func TestDeviceHandler_CanRead_false(t *testing.T) {
	handler := DeviceHandler{}
	assert.False(t, handler.CanRead())
}

func TestDeviceHandler_CanWrite_nil(t *testing.T) {
	var handler *DeviceHandler
	assert.False(t, handler.CanWrite())
}

func TestDeviceHandler_CanWrite_true(t *testing.T) {
	handler := DeviceHandler{
		Write: func(device *Device, data *WriteData) error {
			return nil
		},
	}
	assert.True(t, handler.CanWrite())
}

func TestDeviceHandler_CanWrite_false(t *testing.T) {
	handler := DeviceHandler{}
	assert.False(t, handler.CanWrite())
}

func TestDeviceHandler_CanListen_nil(t *testing.T) {
	var handler *DeviceHandler
	assert.False(t, handler.CanListen())
}

func TestDeviceHandler_CanListen_true(t *testing.T) {
	handler := DeviceHandler{
		Listen: func(device *Device, contexts chan *ReadContext) error {
			return nil
		},
	}
	assert.True(t, handler.CanListen())
}

func TestDeviceHandler_CanListen_false(t *testing.T) {
	handler := DeviceHandler{}
	assert.False(t, handler.CanListen())
}

func TestDeviceHandler_CanBulkRead_nil(t *testing.T) {
	var handler *DeviceHandler
	assert.False(t, handler.CanBulkRead())
}

func TestDeviceHandler_CanBulkRead_true(t *testing.T) {
	handler := DeviceHandler{
		BulkRead: func(devices []*Device) (contexts []*ReadContext, e error) {
			return nil, nil
		},
	}
	assert.True(t, handler.CanBulkRead())
}

func TestDeviceHandler_CanBulkRead_false(t *testing.T) {
	handler := DeviceHandler{}
	assert.False(t, handler.CanBulkRead())
}

func TestDeviceHandler_CanBulkRead_false2(t *testing.T) {
	handler := DeviceHandler{
		Read: func(device *Device) (readings []*output.Reading, e error) {
			return nil, nil
		},
		BulkRead: func(devices []*Device) (contexts []*ReadContext, e error) {
			return nil, nil
		},
	}
	assert.False(t, handler.CanBulkRead())
}

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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/output"
)

func TestNewDeviceFromConfig_nil1(t *testing.T) {
	d, err := NewDeviceFromConfig(nil, &config.DeviceInstance{})
	assert.Error(t, err)
	assert.Nil(t, d)
}

func TestNewDeviceFromConfig_nil2(t *testing.T) {
	d, err := NewDeviceFromConfig(&config.DeviceProto{}, nil)
	assert.Error(t, err)
	assert.Nil(t, d)
}

func TestNewDeviceFromConfig_nil3(t *testing.T) {
	d, err := NewDeviceFromConfig(nil, nil)
	assert.Error(t, err)
	assert.Nil(t, d)
}

func TestNewDeviceFromConfig(t *testing.T) {
	// Tests creating a device where inheritance is enabled, but the
	// instance defines all inheritable things.
	proto := &config.DeviceProto{
		Type: "type1",
		Metadata: map[string]string{
			"a": "b",
		},
		Data: map[string]interface{}{
			"port": 5000,
		},
		Tags:         []string{"default/foo"},
		Handler:      "testhandler",
		WriteTimeout: 3 * time.Second,
	}
	instance := &config.DeviceInstance{
		Type: "type2",
		Info: "testdata",
		Tags: []string{"vapor/io"},
		Data: map[string]interface{}{
			"address": "localhost",
		},
		Output:    "temperature",
		SortIndex: 1,
		Handler:   "testhandler2",
		Alias: &config.DeviceAlias{
			Name: "foo",
		},
		ScalingFactor:      "2",
		WriteTimeout:       5 * time.Second,
		DisableInheritance: false,
	}

	device, err := NewDeviceFromConfig(proto, instance)
	assert.NoError(t, err)
	assert.Equal(t, "type2", device.Type)
	assert.Equal(t, map[string]string{"a": "b"}, device.Metadata)
	assert.Equal(t, "testdata", device.Info)
	assert.Equal(t, 2, len(device.Tags))
	assert.Equal(t, map[string]interface{}{"address": "localhost", "port": 5000}, device.Data)
	assert.Equal(t, "testhandler2", device.Handler)
	assert.Equal(t, int32(1), device.SortIndex)
	assert.Equal(t, "foo", device.Alias)
	assert.Equal(t, "2", device.ScalingFactor)
	assert.Equal(t, 5*time.Second, device.WriteTimeout)
	assert.Equal(t, "temperature", device.Output)
	assert.Equal(t, 0, len(device.fns))
}

func TestNewDeviceFromConfig2(t *testing.T) {
	// Tests creating a device where inheritance is enabled, and the instance will
	// inherit values from the prototype.
	proto := &config.DeviceProto{
		Type: "type1",
		Metadata: map[string]string{
			"a": "b",
		},
		Data: map[string]interface{}{
			"port": 5000,
		},
		Tags:         []string{"default/foo"},
		Handler:      "testhandler",
		WriteTimeout: 3 * time.Second,
	}
	instance := &config.DeviceInstance{
		Info: "testdata",
		Tags: []string{"vapor/io"},
		Data: map[string]interface{}{
			"address": "localhost",
		},
		Output:    "temperature",
		Apply:     []string{"FtoC"},
		SortIndex: 1,
		Alias: &config.DeviceAlias{
			Name: "foo",
		},
		ScalingFactor:      "2",
		DisableInheritance: false,
	}

	device, err := NewDeviceFromConfig(proto, instance)
	assert.NoError(t, err)
	assert.Equal(t, "type1", device.Type)
	assert.Equal(t, map[string]string{"a": "b"}, device.Metadata)
	assert.Equal(t, "testdata", device.Info)
	assert.Equal(t, 2, len(device.Tags))
	assert.Equal(t, map[string]interface{}{"address": "localhost", "port": 5000}, device.Data)
	assert.Equal(t, "testhandler", device.Handler)
	assert.Equal(t, int32(1), device.SortIndex)
	assert.Equal(t, "foo", device.Alias)
	assert.Equal(t, "2", device.ScalingFactor)
	assert.Equal(t, 3*time.Second, device.WriteTimeout)
	assert.Equal(t, "temperature", device.Output)
	assert.Equal(t, 1, len(device.fns))
}

func TestNewDeviceFromConfig3(t *testing.T) {
	// Test when no type is resolved, resulting in error.
	proto := &config.DeviceProto{
		Metadata: map[string]string{
			"a": "b",
		},
		Data: map[string]interface{}{
			"port": 5000,
		},
		Tags:         []string{"default/foo"},
		Handler:      "testhandler",
		WriteTimeout: 3 * time.Second,
	}
	instance := &config.DeviceInstance{
		Info: "testdata",
		Tags: []string{"vapor/io"},
		Data: map[string]interface{}{
			"address": "localhost",
		},
		SortIndex: 1,
		Handler:   "testhandler2",
		Alias: &config.DeviceAlias{
			Name: "foo",
		},
		ScalingFactor:      "2",
		WriteTimeout:       5 * time.Second,
		DisableInheritance: false,
	}

	device, err := NewDeviceFromConfig(proto, instance)
	assert.Error(t, err)
	assert.Nil(t, device)
}

func TestNewDeviceFromConfig4(t *testing.T) {
	// Test inheriting from prototype when there is nothing to inherit.
	proto := &config.DeviceProto{
		Metadata: map[string]string{
			"a": "b",
		},
		Data: map[string]interface{}{
			"port": 5000,
		},
		Tags: []string{"default/foo"},
	}
	instance := &config.DeviceInstance{
		Type: "type2",
		Info: "testdata",
		Tags: []string{"vapor/io"},
		Data: map[string]interface{}{
			"address": "localhost",
		},
		SortIndex: 1,
		Alias: &config.DeviceAlias{
			Name: "foo",
		},
		ScalingFactor:      "2",
		DisableInheritance: false,
	}

	device, err := NewDeviceFromConfig(proto, instance)
	assert.NoError(t, err)
	assert.Equal(t, "type2", device.Type)
	assert.Equal(t, map[string]string{"a": "b"}, device.Metadata)
	assert.Equal(t, "testdata", device.Info)
	assert.Equal(t, 2, len(device.Tags))
	assert.Equal(t, map[string]interface{}{"address": "localhost", "port": 5000}, device.Data)
	assert.Equal(t, "", device.Handler)
	assert.Equal(t, int32(1), device.SortIndex)
	assert.Equal(t, "foo", device.Alias)
	assert.Equal(t, "2", device.ScalingFactor)
	assert.Equal(t, 30*time.Second, device.WriteTimeout) // takes the default value
	assert.Equal(t, "", device.Output)
	assert.Equal(t, 0, len(device.fns))
}

func TestNewDeviceFromConfig5(t *testing.T) {
	// Test disabling inheritance when there are inheritable values.
	proto := &config.DeviceProto{
		Type: "type1",
		Metadata: map[string]string{
			"a": "b",
		},
		Data: map[string]interface{}{
			"port": 5000,
		},
		Tags:         []string{"default/foo"},
		Handler:      "testhandler",
		WriteTimeout: 3 * time.Second,
	}
	instance := &config.DeviceInstance{
		Type: "type2",
		Info: "testdata",
		Tags: []string{"vapor/io"},
		Data: map[string]interface{}{
			"address": "localhost",
		},
		SortIndex: 1,
		Alias: &config.DeviceAlias{
			Name: "foo",
		},
		ScalingFactor:      "2",
		DisableInheritance: true,
	}

	device, err := NewDeviceFromConfig(proto, instance)
	assert.NoError(t, err)
	assert.Equal(t, "type2", device.Type)
	assert.Equal(t, map[string]string{"a": "b"}, device.Metadata)
	assert.Equal(t, "testdata", device.Info)
	assert.Equal(t, 1, len(device.Tags))
	assert.Equal(t, map[string]interface{}{"address": "localhost"}, device.Data)
	assert.Equal(t, "", device.Handler)
	assert.Equal(t, int32(1), device.SortIndex)
	assert.Equal(t, "foo", device.Alias)
	assert.Equal(t, "2", device.ScalingFactor)
	assert.Equal(t, 30*time.Second, device.WriteTimeout) // takes the default value
	assert.Equal(t, "", device.Output)
	assert.Equal(t, 0, len(device.fns))
}

func TestNewDeviceFromConfig6(t *testing.T) {
	// Bad tags specified.
	proto := &config.DeviceProto{
		Type: "type1",
		Metadata: map[string]string{
			"a": "b",
		},
		Data: map[string]interface{}{
			"port": 5000,
		},
		Tags:         []string{"default/foo:"},
		Handler:      "testhandler",
		WriteTimeout: 3 * time.Second,
	}
	instance := &config.DeviceInstance{
		Type: "type2",
		Info: "testdata",
		Tags: []string{"vapor/dot io"},
		Data: map[string]interface{}{
			"address": "localhost",
		},
		SortIndex: 1,
		Handler:   "testhandler2",
		Alias: &config.DeviceAlias{
			Name: "foo",
		},
		ScalingFactor:      "2",
		WriteTimeout:       5 * time.Second,
		DisableInheritance: false,
	}

	device, err := NewDeviceFromConfig(proto, instance)
	assert.Error(t, err)
	assert.Nil(t, device)
}

func TestNewDeviceFromConfig7(t *testing.T) {
	// Bad alias specified.
	proto := &config.DeviceProto{
		Type: "type1",
		Metadata: map[string]string{
			"a": "b",
		},
		Data: map[string]interface{}{
			"port": 5000,
		},
		Tags:         []string{"default/foo"},
		Handler:      "testhandler",
		WriteTimeout: 3 * time.Second,
	}
	instance := &config.DeviceInstance{
		Type: "type2",
		Info: "testdata",
		Tags: []string{"vapor/io"},
		Data: map[string]interface{}{
			"address": "localhost",
		},
		SortIndex: 1,
		Handler:   "testhandler2",
		Alias: &config.DeviceAlias{
			Template: "foo.{{.NotAField}}",
		},
		ScalingFactor:      "2",
		WriteTimeout:       5 * time.Second,
		DisableInheritance: false,
	}

	device, err := NewDeviceFromConfig(proto, instance)
	assert.Error(t, err)
	assert.Nil(t, device)
}

func TestNewDeviceFromConfig8(t *testing.T) {
	// Fail data map merging
	proto := &config.DeviceProto{
		Type: "type1",
		Metadata: map[string]string{
			"a": "b",
		},
		Data: map[string]interface{}{
			"address": 1234,
			"port":    5000,
		},
		Tags:         []string{"default/foo"},
		Handler:      "testhandler",
		WriteTimeout: 3 * time.Second,
	}
	instance := &config.DeviceInstance{
		Type: "type2",
		Info: "testdata",
		Tags: []string{"vapor/io"},
		Data: map[string]interface{}{
			"address": "localhost",
			"port":    []int{5000},
		},
		SortIndex: 1,
		Handler:   "testhandler2",
		Alias: &config.DeviceAlias{
			Name: "foo",
		},
		ScalingFactor:      "2",
		WriteTimeout:       5 * time.Second,
		DisableInheritance: false,
	}

	device, err := NewDeviceFromConfig(proto, instance)
	assert.Error(t, err)
	assert.Nil(t, device)
}

func TestNewDeviceFromConfig9(t *testing.T) {
	// Unknown output type specified
	proto := &config.DeviceProto{
		Type: "type1",
		Metadata: map[string]string{
			"a": "b",
		},
		Data: map[string]interface{}{
			"port": 5000,
		},
		Tags:         []string{"default/foo"},
		Handler:      "testhandler",
		WriteTimeout: 3 * time.Second,
	}
	instance := &config.DeviceInstance{
		Type: "type2",
		Info: "testdata",
		Tags: []string{"vapor/io"},
		Data: map[string]interface{}{
			"address": "localhost",
		},
		SortIndex:          1,
		Handler:            "testhandler2",
		Output:             "unknown-output-name",
		ScalingFactor:      "2",
		WriteTimeout:       5 * time.Second,
		DisableInheritance: false,
	}

	device, err := NewDeviceFromConfig(proto, instance)
	assert.Error(t, err)
	assert.Nil(t, device)
}

func TestNewDeviceFromConfig10(t *testing.T) {
	// Unknown transformation function specified
	proto := &config.DeviceProto{
		Type: "type1",
		Metadata: map[string]string{
			"a": "b",
		},
		Data: map[string]interface{}{
			"port": 5000,
		},
		Tags:         []string{"default/foo"},
		Handler:      "testhandler",
		WriteTimeout: 3 * time.Second,
	}
	instance := &config.DeviceInstance{
		Type: "type2",
		Info: "testdata",
		Tags: []string{"vapor/io"},
		Data: map[string]interface{}{
			"address": "localhost",
		},
		SortIndex:          1,
		Handler:            "testhandler2",
		Apply:              []string{"unknown-fn"},
		ScalingFactor:      "2",
		WriteTimeout:       5 * time.Second,
		DisableInheritance: false,
	}

	device, err := NewDeviceFromConfig(proto, instance)
	assert.Error(t, err)
	assert.Nil(t, device)
}

func TestDevice_setAlias_noConf(t *testing.T) {
	device := Device{}

	err := device.setAlias(nil)
	assert.NoError(t, err)
	assert.Equal(t, "", device.Alias)
}

func TestDevice_setAlias_emptyConf(t *testing.T) {
	device := Device{}

	err := device.setAlias(&config.DeviceAlias{})
	assert.NoError(t, err)
	assert.Equal(t, "", device.Alias)
}

func TestDevice_setAlias_hasName(t *testing.T) {
	device := Device{}

	err := device.setAlias(&config.DeviceAlias{
		Name: "foo",
	})
	assert.NoError(t, err)
	assert.Equal(t, "foo", device.Alias)
}

func TestDevice_setAlias_hasBadTemplate(t *testing.T) {
	device := Device{}

	err := device.setAlias(&config.DeviceAlias{
		Template: "{{{{",
	})
	assert.Error(t, err)
	assert.Equal(t, "", device.Alias)
}

func TestDevice_setAlias_templateExecuteError(t *testing.T) {
	device := Device{}

	err := device.setAlias(&config.DeviceAlias{
		Template: "{{.NotAField}}",
	})
	assert.Error(t, err)
	assert.Equal(t, "", device.Alias)
}

func TestDevice_setAlias_templateOk(t *testing.T) {
	device := Device{
		Type: "testtype",
	}

	err := device.setAlias(&config.DeviceAlias{
		Template: "{{.Device.Type}}",
	})
	assert.NoError(t, err)
	assert.Equal(t, "testtype", device.Alias)
}

func TestDevice_GetMetadata(t *testing.T) {
	device := Device{
		Metadata: map[string]string{
			"foo": "bar",
			"abc": "xyz",
		},
	}

	assert.Equal(t, "", device.GetMetadata("vapor"))
	assert.Equal(t, "bar", device.GetMetadata("foo"))
	assert.Equal(t, "xyz", device.GetMetadata("abc"))
}

func TestDevice_GetHandler(t *testing.T) {
	device := Device{}
	assert.Nil(t, device.GetHandler())
}

func TestDevice_GetHandler2(t *testing.T) {
	handler := DeviceHandler{}
	device := Device{handler: &handler}
	assert.Equal(t, &handler, device.GetHandler())
}

func TestDevice_GetID(t *testing.T) {
	device := Device{}
	assert.Equal(t, "", device.GetID())
}

func TestDevice_GetID2(t *testing.T) {
	device := Device{id: "1234"}
	assert.Equal(t, "1234", device.GetID())
}

func TestDevice_Read_nilDevice(t *testing.T) {
	device := new(Device)
	ctx, err := device.Read()
	assert.Error(t, err)
	assert.Nil(t, ctx)
	assert.IsType(t, &errors.UnsupportedCommandError{}, err)
}

func TestDevice_Read_notReadable(t *testing.T) {
	device := Device{
		handler: &DeviceHandler{},
	}

	ctx, err := device.Read()
	assert.Error(t, err)
	assert.Nil(t, ctx)
	assert.IsType(t, &errors.UnsupportedCommandError{}, err)
}

func TestDevice_Read_ok(t *testing.T) {
	device := Device{
		handler: &DeviceHandler{
			Read: func(device *Device) (readings []*output.Reading, e error) {
				return []*output.Reading{{Value: 1}}, nil
			},
		},
	}

	ctx, err := device.Read()
	assert.NoError(t, err)
	assert.NotNil(t, ctx)
	assert.Len(t, ctx.Reading, 1)
}

func TestDevice_Read_err(t *testing.T) {
	device := Device{
		handler: &DeviceHandler{
			Read: func(device *Device) (readings []*output.Reading, e error) {
				return nil, fmt.Errorf("test error")
			},
		},
	}

	ctx, err := device.Read()
	assert.Error(t, err)
	assert.Nil(t, ctx)
}

func TestDevice_Write_nilDevice(t *testing.T) {
	device := new(Device)
	err := device.Write(&WriteData{})
	assert.Error(t, err)
	assert.IsType(t, &errors.UnsupportedCommandError{}, err)
}

func TestDevice_Write_notWritable(t *testing.T) {
	device := Device{
		handler: &DeviceHandler{},
	}

	err := device.Write(&WriteData{})
	assert.Error(t, err)
	assert.IsType(t, &errors.UnsupportedCommandError{}, err)
}

func TestDevice_Write_ok(t *testing.T) {
	device := Device{
		handler: &DeviceHandler{
			Write: func(device *Device, data *WriteData) error {
				return nil
			},
		},
	}

	err := device.Write(&WriteData{})
	assert.NoError(t, err)
}

func TestDevice_Write_err(t *testing.T) {
	device := Device{
		handler: &DeviceHandler{
			Write: func(device *Device, data *WriteData) error {
				return fmt.Errorf("test error")
			},
		},
	}

	err := device.Write(&WriteData{})
	assert.Error(t, err)
}

func TestDevice_IsReadable_trueReadHandler(t *testing.T) {
	device := Device{
		handler: &DeviceHandler{
			Read: func(device *Device) (readings []*output.Reading, e error) {
				return nil, nil
			},
		},
	}
	assert.True(t, device.IsReadable())
}

func TestDevice_IsReadable_trueBulkReadHandler(t *testing.T) {
	device := Device{
		handler: &DeviceHandler{
			BulkRead: func(devices []*Device) (contexts []*ReadContext, e error) {
				return nil, nil
			},
		},
	}
	assert.True(t, device.IsReadable())
}

func TestDevice_IsReadable_trueListenHandler(t *testing.T) {
	device := Device{
		handler: &DeviceHandler{
			Listen: func(device *Device, contexts chan *ReadContext) error {
				return nil
			},
		},
	}
	assert.True(t, device.IsReadable())
}

func TestDevice_IsReadable_falseNil(t *testing.T) {
	var device *Device
	assert.False(t, device.IsReadable())
}

func TestDevice_IsReadable_false(t *testing.T) {
	device := Device{
		handler: &DeviceHandler{},
	}
	assert.False(t, device.IsReadable())
}

func TestDevice_IsWritable_true(t *testing.T) {
	device := Device{
		handler: &DeviceHandler{
			Write: func(device *Device, data *WriteData) error {
				return nil
			},
		},
	}
	assert.True(t, device.IsWritable())
}

func TestDevice_IsWritable_falseNil(t *testing.T) {
	var device *Device
	assert.False(t, device.IsWritable())
}

func TestDevice_IsWritable_false(t *testing.T) {
	device := Device{
		handler: &DeviceHandler{},
	}
	assert.False(t, device.IsWritable())
}

func TestDevice_encode(t *testing.T) {
	device := Device{
		Type: "foo",
		Metadata: map[string]string{
			"abc": "123",
		},
		Info:    "test",
		Handler: "vapor",
		Tags: []*Tag{
			{Namespace: "1", Annotation: "2", Label: "3"},
		},
		Alias:     "vaportest",
		SortIndex: 1,
		id:        "1234",
		handler: &DeviceHandler{
			Name: "vapor",
			Read: func(device *Device) (readings []*output.Reading, e error) {
				return nil, nil
			},
		},
	}

	encoded := device.encode()
	assert.NotEmpty(t, encoded.Timestamp)
	assert.Equal(t, "1234", encoded.Id)
	assert.Equal(t, "foo", encoded.Type)
	assert.Equal(t, "", encoded.Plugin)
	assert.Equal(t, "test", encoded.Info)
	assert.Equal(t, map[string]string{"abc": "123"}, encoded.Metadata)
	assert.Equal(t, "r", encoded.Capabilities.Mode)
	assert.Nil(t, encoded.Capabilities.Write.Actions)
	assert.Equal(t, 1, len(encoded.Tags))
	assert.Equal(t, 0, len(encoded.Outputs))
	assert.Equal(t, int32(1), encoded.SortIndex)
}

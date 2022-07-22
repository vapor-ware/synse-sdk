// Synse SDK
// Copyright (c) 2017-2022 Vapor IO
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
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/v2/sdk/config"
	"github.com/vapor-ware/synse-sdk/v2/sdk/errors"
	"github.com/vapor-ware/synse-sdk/v2/sdk/output"
)

// A map of DeviceHandlers for the test data. The handlers do nothing.
var testHandlers = map[string]*DeviceHandler{
	"testhandler":  {Name: "testhandler"},
	"testhandler2": {Name: "testhandler2"},
	"type1":        {Name: "type1"},
	"type2":        {Name: "type2"},
}

func TestNewDeviceFromConfig_nil1(t *testing.T) {
	d, err := NewDeviceFromConfig(nil, &config.DeviceInstance{}, testHandlers)
	assert.Error(t, err)
	assert.Nil(t, d)
}

func TestNewDeviceFromConfig_nil2(t *testing.T) {
	d, err := NewDeviceFromConfig(&config.DeviceProto{}, nil, testHandlers)
	assert.Error(t, err)
	assert.Nil(t, d)
}

func TestNewDeviceFromConfig_nil3(t *testing.T) {
	d, err := NewDeviceFromConfig(nil, nil, testHandlers)
	assert.Error(t, err)
	assert.Nil(t, d)
}

func TestNewDeviceFromConfig(t *testing.T) {
	// Tests creating a device where inheritance is enabled, but the
	// instance defines all inheritable things.
	proto := &config.DeviceProto{
		Type: "type1",
		Data: map[string]interface{}{
			"port": 5000,
		},
		Context: map[string]string{
			"foo": "bar",
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
		Context: map[string]string{
			"123": "456",
		},
		Output:    "temperature",
		SortIndex: 1,
		Handler:   "testhandler2",
		Alias: &config.DeviceAlias{
			Name: "foo",
		},
		Transforms: []*config.TransformConfig{
			{Scale: "2"},
			{Apply: "FtoC"},
		},
		WriteTimeout:       5 * time.Second,
		DisableInheritance: false,
	}

	device, err := NewDeviceFromConfig(proto, instance, testHandlers)
	assert.NoError(t, err)
	assert.Equal(t, "type2", device.Type)
	assert.Equal(t, "testdata", device.Info)
	assert.Equal(t, 2, len(device.Tags))
	assert.Equal(t, map[string]interface{}{"address": "localhost", "port": 5000}, device.Data)
	assert.Equal(t, map[string]string{"foo": "bar", "123": "456"}, device.Context)
	assert.Equal(t, "testhandler2", device.Handler)
	assert.Equal(t, int32(1), device.SortIndex)
	assert.Equal(t, "foo", device.Alias)
	assert.Equal(t, 2, len(device.Transforms))
	assert.Equal(t, "scale [2]", device.Transforms[0].Name())
	assert.Equal(t, "apply [FtoC]", device.Transforms[1].Name())
	assert.Equal(t, 5*time.Second, device.WriteTimeout)
	assert.Equal(t, "temperature", device.Output)
}

func TestNewDeviceFromConfig2(t *testing.T) {
	// Tests creating a device where inheritance is enabled, and the instance will
	// inherit values from the prototype.
	proto := &config.DeviceProto{
		Type: "type1",
		Data: map[string]interface{}{
			"port": 5000,
		},
		Context: map[string]string{
			"foo": "bar",
		},
		Tags:         []string{"default/foo"},
		Handler:      "testhandler",
		WriteTimeout: 3 * time.Second,
		Transforms: []*config.TransformConfig{
			{Scale: "2"},
			{Apply: "FtoC"},
		},
	}
	instance := &config.DeviceInstance{
		Info: "testdata",
		Tags: []string{"vapor/io"},
		Data: map[string]interface{}{
			"address": "localhost",
		},
		Output:    "temperature",
		SortIndex: 1,
		Alias: &config.DeviceAlias{
			Name: "foo",
		},
		DisableInheritance: false,
	}

	device, err := NewDeviceFromConfig(proto, instance, testHandlers)
	assert.NoError(t, err)
	assert.Equal(t, "type1", device.Type)
	assert.Equal(t, "testdata", device.Info)
	assert.Equal(t, 2, len(device.Tags))
	assert.Equal(t, map[string]interface{}{"address": "localhost", "port": 5000}, device.Data)
	assert.Equal(t, map[string]string{"foo": "bar"}, device.Context)
	assert.Equal(t, "testhandler", device.Handler)
	assert.Equal(t, int32(1), device.SortIndex)
	assert.Equal(t, "foo", device.Alias)
	assert.Equal(t, 3*time.Second, device.WriteTimeout)
	assert.Equal(t, "temperature", device.Output)
	assert.Equal(t, 2, len(device.Transforms))
	assert.Equal(t, "scale [2]", device.Transforms[0].Name())
	assert.Equal(t, "apply [FtoC]", device.Transforms[1].Name())
}

func TestNewDeviceFromConfig3(t *testing.T) {
	// Test when no type is resolved, resulting in error.
	proto := &config.DeviceProto{
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
		WriteTimeout:       5 * time.Second,
		DisableInheritance: false,
	}

	device, err := NewDeviceFromConfig(proto, instance, testHandlers)
	assert.Error(t, err)
	assert.Nil(t, device)
}

func TestNewDeviceFromConfig4(t *testing.T) {
	// Test inheriting from prototype when there is nothing to inherit.
	proto := &config.DeviceProto{
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
		DisableInheritance: false,
	}

	device, err := NewDeviceFromConfig(proto, instance, testHandlers)
	assert.NoError(t, err)
	assert.Equal(t, "type2", device.Type)
	assert.Equal(t, "testdata", device.Info)
	assert.Equal(t, 2, len(device.Tags))
	assert.Equal(t, map[string]interface{}{"address": "localhost", "port": 5000}, device.Data)
	assert.Empty(t, device.Context)
	assert.Equal(t, "type2", device.Handler)
	assert.Equal(t, int32(1), device.SortIndex)
	assert.Equal(t, "foo", device.Alias)
	assert.Equal(t, 30*time.Second, device.WriteTimeout) // takes the default value
	assert.Equal(t, "", device.Output)
	assert.Equal(t, 0, len(device.Transforms))
}

func TestNewDeviceFromConfig5a(t *testing.T) {
	// Test disabling inheritance when there are inheritable values.
	proto := &config.DeviceProto{
		Type: "type1",
		Data: map[string]interface{}{
			"port": 5000,
		},
		Context: map[string]string{
			"foo": "bar",
		},
		Tags:         []string{"default/foo"},
		Handler:      "testhandler",
		WriteTimeout: 3 * time.Second,
		Transforms: []*config.TransformConfig{
			{Scale: "2"},
			{Apply: "FtoC"},
		},
	}
	instance := &config.DeviceInstance{
		Type: "type2",
		Info: "testdata",
		Tags: []string{"vapor/io"},
		Data: map[string]interface{}{
			"address": "localhost",
		},
		Context: map[string]string{
			"abc": "def",
		},
		SortIndex: 1,
		Alias: &config.DeviceAlias{
			Name: "foo",
		},
		DisableInheritance: true,
	}

	device, err := NewDeviceFromConfig(proto, instance, testHandlers)
	assert.NoError(t, err)
	assert.Equal(t, "type2", device.Type)
	assert.Equal(t, "testdata", device.Info)
	assert.Equal(t, 1, len(device.Tags))
	assert.Equal(t, map[string]interface{}{"address": "localhost"}, device.Data)
	assert.Equal(t, map[string]string{"abc": "def"}, device.Context)
	assert.Equal(t, "type2", device.Handler) // inheritance disabled, does not get proto handler
	assert.Equal(t, int32(1), device.SortIndex)
	assert.Equal(t, "foo", device.Alias)
	assert.Equal(t, 30*time.Second, device.WriteTimeout) // takes the default value
	assert.Equal(t, "", device.Output)
	assert.Equal(t, 0, len(device.Transforms))
}

func TestNewDeviceFromConfig5b(t *testing.T) {
	// Test enabled inheritance when there are inheritable values and the prototype
	// defines a handler, but the instance does not.
	proto := &config.DeviceProto{
		Type: "type1",
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
		DisableInheritance: false,
	}

	device, err := NewDeviceFromConfig(proto, instance, testHandlers)
	assert.NoError(t, err)
	assert.Equal(t, "type2", device.Type)
	assert.Equal(t, "testdata", device.Info)
	assert.Equal(t, 2, len(device.Tags))
	assert.Equal(t, map[string]interface{}{"address": "localhost", "port": 5000}, device.Data)
	assert.Empty(t, device.Context)
	assert.Equal(t, "testhandler", device.Handler) // inheritance enabled, gets proto handler
	assert.Equal(t, int32(1), device.SortIndex)
	assert.Equal(t, "foo", device.Alias)
	assert.Equal(t, 3*time.Second, device.WriteTimeout) // takes the proto value
	assert.Equal(t, "", device.Output)
	assert.Equal(t, 0, len(device.Transforms))
}

func TestNewDeviceFromConfig6(t *testing.T) {
	// Bad tags specified.
	proto := &config.DeviceProto{
		Type: "type1",
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
		WriteTimeout:       5 * time.Second,
		DisableInheritance: false,
	}

	device, err := NewDeviceFromConfig(proto, instance, testHandlers)
	assert.Error(t, err)
	assert.Nil(t, device)
}

func TestNewDeviceFromConfig7(t *testing.T) {
	// Bad alias specified.
	proto := &config.DeviceProto{
		Type: "type1",
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
		WriteTimeout:       5 * time.Second,
		DisableInheritance: false,
	}

	device, err := NewDeviceFromConfig(proto, instance, testHandlers)
	assert.Error(t, err)
	assert.Nil(t, device)
}

func TestNewDeviceFromConfig8(t *testing.T) {
	// Fail data map merging
	proto := &config.DeviceProto{
		Type: "type1",
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
		WriteTimeout:       5 * time.Second,
		DisableInheritance: false,
	}

	device, err := NewDeviceFromConfig(proto, instance, testHandlers)
	assert.Error(t, err)
	assert.Nil(t, device)
}

func TestNewDeviceFromConfig9(t *testing.T) {
	// Unknown output type specified
	proto := &config.DeviceProto{
		Type: "type1",
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
		WriteTimeout:       5 * time.Second,
		DisableInheritance: false,
	}

	device, err := NewDeviceFromConfig(proto, instance, testHandlers)
	assert.Error(t, err)
	assert.Nil(t, device)
}

func TestNewDeviceFromConfig10(t *testing.T) {
	// Proto and instance both define transformers - ensure they merge correctly.
	proto := &config.DeviceProto{
		Type: "type1",
		Data: map[string]interface{}{
			"port": 5000,
		},
		Context: map[string]string{
			"foo": "bar",
		},
		Tags:         []string{"default/foo"},
		Handler:      "testhandler",
		WriteTimeout: 3 * time.Second,
		Transforms: []*config.TransformConfig{
			{Apply: "FtoC"},
			{Scale: "3"},
		},
	}
	instance := &config.DeviceInstance{
		Info: "testdata",
		Tags: []string{"vapor/io"},
		Data: map[string]interface{}{
			"address": "localhost",
		},
		Output:    "temperature",
		SortIndex: 1,
		Alias: &config.DeviceAlias{
			Name: "foo",
		},
		DisableInheritance: false,
		Transforms: []*config.TransformConfig{
			{Scale: "2"},
			{Apply: "FtoC"},
		},
	}

	device, err := NewDeviceFromConfig(proto, instance, testHandlers)
	assert.NoError(t, err)
	assert.Equal(t, "type1", device.Type)
	assert.Equal(t, "testdata", device.Info)
	assert.Equal(t, 2, len(device.Tags))
	assert.Equal(t, map[string]interface{}{"address": "localhost", "port": 5000}, device.Data)
	assert.Equal(t, map[string]string{"foo": "bar"}, device.Context)
	assert.Equal(t, "testhandler", device.Handler)
	assert.Equal(t, int32(1), device.SortIndex)
	assert.Equal(t, "foo", device.Alias)
	assert.Equal(t, 3*time.Second, device.WriteTimeout)
	assert.Equal(t, "temperature", device.Output)

	assert.Equal(t, 4, len(device.Transforms))
	assert.Equal(t, "apply [FtoC]", device.Transforms[0].Name())
	assert.Equal(t, "scale [3]", device.Transforms[1].Name())
	assert.Equal(t, "scale [2]", device.Transforms[2].Name())
	assert.Equal(t, "apply [FtoC]", device.Transforms[3].Name())
}

func TestNewDeviceFromConfig11a(t *testing.T) {
	// Invalid instance transformer config provided (specified multiple operations)
	proto := &config.DeviceProto{
		Type: "type1",
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
		Transforms: []*config.TransformConfig{{
			Apply: "FtoC",
			Scale: "2",
		}},
		WriteTimeout:       5 * time.Second,
		DisableInheritance: false,
	}

	device, err := NewDeviceFromConfig(proto, instance, testHandlers)
	assert.Nil(t, device)
	assert.Error(t, err)
}

func TestNewDeviceFromConfig11b(t *testing.T) {
	// Invalid prototype transformer config provided (specified multiple operations)
	proto := &config.DeviceProto{
		Type: "type1",
		Data: map[string]interface{}{
			"port": 5000,
		},
		Tags:         []string{"default/foo"},
		Handler:      "testhandler",
		WriteTimeout: 3 * time.Second,
		Transforms: []*config.TransformConfig{{
			Apply: "FtoC",
			Scale: "2",
		}},
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
		WriteTimeout:       5 * time.Second,
		DisableInheritance: false,
	}

	device, err := NewDeviceFromConfig(proto, instance, testHandlers)
	assert.Nil(t, device)
	assert.Error(t, err)
}

func TestNewDeviceFromConfig12(t *testing.T) {
	// Tests creating a device where inheritance is enabled and the instance context
	// overrides some values in the prototype context.
	proto := &config.DeviceProto{
		Type: "type1",
		Data: map[string]interface{}{
			"port": 5000,
		},
		Context: map[string]string{
			"foo": "bar",
			"123": "456",
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
		Context: map[string]string{
			"123": "abc",
			"xyz": "456",
		},
		Output:    "temperature",
		SortIndex: 1,
		Handler:   "testhandler2",
		Alias: &config.DeviceAlias{
			Name: "foo",
		},
		WriteTimeout:       5 * time.Second,
		DisableInheritance: false,
	}

	device, err := NewDeviceFromConfig(proto, instance, testHandlers)
	assert.NoError(t, err)
	assert.Equal(t, "type2", device.Type)
	assert.Equal(t, "testdata", device.Info)
	assert.Equal(t, 2, len(device.Tags))
	assert.Equal(t, map[string]interface{}{"address": "localhost", "port": 5000}, device.Data)
	assert.Equal(t, map[string]string{"foo": "bar", "123": "abc", "xyz": "456"}, device.Context)
	assert.Equal(t, "testhandler2", device.Handler)
	assert.Equal(t, int32(1), device.SortIndex)
	assert.Equal(t, "foo", device.Alias)
	assert.Equal(t, 5*time.Second, device.WriteTimeout)
	assert.Equal(t, "temperature", device.Output)
	assert.Equal(t, 0, len(device.Transforms))
}

func TestNewDeviceFromConfig13(t *testing.T) {
	// A regression test for a bug where having multiple instances with different
	// contexts and shared prototype context would lead to context overrwrite.
	proto := &config.DeviceProto{
		Type: "type1",
		Data: map[string]interface{}{
			"port": 5000,
		},
		Context: map[string]string{
			"foo": "bar",
		},
		Tags:         []string{"default/foo"},
		Handler:      "testhandler",
		WriteTimeout: 3 * time.Second,
	}
	instance1 := &config.DeviceInstance{
		Type: "type2",
		Info: "test device 1",
		Tags: []string{
			"vapor/io",
			"tag/1",
			"abc/123",
		},
		Data: map[string]interface{}{
			"address": "localhost",
			"value":   123,
			"foo":     "a",
		},
		Context: map[string]string{
			"alpha": "value-1",
			"bravo": "value-2",
		},
		Output:    "temperature",
		SortIndex: 1,
		Handler:   "testhandler2",
		Alias: &config.DeviceAlias{
			Name: "foo",
		},
		WriteTimeout:       5 * time.Second,
		DisableInheritance: false,
	}
	instance2 := &config.DeviceInstance{
		Type: "type3",
		Info: "test device 2",
		Tags: []string{
			"vapor/io",
			"tag/2",
			"def/456",
		},
		Data: map[string]interface{}{
			"address": "localhost",
			"value":   456,
			"bar":     "b",
		},
		Context: map[string]string{
			"bravo":   "value-3",
			"charlie": "value-4",
		},
		Output:    "temperature",
		SortIndex: 2,
		Handler:   "testhandler2",
		Alias: &config.DeviceAlias{
			Name: "bar",
		},
		WriteTimeout:       5 * time.Second,
		DisableInheritance: false,
	}

	t1, _ := NewTag("default/foo")
	t2, _ := NewTag("vapor/io")
	t3, _ := NewTag("tag/1")
	t4, _ := NewTag("tag/2")
	t5, _ := NewTag("abc/123")
	t6, _ := NewTag("def/456")

	dev1ExpectedTags := []*Tag{t1, t2, t3, t5}
	dev2ExpectedTags := []*Tag{t1, t2, t4, t6}

	dev1, err := NewDeviceFromConfig(proto, instance1, testHandlers)
	assert.NoError(t, err)
	assert.Equal(t, "type2", dev1.Type)
	assert.Equal(t, "test device 1", dev1.Info)
	assert.Equal(t, 4, len(dev1.Tags))
	assert.Equal(t, map[string]interface{}{"address": "localhost", "port": 5000, "value": 123, "foo": "a"}, dev1.Data)
	assert.Equal(t, map[string]string{"foo": "bar", "alpha": "value-1", "bravo": "value-2"}, dev1.Context)
	assert.Equal(t, "testhandler2", dev1.Handler)
	assert.Equal(t, int32(1), dev1.SortIndex)
	assert.Equal(t, "foo", dev1.Alias)
	assert.Equal(t, 5*time.Second, dev1.WriteTimeout)
	assert.Equal(t, "temperature", dev1.Output)
	assert.Equal(t, 0, len(dev1.Transforms))
	for i, tag := range dev1.Tags {
		assert.Equal(t, dev1ExpectedTags[i], tag)
	}

	// Use the same prototype for the next instance
	dev2, err := NewDeviceFromConfig(proto, instance2, testHandlers)
	assert.NoError(t, err)
	assert.Equal(t, "type3", dev2.Type)
	assert.Equal(t, "test device 2", dev2.Info)
	assert.Equal(t, 4, len(dev2.Tags))
	assert.Equal(t, map[string]interface{}{"address": "localhost", "port": 5000, "value": 456, "bar": "b"}, dev2.Data)
	assert.Equal(t, map[string]string{"foo": "bar", "charlie": "value-4", "bravo": "value-3"}, dev2.Context)
	assert.Equal(t, "testhandler2", dev2.Handler)
	assert.Equal(t, int32(2), dev2.SortIndex)
	assert.Equal(t, "bar", dev2.Alias)
	assert.Equal(t, 5*time.Second, dev2.WriteTimeout)
	assert.Equal(t, "temperature", dev2.Output)
	assert.Equal(t, 0, len(dev2.Transforms))
	for i, tag := range dev2.Tags {
		assert.Equal(t, dev2ExpectedTags[i], tag)
	}
}

func TestNewDeviceFromConfig14(t *testing.T) {
	proto := &config.DeviceProto{
		Type: "type1",
		Data: map[string]interface{}{
			"port": 5000,
		},
		Handler: "handler-that-does-not-exist",
	}
	instance := &config.DeviceInstance{
		Info:               "testdata",
		Tags:               []string{"vapor/io"},
		Output:             "temperature",
		WriteTimeout:       5 * time.Second,
		DisableInheritance: false,
	}

	device, err := NewDeviceFromConfig(proto, instance, testHandlers)
	assert.Nil(t, device)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown handler specified")
}

func TestNewDeviceFromConfig15(t *testing.T) {
	// Tests creating a device with tags and context including templates
	proto := &config.DeviceProto{
		Type: "type1",
		Data: map[string]interface{}{
			"port": 5000,
		},
		Context: map[string]string{
			"foo": "bar",
			"123": "456",
		},
		Tags:         []string{`default/{{ env "FOO" }}`},
		Handler:      "testhandler",
		WriteTimeout: 3 * time.Second,
	}
	instance := &config.DeviceInstance{
		Type: "type2",
		Info: "testdata",
		Data: map[string]interface{}{
			"address": "localhost",
		},
		Context: map[string]string{
			"123": "abc",
			"xyz": `{{ env "BAR" }}`,
		},
		Output:    "temperature",
		SortIndex: 1,
		Handler:   "testhandler2",
		Alias: &config.DeviceAlias{
			Name: "foo",
		},
		WriteTimeout:       5 * time.Second,
		DisableInheritance: false,
	}

	// Set ENV vars for the test case.
	testEnv := map[string]string{
		"FOO": "foo",
		"BAR": "bar",
	}
	// Setup the environment for the test case.
	for k, v := range testEnv {
		err := os.Setenv(k, v)
		assert.NoError(t, err)
	}
	defer func() {
		for k := range testEnv {
			err := os.Unsetenv(k)
			assert.NoError(t, err)
		}
	}()

	t1, _ := NewTag("default/foo")

	device, err := NewDeviceFromConfig(proto, instance, testHandlers)
	assert.NoError(t, err)
	assert.Equal(t, "type2", device.Type)
	assert.Equal(t, "testdata", device.Info)
	assert.Equal(t, 1, len(device.Tags))
	assert.Equal(t, t1, device.Tags[0])
	assert.Equal(t, map[string]interface{}{"address": "localhost", "port": 5000}, device.Data)
	assert.Equal(t, map[string]string{"foo": "bar", "123": "abc", "xyz": "bar"}, device.Context)
	assert.Equal(t, "testhandler2", device.Handler)
	assert.Equal(t, int32(1), device.SortIndex)
	assert.Equal(t, "foo", device.Alias)
	assert.Equal(t, 5*time.Second, device.WriteTimeout)
	assert.Equal(t, "temperature", device.Output)
	assert.Equal(t, 0, len(device.Transforms))
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

func TestDevice_GetContext(t *testing.T) {
	device := Device{
		Context: map[string]string{
			"foo": "bar",
			"abc": "xyz",
		},
	}

	assert.Equal(t, "", device.GetContext("vapor"))
	assert.Equal(t, "bar", device.GetContext("foo"))
	assert.Equal(t, "xyz", device.GetContext("abc"))
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

func TestDevice_Write_hasMatchingAction(t *testing.T) {
	device := Device{
		id: "123",
		handler: &DeviceHandler{
			Write: func(device *Device, data *WriteData) error {
				return nil
			},
			Actions: []string{"test-action"},
		},
	}

	err := device.Write(&WriteData{Action: "test-action"})
	assert.NoError(t, err)
}

func TestDevice_Write_hasNonMatchingAction(t *testing.T) {
	device := Device{
		id: "123",
		handler: &DeviceHandler{
			Write: func(device *Device, data *WriteData) error {
				return nil
			},
			Actions: []string{"test-action"},
		},
	}

	err := device.Write(&WriteData{Action: "unsupported-action"})
	assert.Error(t, err)
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
		Context: map[string]string{
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

func TestDevice_encode_2(t *testing.T) {
	// Encode when there are handler actions, but no Write handler
	device := Device{
		Type: "foo",
		Context: map[string]string{
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
			Actions: []string{"action-1", "action-2"},
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

func TestDevice_encode_3(t *testing.T) {
	// Encode when there are handler actions and a Write handler is specified
	device := Device{
		Type: "foo",
		Context: map[string]string{
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
			Write: func(device *Device, data *WriteData) error {
				return nil
			},
			Actions: []string{"action-1", "action-2"},
		},
	}

	encoded := device.encode()
	assert.NotEmpty(t, encoded.Timestamp)
	assert.Equal(t, "1234", encoded.Id)
	assert.Equal(t, "foo", encoded.Type)
	assert.Equal(t, "", encoded.Plugin)
	assert.Equal(t, "test", encoded.Info)
	assert.Equal(t, map[string]string{"abc": "123"}, encoded.Metadata)
	assert.Equal(t, "rw", encoded.Capabilities.Mode)
	assert.Equal(t, []string{"action-1", "action-2"}, encoded.Capabilities.Write.Actions)
	assert.Equal(t, 1, len(encoded.Tags))
	assert.Equal(t, 0, len(encoded.Outputs))
	assert.Equal(t, int32(1), encoded.SortIndex)
}

func TestDevice_parseContext(t *testing.T) {
	tests := []struct {
		name     string
		ctx      map[string]string
		expected map[string]string
	}{
		{
			name:     "no template",
			ctx:      map[string]string{"foo": "bar", "abc": "123"},
			expected: map[string]string{"foo": "bar", "abc": "123"},
		},
		{
			name:     "template whole value",
			ctx:      map[string]string{"foo": `{{ env "BAR" }}`},
			expected: map[string]string{"foo": "bar"},
		},
		{
			name:     "template part of value",
			ctx:      map[string]string{"foo": `value-{{ env "TEST_ENV_VAL_1" }}`},
			expected: map[string]string{"foo": "value-1"},
		},
		{
			name: "multiple env template",
			ctx: map[string]string{
				"first":  `{{env "FOO"}}`,
				"second": `val-{{env "FOO"}}-{{ env "BAR" }}`,
				"third":  `{{ env "FOO" }}.{{ env "TEST_ENV_VAL_1" }}`,
			},
			expected: map[string]string{
				"first":  "foo",
				"second": "val-foo-bar",
				"third":  "foo.1",
			},
		},
		{
			name:     "no env set",
			ctx:      map[string]string{"foo": `{{ env "ENV_VALUE_NOT_SET" }}`},
			expected: map[string]string{"foo": ""},
		},
	}

	testEnv := map[string]string{
		"FOO":            "foo",
		"BAR":            "bar",
		"TEST_ENV_VAL_1": "1",
	}
	// Setup the environment for the test case.
	for k, v := range testEnv {
		err := os.Setenv(k, v)
		assert.NoError(t, err)
	}
	defer func() {
		for k := range testEnv {
			err := os.Unsetenv(k)
			assert.NoError(t, err)
		}
	}()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := parseContext(test.ctx)
			assert.NoError(t, err, test.name)
			assert.Equal(t, test.expected, test.ctx, test.name)
		})
	}
}

func TestDevice_parseContextError(t *testing.T) {
	ctx := map[string]string{
		"foo": `{{ foobar "ENV_VALUE_NOT_SET" }}`, // no such function
	}
	err := parseContext(ctx)
	assert.Error(t, err)
}

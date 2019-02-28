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

	"github.com/vapor-ware/synse-sdk/internal/test"
	"github.com/vapor-ware/synse-sdk/sdk/policy"

	"github.com/google/uuid"
	"github.com/vapor-ware/synse-sdk/sdk/output"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/sdk/config"
)

func Test_newDeviceManager_nil(t *testing.T) {
	assert.Panics(t, func() {
		newDeviceManager(nil)
	})
}

func Test_newDeviceManager_ok(t *testing.T) {
	m := newDeviceManager(&Plugin{
		config: &config.Plugin{},
	})
	assert.Empty(t, m.devices)
	assert.Empty(t, m.handlers)
}

func TestDeviceManager_Start(t *testing.T) {
	plugin := Plugin{}
	m := deviceManager{}

	err := m.Start(&plugin)
	assert.NoError(t, err)
}

func TestDeviceManager_loadDynamicConfig_noConfig(t *testing.T) {
	m := deviceManager{
		config:        &config.Devices{},
		dynamicConfig: &config.DynamicRegistrationSettings{},
	}

	err := m.loadDynamicConfig()
	assert.NoError(t, err)
	assert.Empty(t, m.config.Devices)
}

func TestDeviceManager_loadDynamicConfig_ok(t *testing.T) {
	m := deviceManager{
		config: &config.Devices{},
		dynamicConfig: &config.DynamicRegistrationSettings{
			Config: []map[string]interface{}{{}},
		},
		pluginHandlers: &PluginHandlers{
			DynamicConfigRegistrar: func(i map[string]interface{}) (protos []*config.DeviceProto, e error) {
				return []*config.DeviceProto{
					{Type: "foo"},
					{Type: "bar"},
				}, nil
			},
		},
	}

	err := m.loadDynamicConfig()
	assert.NoError(t, err)
	assert.Len(t, m.config.Devices, 2)
}

func TestDeviceManager_loadDynamicConfig_errUnknownPolicy(t *testing.T) {
	m := deviceManager{
		config: &config.Devices{},
		dynamicConfig: &config.DynamicRegistrationSettings{
			Config: []map[string]interface{}{{}},
		},
		pluginHandlers: &PluginHandlers{
			DynamicConfigRegistrar: func(i map[string]interface{}) (protos []*config.DeviceProto, e error) {
				return nil, fmt.Errorf("test error")
			},
		},
		policies: &policy.Policies{
			DynamicDeviceConfig: policy.Policy("unknown"),
		},
	}

	err := m.loadDynamicConfig()
	assert.Error(t, err)
	assert.Empty(t, m.config.Devices)
}

func TestDeviceManager_loadDynamicConfig_errOptionalPolicy(t *testing.T) {
	m := deviceManager{
		config: &config.Devices{},
		dynamicConfig: &config.DynamicRegistrationSettings{
			Config: []map[string]interface{}{{}},
		},
		pluginHandlers: &PluginHandlers{
			DynamicConfigRegistrar: func(i map[string]interface{}) (protos []*config.DeviceProto, e error) {
				return nil, fmt.Errorf("test error")
			},
		},
		policies: &policy.Policies{
			DynamicDeviceConfig: policy.Optional,
		},
	}

	err := m.loadDynamicConfig()
	assert.NoError(t, err)
	assert.Empty(t, m.config.Devices)
}

func TestDeviceManager_loadDynamicConfig_errRequiredPolicy(t *testing.T) {
	m := deviceManager{
		config: &config.Devices{},
		dynamicConfig: &config.DynamicRegistrationSettings{
			Config: []map[string]interface{}{{}},
		},
		pluginHandlers: &PluginHandlers{
			DynamicConfigRegistrar: func(i map[string]interface{}) (protos []*config.DeviceProto, e error) {
				return nil, fmt.Errorf("test error")
			},
		},
		policies: &policy.Policies{
			DynamicDeviceConfig: policy.Required,
		},
	}

	err := m.loadDynamicConfig()
	assert.Error(t, err)
	assert.Empty(t, m.config.Devices)
}

func TestDeviceManager_createDynamicDevices_noConfig(t *testing.T) {
	m := deviceManager{
		config:        &config.Devices{},
		dynamicConfig: &config.DynamicRegistrationSettings{},
		devices:       map[string]*Device{},
	}

	err := m.createDynamicDevices()
	assert.NoError(t, err)
	assert.Empty(t, m.devices)
}

func TestDeviceManager_createDynamicDevices_ok(t *testing.T) {
	m := deviceManager{
		tagCache: NewTagCache(),
		config:   &config.Devices{},
		dynamicConfig: &config.DynamicRegistrationSettings{
			Config: []map[string]interface{}{{}},
		},
		pluginHandlers: &PluginHandlers{
			DynamicRegistrar: func(i map[string]interface{}) (devices []*Device, e error) {
				return []*Device{
					{
						Type:    "foo",
						Handler: "foo",
						handler: &DeviceHandler{Name: "foo"},
						id:      "12345",
					},
				}, nil
			},
		},
		devices: map[string]*Device{},
	}

	err := m.createDynamicDevices()
	assert.NoError(t, err)
	assert.Len(t, m.devices, 1)
	assert.Contains(t, m.devices, "12345")
}

func TestDeviceManager_createDynamicDevices_errUnknownPolicy(t *testing.T) {
	m := deviceManager{
		config: &config.Devices{},
		dynamicConfig: &config.DynamicRegistrationSettings{
			Config: []map[string]interface{}{{}},
		},
		pluginHandlers: &PluginHandlers{
			DynamicRegistrar: func(i map[string]interface{}) (devices []*Device, e error) {
				return nil, fmt.Errorf("test error")
			},
		},
		policies: &policy.Policies{
			DynamicDeviceConfig: policy.Policy("unknown"),
		},
		devices: map[string]*Device{},
	}

	err := m.createDynamicDevices()
	assert.Error(t, err)
	assert.Empty(t, m.devices)
}

func TestDeviceManager_createDynamicDevices_errOptionalPolicy(t *testing.T) {
	m := deviceManager{
		config: &config.Devices{},
		dynamicConfig: &config.DynamicRegistrationSettings{
			Config: []map[string]interface{}{{}},
		},
		pluginHandlers: &PluginHandlers{
			DynamicRegistrar: func(i map[string]interface{}) (devices []*Device, e error) {
				return nil, fmt.Errorf("test error")
			},
		},
		policies: &policy.Policies{
			DynamicDeviceConfig: policy.Optional,
		},
		devices: map[string]*Device{},
	}

	err := m.createDynamicDevices()
	assert.NoError(t, err)
	assert.Empty(t, m.devices)
}

func TestDeviceManager_createDynamicDevices_errRequiredPolicy(t *testing.T) {
	m := deviceManager{
		config: &config.Devices{},
		dynamicConfig: &config.DynamicRegistrationSettings{
			Config: []map[string]interface{}{{}},
		},
		pluginHandlers: &PluginHandlers{
			DynamicRegistrar: func(i map[string]interface{}) (devices []*Device, e error) {
				return nil, fmt.Errorf("test error")
			},
		},
		policies: &policy.Policies{
			DynamicDeviceConfig: policy.Required,
		},
		devices: map[string]*Device{},
	}

	err := m.createDynamicDevices()
	assert.Error(t, err)
	assert.Empty(t, m.devices)
}

func TestDeviceManager_createDynamicDevices_errAddDevice(t *testing.T) {
	m := deviceManager{
		tagCache: NewTagCache(),
		config:   &config.Devices{},
		dynamicConfig: &config.DynamicRegistrationSettings{
			Config: []map[string]interface{}{{}},
		},
		pluginHandlers: &PluginHandlers{
			DynamicRegistrar: func(i map[string]interface{}) (devices []*Device, e error) {
				return []*Device{
					{
						Type:    "foo",
						Handler: "foo",
						handler: &DeviceHandler{Name: "foo"},
						id:      "12345",
					},
				}, nil
			},
		},
		devices: map[string]*Device{
			"12345": {id: "12345"},
		},
	}

	err := m.createDynamicDevices()
	assert.Error(t, err)
	assert.Len(t, m.devices, 1)
	assert.Contains(t, m.devices, "12345")
}

func TestDeviceManager_GetDevice_notFound(t *testing.T) {
	m := deviceManager{
		config:  &config.Devices{},
		devices: map[string]*Device{},
	}

	device := m.GetDevice("123")
	assert.Nil(t, device)
}

func TestDeviceManager_GetDevice_exists(t *testing.T) {
	m := deviceManager{
		devices: map[string]*Device{
			"123": {id: "123"},
		},
	}

	device := m.GetDevice("123")
	assert.NotNil(t, device)
	assert.Equal(t, "123", device.id)
}

func TestDeviceManager_GetDevices(t *testing.T) {
	m := deviceManager{
		tagCache: &TagCache{
			cache: map[string]map[string]map[string][]*Device{
				"foo": {
					"": {"c": {&Device{id: "123"}, &Device{id: "789"}}},
				},
				"bar": {
					"":    {"b": {&Device{id: "456"}}},
					"baz": {"a": {&Device{id: "789"}}},
				},
			},
		},
	}

	devices := m.GetDevices(&Tag{Namespace: "foo", Label: "c"})
	assert.Len(t, devices, 2)

	devices = m.GetDevices(&Tag{Namespace: "bar", Label: "b"})
	assert.Len(t, devices, 1)

	devices = m.GetDevices(
		&Tag{Namespace: "foo", Label: "c"},
		&Tag{Namespace: "bar", Annotation: "baz", Label: "a"},
	)
	assert.Len(t, devices, 1)
}

func TestDeviceManager_GetDevicesByTagNamespace_notFound(t *testing.T) {
	m := deviceManager{
		tagCache: &TagCache{
			cache: map[string]map[string]map[string][]*Device{},
		},
	}

	devices := m.GetDevicesByTagNamespace("foo")
	assert.Empty(t, devices)
}

func TestDeviceManager_GetDevicesByTagNamespace_ok(t *testing.T) {
	m := deviceManager{
		tagCache: &TagCache{
			cache: map[string]map[string]map[string][]*Device{
				"foo": {
					"": {"": {&Device{id: "123"}}},
				},
				"bar": {
					"": {"": {&Device{id: "456"}}},
				},
			},
		},
	}

	devices := m.GetDevicesByTagNamespace("foo")
	assert.Len(t, devices, 1)

	devices = m.GetDevicesByTagNamespace("bar")
	assert.Len(t, devices, 1)

	devices = m.GetDevicesByTagNamespace("foo", "bar")
	assert.Len(t, devices, 2)
}

func TestDeviceManager_GetAllDevices_empty(t *testing.T) {
	m := deviceManager{
		devices: map[string]*Device{},
	}

	devices := m.GetAllDevices()
	assert.Empty(t, devices)
}

func TestDeviceManager_GetAllDevices(t *testing.T) {
	m := deviceManager{
		devices: map[string]*Device{
			"123": {},
			"456": {},
			"789": {},
		},
	}

	devices := m.GetAllDevices()
	assert.Len(t, devices, 3)
}

func TestDeviceManager_IsDeviceReadable_nil(t *testing.T) {
	m := deviceManager{
		devices: map[string]*Device{},
	}
	assert.False(t, m.IsDeviceReadable("1234"))
}

func TestDeviceManager_IsDeviceReadable_true(t *testing.T) {
	m := deviceManager{
		devices: map[string]*Device{
			"1234": {
				handler: &DeviceHandler{
					Read: func(device *Device) (readings []*output.Reading, e error) {
						return nil, nil
					},
				},
			},
		},
	}
	assert.True(t, m.IsDeviceReadable("1234"))
}

func TestDeviceManager_IsDeviceReadable_false(t *testing.T) {
	m := deviceManager{
		devices: map[string]*Device{
			"1234": {
				handler: &DeviceHandler{},
			},
		},
	}
	assert.False(t, m.IsDeviceReadable("1234"))
}

func TestDeviceManager_IsDeviceWritable_nil(t *testing.T) {
	m := deviceManager{
		devices: map[string]*Device{},
	}
	assert.False(t, m.IsDeviceWritable("1234"))
}

func TestDeviceManager_IsDeviceWritable_true(t *testing.T) {
	m := deviceManager{
		devices: map[string]*Device{
			"1234": {
				handler: &DeviceHandler{
					Write: func(device *Device, data *WriteData) error {
						return nil
					},
				},
			},
		},
	}
	assert.True(t, m.IsDeviceWritable("1234"))
}

func TestDeviceManager_IsDeviceWritable_false(t *testing.T) {
	m := deviceManager{
		devices: map[string]*Device{
			"1234": {
				handler: &DeviceHandler{},
			},
		},
	}
	assert.False(t, m.IsDeviceWritable("1234"))
}

func TestDeviceManager_HasReadHandlers_empty(t *testing.T) {
	m := deviceManager{
		handlers: map[string]*DeviceHandler{},
	}
	assert.False(t, m.HasReadHandlers())
}

func TestDeviceManager_HasReadHandlers_true(t *testing.T) {
	m := deviceManager{
		handlers: map[string]*DeviceHandler{
			"a": {},
			"b": {Read: func(device *Device) (readings []*output.Reading, e error) {
				return nil, nil
			}},
			"c": {},
		},
	}
	assert.True(t, m.HasReadHandlers())
}

func TestDeviceManager_HasReadHandlers_false(t *testing.T) {
	m := deviceManager{
		handlers: map[string]*DeviceHandler{
			"a": {},
			"b": {},
			"c": {},
		},
	}
	assert.False(t, m.HasReadHandlers())
}

func TestDeviceManager_HasWriteHandlers_empty(t *testing.T) {
	m := deviceManager{
		handlers: map[string]*DeviceHandler{},
	}
	assert.False(t, m.HasWriteHandlers())
}
func TestDeviceManager_HasWriteHandlers_true(t *testing.T) {
	m := deviceManager{
		handlers: map[string]*DeviceHandler{
			"a": {},
			"b": {
				Write: func(device *Device, data *WriteData) error {
					return nil
				},
			},
			"c": {},
		},
	}
	assert.True(t, m.HasWriteHandlers())
}
func TestDeviceManager_HasWriteHandlers_false(t *testing.T) {
	m := deviceManager{
		handlers: map[string]*DeviceHandler{
			"a": {},
			"b": {},
			"c": {},
		},
	}
	assert.False(t, m.HasWriteHandlers())
}

func TestDeviceManager_HasListenerHandlers_empty(t *testing.T) {
	m := deviceManager{
		handlers: map[string]*DeviceHandler{},
	}
	assert.False(t, m.HasListenerHandlers())
}
func TestDeviceManager_HasListenerHandlers_true(t *testing.T) {
	m := deviceManager{
		handlers: map[string]*DeviceHandler{
			"a": {},
			"b": {Listen: func(device *Device, contexts chan *ReadContext) error {
				return nil
			}},
			"c": {},
		},
	}
	assert.True(t, m.HasListenerHandlers())
}
func TestDeviceManager_HasListenerHandlers_false(t *testing.T) {
	m := deviceManager{
		handlers: map[string]*DeviceHandler{
			"a": {},
			"b": {},
			"c": {},
		},
	}
	assert.False(t, m.HasListenerHandlers())
}

func TestDeviceManager_AddDevice_nil(t *testing.T) {
	m := deviceManager{}

	err := m.AddDevice(nil)
	assert.Error(t, err)
}

func TestDeviceManager_AddDevice_noHandler(t *testing.T) {
	m := deviceManager{}
	device := Device{}

	err := m.AddDevice(&device)
	assert.Error(t, err)
}

func TestDeviceManager_AddDevice_noSuchHandler(t *testing.T) {
	m := deviceManager{
		handlers: map[string]*DeviceHandler{},
	}
	device := Device{
		Handler: "foo",
	}

	err := m.AddDevice(&device)
	assert.Error(t, err)
}

func TestDeviceManager_AddDevice_idExists(t *testing.T) {
	m := deviceManager{
		handlers: map[string]*DeviceHandler{
			"foo": {Name: "foo"},
		},
		devices: map[string]*Device{
			"1234": {id: "1234"},
		},
	}
	device := Device{
		Handler: "foo",
		id:      "1234",
	}

	err := m.AddDevice(&device)
	assert.Error(t, err)
}

func TestDeviceManager_AddDevice(t *testing.T) {
	handler := DeviceHandler{Name: "foo"}
	m := deviceManager{
		tagCache:       NewTagCache(),
		id:             &pluginID{uuid: uuid.NewSHA1(uuid.NameSpaceDNS, []byte("test"))},
		pluginHandlers: NewDefaultPluginHandlers(),
		handlers: map[string]*DeviceHandler{
			"foo": &handler,
		},
		devices: map[string]*Device{},
	}
	device := Device{
		Type:    "testtype",
		Handler: "foo",
		Data: map[string]interface{}{
			"id":  1,
			"foo": "bar",
		},
		Tags: []*Tag{
			{Namespace: "default", Label: "foo"},
		},
	}

	// Before we add the device, make sure the state is empty.
	assert.Empty(t, m.tagCache.cache)
	assert.Empty(t, m.devices)

	err := m.AddDevice(&device)
	assert.NoError(t, err)

	// Make sure that the device was added to the manager, and its
	// tags were updated in the tag cache.
	expectedID := "81c0d156-06c0-50de-8e37-410cdb881eaf"
	assert.Len(t, m.devices, 1)
	assert.Contains(t, m.devices, expectedID)
	assert.Equal(t, &device, m.devices[expectedID])

	assert.Len(t, m.tagCache.cache, 2)
	assert.Contains(t, m.tagCache.cache, "default")
	assert.Contains(t, m.tagCache.cache, "system")

	// Make sure the device was updated with its pertinent fields.
	assert.Equal(t, &handler, device.handler)
	assert.Equal(t, "testtype.foo.bar1", device.idName)
	assert.Equal(t, expectedID, device.id)
	assert.Len(t, device.Tags, 3) // two additional system-generated tags added
}

func TestDeviceManager_AddHandlers(t *testing.T) {
	m := deviceManager{
		handlers: map[string]*DeviceHandler{},
	}

	err := m.AddHandlers(
		&DeviceHandler{Name: "foo"},
		&DeviceHandler{Name: "bar"},
	)
	assert.NoError(t, err)
	assert.Len(t, m.handlers, 2)
}

func TestDeviceManager_AddHandlers_err(t *testing.T) {
	m := deviceManager{
		handlers: map[string]*DeviceHandler{},
	}

	err := m.AddHandlers(
		&DeviceHandler{Name: "foo"},
		&DeviceHandler{Name: "foo"},
	)
	assert.Error(t, err)
	assert.Len(t, m.handlers, 1)
}

func TestDeviceManager_GetDevicesForHandler(t *testing.T) {
	m := deviceManager{
		handlers: map[string]*DeviceHandler{
			"foo": {Name: "foo"},
			"bar": {Name: "bar"},
		},
		devices: map[string]*Device{
			"123": {id: "123", Type: "foo", Handler: "foo"},
			"456": {id: "456", Type: "bar", Handler: "bar"},
			"678": {id: "678", Type: "foo", Handler: "foo"},
		},
	}

	devices := m.GetDevicesForHandler("foo")
	assert.Len(t, devices, 2)

	devices = m.GetDevicesForHandler("bar")
	assert.Len(t, devices, 1)

	devices = m.GetDevicesForHandler("baz")
	assert.Len(t, devices, 0)
}

func TestDeviceManager_GetHandler_notExists(t *testing.T) {
	m := deviceManager{
		handlers: map[string]*DeviceHandler{},
	}

	handler, err := m.GetHandler("foo")
	assert.Error(t, err)
	assert.Nil(t, handler)

}

func TestDeviceManager_GetHandler_exists(t *testing.T) {
	m := deviceManager{
		handlers: map[string]*DeviceHandler{
			"foo": {Name: "foo"},
		},
	}

	handler, err := m.GetHandler("foo")
	assert.NoError(t, err)
	assert.NotNil(t, handler)
	assert.Equal(t, "foo", handler.Name)
}

func TestDeviceManager_AddDeviceSetupActions_ok(t *testing.T) {
	m := deviceManager{}
	assert.Empty(t, m.setupActions)

	err := m.AddDeviceSetupActions(
		&DeviceAction{
			Name:   "foo",
			Filter: map[string][]string{"type": {"foo"}},
			Action: func(p *Plugin, d *Device) error {
				return nil
			},
		},
		&DeviceAction{
			Name:   "bar",
			Filter: map[string][]string{"type": {"bar"}},
			Action: func(p *Plugin, d *Device) error {
				return nil
			},
		},
	)

	assert.NoError(t, err)
	assert.Len(t, m.setupActions, 2)
}

func TestDeviceManager_AddDeviceSetupActions_okEmpty(t *testing.T) {
	m := deviceManager{}
	assert.Empty(t, m.setupActions)

	err := m.AddDeviceSetupActions()

	assert.NoError(t, err)
	assert.Empty(t, m.setupActions)
}

func TestDeviceManager_AddDeviceSetupActions_error(t *testing.T) {
	m := deviceManager{}
	assert.Empty(t, m.setupActions)

	err := m.AddDeviceSetupActions(
		&DeviceAction{
			Name: "foo",
			// no filter specified
			Action: func(p *Plugin, d *Device) error {
				return nil
			},
		},
	)

	assert.Error(t, err)
	assert.Empty(t, m.setupActions)
}

func TestDeviceManager_FilterDevices_ok(t *testing.T) {
	m := deviceManager{
		devices: map[string]*Device{
			"123": {id: "123", Type: "foo"},
			"456": {id: "456", Type: "bar"},
			"678": {id: "678", Type: "foo"},
		},
	}
	filter := map[string][]string{
		"type": {"foo"},
	}

	devices, err := m.FilterDevices(filter)
	assert.NoError(t, err)
	assert.Len(t, devices, 2)
}

func TestDeviceManager_FilterDevices_ok2(t *testing.T) {
	m := deviceManager{
		devices: map[string]*Device{
			"123": {id: "123", Type: "foo"},
			"456": {id: "456", Type: "bar"},
			"678": {id: "678", Type: "foo"},
		},
	}
	filter := map[string][]string{
		"type": {"bar"},
	}

	devices, err := m.FilterDevices(filter)
	assert.NoError(t, err)
	assert.Len(t, devices, 1)
}

func TestDeviceManager_FilterDevices_ok3(t *testing.T) {
	m := deviceManager{
		devices: map[string]*Device{
			"123": {id: "123", Type: "foo"},
			"456": {id: "456", Type: "bar"},
			"678": {id: "678", Type: "foo"},
		},
	}
	filter := map[string][]string{
		"type": {"baz"},
	}

	devices, err := m.FilterDevices(filter)
	assert.NoError(t, err)
	assert.Len(t, devices, 0)
}

func TestDeviceManager_FilterDevices_ok4(t *testing.T) {
	m := deviceManager{
		devices: map[string]*Device{
			"123": {id: "123", Type: "foo"},
			"456": {id: "456", Type: "bar"},
			"678": {id: "678", Type: "foo"},
		},
	}
	filter := map[string][]string{
		"type": {"*"},
	}

	devices, err := m.FilterDevices(filter)
	assert.NoError(t, err)
	assert.Len(t, devices, 3)
}

func TestDeviceManager_FilterDevices_error(t *testing.T) {
	m := deviceManager{
		devices: map[string]*Device{
			"123": {id: "123", Type: "foo"},
			"456": {id: "456", Type: "bar"},
			"678": {id: "678", Type: "foo"},
		},
	}
	filter := map[string][]string{ // bad filter
		"something": {"foo"},
	}

	devices, err := m.FilterDevices(filter)
	assert.Error(t, err)
	assert.Empty(t, devices)
}

func TestDeviceManager_createDevices_noConfig(t *testing.T) {
	m := deviceManager{
		devices: map[string]*Device{},
	}

	err := m.createDevices()
	assert.Error(t, err)
	assert.Empty(t, m.devices)
}

func TestDeviceManager_createDevices_failedCreate(t *testing.T) {
	cfg := &config.Devices{
		Devices: []*config.DeviceProto{
			{
				Instances: []*config.DeviceInstance{nil},
			},
		},
	}
	m := deviceManager{
		config:  cfg,
		devices: map[string]*Device{},
	}

	err := m.createDevices()
	assert.Error(t, err)
	assert.Empty(t, m.devices)
}

func TestDeviceManager_createDevices_failedAdd(t *testing.T) {
	cfg := &config.Devices{
		Devices: []*config.DeviceProto{
			{
				Instances: []*config.DeviceInstance{
					{
						Type: "foo", // will fail to add, no Handler defined
					},
				},
			},
		},
	}
	m := deviceManager{
		config:  cfg,
		devices: map[string]*Device{},
	}

	err := m.createDevices()
	assert.Error(t, err)
	assert.Empty(t, m.devices)
}

func TestDeviceManager_createDevices_ok(t *testing.T) {
	cfg := &config.Devices{
		Devices: []*config.DeviceProto{
			{
				Instances: []*config.DeviceInstance{
					{
						Type:    "foo",
						Handler: "foo",
					},
				},
			},
		},
	}
	m := deviceManager{
		config:         cfg,
		tagCache:       NewTagCache(),
		pluginHandlers: NewDefaultPluginHandlers(),
		id:             &pluginID{uuid: uuid.NewSHA1(uuid.NameSpaceDNS, []byte("test"))},
		handlers: map[string]*DeviceHandler{
			"foo": {Name: "foo"},
		},
		devices: map[string]*Device{},
	}

	err := m.createDevices()
	assert.NoError(t, err)
	assert.Len(t, m.devices, 1)
}

func TestDeviceManager_loadConfig_noCfgOptional(t *testing.T) {
	origLocal := localDeviceConfig
	origDefault := defaultDeviceConfig
	d, closer := test.TempDir(t)
	defer func() {
		localDeviceConfig = origLocal
		defaultDeviceConfig = origDefault
		closer()
	}()
	localDeviceConfig = d

	m := deviceManager{
		config: new(config.Devices),
		policies: &policy.Policies{
			DeviceConfig: policy.Optional,
		},
	}

	assert.Empty(t, m.config)

	err := m.loadConfig()
	assert.NoError(t, err)
	assert.Empty(t, m.config)
}

func TestDeviceManager_loadConfig_noCfgRequired(t *testing.T) {
	origLocal := localDeviceConfig
	origDefault := defaultDeviceConfig
	d, closer := test.TempDir(t)
	defer func() {
		localDeviceConfig = origLocal
		defaultDeviceConfig = origDefault
		closer()
	}()
	localDeviceConfig = d

	m := deviceManager{
		config: new(config.Devices),
		policies: &policy.Policies{
			DeviceConfig: policy.Required,
		},
	}

	assert.Empty(t, m.config)

	err := m.loadConfig()
	assert.Error(t, err)
	assert.Empty(t, m.config)
}

func TestDeviceManager_loadConfig_cfgOptional(t *testing.T) {
	origLocal := localDeviceConfig
	defer func() {
		localDeviceConfig = origLocal
	}()
	localDeviceConfig = "./testdata/device"

	m := deviceManager{
		config: new(config.Devices),
		policies: &policy.Policies{
			DeviceConfig: policy.Optional,
		},
	}

	assert.Empty(t, m.config)

	err := m.loadConfig()
	assert.NoError(t, err)
	assert.NotEmpty(t, m.config)
	assert.Equal(t, 3, m.config.Version)
	assert.Len(t, m.config.Devices, 1)
	assert.Len(t, m.config.Devices[0].Instances, 3)
}

func TestDeviceManager_loadConfig_cfgRequired(t *testing.T) {
	origLocal := localDeviceConfig
	defer func() {
		localDeviceConfig = origLocal
	}()
	localDeviceConfig = "./testdata/device"

	m := deviceManager{
		config: new(config.Devices),
		policies: &policy.Policies{
			DeviceConfig: policy.Required,
		},
	}

	assert.Empty(t, m.config)

	err := m.loadConfig()
	assert.NoError(t, err)
	assert.NotEmpty(t, m.config)
	assert.Equal(t, 3, m.config.Version)
	assert.Len(t, m.config.Devices, 1)
	assert.Len(t, m.config.Devices[0].Instances, 3)
}

func TestDeviceManager_execDeviceSetupActions_noActions(t *testing.T) {
	p := &Plugin{}
	m := deviceManager{
		setupActions: []*DeviceAction{},
	}

	err := m.execDeviceSetupActions(p)
	assert.NoError(t, err)
}

func TestDeviceManager_execDeviceSetupActions_withError(t *testing.T) {
	p := &Plugin{}
	m := deviceManager{
		setupActions: []*DeviceAction{
			{
				Name: "failing",
				Filter: map[string][]string{
					"type": {"foo"},
				},
				Action: func(p *Plugin, d *Device) error {
					return fmt.Errorf("test error")
				},
			},
		},
		devices: map[string]*Device{
			"123": {id: "123", Type: "foo"},
			"456": {id: "456", Type: "bar"},
			"678": {id: "678", Type: "foo"},
		},
	}

	err := m.execDeviceSetupActions(p)
	assert.Error(t, err)
}

func TestDeviceManager_execDeviceSetupActions_withBadFilter(t *testing.T) {
	counter := 0
	p := &Plugin{}
	m := deviceManager{
		setupActions: []*DeviceAction{
			{
				Name: "ok",
				Filter: map[string][]string{
					"something": {"foo"},
				},
				Action: func(p *Plugin, d *Device) error {
					counter += 1
					return nil
				},
			},
		},
		devices: map[string]*Device{
			"123": {id: "123", Type: "foo"},
			"456": {id: "456", Type: "bar"},
			"678": {id: "678", Type: "foo"},
		},
	}

	err := m.execDeviceSetupActions(p)
	assert.Error(t, err)
	assert.Equal(t, 0, counter)
}

func TestDeviceManager_execDeviceSetupActions_ok(t *testing.T) {
	counter := 0
	p := &Plugin{}
	m := deviceManager{
		setupActions: []*DeviceAction{
			{
				Name: "ok",
				Filter: map[string][]string{
					"type": {"foo"},
				},
				Action: func(p *Plugin, d *Device) error {
					counter += 1
					return nil
				},
			},
		},
		devices: map[string]*Device{
			"123": {id: "123", Type: "foo"},
			"456": {id: "456", Type: "bar"},
			"678": {id: "678", Type: "foo"},
		},
	}

	err := m.execDeviceSetupActions(p)
	assert.NoError(t, err)
	assert.Equal(t, 2, counter)
}

func TestDeviceManager_execDeviceSetupActions_ok2(t *testing.T) {
	counter := 0
	p := &Plugin{}
	m := deviceManager{
		setupActions: []*DeviceAction{
			{
				Name: "ok",
				Filter: map[string][]string{
					"type": {"foo", "bar"},
				},
				Action: func(p *Plugin, d *Device) error {
					counter += 1
					return nil
				},
			},
		},
		devices: map[string]*Device{
			"123": {id: "123", Type: "foo"},
			"456": {id: "456", Type: "bar"},
			"678": {id: "678", Type: "foo"},
		},
	}

	err := m.execDeviceSetupActions(p)
	assert.NoError(t, err)
	assert.Equal(t, 3, counter)
}

func TestDeviceManager_execDeviceSetupActions_ok3(t *testing.T) {
	counter := 0
	p := &Plugin{}
	m := deviceManager{
		setupActions: []*DeviceAction{
			{
				Name: "ok",
				Action: func(p *Plugin, d *Device) error {
					counter += 1
					return nil
				},
			},
		},
		devices: map[string]*Device{
			"123": {id: "123", Type: "foo"},
			"456": {id: "456", Type: "bar"},
			"678": {id: "678", Type: "foo"},
		},
	}

	err := m.execDeviceSetupActions(p)
	assert.NoError(t, err)
	assert.Equal(t, 0, counter)
}

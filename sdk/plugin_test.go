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
	"testing"
	"time"

	"github.com/google/uuid"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/v2/internal/test"
	"github.com/vapor-ware/synse-sdk/v2/sdk/config"
	"github.com/vapor-ware/synse-sdk/v2/sdk/health"
	"github.com/vapor-ware/synse-sdk/v2/sdk/policy"
)

func TestNewPlugin(t *testing.T) {
	// check that logging gets set to debug
	flagDebug = true
	origPath := currentDirConfig
	metadata = PluginMetadata{Name: "test"}
	defer func() {
		currentDirConfig = origPath
		metadata = PluginMetadata{}
		flagDebug = false
	}()
	currentDirConfig = "./testdata/plugin"

	p, err := NewPlugin()
	assert.NoError(t, err)
	assert.NotNil(t, p)

	assert.Equal(t, version.SDKVersion, p.version.SDKVersion)
	assert.Equal(t, metadata.Name, p.info.Name)
	assert.NotEmpty(t, p.config)
	assert.NotNil(t, p.quit)
	assert.Equal(t, policy.Optional, p.policies.PluginConfig)
	assert.Equal(t, policy.Optional, p.policies.DynamicDeviceConfig)
	assert.Equal(t, policy.Required, p.policies.DeviceConfig)
	assert.NotNil(t, p.pluginHandlers)
	assert.Equal(t, log.DebugLevel, log.GetLevel())
	assert.NotNil(t, p.id)
	assert.NotNil(t, p.health)
	assert.NotNil(t, p.state)
	assert.NotNil(t, p.device)
	assert.NotNil(t, p.scheduler)
	assert.NotNil(t, p.server)
}

func TestNewPlugin_withOptions(t *testing.T) {
	origPath := currentDirConfig
	metadata = PluginMetadata{Name: "test"}
	defer func() {
		currentDirConfig = origPath
		metadata = PluginMetadata{}
		flagDebug = false
	}()
	currentDirConfig = "./testdata/plugin"

	p, err := NewPlugin(
		PluginConfigRequired(),
		DeviceConfigOptional(),
	)

	assert.NoError(t, err)
	assert.NotNil(t, p)

	assert.Equal(t, version.SDKVersion, p.version.SDKVersion)
	assert.Equal(t, metadata.Name, p.info.Name)
	assert.NotEmpty(t, p.config)
	assert.NotNil(t, p.quit)
	assert.Equal(t, policy.Required, p.policies.PluginConfig)
	assert.Equal(t, policy.Optional, p.policies.DynamicDeviceConfig)
	assert.Equal(t, policy.Optional, p.policies.DeviceConfig)
	assert.NotNil(t, p.pluginHandlers)
	assert.Equal(t, log.DebugLevel, log.GetLevel()) // set via config file
	assert.NotNil(t, p.id)
	assert.NotNil(t, p.health)
	assert.NotNil(t, p.state)
	assert.NotNil(t, p.device)
	assert.NotNil(t, p.scheduler)
	assert.NotNil(t, p.server)
}

func TestNewPlugin_noMetadata(t *testing.T) {
	p, err := NewPlugin()
	assert.Error(t, err)
	assert.Nil(t, p)
}

func TestNewPlugin_noConfig(t *testing.T) {
	metadata = PluginMetadata{Name: "test"}
	defer func() {
		metadata = PluginMetadata{}
	}()

	p, err := NewPlugin()
	assert.Error(t, err)
	assert.Nil(t, p)
}

func TestPlugin_RegisterHealthChecks_noneRegistered(t *testing.T) {
	p := Plugin{
		health: health.NewManager(&config.HealthSettings{}),
	}
	assert.Equal(t, 0, p.health.Count())

	err := p.RegisterHealthChecks()
	assert.NoError(t, err)
	assert.Equal(t, 0, p.health.Count())
}

func TestPlugin_RegisterHealthChecks_oneRegistered(t *testing.T) {
	p := Plugin{
		health: health.NewManager(&config.HealthSettings{}),
	}
	assert.Equal(t, 0, p.health.Count())

	check := health.NewPeriodicHealthCheck("test", 1*time.Second, func() error { return nil })

	err := p.RegisterHealthChecks(check)
	assert.NoError(t, err)
	assert.Equal(t, 1, p.health.Count())
}

func TestPlugin_RegisterHealthChecks_multipleRegistered(t *testing.T) {
	p := Plugin{
		health: health.NewManager(&config.HealthSettings{}),
	}
	assert.Equal(t, 0, p.health.Count())

	check1 := health.NewPeriodicHealthCheck("test-1", 1*time.Second, func() error { return nil })
	check2 := health.NewPeriodicHealthCheck("test-2", 1*time.Second, func() error { return nil })
	check3 := health.NewPeriodicHealthCheck("test-3", 1*time.Second, func() error { return nil })

	err := p.RegisterHealthChecks(check1, check2, check3)
	assert.NoError(t, err)
	assert.Equal(t, 3, p.health.Count())
}

func TestPlugin_RegisterHealthChecks_badCheck(t *testing.T) {
	p := Plugin{
		health: health.NewManager(&config.HealthSettings{}),
	}
	assert.Equal(t, 0, p.health.Count())

	check := health.NewPeriodicHealthCheck("", 1*time.Second, func() error { return nil })

	err := p.RegisterHealthChecks(check)
	assert.Error(t, err)
	assert.Equal(t, 0, p.health.Count())
}

func TestPlugin_RegisterOutputs_noOutputs(t *testing.T) {
	p := Plugin{}

	err := p.RegisterOutputs()
	assert.NoError(t, err)
}

func TestPlugin_RegisterPreRunActions_noneRegistered(t *testing.T) {
	p := Plugin{
		preRun: []*PluginAction{},
	}
	assert.Empty(t, p.preRun)

	p.RegisterPreRunActions()
	assert.Empty(t, p.preRun)
}

func TestPlugin_RegisterPreRunActions_oneRegistered(t *testing.T) {
	p := Plugin{
		preRun: []*PluginAction{},
	}
	assert.Empty(t, p.preRun)

	p.RegisterPreRunActions(
		&PluginAction{
			Name:   "action-1",
			Action: func(p *Plugin) error { return nil },
		},
	)
	assert.Len(t, p.preRun, 1)
}

func TestPlugin_RegisterPreRunActions_multipleRegistered(t *testing.T) {
	p := Plugin{
		preRun: []*PluginAction{},
	}
	assert.Empty(t, p.preRun)

	p.RegisterPreRunActions(
		&PluginAction{
			Name:   "action-1",
			Action: func(p *Plugin) error { return nil },
		},
		&PluginAction{
			Name:   "action-2",
			Action: func(p *Plugin) error { return nil },
		},
		&PluginAction{
			Name:   "action-3",
			Action: func(p *Plugin) error { return nil },
		},
	)
	assert.Len(t, p.preRun, 3)
}

func TestPlugin_RegisterPostRunActions_noneRegistered(t *testing.T) {
	p := Plugin{
		postRun: []*PluginAction{},
	}
	assert.Empty(t, p.postRun)

	p.RegisterPostRunActions()
	assert.Empty(t, p.postRun)
}

func TestPlugin_RegisterPostRunActions_oneRegistered(t *testing.T) {
	p := Plugin{
		postRun: []*PluginAction{},
	}
	assert.Empty(t, p.postRun)

	p.RegisterPostRunActions(
		&PluginAction{
			Name:   "action-1",
			Action: func(p *Plugin) error { return nil },
		},
	)
	assert.Len(t, p.postRun, 1)
}

func TestPlugin_RegisterPostRunActions_multipleRegistered(t *testing.T) {
	p := Plugin{
		postRun: []*PluginAction{},
	}
	assert.Empty(t, p.postRun)

	p.RegisterPostRunActions(
		&PluginAction{
			Name:   "action-1",
			Action: func(p *Plugin) error { return nil },
		},
		&PluginAction{
			Name:   "action-2",
			Action: func(p *Plugin) error { return nil },
		},
		&PluginAction{
			Name:   "action-3",
			Action: func(p *Plugin) error { return nil },
		},
	)
	assert.Len(t, p.postRun, 3)
}

func TestPlugin_RegisterDeviceHandlers_noneRegistered(t *testing.T) {
	p := Plugin{
		device: &deviceManager{
			handlers: map[string]*DeviceHandler{},
		},
	}
	assert.Empty(t, p.device.handlers)

	err := p.RegisterDeviceHandlers()
	assert.NoError(t, err)
	assert.Empty(t, p.device.handlers)
}

func TestPlugin_RegisterDeviceHandlers_oneRegistered(t *testing.T) {
	p := Plugin{
		device: &deviceManager{
			handlers: map[string]*DeviceHandler{},
		},
	}
	assert.Empty(t, p.device.handlers)

	err := p.RegisterDeviceHandlers(
		&DeviceHandler{Name: "foo"},
	)
	assert.NoError(t, err)
	assert.Len(t, p.device.handlers, 1)
}

func TestPlugin_RegisterDeviceHandlers_multipleRegistered(t *testing.T) {
	p := Plugin{
		device: &deviceManager{
			handlers: map[string]*DeviceHandler{},
		},
	}
	assert.Empty(t, p.device.handlers)

	err := p.RegisterDeviceHandlers(
		&DeviceHandler{Name: "foo"},
		&DeviceHandler{Name: "bar"},
		&DeviceHandler{Name: "baz"},
	)
	assert.NoError(t, err)
	assert.Len(t, p.device.handlers, 3)
}

func TestPlugin_RegisterDeviceHandlers_conflictingHandlers(t *testing.T) {
	p := Plugin{
		device: &deviceManager{
			handlers: map[string]*DeviceHandler{
				"foo": {Name: "foo"},
			},
		},
	}
	assert.Len(t, p.device.handlers, 1)

	err := p.RegisterDeviceHandlers(
		&DeviceHandler{Name: "foo"},
	)
	assert.Error(t, err)
	assert.Len(t, p.device.handlers, 1)
}

func TestPlugin_RegisterDeviceSetupActions_noneRegistered(t *testing.T) {
	p := Plugin{
		device: &deviceManager{
			setupActions: []*DeviceAction{},
		},
	}
	assert.Empty(t, p.device.setupActions)

	err := p.RegisterDeviceSetupActions()
	assert.NoError(t, err)
	assert.Empty(t, p.device.setupActions)
}

func TestPlugin_RegisterDeviceSetupActions_oneRegistered(t *testing.T) {
	p := Plugin{
		device: &deviceManager{
			setupActions: []*DeviceAction{},
		},
	}
	assert.Empty(t, p.device.setupActions)

	err := p.RegisterDeviceSetupActions(
		&DeviceAction{
			Name:   "action-1",
			Filter: map[string][]string{"type": {"foo"}},
			Action: func(p *Plugin, d *Device) error { return nil },
		},
	)
	assert.NoError(t, err)
	assert.Len(t, p.device.setupActions, 1)
}

func TestPlugin_RegisterDeviceSetupActions_multipleRegistered(t *testing.T) {
	p := Plugin{
		device: &deviceManager{
			setupActions: []*DeviceAction{},
		},
	}
	assert.Empty(t, p.device.setupActions)

	err := p.RegisterDeviceSetupActions(
		&DeviceAction{
			Name:   "action-1",
			Filter: map[string][]string{"type": {"foo"}},
			Action: func(p *Plugin, d *Device) error { return nil },
		},
		&DeviceAction{
			Name:   "action-2",
			Filter: map[string][]string{"type": {"foo"}},
			Action: func(p *Plugin, d *Device) error { return nil },
		},
		&DeviceAction{
			Name:   "action-3",
			Filter: map[string][]string{"type": {"foo"}},
			Action: func(p *Plugin, d *Device) error { return nil },
		},
	)
	assert.NoError(t, err)
	assert.Len(t, p.device.setupActions, 3)
}

func TestPlugin_RegisterDeviceSetupActions_badAction(t *testing.T) {
	p := Plugin{
		device: &deviceManager{
			setupActions: []*DeviceAction{},
		},
	}
	assert.Empty(t, p.device.setupActions)

	err := p.RegisterDeviceSetupActions(&DeviceAction{
		Name: "foo",
		// no filter specified
		Action: func(p *Plugin, d *Device) error {
			return nil
		},
	})
	assert.Error(t, err)
	assert.Empty(t, p.device.setupActions)
}

func TestPlugin_execPreRun_noActions(t *testing.T) {
	p := Plugin{
		preRun: []*PluginAction{},
	}

	err := p.execPreRun()
	assert.NoError(t, err)
}

func TestPlugin_execPreRun_actionsNoError(t *testing.T) {
	p := Plugin{
		preRun: []*PluginAction{
			{
				Name: "test ok action",
				Action: func(p *Plugin) error {
					return nil
				},
			},
		},
	}

	err := p.execPreRun()
	assert.NoError(t, err)
}

func TestPlugin_execPreRun_actionsWithError(t *testing.T) {
	p := Plugin{
		preRun: []*PluginAction{
			{
				Name: "test error action",
				Action: func(p *Plugin) error {
					return fmt.Errorf("test error")
				},
			},
		},
	}

	err := p.execPreRun()
	assert.Error(t, err)
}

func TestPlugin_execPostRun_noActions(t *testing.T) {
	p := Plugin{
		postRun: []*PluginAction{},
	}

	err := p.execPostRun()
	assert.NoError(t, err)
}

func TestPlugin_execPostRun_actionNoError(t *testing.T) {
	p := Plugin{
		postRun: []*PluginAction{
			{
				Name: "test ok action",
				Action: func(p *Plugin) error {
					return nil
				},
			},
		},
	}

	err := p.execPostRun()
	assert.NoError(t, err)
}

func TestPlugin_execPostRun_actionWithError(t *testing.T) {
	p := Plugin{
		postRun: []*PluginAction{
			{
				Name: "test error action",
				Action: func(p *Plugin) error {
					return fmt.Errorf("test error")
				},
			},
		},
	}

	err := p.execPostRun()
	assert.Error(t, err)
}

func TestPlugin_loadConfig_noCfgOptional(t *testing.T) {
	origPath := currentDirConfig
	d, closer := test.TempDir(t)
	defer func() {
		currentDirConfig = origPath
		closer()
	}()
	currentDirConfig = d

	p := Plugin{
		config: new(config.Plugin),
		policies: &policy.Policies{
			PluginConfig: policy.Optional,
		},
	}

	assert.Empty(t, p.config)

	err := p.loadConfig()
	assert.NoError(t, err)
	assert.Empty(t, p.config)
}

func TestPlugin_loadConfig_noCfgRequired(t *testing.T) {
	origPath := currentDirConfig
	d, closer := test.TempDir(t)
	defer func() {
		currentDirConfig = origPath
		closer()
	}()
	currentDirConfig = d

	p := Plugin{
		config: new(config.Plugin),
		policies: &policy.Policies{
			PluginConfig: policy.Required,
		},
	}

	assert.Empty(t, p.config)

	err := p.loadConfig()
	assert.Error(t, err)
	assert.Empty(t, p.config)
}

func TestPlugin_loadConfig_cfgOptional(t *testing.T) {
	origPath := currentDirConfig
	defer func() {
		currentDirConfig = origPath
	}()
	currentDirConfig = "./testdata/plugin"

	p := Plugin{
		config: new(config.Plugin),
		policies: &policy.Policies{
			PluginConfig: policy.Optional,
		},
	}

	assert.Empty(t, p.config)
	err := p.loadConfig()
	assert.NoError(t, err)
	assert.NotEmpty(t, p.config)
	assert.Equal(t, 3, p.config.Version)
	assert.Equal(t, true, p.config.Debug)
}

func TestPlugin_loadConfig_cfgRequired(t *testing.T) {
	origPath := currentDirConfig
	defer func() {
		currentDirConfig = origPath
	}()
	currentDirConfig = "./testdata/plugin"

	p := Plugin{
		config: new(config.Plugin),
		policies: &policy.Policies{
			PluginConfig: policy.Required,
		},
	}

	assert.Empty(t, p.config)

	err := p.loadConfig()
	assert.NoError(t, err)
	assert.NotEmpty(t, p.config)
	assert.Equal(t, 3, p.config.Version)
	assert.Equal(t, true, p.config.Debug)
}

func TestPlugin_initialize_ok(t *testing.T) {
	p := Plugin{
		device: &deviceManager{
			config: &config.Devices{},
			policies: &policy.Policies{
				DeviceConfig:        policy.Optional,
				DynamicDeviceConfig: policy.Optional,
			},
		},
		server: &server{
			conf: &config.NetworkSettings{
				Type:    "tcp",
				Address: "localhost:5001",
			},
		},
		health: health.NewManager(&config.HealthSettings{}),
	}

	err := p.initialize()
	assert.NoError(t, err)
}

func TestPlugin_initialize_fail1(t *testing.T) {
	// fail initializing the device manager (missing config)
	p := Plugin{
		device: &deviceManager{
			policies: &policy.Policies{
				DeviceConfig:        policy.Optional,
				DynamicDeviceConfig: policy.Optional,
			},
		},
		server: &server{
			conf: &config.NetworkSettings{
				Type:    "tcp",
				Address: "localhost:5001",
			},
		},
	}

	err := p.initialize()
	assert.Error(t, err)
}

func TestPlugin_initialize_fail2(t *testing.T) {
	// fail initializing the server (missing config)
	p := Plugin{
		device: &deviceManager{
			config: &config.Devices{},
			policies: &policy.Policies{
				DeviceConfig:        policy.Optional,
				DynamicDeviceConfig: policy.Optional,
			},
		},
		server: &server{},
	}

	err := p.initialize()
	assert.Error(t, err)
}

func Test_handleRunOptions(t *testing.T) {
	// TODO: need to override exiter..
}

func TestPlugin_NewDevice(t *testing.T) {
	p := Plugin{
		device: &deviceManager{
			handlers: map[string]*DeviceHandler{
				"type1": {Name: "type1"},
				"type2": {Name: "type2"},
			},
		},
	}

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
		Handler:   "type2",
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

	device, err := p.NewDevice(proto, instance)
	assert.NoError(t, err)
	assert.Equal(t, "type2", device.Type)
	assert.Equal(t, "testdata", device.Info)
	assert.Equal(t, 2, len(device.Tags))
	assert.Equal(t, map[string]interface{}{"address": "localhost", "port": 5000}, device.Data)
	assert.Equal(t, map[string]string{"foo": "bar", "123": "456"}, device.Context)
	assert.Equal(t, "type2", device.Handler)
	assert.Equal(t, int32(1), device.SortIndex)
	assert.Equal(t, "foo", device.Alias)
	assert.Equal(t, 2, len(device.Transforms))
	assert.Equal(t, "scale [2]", device.Transforms[0].Name())
	assert.Equal(t, "apply [FtoC]", device.Transforms[1].Name())
	assert.Equal(t, 5*time.Second, device.WriteTimeout)
	assert.Equal(t, "temperature", device.Output)
}

func TestPlugin_AddDevice(t *testing.T) {
	handler := DeviceHandler{Name: "foo"}
	pluginid := &pluginID{uuid: uuid.NewSHA1(uuid.NameSpaceDNS, []byte("test"))}
	p := Plugin{
		pluginHandlers: NewDefaultPluginHandlers(),
		id:             pluginid,
		device: &deviceManager{
			aliasCache:     NewAliasCache(),
			tagCache:       NewTagCache(),
			id:             pluginid,
			pluginHandlers: NewDefaultPluginHandlers(),
			handlers: map[string]*DeviceHandler{
				"foo": &handler,
			},
			devices: map[string]*Device{},
		},
	}
	p.device.plugin = &p
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
		Alias: "example-alias-1",
	}

	// Before we add the device, make sure the state is empty.
	assert.Empty(t, p.device.tagCache.cache)
	assert.Empty(t, p.device.aliasCache.cache)
	assert.Empty(t, p.device.devices)

	err := p.AddDevice(&device)
	assert.NoError(t, err)

	// Make sure that the device was added to the manager, and its
	// tags were updated in the tag cache.
	expectedID := "81c0d156-06c0-50de-8e37-410cdb881eaf"
	assert.Len(t, p.device.devices, 1)
	assert.Contains(t, p.device.devices, expectedID)
	assert.Equal(t, &device, p.device.devices[expectedID])

	assert.Len(t, p.device.tagCache.cache, 2)
	assert.Contains(t, p.device.tagCache.cache, "default")
	assert.Contains(t, p.device.tagCache.cache, "system")

	assert.Len(t, p.device.aliasCache.cache, 1)
	assert.Contains(t, p.device.aliasCache.cache, "example-alias-1")

	// Make sure the device was updated with its pertinent fields.
	assert.Equal(t, &handler, device.handler)
	assert.Equal(t, "testtype.foo.bar1", device.idName)
	assert.Equal(t, expectedID, device.id)
	assert.Len(t, device.Tags, 3) // two additional system-generated tags added
}

func TestPlugin_GetDevice(t *testing.T) {
	p := Plugin{
		device: &deviceManager{
			devices: map[string]*Device{
				"123": {id: "123"},
			},
		},
	}

	device := p.GetDevice("123")
	assert.NotNil(t, device)
	assert.Equal(t, "123", device.id)
}

func TestPlugin_GenerateDeviceID(t *testing.T) {
	p := Plugin{
		pluginHandlers: NewDefaultPluginHandlers(),
		id:             &pluginID{uuid: uuid.NewSHA1(uuid.NameSpaceDNS, []byte("test"))},
	}
	d := Device{
		Type:    "foo",
		Handler: "bar",
		Data: map[string]interface{}{
			"key1": "value1",
			"key2": 2,
		},
	}

	devID := p.GenerateDeviceID(&d)
	assert.Equal(t, "e534b6b2-006e-5f61-93c0-b00ae7535155", devID)
}

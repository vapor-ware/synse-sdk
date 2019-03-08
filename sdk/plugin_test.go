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

	"github.com/vapor-ware/synse-sdk/sdk/health"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/internal/test"
	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/policy"
)

func TestNewPlugin(t *testing.T) {

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
	assert.Equal(t, false, p.config.Debug)
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
	assert.Equal(t, false, p.config.Debug)
}

func Test_handleRunOptions(t *testing.T) {
	// TODO: need to override exiter..
}

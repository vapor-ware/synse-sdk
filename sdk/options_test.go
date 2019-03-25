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
	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/policy"
)

// TestCustomDeviceIdentifier tests creating a PluginOption for a custom
// device identifier.
func TestCustomDeviceIdentifier(t *testing.T) {
	opt := CustomDeviceIdentifier(
		func(data map[string]interface{}) string {
			return "foo"
		},
	)
	plugin := Plugin{
		pluginHandlers: &PluginHandlers{},
	}
	assert.Nil(t, plugin.pluginHandlers.DeviceIdentifier)

	opt(&plugin)
	assert.NotNil(t, plugin.pluginHandlers.DeviceIdentifier)
}

// TestCustomDynamicDeviceRegistration tests creating a PluginOption for
// a custom device registration function.
func TestCustomDynamicDeviceRegistration(t *testing.T) {
	opt := CustomDynamicDeviceRegistration(
		func(data map[string]interface{}) ([]*Device, error) {
			return []*Device{}, nil
		},
	)
	plugin := Plugin{
		pluginHandlers: &PluginHandlers{},
	}
	assert.Nil(t, plugin.pluginHandlers.DynamicRegistrar)

	opt(&plugin)
	assert.NotNil(t, plugin.pluginHandlers.DynamicRegistrar)
}

// TestCustomDynamicDeviceConfigRegistration tests creating a PluginOption
// for a custom device config registration function.
func TestCustomDynamicDeviceConfigRegistration(t *testing.T) {
	opt := CustomDynamicDeviceConfigRegistration(
		func(data map[string]interface{}) ([]*config.DeviceProto, error) {
			return []*config.DeviceProto{}, nil
		},
	)
	plugin := Plugin{
		pluginHandlers: &PluginHandlers{},
	}
	assert.Nil(t, plugin.pluginHandlers.DynamicConfigRegistrar)

	opt(&plugin)
	assert.NotNil(t, plugin.pluginHandlers.DynamicConfigRegistrar)
}

// TestCustomDeviceDataValidator tests creating a PluginOption for a custom
// device data validator function.
func TestCustomDeviceDataValidator(t *testing.T) {
	opt := CustomDeviceDataValidator(
		func(i map[string]interface{}) error {
			return nil
		},
	)
	plugin := Plugin{
		pluginHandlers: &PluginHandlers{},
	}
	assert.Nil(t, plugin.pluginHandlers.DeviceDataValidator)

	opt(&plugin)
	assert.NotNil(t, plugin.pluginHandlers.DeviceDataValidator)
}

func TestPluginConfigRequired(t *testing.T) {
	opt := PluginConfigRequired()
	plugin := Plugin{
		policies: &policy.Policies{},
	}
	assert.Empty(t, plugin.policies.PluginConfig)

	opt(&plugin)
	assert.Equal(t, policy.Required, plugin.policies.PluginConfig)
}

func TestDeviceConfigOptional(t *testing.T) {
	opt := DeviceConfigOptional()
	plugin := Plugin{
		policies: &policy.Policies{},
	}
	assert.Empty(t, plugin.policies.DeviceConfig)

	opt(&plugin)
	assert.Equal(t, policy.Optional, plugin.policies.DeviceConfig)
}

func TestDynamicConfigRequired(t *testing.T) {
	opt := DynamicConfigRequired()
	plugin := Plugin{
		policies: &policy.Policies{},
	}
	assert.Empty(t, plugin.policies.DynamicDeviceConfig)

	opt(&plugin)
	assert.Equal(t, policy.Required, plugin.policies.DynamicDeviceConfig)
}

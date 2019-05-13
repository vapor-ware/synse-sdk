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
	log "github.com/sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk/policy"
)

// A PluginOption sets optional configurations or functional capabilities for
// a plugin. This includes things like device identification and device
// registration behaviors.
type PluginOption func(*Plugin)

// CustomDeviceIdentifier lets you set a custom function for creating a deterministic
// identifier for a device using the config data for the device.
func CustomDeviceIdentifier(identifier DeviceIdentifier) PluginOption {
	log.Debug("[options] using custom device identifier")
	return func(plugin *Plugin) {
		plugin.pluginHandlers.DeviceIdentifier = identifier
	}
}

// CustomDynamicDeviceRegistration lets you set a custom function for dynamically registering
// Device instances using the data from the "dynamic registration" field in the Plugin config.
func CustomDynamicDeviceRegistration(registrar DynamicDeviceRegistrar) PluginOption {
	log.Debug("[options] using custom device registration")
	return func(plugin *Plugin) {
		plugin.pluginHandlers.DynamicRegistrar = registrar
	}
}

// CustomDynamicDeviceConfigRegistration lets you set a custom function for dynamically
// registering DeviceConfig instances using the data from the "dynamic registration" field
// in the Plugin config.
func CustomDynamicDeviceConfigRegistration(registrar DynamicDeviceConfigRegistrar) PluginOption {
	log.Debug("[options] using custom device config registration")
	return func(plugin *Plugin) {
		plugin.pluginHandlers.DynamicConfigRegistrar = registrar
	}
}

// CustomDeviceDataValidator lets you set a custom function for validating the Data field
// of a device's config. By default, this data is not validated by the SDK, since it is
// plugin-specific.
func CustomDeviceDataValidator(validator DeviceDataValidator) PluginOption {
	log.Debug("[options] using custom data validator")
	return func(plugin *Plugin) {
		plugin.pluginHandlers.DeviceDataValidator = validator
	}
}

// PluginConfigRequired is a PluginOption which designates that a Plugin should require
// a plugin config and will fail if it does not detect one. By default, a Plugin considers
// them optional and will use a set of default configurations if no config is found.
func PluginConfigRequired() PluginOption {
	return func(plugin *Plugin) {
		plugin.policies.PluginConfig = policy.Required
	}
}

// DeviceConfigOptional is a PluginOption which designates that a Plugin should not require
// device configurations to be required, as they are by default. This can be the case when
// a plugin may support dynamic device configuration, so pre-defined device configs may not
// be specified.
func DeviceConfigOptional() PluginOption {
	return func(plugin *Plugin) {
		plugin.policies.DeviceConfig = policy.Optional
	}
}

// DynamicConfigRequired is a PluginOption which designates that a Plugin requires dynamic
// device configuration. By default, dynamic device configuration is optional. This can
// be set if a plugin is designed to only load devices in a dynamic fashion, and not through
// pre-defined config files.
func DynamicConfigRequired() PluginOption {
	return func(plugin *Plugin) {
		plugin.policies.DynamicDeviceConfig = policy.Required
	}
}

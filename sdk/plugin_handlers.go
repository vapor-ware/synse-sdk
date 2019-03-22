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
	"reflect"
	"sort"

	log "github.com/sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk/config"
)

// DeviceIdentifier is a handler function that produces a string that can be used to
// identify a device deterministically. The returned string should be a composite
// from the Device's config data.
type DeviceIdentifier func(map[string]interface{}) string

// DynamicDeviceRegistrar is a handler function that takes a Plugin config's "dynamic
// registration" data and generates Device instances from it. How this is done
// is specific to the plugin/protocol.
type DynamicDeviceRegistrar func(map[string]interface{}) ([]*Device, error)

// DynamicDeviceConfigRegistrar is a handler function that takes a Plugin config's "dynamic
// registration" data and generates Devices config instances from it. How this is done
// is specific to the plugin/protocol.
type DynamicDeviceConfigRegistrar func(map[string]interface{}) ([]*config.DeviceProto, error)

// DeviceDataValidator is a handler function that takes the `Data` field of a device config
// and performs some validation on it. This allows users to provide validation on the
// plugin-specific config fields.
type DeviceDataValidator func(map[string]interface{}) error

type PluginHandlers struct {
	// DeviceIdentifier is a plugin-defined function for uniquely identifying devices.
	DeviceIdentifier DeviceIdentifier

	// DynamicRegistrar is a plugin-defined function which generates devices dynamically.
	DynamicRegistrar DynamicDeviceRegistrar

	// DynamicConfigRegistrar is a plugin-defined function which generates device configs dynamically.
	DynamicConfigRegistrar DynamicDeviceConfigRegistrar

	// DeviceDataValidator is a plugin-defined function that can be used to validate a
	// Device's Data field.
	DeviceDataValidator DeviceDataValidator
}

// NewDefaultPluginHandlers returns the default set of plugin handlers.
func NewDefaultPluginHandlers() *PluginHandlers {
	return &PluginHandlers{
		DeviceIdentifier:       defaultDeviceIdentifier,
		DynamicRegistrar:       defaultDynamicDeviceRegistration,
		DynamicConfigRegistrar: defaultDynamicDeviceConfigRegistration,
		DeviceDataValidator:    defaultDeviceDataValidator,
	}
}

// defaultDeviceIdentifier is the default implementation that fulfils the DeviceIdentifier
// type for a plugin context.
//
// This implementation creates a string by joining all values found in the provided
// device data map. Non-string values in the map are cast to a string.
func defaultDeviceIdentifier(data map[string]interface{}) string {
	var identifier string

	// To ensure that we get the same identifier reliably, we want to make sure
	// we append the components reliably, so we will sort the keys.
	var keys []string
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := data[key]

		// Check if the value is a map. If so, ignore it. Since maps are
		// not ordered, we cannot use them to create a stable device id.
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Map {
			log.Debug("[sdk] default device identifier - data value is map; skipping")
			continue
		}

		// Instead of implementing our own type checking and casting, just
		// use Sprint. Note that this may be meaningless for complex types.
		identifier += fmt.Sprint(data[key])
	}
	return identifier
}

// defaultDynamicDeviceRegistration is the default implementation that fulfils the
// DynamicDeviceRegistrar type for a plugin context.
//
// This implementation simply returns an empty slice. A plugin will not do any dynamic
// registration by default.
func defaultDynamicDeviceRegistration(_ map[string]interface{}) ([]*Device, error) {
	return []*Device{}, nil
}

// defaultDynamicDeviceConfigRegistration is the default implementation that fulfils the
// DynamicDeviceConfigRegistrar type for a plugin context.
//
// This implementation simply returns an empty slice. A plugin will not do any dynamic
// registration by default.
func defaultDynamicDeviceConfigRegistration(_ map[string]interface{}) ([]*config.DeviceProto, error) {
	return []*config.DeviceProto{}, nil
}

// defaultDeviceDataValidator is the default implementation that fulfils the
// DeviceDataValidator type for a plugin context.
//
// This implementation simply returns nil. By default, this will not do any custom
// validation.
func defaultDeviceDataValidator(_ map[string]interface{}) error {
	return nil
}

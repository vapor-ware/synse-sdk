// Synse SDK
// Copyright (c) 2019-2020 Vapor IO
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

package config

import (
	"errors"
	"time"
)

// Errors for device configuration validation.
var (
	ErrInvalidTransform = errors.New("invalid transform config: must have only one of: 'apply', 'scale'")
)

// Devices is the top-level configuration for devices for Synse plugins.
//
// Devices can be specified in a single configuration file, or in multiple
// configuration files. Each Device config can be merged simply by joining
// all of their `Devices` fields together.
type Devices struct {
	// Version is the major version of the device configuration.
	Version int `yaml:"version,omitempty"`

	// Devices is the collection of devices defined in the configuration.
	Devices []*DeviceProto `yaml:"devices,omitempty"`
}

// DeviceProto defines the "prototype" of a device. It contains some high-level
// information which applies to each of its device instances.
type DeviceProto struct {
	// Type is the type of device. Device types are not strictly defined and
	// are primarily used as metadata for the high-level consumer to help
	// identify and categorize the device. Example types are: LED, fan,
	// temperature, humidity, power, etc.
	//
	// The type should be descriptive and categorical, but is not well-defined,
	// meaning that devices are free to define their own types.
	Type string `yaml:"type,omitempty"`

	// Tags contains the set of tags to apply to each of the devices that
	// are instances of this prototype. It is not required to define tags.
	// All devices will get system-generated tags, so these are supplemental.
	Tags []string `yaml:"tags,omitempty"`

	// Data is any data that can be applied to each of the devices that are
	// instances of this prototype. If specified, this data will be merged
	// with any instance data, where the instance data will override any
	// conflicting values.
	Data map[string]interface{} `yaml:"data,omitempty"`

	// Handler is the name of the plugin's DeviceHandler that will be used
	// for instances of the device prototype. All instances will inherit
	// this handler, but can override.
	Handler string `yaml:"handler,omitempty"`

	// WriteTimeout defines a custom write timeout for all instances of
	// the device prototype. This is the time within which the write
	// transaction will remain valid. If left unspecified, it will fall
	// back to the default value of 30s.
	WriteTimeout time.Duration `yaml:"writeTimeout,omitempty"`

	// Context defines any context information which should be associated
	// with a device instance's reading(s). If specified here, all prototype
	// instances will inherit the context, unless inheritance is disabled.
	Context map[string]string `yaml:"context,omitempty"`

	// Transforms define a collection of operations to apply to the device's
	// reading values to transform it. This could be done for scaling, conversion,
	// etc. See the TransformConfig godoc for details on its configuration.
	//
	// Transforms are applied in the order in which they are specified. That is
	// to say, with the transforms [scale, apply], the scale operation would happen
	// before the apply operation.
	//
	// The transforms defined here will inherited by all instances, unless
	// inheritance is disabled. If both the prototype and the instance specify
	// transform, the prototype transforms are applied first in order, followed
	// by the instance transforms in order.
	Transforms []*TransformConfig

	// Instances contains the data for all configured instances of the
	// device prototype.
	Instances []*DeviceInstance `yaml:"instances,omitempty"`
}

// DeviceInstance defines the instance-specific configuration for a device.
type DeviceInstance struct {
	// Type is the type of device. Device types are not strictly defined and
	// are primarily used as metadata for the high-level consumer to help
	// identify and categorize the device. Example types are: LED, fan,
	// temperature, humidity, power, etc.
	//
	// The type should be descriptive and categorical, but is not well-defined,
	// meaning that devices are free to define their own types.
	Type string `yaml:"type,omitempty"`

	// Info is a string which provides a short human-understandable description
	// or summary of the device instance.
	Info string `yaml:"info,omitempty"`

	// Tags contains the set of tags which apply to the device instance. It
	// is not required to define tags. All devices will get system-generated
	// tags, so these are supplemental.
	Tags []string `yaml:"tags,omitempty"`

	// Context defines any context information which should be associated with
	// the device instance's reading(s). Any values specified here will be
	// applied to the reading context automatically by the SDK.
	Context map[string]string `yaml:"context,omitempty"`

	// Data contains any protocol/plugin/device-specific configuration that
	// is associated with the device instance. It is the responsibility of the
	// plugin to handle these values correctly.
	Data map[string]interface{} `yaml:"data,omitempty"`

	// Output specifies the name of the Output that this device instance
	// will use. This is not needed for all devices/plugins, as many DeviceHandlers
	// will already know which output to use. This field is used in cases of
	// generalized plugins, such as Modbus-IP, where a generalized handler
	// will need to map something (like a set of registers) to a reading output.
	Output string `yaml:"output,omitempty"`

	// SortIndex is a 1-based index that can be used to sort devices in a
	// Synse Server scan. The zero value (0) designates no special sorting
	// for the device.
	SortIndex int32 `yaml:"sortIndex,omitempty"`

	// Handler is the name of the plugin's DeviceHandler that will be used to
	// interface with this device.
	Handler string `yaml:"handler,omitempty"`

	// Alias defines an alias that can be used to reference the device. The
	// alias can either be a pre-defined string, or a template which will
	// be rendered by the SDK.
	//
	// It is up to the configurer to ensure that there are no alias collisions.
	// The SDK can check to ensure no collisions within a single plugin, but
	// can not do so across multiple plugins which may be active in the system.
	Alias *DeviceAlias `yaml:"alias,omitempty"`

	// Transforms define a collection of operations to apply to the device's
	// reading values to transform it. This could be done for scaling, conversion,
	// etc. See the TransformConfig godoc for details on its configuration.
	//
	// Transforms are applied in the order in which they are specified. That is
	// to say, with the transforms [scale, apply], the scale operation would happen
	// before the apply operation.
	//
	// If both the prototype and the instance specify transform, the prototype
	// transforms are applied first in order, followed by the instance
	// transforms in order.
	Transforms []*TransformConfig

	// WriteTimeout defines a custom write timeout for the device instance. This
	// is the time within which the write transaction will remain valid. If left
	// unspecified, it will fall back to the default value of 30s.
	WriteTimeout time.Duration `yaml:"writeTimeout,omitempty"`

	// DisableInheritance determines whether the device instance should inherit
	// from its device prototype.
	DisableInheritance bool `default:"false" yaml:"disableInheritance,omitempty"`
}

// DeviceAlias defines the configuration for setting a device alias.
type DeviceAlias struct {
	// Name is the pre-defined, hardcoded name for the alias.
	Name string `yaml:"name,omitempty"`

	// Template is a Go template string that will be rendered into
	// an alias string by the SDK.
	Template string `yaml:"template,omitempty"`
}

// TransformConfig defines the configuration options for transformations to be
// applied to a device's readings. This is typically used more in general-purpose
// plugin implementations, where the handler does not bake-in any scaling or
// conversions.
//
// TransformConfigs are specified as a list of items underneath the 'transforms'
// key of either the device prototype config or device instance config. Instances
// of the transform config should only define a single field, e.g.
//
//    transforms:
//		- apply: foo
//		- scale: 1
//
// It should never specify multiple, as then the order in which the operations
// should be executed are ambiguous, e.g.
//
//    transforms:
//      - apply: foo
//        scale: 1
//
// The Validate method can be called to ensure that the config adheres to the
// above requirement.
type TransformConfig struct {
	// Apply defines a function to be applied to the device's reading value(s).
	// The function to apply could be anything, e.g. a unit conversion. The SDK
	// defines built-in functions in the 'funcs' package. A plugin may also register
	// custom functions. Functions are referenced here by name.
	Apply string `yaml:"apply,omitempty"`

	// Scale defines a scaling transformation value to be applied to a device's
	// reading(s). The scaling factor defined here is multiplied with the device
	// reading. This allows it to be scaled up (multiplication, e.g. "* 2"), or
	// scaled down (division, e.g. "/ 2" == "* 0.5").
	//
	// This value is specified as a string, but should resolve to a numeric. By
	// default, it will have a value of 1 (e.g. no-op). Negative values and
	// fractional values are supported. This can be the value itself, e.g. "0.01",
	// or a mathematical representation of the value, e.g. "1e-2".
	Scale string `yaml:"scale,omitempty"`
}

// Validate that the TransformConfig adheres to its configuration restrictions.
func (c *TransformConfig) Validate() error {
	if c.Apply != "" && c.Scale != "" {
		return ErrInvalidTransform
	}
	return nil
}

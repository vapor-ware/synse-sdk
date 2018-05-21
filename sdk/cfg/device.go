package cfg

import (
	"fmt"
	"os"

	"github.com/vapor-ware/synse-sdk/sdk/logger"
)

// DeviceConfig holds the configuration for the kinds of devices and the
// instances of those kinds which a plugin will manage.
type DeviceConfig struct {

	// Locations are all of the locations that are defined by the configuration
	// for device instances to reference.
	Locations []Location

	// Devices are all of the DeviceKinds (and subsequently, all of the
	// DeviceInstances) that are defined by the configuration.
	Devices []DeviceKind
}

// Location defines a location (rack, board) which will be associated with
// DeviceInstances. The locational information defined here is used by Synse
// Server to route commands to the proper device instance.
type Location struct {
	Name  string
	Rack  LocationData
	Board LocationData
}

// LocationData defines the name of a locational routing component.
//
// The name of a Location component can either be defined directly via the
// Name field, or from the environment via the FromEnv field.
type LocationData struct {
	Name    string `yaml:"name,omitempty"`
	FromEnv string `yaml:"fromEnv,omitempty"`
}

// Get returns the resolved location data.
//
// This is the preferred method of getting the location component value.
func (locData *LocationData) Get() (string, error) {
	var location string

	if locData.Name != "" {
		location = locData.Name
	}

	if locData.FromEnv != "" {
		// If we already have the location info from the Name field, we
		// will not resolve the FromEnv field and will log out a warning.
		if location != "" {
			logger.Warnf("location fields 'fromEnv' and 'name' are both specified, ignoring 'fromEnv': %+v", locData)
		} else {
			l, ok := os.LookupEnv(locData.FromEnv)
			if !ok {
				return "", fmt.Errorf("no value found for location data from env: %s", locData.FromEnv)
			}
			location = l
		}
	}
	return location, nil
}

// DeviceKind is a kind of device that it being defined.
//
// DeviceKinds are configured as elements of a list under the  "devices" field
// of a DeviceConfig.
type DeviceKind struct {
	// Name is the fully qualified name of the device.
	//
	// The Name of a DeviceKind minimally describes the type of the device, e.g.
	// "temperature". To avoid collisions with DeviceKinds of potentially similar
	// or identical types, the name can be namespaced using '.' as the delimiter,
	// e.g. "foo.temperature".
	//
	// There is no limit to the number of namespace elements. The terminating
	// namespace element should always be the type.
	Name string

	// Metadata contains any metainformation provided for the device. Metadata
	// does not need to be set for a device, but it is recommended, as it makes
	// it easier to identify devices to plugin consumers.
	//
	// There is no restriction on what data can be supplied as metadata.
	Metadata map[string]string

	// Instances contains the configuration data for instances of this DeviceKind.
	Instances []DeviceInstance

	// Outputs describes the reading type outputs provided by instances for this
	// DeviceKind.
	//
	// By default, all DeviceInstances for a DeviceKind will inherit these outputs.
	// This behavior can be changed by setting the DeviceInstance.InheritKindOutputs
	// flag to false.
	Outputs []DeviceOutput
}

// DeviceInstance describes an individual instance of a given DeviceKind.
type DeviceInstance struct {
	// Info is a string that provides a short human-understandable label, description,
	// or summary of the device instance.
	Info string

	// Location is a string that references a named location entry from the
	// "locations" section of the config. It is required, as Synse Server,
	// the consumer of the plugins, routes requests based on this locational
	// information.
	//
	// Note: In future versions of Synse, Location will be deprecated and
	// replaced with a notion of "tags".
	Location string

	// Data contains any protocol/plugin specific configuration associated
	// with the device instance.
	//
	// It is the responsibility of the plugin to handle these values correctly.
	Data map[string]interface{}

	// Outputs describes the reading type output provided by this device instance.
	Outputs []DeviceOutput

	// InheritKindOutputs determines whether the device instance should inherit
	// the Outputs defined in it's DeviceKind. This should be true by default.
	//
	// If this is true, it will inherit all outputs defined by its DeviceKind.
	// If it specifies an output of the same type, the one defined by the
	// DeviceInstance will override the one defined by the DeviceKind, for the
	// DeviceInstance. If the DeviceKind has no outputs defined and this is true,
	// it simply will not inherit anything.
	//
	// If false, this will not inherit any of the DeviceKind's outputs.
	InheritKindOutputs bool
}

// DeviceOutput describes a valid output for the DeviceInstance.
type DeviceOutput struct {
	// Info is a string that provides a short human-understandable label, description,
	// or summary of the device output.
	//
	// This is optional. If this is not set, the Info from its corresponding
	// DeviceInstance is used.
	Info string

	// Type is the name of the ReadingType that describes the expected output format
	// for this device output.
	Type string

	// Data contains any protocol/output specific configuration associated with
	// the device output.
	//
	// Not all device outputs will need their own configuration, in which case, this
	// will remain empty.
	//
	// It is the responsibility of the plugin to handle these values correctly.
	Data map[string]interface{}
}

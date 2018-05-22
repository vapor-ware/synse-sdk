package cfg

import (
	"fmt"
	"os"

	"github.com/vapor-ware/synse-sdk/sdk/logger"
)

/*
TODO:
---------------
- functionality to read in from file
- functionality for default search paths.. '.', '/synse/etc/config', ...

- error handling w/ context (pkg/errors)
- more context w/ error messages (e.g. which file it came from)

- for validation, consider completely validating before returning an error
  so a user can get a list of all issues at once

- think about having some kind of ConfigContext struct that can be associated
  with configs.
	- could describe where the config came from (file, dynamic, env, ...)
	- could provide info depending on where it came from (which env variable(s), which file, ...)

*/

// DeviceConfig holds the configuration for the kinds of devices and the
// instances of those kinds which a plugin will manage.
type DeviceConfig struct {

	// Version is the version of the configuration scheme.
	Version ConfigVersion `yaml:"version,omitempty" addedIn:"1.0"`

	// Locations are all of the locations that are defined by the configuration
	// for device instances to reference.
	Locations []Location `yaml:"locations,omitempty" addedIn:"1.0"`

	// Devices are all of the DeviceKinds (and subsequently, all of the
	// DeviceInstances) that are defined by the configuration.
	Devices []DeviceKind `yaml:"devices,omitempty" addedIn:"1.0"`
}

// Validate validates that the DeviceConfig has no configuration errors.
//
// This is called before Devices are created.
func (deviceConfig *DeviceConfig) Validate() (err error) {
	err = deviceConfig.Version.Validate()
	if err != nil {
		return
	}

	// Validate all of the Locations that the DeviceConfig contains.
	for _, location := range deviceConfig.Locations {
		err = location.Validate()
		if err != nil {
			return
		}
	}

	// Validate all of the DeviceKinds that the DeviceConfig contains.
	for _, deviceKind := range deviceConfig.Devices {
		err = deviceKind.Validate()
		if err != nil {
			return
		}
	}
	return
}

// Location defines a location (rack, board) which will be associated with
// DeviceInstances. The locational information defined here is used by Synse
// Server to route commands to the proper device instance.
type Location struct {
	Name  string       `yaml:"name,omitempty" addedIn:"1.0"`
	Rack  LocationData `yaml:"rack,omitempty" addedIn:"1.0"`
	Board LocationData `yaml:"board,omitempty" addedIn:"1.0"`
}

// Validate validates that the Location has no configuration errors.
func (location *Location) Validate() (err error) {
	// All locations must have a name.
	if location.Name == "" {
		return fmt.Errorf("location has no 'name' set, but is required")
	}

	err = location.Rack.Validate()
	if err != nil {
		return
	}
	err = location.Board.Validate()
	if err != nil {
		return
	}
	return
}

// LocationData defines the name of a locational routing component.
//
// The name of a Location component can either be defined directly via the
// Name field, or from the environment via the FromEnv field.
type LocationData struct {
	Name    string `yaml:"name,omitempty" addedIn:"1.0"`
	FromEnv string `yaml:"fromEnv,omitempty" addedIn:"1.0"`
}

// Validate validates that the LocationData has no configuration errors.
func (locData *LocationData) Validate() error {
	if locData.Name == "" && locData.FromEnv == "" {
		return fmt.Errorf("location requires one of 'name' or 'fromEnv' to be specified, but found neither")
	}
	value, err := locData.Get()
	if err != nil {
		return err
	}
	if value == "" {
		return fmt.Errorf("got empty location info, but location requires a value")
	}

	return nil
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
	Name string `yaml:"name,omitempty" addedIn:"1.0"`

	// Metadata contains any metainformation provided for the device. Metadata
	// does not need to be set for a device, but it is recommended, as it makes
	// it easier to identify devices to plugin consumers.
	//
	// There is no restriction on what data can be supplied as metadata.
	Metadata map[string]string `yaml:"metadata,omitempty" addedIn:"1.0"`

	// Instances contains the configuration data for instances of this DeviceKind.
	Instances []DeviceInstance `yaml:"instances,omitempty" addedIn:"1.0"`

	// Outputs describes the reading type outputs provided by instances for this
	// DeviceKind.
	//
	// By default, all DeviceInstances for a DeviceKind will inherit these outputs.
	// This behavior can be changed by setting the DeviceInstance.InheritKindOutputs
	// flag to false.
	Outputs []DeviceOutput `yaml:"outputs,omitempty" addedIn:"1.0"`
}

// Validate validates that the DeviceKind has no configuration errors.
func (deviceKind *DeviceKind) Validate() (err error) {
	if deviceKind.Name == "" {
		err = fmt.Errorf("device kind requires 'name', but is empty")
		return
	}

	// Validate all of the DeviceInstances that the DeviceKind contains.
	for _, instance := range deviceKind.Instances {
		err = instance.Validate()
		if err != nil {
			return
		}
	}

	// Validate all of the DeviceOutputs that the DeviceKind contains.
	for _, output := range deviceKind.Outputs {
		err = output.Validate()
		if err != nil {
			return
		}
	}
	return
}

// DeviceInstance describes an individual instance of a given DeviceKind.
type DeviceInstance struct {
	// Info is a string that provides a short human-understandable label, description,
	// or summary of the device instance.
	Info string `yaml:"info,omitempty" addedIn:"1.0"`

	// Location is a string that references a named location entry from the
	// "locations" section of the config. It is required, as Synse Server,
	// the consumer of the plugins, routes requests based on this locational
	// information.
	//
	// Note: In future versions of Synse, Location will be deprecated and
	// replaced with a notion of "tags".
	Location string `yaml:"location,omitempty" addedIn:"1.0"`

	// Data contains any protocol/plugin specific configuration associated
	// with the device instance.
	//
	// It is the responsibility of the plugin to handle these values correctly.
	Data map[string]interface{} `yaml:"data,omitempty" addedIn:"1.0"`

	// Outputs describes the reading type output provided by this device instance.
	Outputs []DeviceOutput `yaml:"outputs,omitempty" addedIn:"1.0"`

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
	InheritKindOutputs bool `yaml:"inheritKindOutputs,omitempty" addedIn:"1.0"`
}

// Validate validates that the DeviceInstance has no configuration errors.
func (deviceInstance *DeviceInstance) Validate() (err error) {
	if deviceInstance.Location == "" {
		return fmt.Errorf("device kind requires 'location', but is empty")
	}

	// Validate all of the DeviceOutputs that the DeviceInstance contains.
	for _, output := range deviceInstance.Outputs {
		err = output.Validate()
		if err != nil {
			return
		}
	}
	return
}

// DeviceOutput describes a valid output for the DeviceInstance.
type DeviceOutput struct {
	// Type is the name of the ReadingType that describes the expected output format
	// for this device output.
	Type string `yaml:"type,omitempty" addedIn:"1.0"`

	// Info is a string that provides a short human-understandable label, description,
	// or summary of the device output.
	//
	// This is optional. If this is not set, the Info from its corresponding
	// DeviceInstance is used.
	Info string `yaml:"info,omitempty" addedIn:"1.0"`

	// Data contains any protocol/output specific configuration associated with
	// the device output.
	//
	// Not all device outputs will need their own configuration, in which case, this
	// will remain empty.
	//
	// It is the responsibility of the plugin to handle these values correctly.
	Data map[string]interface{} `yaml:"data,omitempty" addedIn:"1.0"`
}

// Validate validates that the DeviceOutput has no configuration errors.
func (deviceOutput *DeviceOutput) Validate() error {
	if deviceOutput.Type == "" {
		return fmt.Errorf("device output requires 'type', but is empty")
	}
	return nil
}

package cfg

import (
	"fmt"
	"os"

	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
)

// DeviceConfig holds the configuration for the kinds of devices and the
// instances of those kinds which a plugin will manage.
type DeviceConfig struct {

	// ConfigVersion is the version of the configuration scheme.
	ConfigVersion `yaml:",inline"`

	// Locations are all of the locations that are defined by the configuration
	// for device instances to reference.
	Locations []*Location `yaml:"locations,omitempty" addedIn:"1.0"`

	// Devices are all of the DeviceKinds (and subsequently, all of the
	// DeviceInstances) that are defined by the configuration.
	Devices []*DeviceKind `yaml:"devices,omitempty" addedIn:"1.0"`
}

// Validate validates that the DeviceConfig has no configuration errors.
//
// This is called before Devices are created.
func (config DeviceConfig) Validate(multiErr *errors.MultiError) {
	// A version must be specified and it must be of the correct format.
	_, err := config.GetSchemeVersion()
	if err != nil {
		multiErr.Add(errors.NewValidationError(multiErr.Context["source"], err.Error()))
	}
}

// Location defines a location (rack, board) which will be associated with
// DeviceInstances. The locational information defined here is used by Synse
// Server to route commands to the proper device instance.
type Location struct {
	Name  string        `yaml:"name,omitempty"  addedIn:"1.0"`
	Rack  *LocationData `yaml:"rack,omitempty"  addedIn:"1.0"`
	Board *LocationData `yaml:"board,omitempty" addedIn:"1.0"`
}

// Validate validates that the Location has no configuration errors.
func (location Location) Validate(multiErr *errors.MultiError) {
	// All locations must have a name.
	if location.Name == "" {
		multiErr.Add(errors.NewFieldRequiredError(multiErr.Context["source"], "location.name"))
	}

	// Something must be specified for rack
	if location.Rack == nil || *location.Rack == (LocationData{}) {
		multiErr.Add(errors.NewFieldRequiredError(multiErr.Context["source"], "location.rack"))
	}

	// Something must be specified for board
	if location.Board == nil || *location.Board == (LocationData{}) {
		multiErr.Add(errors.NewFieldRequiredError(multiErr.Context["source"], "location.board"))
	}
}

// LocationData defines the name of a locational routing component.
//
// The name of a Location component can either be defined directly via the
// Name field, or from the environment via the FromEnv field.
type LocationData struct {
	Name    string `yaml:"name,omitempty"    addedIn:"1.0"`
	FromEnv string `yaml:"fromEnv,omitempty" addedIn:"1.0"`
}

// Validate validates that the LocationData has no configuration errors.
func (locData LocationData) Validate(multiErr *errors.MultiError) {
	if locData.Name == "" && locData.FromEnv == "" {
		multiErr.Add(errors.NewFieldRequiredError(multiErr.Context["source"], "LocationDat.{type,fromEnv}"))
	}
	value, err := locData.Get()
	if err != nil {
		multiErr.Add(errors.NewValidationError(multiErr.Context["source"], err.Error()))
	}
	if value == "" {
		multiErr.Add(errors.NewValidationError(
			multiErr.Context["source"],
			"got empty location info, but location requires a value",
		))
	}
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
	Instances []*DeviceInstance `yaml:"instances,omitempty" addedIn:"1.0"`

	// Outputs describes the reading type outputs provided by instances for this
	// DeviceKind.
	//
	// By default, all DeviceInstances for a DeviceKind will inherit these outputs.
	// This behavior can be changed by setting the DeviceInstance.InheritKindOutputs
	// flag to false.
	Outputs []*DeviceOutput `yaml:"outputs,omitempty" addedIn:"1.0"`
}

// Validate validates that the DeviceKind has no configuration errors.
func (deviceKind DeviceKind) Validate(multiErr *errors.MultiError) {
	if deviceKind.Name == "" {
		multiErr.Add(errors.NewFieldRequiredError(multiErr.Context["source"], "deviceKind.name"))
	}
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
	Outputs []*DeviceOutput `yaml:"outputs,omitempty" addedIn:"1.0"`

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
func (deviceInstance DeviceInstance) Validate(multiErr *errors.MultiError) {
	// All device instances must be associated with a location
	if deviceInstance.Location == "" {
		multiErr.Add(errors.NewFieldRequiredError(multiErr.Context["source"], "deviceInstance.location"))
	}
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
func (deviceOutput DeviceOutput) Validate(multiErr *errors.MultiError) {
	// All device outputs need to be associated with an output type.
	if deviceOutput.Type == "" {
		multiErr.Add(errors.NewFieldRequiredError(multiErr.Context["source"], "deviceOutput.type"))
	}
}

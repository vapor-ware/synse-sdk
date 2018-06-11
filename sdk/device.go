package sdk

import (
	"fmt"
	"os"
	"strings"

	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
	"github.com/vapor-ware/synse-server-grpc/go"
)

// The current (latest) version of the device config scheme.
var currentDeviceSchemeVersion = "1.0"

// DeviceHandler specifies the read and write handlers for a Device
// based on its type and model.
type DeviceHandler struct {

	// Name is the name of the handler. This is how the handler will be referenced
	// and associated with Device instances via their DeviceConfig. This name should
	// be the same as the "Kind" of the device which it corresponds with.
	//
	// Additionally, there are cases where we may not want the DeviceHandler to match
	// on the name of the Kind, or we may want a subset of a Device Kind's instances
	// to match to a different handler. In that case, the "handlerName" field can be
	// set in the DeviceConfig at either the DeviceKind level (where it would apply
	// for all instances of that kind), or at the DeviceInstance level (where it would
	// apply for only that instance.
	//
	// If the "handlerName" field is specified, it will be used to match against
	// this Name field. Otherwise, the Kind of the device will be used to match
	// against this Name field.
	Name string

	// Write is a function that handles Write requests for the device. If the
	// device does not support writing, this can be left as nil.
	Write func(*Device, *WriteData) error

	// Read is a function that handles Read requests for the device. If the device
	// does not support reading, this can be left as nil.
	Read func(*Device) ([]*Reading, error)

	// BulkRead is a function that handles bulk reading for the device. A bulk read
	// is where all devices of a given kind are read at once, instead of individually.
	// If a device does not support bulk read, this can be left as nil. Additionally,
	// a device can only be bulk read if there is no Read handler set.
	BulkRead func([]*Device) ([]*ReadContext, error)
}

// supportsBulkRead checks if the handler supports bulk reading for its Devices.
//
// If BulkRead is set for the device handler and Read is not, then the handler
// supports bulk reading. If both BulkRead and Read are defined, bulk reading
// will not be considered supported and the handler will default to individual
// reads.
func (deviceHandler *DeviceHandler) supportsBulkRead() bool {
	return deviceHandler.Read == nil && deviceHandler.BulkRead != nil
}

// getDevicesForHandler gets a list of all the devices which use the DeviceHandler.
func (deviceHandler *DeviceHandler) getDevicesForHandler() []*Device {
	var devices []*Device

	for _, v := range ctx.devices {
		if v.Handler == deviceHandler {
			devices = append(devices, v)
		}
	}
	return devices
}

// getHandlerForDevice gets the DeviceHandler for a device, based on the handler name.
func getHandlerForDevice(handlerName string) (*DeviceHandler, error) {
	for _, handler := range ctx.deviceHandlers {
		if handler.Name == handlerName {
			return handler, nil
		}
	}
	return nil, fmt.Errorf("no handler found with name: %s", handlerName)
}

// Device is the internal model for a single device (physical or virtual) that
// a plugin can read to or write from.
type Device struct {
	// The name of the device kind. This is essentially the identifier
	// for the device type.
	Kind string

	// Any metadata associated with the device kind.
	Metadata map[string]string

	// The name of the plugin this device is managed by.
	Plugin string

	// Device-level information specified in the Device's config.
	Info string

	// The location of the Device.
	Location *Location

	// Any plugin-specific configuration data associated with the Device.
	Data map[string]interface{}

	// The outputs supported by the device. A device output may supply more
	// info, such as Data, Info, Type, etc. It is up to the user to extract
	// and use that output info when they perform reads for the Device outputs.
	Outputs []*Output

	// The read/write handler for the device. Handlers should be registered globally.
	Handler *DeviceHandler

	// id is the deterministic id of the device
	id string

	// bulkRead is a flag that determines whether or not the device should be
	// read in bulk, i.e. in a batch with other devices of the same kind.
	bulkRead bool
}

// GetType gets the type of the device. The type of the device is the last
// element in its Kind namespace. For example, with the Kind "foo.bar.temperature",
// the type would be "temperature".
func (device *Device) GetType() string {
	if strings.Contains(device.Kind, ".") {
		nameSpace := strings.Split(device.Kind, ".")
		return nameSpace[len(nameSpace)-1]
	}
	return device.Kind
}

// GetOutput gets the named Output from the Device's output list. If the Output
// is not found, nil is returned.
func (device *Device) GetOutput(name string) *Output {
	for _, output := range device.Outputs {
		if output.Name == name {
			return output
		}
	}
	return nil
}

// makeDevices creates Device instances from a DeviceConfig. The DeviceConfig
// used here should be a unified config, meaning that all DeviceConfigs (either from
// different files or from file and dynamic registration) are merged into a single
// DeviceConfig. This should only be called once all configs have been parsed and
// validated to ensure that the information we have is all correct.
func makeDevices(config *DeviceConfig) ([]*Device, error) {
	var devices []*Device

	// The DeviceConfig we get here should be the unified config.
	for _, kind := range config.Devices {
		for _, instance := range kind.Instances {

			// Get the outputs for the instance.
			instanceOutputs, err := getInstanceOutputs(kind, instance)
			if err != nil {
				return nil, err
			}

			// Get the location
			l, err := config.GetLocation(instance.Location)
			if err != nil {
				return nil, err
			}
			location, err := l.Resolve()
			if err != nil {
				return nil, err
			}

			// Get the DeviceHandler. If a specific handlerName is set in the config,
			// we will use that as the definitive handler. Otherwise, use the kind.
			handlerName := kind.Name
			if kind.HandlerName != "" {
				handlerName = kind.HandlerName
			}
			if instance.HandlerName != "" {
				handlerName = instance.HandlerName
			}
			handler, err := getHandlerForDevice(handlerName)
			if err != nil {
				return nil, err
			}

			device := &Device{
				Kind:     kind.Name,
				Metadata: kind.Metadata,
				Plugin:   metainfo.Name,
				Info:     instance.Info,
				Location: location,
				Data:     instance.Data,
				Outputs:  instanceOutputs,
				Handler:  handler,
			}
			devices = append(devices, device)
		}
	}
	return devices, nil
}

// getInstanceOutputs get the Outputs for a single device instance. It converts
// the instance's DeviceOutput to an Output type, and by doing so unifies that
// output with its corresponding OutputType information.
//
// If output inheritance is enable for the instance (which is it by default),
// this will also take the DeviceOutputs defined by the instance's kind.
func getInstanceOutputs(kind *DeviceKind, instance *DeviceInstance) ([]*Output, error) {
	var instanceOutputs []*Output

	// Create the outputs specific to the instance first.
	for _, o := range instance.Outputs {
		output, err := NewOutputFromConfig(o)
		if err != nil {
			return nil, err
		}
		instanceOutputs = append(instanceOutputs, output)
	}

	// If output inheritance is not disabled, we will take any outputs
	// from the DeviceKind as well. If there is an output with the same
	// name already set from the instance config, we will ignore it.
	if !instance.DisableOutputInheritance {
		for _, o := range kind.Outputs {
			output, err := NewOutputFromConfig(o)
			if err != nil {
				return nil, err
			}
			// Check if the output is already being tracked
			duplicate := false
			for _, tracked := range instanceOutputs {
				if tracked.Name == output.Name {
					duplicate = true
					break
				}
			}
			if !duplicate {
				instanceOutputs = append(instanceOutputs, output)
			}
		}
	}
	return instanceOutputs, nil
}

// Location holds the location information for a Device. This is essentially just
// the config.Location struct, but with all fields fully resolved.
type Location struct {
	Rack  string
	Board string
}

// encode translates the Location to the corresponding gRPC Location message.
func (location *Location) encode() *synse.Location {
	return &synse.Location{
		Rack:  location.Rack,
		Board: location.Board,
	}
}

// Output defines a single output that a device can support. It is the DeviceConfig's
// Output merged with its associated output type.
type Output struct {
	OutputType

	Info string
	Data map[string]interface{}
}

// MakeReading makes a reading for the Output. This is a wrapper around `NewReading`.
func (output *Output) MakeReading(value interface{}) *Reading {
	return NewReading(output, value)
}

// encode translates the Output to the corresponding gRPC Output message.
func (output *Output) encode() *synse.Output {
	sf, err := output.GetScalingFactor()
	if err != nil {
		logger.Errorf("error getting scaling factor: %v", err)
	}

	return &synse.Output{
		Name:          output.Name,
		Type:          output.Type(),
		Precision:     int32(output.Precision),
		ScalingFactor: sf,
		Unit:          output.Unit.Encode(),
	}
}

// NewOutputFromConfig creates a new Output from the DeviceOutput config struct.
func NewOutputFromConfig(config *DeviceOutput) (*Output, error) {
	t, err := getTypeByName(config.Type)
	if err != nil {
		return nil, err
	}

	return &Output{
		OutputType: *t,
		Info:       config.Info,
		Data:       config.Data,
	}, nil
}

// Read performs the read action for the device, as set by its DeviceHandler.
//
// If reading is not supported on the device, an UnsupportedCommandError is
// returned.
// FIXME: should we update the unsupported command error to be more descriptive?
func (device *Device) Read() (*ReadContext, error) {
	if device.IsReadable() {
		readings, err := device.Handler.Read(device)
		if err != nil {
			return nil, err
		}

		return NewReadContext(device, readings), nil
	}
	return nil, &errors.UnsupportedCommandError{}
}

// Write performs the write action for the device, as set by its DeviceHandler.
//
// If writing is not supported on the device, an UnsupportedCommandError is
// returned.
// FIXME: should we update the unsupported command error to be more descriptive?
func (device *Device) Write(data *WriteData) error {
	if device.IsWritable() {
		return device.Handler.Write(device, data)
	}
	return &errors.UnsupportedCommandError{}
}

// IsReadable checks if the Device is readable based on the presence/absence
// of a Read/BulkRead action defined in its DeviceHandler.
func (device *Device) IsReadable() bool {
	return device.Handler.Read != nil || device.Handler.BulkRead != nil
}

// IsWritable checks if the Device is writable based on the presence/absence
// of a Write action defined in its DeviceHandler.
func (device *Device) IsWritable() bool {
	return device.Handler.Write != nil
}

// ID generates the deterministic ID for the Device using its config values.
func (device *Device) ID() string {
	if device.id == "" {
		protocolComp := ctx.deviceIdentifier(device.Data)
		device.id = newUID(device.Plugin, device.Kind, protocolComp)
	}
	return device.id
}

// GUID generates a globally unique ID string by creating a composite
// string from the rack, board, and device UID.
func (device *Device) GUID() string {
	return makeIDString(
		device.Location.Rack,
		device.Location.Board,
		device.ID(),
	)
}

// encode translates the Device to the corresponding gRPC Device message.
func (device *Device) encode() *synse.Device {
	var output []*synse.Output
	for _, out := range device.Outputs {
		output = append(output, out.encode())
	}
	return &synse.Device{
		Timestamp: GetCurrentTime(),
		Uid:       device.ID(),
		Kind:      device.Kind,
		Metadata:  device.Metadata,
		Plugin:    device.Plugin,
		Info:      device.Info,
		Location:  device.Location.encode(),
		Output:    output,
	}
}

// updateDeviceMap updates the global device map with the provided Devices.
// If duplicate IDs are detected, the plugin will terminate.
func updateDeviceMap(devices []*Device) {
	for _, d := range devices {
		if _, hasDevice := ctx.devices[d.GUID()]; hasDevice {
			// If we have devices with the same ID, there is something very wrong
			// happening and we will not want to proceed, since we won't be able
			// to route to devices correctly.
			logger.Fatalf("duplicate device id found: %s", d.GUID())
		}
		ctx.devices[d.GUID()] = d
	}
}

// DeviceConfig holds the configuration for the kinds of devices and the
// instances of those kinds which a plugin will manage.
type DeviceConfig struct {

	// SchemeVersion is the version of the configuration scheme.
	SchemeVersion `yaml:",inline"`

	// Locations are all of the locations that are defined by the configuration
	// for device instances to reference.
	Locations []*LocationConfig `yaml:"locations,omitempty" addedIn:"1.0"`

	// Devices are all of the DeviceKinds (and subsequently, all of the
	// DeviceInstances) that are defined by the configuration.
	Devices []*DeviceKind `yaml:"devices,omitempty" addedIn:"1.0"`
}

// NewDeviceConfig returns a new instance of a DeviceConfig with the SchemeVersion
// set to the latest (most current) device config scheme version, and the Locations
// and Devices fields initialized, but not filled.
func NewDeviceConfig() *DeviceConfig {
	return &DeviceConfig{
		SchemeVersion: SchemeVersion{
			Version: currentDeviceSchemeVersion,
		},
		Locations: []*LocationConfig{},
		Devices:   []*DeviceKind{},
	}
}

// ValidateDeviceConfigData validates the `Data` field(s) of a Device Config to
// ensure that they are correct. The `Data` fields are plugin-specific, so its
// up to the user to provide us with a validation function.
func (config *DeviceConfig) ValidateDeviceConfigData(validator func(map[string]interface{}) error) *errors.MultiError {
	multiErr := errors.NewMultiError("device config 'data' field validation")

	for _, device := range config.Devices {
		// Verify that the DeviceKind Instances' `Data` field is correct
		for _, instance := range device.Instances {
			err := validator(instance.Data)
			if err != nil {
				multiErr.Add(err)
			}
			// Instance Outputs can have their own data too. Verify instance
			// output data.
			for _, output := range instance.Outputs {
				err := validator(output.Data)
				if err != nil {
					multiErr.Add(err)
				}
			}
		}

		// Device kind outputs can have their own data too. Verify the
		// device kind output data.
		for _, output := range device.Outputs {
			err := validator(output.Data)
			if err != nil {
				multiErr.Add(err)
			}
		}
	}
	return multiErr
}

// Validate validates that the DeviceConfig has no configuration errors.
//
// This is called before Devices are created.
func (config DeviceConfig) Validate(multiErr *errors.MultiError) {
	// A version must be specified and it must be of the correct format.
	_, err := config.GetVersion()
	if err != nil {
		multiErr.Add(errors.NewValidationError(multiErr.Context["source"], err.Error()))
	}
}

// GetLocation gets a location from the DeviceConfig by name, if it exists.
// If the specified location name is not associated with any location in the
// DeviceConfig, an error is returned.
func (config *DeviceConfig) GetLocation(name string) (*LocationConfig, error) {
	for _, l := range config.Locations {
		if l.Name == name {
			return l, nil
		}
	}
	return nil, fmt.Errorf("no location with name '%s' was found", name)
}

// LocationConfig defines a location (rack, board) which will be associated with
// DeviceInstances. The locational information defined here is used by Synse
// Server to route commands to the proper device instance.
type LocationConfig struct {
	Name  string        `yaml:"name,omitempty"  addedIn:"1.0"`
	Rack  *LocationData `yaml:"rack,omitempty"  addedIn:"1.0"`
	Board *LocationData `yaml:"board,omitempty" addedIn:"1.0"`
}

// Validate validates that the Location has no configuration errors.
func (location LocationConfig) Validate(multiErr *errors.MultiError) {
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

// Equals checks if another Location is equal to this Location.
func (location *LocationConfig) Equals(other *LocationConfig) bool {
	if location == other {
		return true
	}
	if location.Name == other.Name && location.Rack.Equals(other.Rack) && location.Board.Equals(other.Board) {
		return true
	}
	return false
}

// Resolve resolves the LocationConfig into a Location.
func (location *LocationConfig) Resolve() (*Location, error) {
	multiErr := errors.NewMultiError("creating new Location from config")
	rack, err := location.Rack.Get()
	if err != nil {
		multiErr.Add(err)
	}
	board, err := location.Board.Get()
	if err != nil {
		multiErr.Add(err)
	}

	if multiErr.HasErrors() {
		return nil, multiErr
	}

	return &Location{
		Rack:  rack,
		Board: board,
	}, nil
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

// Equals checks if another LocationData is equal to this LocationData.
func (locData *LocationData) Equals(other *LocationData) bool {
	if locData == other {
		return true
	}
	if locData.Name == other.Name && locData.FromEnv == other.FromEnv {
		return true
	}
	return false
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

	// HandlerName specifies the name of the DeviceHandler to match this DeviceKind
	// with. By default, a DeviceKind will match with a DeviceHandler using its
	// `Kind` field. This field can be set to override that behavior.
	HandlerName string `yaml:"handlerName,omitempty" addedIn:"1.0"`
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

	// DisableOutputInheritance determines whether the device instance should inherit
	// the Outputs defined in it's DeviceKind. This is false by default, meaning that
	// instances will inherit outputs from their DeviceKind. If it specifies an output
	// of the same type, the one defined by the DeviceInstance will override the one
	// defined by the DeviceKind, for the DeviceInstance. If the DeviceKind has no
	// outputs defined, it simply will not inherit anything.
	//
	// If this is true, this instance will not inherit any outputs defined by its
	// DeviceKind.
	DisableOutputInheritance bool `yaml:"disableOutputInheritance,omitempty" addedIn:"1.0"`

	// HandlerName specifies the name of the DeviceHandler to match this DeviceInstance
	// with. By default, a DeviceInstance will match with a DeviceHandler using
	// the `Kind` field of its DeviceKind. This field can be set to override
	// that behavior.
	HandlerName string `yaml:"handlerName,omitempty" addedIn:"1.0"`
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
	// Type is the name of the OutputType that describes the expected output format
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

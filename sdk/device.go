package sdk

import (
	"encoding/json"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-server-grpc/go"
)

// The current (latest) version of the device config scheme.
var currentDeviceSchemeVersion = 3

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

	// Listen is a function that will listen for push-based data from the device.
	// This function is called one per device using the handler, even if there are
	// other handler functions (e.g. read, write) defined. The listener function
	// will run in a separate goroutine for each device. The goroutines are started
	// before the read/write loops.
	Listen func(*Device, chan *ReadContext) error
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
	Handler *DeviceHandler `json:"-"`

	// id is the deterministic id of the device
	id string

	// bulkRead is a flag that determines whether or not the device should be
	// read in bulk, i.e. in a batch with other devices of the same kind.
	bulkRead bool

	// SortOrdinal is a one based sort ordinal for a device in a scan. Zero for
	// don't care.
	SortOrdinal int32
}

// JSON encodes the device as JSON. This can be useful for logging and debugging.
func (device *Device) JSON() (string, error) {
	bytes, err := json.Marshal(device)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
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
func makeDevices(config *DeviceConfig) ([]*Device, error) { // nolint: gocyclo
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
				Kind:        kind.Name,
				Metadata:    kind.Metadata,
				Plugin:      metainfo.Name,
				Info:        instance.Info,
				Location:    location,
				Data:        instance.Data,
				Outputs:     instanceOutputs,
				Handler:     handler,
				SortOrdinal: instance.SortOrdinal,
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
func (output *Output) MakeReading(value interface{}) (reading *Reading, err error) {
	return NewReading(output, value)
}

// encode translates the Output to the corresponding gRPC Output message.
func (output *Output) encode() *synse.Output {
	sf, err := output.GetScalingFactor()
	if err != nil {
		log.Errorf("[sdk] error getting scaling factor: %v", err)
	}

	return &synse.Output{
		Name:          output.Name,
		Type:          output.Type(),
		Precision:     int32(output.Precision),
		ScalingFactor: sf,
		Unit:          output.Unit.encode(),
	}
}

// NewOutputFromConfig creates a new Output from the DeviceOutput config struct.
func NewOutputFromConfig(config *DeviceOutput) (*Output, error) {
	t, err := GetTypeByName(config.Type)
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
	// Bulk read is handled elsewhere.
	// Device may only support bulk read.
	if device == nil {
		return nil, fmt.Errorf("device is nil")
	}
	if device.Handler == nil {
		return nil, fmt.Errorf("device.Handler is nil")
	}
	if device.Handler.Read != nil {
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
	return device.Handler.Read != nil || device.Handler.BulkRead != nil || device.Handler.Listen != nil
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
		Timestamp:   GetCurrentTime(),
		Uid:         device.ID(),
		Kind:        device.Kind,
		Metadata:    device.Metadata,
		Plugin:      device.Plugin,
		Info:        device.Info,
		Location:    device.Location.encode(),
		SortOrdinal: device.SortOrdinal,
		Output:      output,
	}
}

// updateDeviceMap updates the global device map with the provided Devices.
// If duplicate IDs are detected, the plugin will terminate.
func updateDeviceMap(devices []*Device) {
	var foundDuplicates bool
	for _, d := range devices {
		if existing, hasDevice := ctx.devices[d.GUID()]; hasDevice {
			// If we have devices with the same ID, there is something very wrong
			// happening and we will not want to proceed, since we won't be able
			// to route to devices correctly.
			log.WithField("id", d.ID()).Error("[sdk] duplicate device found")
			foundDuplicates = true

			// Get a dump of the device data, including all nested structs
			existingJSON, err := existing.JSON()
			if err != nil {
				log.Errorf("[sdk] failed to dump device to JSON: %v", err)
				log.Errorf("[sdk] existing device: %v", existing)
			} else {
				log.Errorf("[sdk] existing device: %v", existingJSON)
			}
			duplicateJSON, err := d.JSON()
			if err != nil {
				log.Errorf("[sdk] failed to dump device to JSON: %v", err)
				log.Errorf("[sdk] duplicate device: %v", d)
			} else {
				log.Errorf("[sdk] duplicate device: %v", duplicateJSON)
			}
		}
		ctx.devices[d.GUID()] = d
	}
	if foundDuplicates {
		log.Panic("[sdk] unable to run plugin with duplicate device configurations")
	}
}

//// ValidateDeviceConfigData validates the `Data` field(s) of a Device Config to
//// ensure that they are correct. The `Data` fields are plugin-specific, so its
//// up to the user to provide us with a validation function.
//func (config *DeviceConfig) ValidateDeviceConfigData(validator func(map[string]interface{}) error) *errors.MultiError {
//	multiErr := errors.NewMultiError("device config 'data' field validation")
//
//	for _, device := range config.Devices {
//		// Verify that the DeviceKind Instances' `Data` field is correct
//		for _, instance := range device.Instances {
//			err := validator(instance.Data)
//			if err != nil {
//				multiErr.Add(err)
//			}
//			// Instance Outputs can have their own data too. Verify instance
//			// output data.
//			for _, output := range instance.Outputs {
//				err := validator(output.Data)
//				if err != nil {
//					multiErr.Add(err)
//				}
//			}
//		}
//
//		// Device kind outputs can have their own data too. Verify the
//		// device kind output data.
//		for _, output := range device.Outputs {
//			err := validator(output.Data)
//			if err != nil {
//				multiErr.Add(err)
//			}
//		}
//	}
//	return multiErr
//}

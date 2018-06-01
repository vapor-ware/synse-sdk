package sdk

import (
	"github.com/vapor-ware/synse-server-grpc/go"

	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
)

var (
	// The deviceMap holds all of the known devices configured for the plugin.
	deviceMap map[string]*Device

	// The deviceHandlers list holds all of the DeviceHandlers that are registered
	// with the plugin.
	deviceHandlers []*DeviceHandler
)

func init() {
	// Initialize the global variables so they are never nil.
	deviceMap = map[string]*Device{}
	deviceHandlers = []*DeviceHandler{}
}

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

// getDevicesForHandler gets a list of all the devices which use the DeviceHandler.
func (deviceHandler *DeviceHandler) getDevicesForHandler() []*Device {
	var devices []*Device

	for _, v := range deviceMap {
		if v.Handler == deviceHandler {
			devices = append(devices, v)
		}
	}
	return devices
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

// Device is the internal model for a single device (physical or virtual) that
// a plugin can read to or write from.
type Device struct {
	Kind string

	Metadata map[string]string

	Plugin string

	Info string

	Location *Location

	Data map[string]interface{}

	Outputs []*Output

	Handler *DeviceHandler

	// id is the deterministic id of the device
	id string

	// bulkRead is a flag that determines whether or not the device should be
	// read in bulk, i.e. in a batch with other devices of the same kind.
	bulkRead bool
}

// Location holds the location information for a Device. This is essentially just
// the config.Location struct, but with all fields fully resolved.
type Location struct {
	Rack  string
	Board string
}

// encode translates the SDK Location type to the corresponding gRPC Location type.
func (location *Location) encode() *synse.Location {
	return &synse.Location{
		Rack:  location.Rack,
		Board: location.Board,
	}
}

// NewLocationFromConfig creates a new Location from the DeviceOutput location struct.
func NewLocationFromConfig(config *config.Location) (*Location, error) {
	multiErr := errors.NewMultiError("creating new Location from config")
	rack, err := config.Rack.Get()
	if err != nil {
		multiErr.Add(err)
	}
	board, err := config.Board.Get()
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

// GetOutput gets the named output from the Device's output list. If the output
// is not there, nil is returned.
func (device *Device) GetOutput(name string) *Output {
	for _, output := range device.Outputs {
		if output.Name == name {
			return output
		}
	}
	return nil
}

// Output defines a single output that a device can support. It is the DeviceConfig's
// Output merged with its associated output type.
type Output struct {
	ReadingType

	Info string
	Data map[string]interface{}
}

// MakeReading makes a reading for this output. This is a wrapper around NewReading.
func (output *Output) MakeReading(value interface{}) *Reading {
	return NewReading(output, value)
}

// encode translates the SDK Output type to the corresponding gRPC Output type.
func (output *Output) encode() *synse.Output {
	sf, err := output.GetScalingFactor()
	if err != nil {
		logger.Errorf("error getting scaling factor: %v", err)
	}

	return &synse.Output{
		Name:          output.Name,
		Type:          output.Type(),
		DataType:      output.DataType,
		Precision:     int32(output.Precision),
		ScalingFactor: sf,
		Unit:          output.Unit.encode(),
	}
}

// NewOutputFromConfig creates a new Output from the DeviceOutput config struct.
func NewOutputFromConfig(config *config.DeviceOutput) (*Output, error) {
	t, err := getTypeByName(config.Type)
	if err != nil {
		return nil, err
	}

	return &Output{
		ReadingType: *t,
		Info:        config.Info,
		Data:        config.Data,
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

		return NewReadContext(device, readings)
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
// of a Read action defined in its DeviceHandler.
func (device *Device) IsReadable() bool {
	return device.Handler.Read != nil || device.Handler.BulkRead != nil
}

// IsWritable checks if the Device is writable based on the presence/absence
// of a Write action defined in its DeviceHandler.
func (device *Device) IsWritable() bool {
	return device.Handler.Write != nil
}

// ID generates the ID for the Device.
func (device *Device) ID() string {
	if device.id == "" {
		protocolComp := Context.deviceIdentifier(device.Data)
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

// encode translates the SDK Device to its corresponding gRPC Device.
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
		if _, hasDevice := deviceMap[d.GUID()]; hasDevice {
			// If we have devices with the same ID, there is something very wrong
			// happening and we will not want to proceed, since we won't be able
			// to route to devices correctly.
			logger.Fatalf("duplicate device id found: %s", d.GUID())
		}
		deviceMap[d.GUID()] = d
	}
}

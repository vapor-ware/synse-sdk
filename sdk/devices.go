package sdk

import (
	"github.com/vapor-ware/synse-server-grpc/go"

	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/types"
)

// The deviceMap holds all of the known devices configured for the plugin.
var deviceMap = make(map[string]*Device)

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

	Location *config.Location

	Data map[string]interface{}

	Outputs []*Output

	Handler *DeviceHandler

	// id is the deterministic id of the device
	id string

	// bulkRead is a flag that determines whether or not the device should be
	// read in bulk, i.e. in a batch with other devices of the same kind.
	bulkRead bool
}

type Output struct {
	types.ReadingType

	Info string
	Data map[string]interface{}
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

//// NewDevice creates a new instance of a Device.
////
//// A Device serves as the internal model for a single physical or virtual
//// device that a plugin manages, e.g. a temperature sensor. The Device
//// meta information is joined from the device's prototype config and its
//// instance config.
//func NewDevice(p *config.PrototypeConfig, d *config.DeviceConfig, h *DeviceHandler, plugin *Plugin) (*Device, error) {
//	if plugin.handlers.DeviceIdentifier == nil {
//		return nil, fmt.Errorf("identifier function not defined for device")
//	}
//
//	if p.Type != d.Type {
//		return nil, fmt.Errorf("prototype and instance config mismatch (type): %v != %v", p.Type, d.Type)
//	}
//
//	if p.Model != d.Model {
//		return nil, fmt.Errorf("prototype and instance config mismatch (model): %v != %v", p.Model, d.Model)
//	}
//
//	dev := Device{
//		Type:         p.Type,
//		Model:        p.Model,
//		Manufacturer: p.Manufacturer,
//		Protocol:     p.Protocol,
//		Output:       p.Output,
//		Location:     d.Location,
//		Data:         d.Data,
//		Handler:      h,
//		Identifier:   plugin.handlers.DeviceIdentifier,
//		pconfig:      p,
//		dconfig:      d,
//		bulkRead:     h.doesBulkRead(),
//	}
//	return &dev, nil
//}

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
	return nil, &UnsupportedCommandError{}
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
	return &UnsupportedCommandError{}
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
		protocolComp := device.Identifier(device.Data)
		device.id = newUID(device.Protocol, device.Type, device.Model, protocolComp)
	}
	return device.id
}

// GUID generates a globally unique ID string by creating a composite
// string from the rack, board, and device UID.
func (device *Device) GUID() string {
	rack, _ := device.Location.GetRack()
	return makeIDString(rack, device.Location.Board, device.ID())
}

// encode translates the SDK Device to its corresponding gRPC Device.
func (device *Device) encode() *synse.Device {
	var output []*synse.Output
	for _, out := range device.Outputs {
		output = append(output, out.Encode())
	}
	return &synse.Device{
		Timestamp: GetCurrentTime(),
		Uid:       device.ID(),
		Kind:      device.Kind,
		Metadata:  device.Metadata,
		Plugin:    device.Plugin,
		Info:      device.Info,
		Location:  device.Location.Encode(),
		Output:    output,
	}
}

//// devicesFromConfig generates device instance configurations from YAML, if any
//// are specified.
//func devicesFromConfig() ([]*config.DeviceConfig, error) {
//	logger.Debug("devicesFromConfig start")
//
//	var configs []*config.DeviceConfig
//	deviceConfig, err := config.ParseDeviceConfig()
//	if err != nil {
//		logger.Errorf("error when parsing device configs: %v", err)
//		return nil, err
//	}
//	configs = append(configs, deviceConfig...)
//
//	return configs, nil
//}
//
//// devicesFromAutoEnum generates device instance configurations based on the auto
//// enumeration configuration and handler specified for the plugin.
//func devicesFromAutoEnum(plugin *Plugin) ([]*config.DeviceConfig, error) {
//	var configs []*config.DeviceConfig
//
//	// get any instance configurations from the enumerator function registered
//	// with the plugin, if any is registered.
//	autoEnum := plugin.Config.AutoEnumerate
//	if len(autoEnum) > 0 {
//		if plugin.handlers.DeviceEnumerator == nil {
//			logger.Errorf("no device enumerator function registered with the plugin")
//			return nil, fmt.Errorf("no device enumerator function registered with the plugin")
//		}
//
//		for _, c := range autoEnum {
//			deviceConfigs, err := plugin.handlers.DeviceEnumerator(c)
//			if err != nil {
//				logger.Errorf("failed to enumerate devices with %#v: %v", c, err)
//			} else {
//				configs = append(configs, deviceConfigs...)
//			}
//		}
//	}
//	logger.Debugf("device configs from auto-enumeration: %v", configs)
//	return configs, nil
//}
//
//// registerDevices registers all devices specified in the given device configurations
//// with the Plugin by matching them up with their corresponding prototype config and
//// creating new Devices for each of those devices.
//func registerDevices(plugin *Plugin, deviceConfigs []*config.DeviceConfig) error {
//	logger.Debugf("registering devices with the plugin")
//
//	// get the prototype configuration from YAML
//	protoConfigs, err := config.ParsePrototypeConfig()
//	if err != nil {
//		logger.Errorf("failed to parse the device prototype configuration: %v", err)
//		return err
//	}
//
//	devices, err := makeDevices(deviceConfigs, protoConfigs, plugin)
//	if err != nil {
//		logger.Errorf("failed to make devices from found configs: %v", err)
//		return err
//	}
//
//	for _, device := range devices {
//		deviceMap[device.GUID()] = device
//		deviceMapOrder = append(deviceMapOrder, device.GUID())
//	}
//	logger.Debugf("finished registering devices")
//	return nil
//}

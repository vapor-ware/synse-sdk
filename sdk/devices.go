package sdk

import (
	"fmt"

	"github.com/vapor-ware/synse-server-grpc/go"

	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
)

// The deviceMap holds all of the known devices configured for the plugin.
var deviceMap = make(map[string]*Device)

// deviceMapOrder holds all of the known deviceIds in the order of insertion.
// This allows deviceMap iterations in insertion order.
var deviceMapOrder []string

// DeviceRead is a function that defines the read behavior for a Device. It
// is used by the DeviceHandler, and thus is used to handle reads for all device
// instances of a given type and model.
type DeviceRead func(*Device) ([]*Reading, error)

// DeviceWrite is a function that defines the write behavior for a Device. It
// is used by the DeviceHandler, and thus is used to handle writes for all device
// instances of a given type and model.
type DeviceWrite func(*Device, *WriteData) error

// DeviceHandler specifies the read and write handlers for a Device
// based on its type and model.
type DeviceHandler struct {
	Type  string
	Model string

	Write DeviceWrite
	Read  DeviceRead

	BulkRead func([]*Device) ([]*ReadContext, error)
}

// getDevicesForHandler gets a list of all the devices which use the
// DeviceHandler.
func (deviceHandler *DeviceHandler) getDevicesForHandler() []*Device {
	var devices []*Device

	for _, v := range deviceMap {
		if v.Handler == deviceHandler {
			devices = append(devices, v)
		}
	}
	return devices
}

// doesBulkRead checks if the handler does bulk reads for its Devices.
//
// If BulkRead is set for the device handler and Read is not, then
// we will have enabled bulk reads. If both are defined, this will
// not be set, so we will never bulk read.
func (deviceHandler *DeviceHandler) doesBulkRead() bool {
	return deviceHandler.Read == nil && deviceHandler.BulkRead != nil
}

// NewDevice creates a new instance of a Device.
//
// A Device serves as the internal model for a single physical or virtual
// device that a plugin manages, e.g. a temperature sensor. The Device
// meta information is joined from the device's prototype config and its
// instance config.
func NewDevice(p *config.PrototypeConfig, d *config.DeviceConfig, h *DeviceHandler, plugin *Plugin) (*Device, error) {
	if plugin.handlers.DeviceIdentifier == nil {
		return nil, fmt.Errorf("identifier function not defined for device")
	}

	if p.Type != d.Type {
		return nil, fmt.Errorf("prototype and instance config mismatch (type): %v != %v", p.Type, d.Type)
	}

	if p.Model != d.Model {
		return nil, fmt.Errorf("prototype and instance config mismatch (model): %v != %v", p.Model, d.Model)
	}

	dev := Device{
		Type:         p.Type,
		Model:        p.Model,
		Manufacturer: p.Manufacturer,
		Protocol:     p.Protocol,
		Output:       p.Output,
		Location:     d.Location,
		Data:         d.Data,
		Handler:      h,
		Identifier:   plugin.handlers.DeviceIdentifier,
		pconfig:      p,
		dconfig:      d,
		bulkRead:     h.doesBulkRead(),
	}
	return &dev, nil
}

// Device is the internal model for a single device (physical or virtual) that
// a plugin can read to or write from.
type Device struct {
	// prototype
	pconfig *config.PrototypeConfig

	Type         string
	Model        string
	Manufacturer string
	Protocol     string
	Output       []config.DeviceOutput

	// instance
	dconfig  *config.DeviceConfig
	Location config.Location
	Data     map[string]string

	Handler    *DeviceHandler
	Identifier DeviceIdentifier

	// flag to determine whether this device is read in bulk
	// or if it is read individually
	bulkRead bool

	id string
}

// Read performs the read action for the device, as set by its DeviceHandler.
// If reading is not supported on the device, an UnsupportedCommandError is
// returned.
func (d *Device) Read() (*ReadContext, error) {
	if d.IsReadable() {
		readings, err := d.Handler.Read(d)
		if err != nil {
			return nil, err
		}

		return NewReadContext(d, readings)
	}
	return nil, &UnsupportedCommandError{}
}

// Write performs the write action for the device, as set by its DeviceHandler.
// If writing is not supported on the device, an UnsupportedCommandError is
// returned.
func (d *Device) Write(data *WriteData) error {
	if d.IsWritable() {
		return d.Handler.Write(d, data)
	}
	return &UnsupportedCommandError{}
}

// IsReadable checks if the Device is readable based on the presence/absence
// of a Read action defined in its DeviceHandler.
func (d *Device) IsReadable() bool {
	return d.Handler.Read != nil || d.Handler.BulkRead != nil
}

// IsWritable checks if the Device is writable based on the presence/absence
// of a Write action defined in its DeviceHandler.
func (d *Device) IsWritable() bool {
	return d.Handler.Write != nil
}

// ID generates the ID for the Device.
func (d *Device) ID() string {
	if d.id == "" {
		protocolComp := d.Identifier(d.Data)
		d.id = newUID(d.Protocol, d.Type, d.Model, protocolComp)
	}
	return d.id
}

// GUID generates a globally unique ID string by creating a composite
// string from the rack, board, and device UID.
func (d *Device) GUID() string {
	rack, _ := d.Location.GetRack()
	return makeIDString(rack, d.Location.Board, d.ID())
}

// encode translates the Device to a corresponding gRPC MetainfoResponse.
func (d *Device) encode() *synse.MetainfoResponse {
	var output []*synse.MetaOutput
	for _, out := range d.Output {
		mo := out.Encode()
		output = append(output, mo)
	}

	return &synse.MetainfoResponse{
		Timestamp:    GetCurrentTime(),
		Uid:          d.ID(),
		Type:         d.Type,
		Model:        d.Model,
		Manufacturer: d.Manufacturer,
		Protocol:     d.Protocol,
		Info:         d.Data["info"],
		Comment:      d.Data["comment"],
		Location:     d.Location.Encode(),
		Output:       output,
	}
}

// devicesFromConfig generates device instance configurations from YAML, if any
// are specified.
func devicesFromConfig() ([]*config.DeviceConfig, error) {
	logger.Debug("devicesFromConfig start")

	var configs []*config.DeviceConfig
	deviceConfig, err := config.ParseDeviceConfig()
	if err != nil {
		logger.Errorf("error when parsing device configs: %v", err)
		return nil, err
	}
	configs = append(configs, deviceConfig...)

	return configs, nil
}

// devicesFromAutoEnum generates device instance configurations based on the auto
// enumeration configuration and handler specified for the plugin.
func devicesFromAutoEnum(plugin *Plugin) ([]*config.DeviceConfig, error) {
	var configs []*config.DeviceConfig

	// get any instance configurations from the enumerator function registered
	// with the plugin, if any is registered.
	autoEnum := plugin.Config.AutoEnumerate
	if len(autoEnum) > 0 {
		if plugin.handlers.DeviceEnumerator == nil {
			logger.Errorf("no device enumerator function registered with the plugin")
			return nil, fmt.Errorf("no device enumerator function registered with the plugin")
		}

		for _, c := range autoEnum {
			deviceConfigs, err := plugin.handlers.DeviceEnumerator(c)
			if err != nil {
				logger.Errorf("failed to enumerate devices with %#v: %v", c, err)
			} else {
				configs = append(configs, deviceConfigs...)
			}
		}
	}
	logger.Debugf("device configs from auto-enumeration: %v", configs)
	return configs, nil
}

// registerDevices registers all devices specified in the given device configurations
// with the Plugin by matching them up with their corresponding prototype config and
// creating new Devices for each of those devices.
func registerDevices(plugin *Plugin, deviceConfigs []*config.DeviceConfig) error {
	logger.Debugf("registering devices with the plugin")

	// get the prototype configuration from YAML
	protoConfigs, err := config.ParsePrototypeConfig()
	if err != nil {
		logger.Errorf("failed to parse the device prototype configuration: %v", err)
		return err
	}

	devices, err := makeDevices(deviceConfigs, protoConfigs, plugin)
	if err != nil {
		logger.Errorf("failed to make devices from found configs: %v", err)
		return err
	}

	for _, device := range devices {
		deviceMap[device.GUID()] = device
		deviceMapOrder = append(deviceMapOrder, device.GUID())
	}
	logger.Debugf("finished registering devices")
	return nil
}

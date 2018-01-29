package sdk

import (
	"time"

	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
	"github.com/vapor-ware/synse-server-grpc/go"
	"fmt"
)

// The deviceMap holds all of the known devices configured for the plugin.
var deviceMap = make(map[string]*Device)

// DeviceRead is a function that defines the read behavior for a Device.
type DeviceRead func(*Device) ([]*Reading, error)

// DeviceWrite is a function that defines the write behavior for a Device.
type DeviceWrite func(*Device, *WriteData) error

// DeviceHandler specifies the read and write handlers for a certain device
// based on its type and model.
type DeviceHandler struct {
	Type  string
	Model string

	Write DeviceWrite
	Read  DeviceRead
}

// NewDevice creates a new instance of a Device.
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
	}
	return &dev, nil
}

// Device is the internal model for a device (whether physical or virtual)
// that a plugin can read to or write from.
type Device struct {
	// prototype
	pconfig      *config.PrototypeConfig
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

	id string
}

// Read performs the read action for the device, as set by its DeviceHandler
// implementation. If reading is not supported on the device, an Unsupported
// Command Error is returned.
func (d *Device) Read() (*ReadContext, error) {
	if d.IsReadable() {
		readings, err := d.Handler.Read(d)
		if err != nil {
			return nil, err
		}
		return &ReadContext{
			Device:  d.ID(),
			Board:   d.Location.Board,
			Rack:    d.Location.Rack,
			Reading: readings,
		}, nil

	}
	return nil, &UnsupportedCommandError{}
}

// Write performs the write action for the device, as set by its DeviceHandler
// implementation. If writing is not supported on the device, an Unsupported
// Command Error is returned.
func (d *Device) Write(data *WriteData) error {
	if d.IsWritable() {
		return d.Handler.Write(d, data)
	}
	return &UnsupportedCommandError{}
}

// IsReadable checks if the Device is readable via the presence/absence of
// a Read action defined in its DeviceHandler.
func (d *Device) IsReadable() bool {
	return d.Handler.Read != nil
}

// IsWritable checks if the Device is writable via the presence/absence of
// a Write action defined in its DeviceHandler.
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
	return makeIDString(d.Location.Rack, d.Location.Board, d.ID())
}

// encode translates the Device to a corresponding gRPC MetainfoResponse.
func (d *Device) encode() *synse.MetainfoResponse {
	var output []*synse.MetaOutput
	for _, out := range d.Output {
		mo := out.Encode()
		output = append(output, mo)
	}

	return &synse.MetainfoResponse{
		Timestamp:    time.Now().String(),
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

// registerDevicesFromConfig reads in the device configuration files and generates
// Device instances based on those configurations.
func registerDevicesFromConfig(handlers *Handlers, devHandlers []*DeviceHandler, autoEnumCfg []map[string]interface{}, plugin *Plugin) error {
	var instanceCfg []*config.DeviceConfig

	// get any instance configurations from plugin-defined enumeration function
	// FIXME - maybe this should be its own fn, and registering from config is its own fn
	//   then the results of the two can be merged and passed along to the "make devices"
	//   function? doing so might clean up the usage pattern here and not make the function
	//   signature so ugly.
	for _, enumCfg := range autoEnumCfg {
		deviceEnum, err := handlers.DeviceEnumerator(enumCfg)
		if err != nil {
			logger.Errorf("Error enumerating devices with %+v: %v", enumCfg, err)
		} else {
			instanceCfg = append(instanceCfg, deviceEnum...)
		}
	}

	// get any instance configurations from YAML
	deviceCfg, err := config.ParseDeviceConfig()
	if err != nil {
		return err
	}
	instanceCfg = append(instanceCfg, deviceCfg...)

	// get the prototype configurations from YAML
	protoCfg, err := config.ParsePrototypeConfig()
	if err != nil {
		return err
	}

	// make the composite device records
	devices, err := makeDevices(instanceCfg, protoCfg, handlers, devHandlers, plugin)
	if err != nil {
		return err
	}

	for _, device := range devices {
		deviceMap[device.GUID()] = device
	}
	return nil
}

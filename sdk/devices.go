package sdk

import (
	"time"

	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-server-grpc/go"
)

// The deviceMap holds all of the known devices configured for the plugin.
var deviceMap = make(map[string]*Device)

// Device describes a single configured device for the plugin.
type Device struct {
	Prototype *config.PrototypeConfig
	Instance  *config.DeviceConfig
	Handler   DeviceHandler
}

// Type gets the configured type of the Device.
func (d *Device) Type() string {
	return d.Prototype.Type
}

// Model gets the configured model of the Device.
func (d *Device) Model() string {
	return d.Prototype.Model
}

// Manufacturer gets the configured manufacturer of the Device.
func (d *Device) Manufacturer() string {
	return d.Prototype.Manufacturer
}

// Protocol gets the configured protocol of the Device.
func (d *Device) Protocol() string {
	return d.Prototype.Protocol
}

// ID gets the id for the Device.
func (d *Device) ID() string {
	protocolComp := d.Handler.GetProtocolIdentifiers(d.Data())
	return newUID(d.Protocol(), d.Type(), d.Model(), protocolComp)
}

// GUID generates a globally unique ID string by creating a composite
// string from the rack, board, and device UID.
func (d *Device) GUID() string {
	return makeIDString(d.Location().Rack, d.Location().Board, d.ID())
}

// Output gets the list of configured reading outputs for the Device.
func (d *Device) Output() []config.DeviceOutput {
	return d.Prototype.Output
}

// Location gets the configured location of the Device.
func (d *Device) Location() config.Location {
	return d.Instance.Location
}

// Data gets the plugin-specific data for the device. This is left as a map
// of string to string (how it is read from the config YAML) and is left to
// the plugin itself to parse further.
func (d *Device) Data() map[string]string {
	return d.Instance.Data
}

// encode translates the Device to a corresponding gRPC MetainfoResponse.
func (d *Device) encode() *synse.MetainfoResponse {

	location := d.Location()

	var output []*synse.MetaOutput
	for _, out := range d.Output() {
		mo := out.Encode()
		output = append(output, mo)
	}

	return &synse.MetainfoResponse{
		Timestamp:    time.Now().String(),
		Uid:          d.ID(),
		Type:         d.Type(),
		Model:        d.Model(),
		Manufacturer: d.Manufacturer(),
		Protocol:     d.Protocol(),
		Info:         d.Data()["info"],
		Comment:      d.Data()["comment"],
		Location:     location.Encode(),
		Output:       output,
	}
}

// registerDevicesFromConfig reads in the device configuration files and generates
// Device instances based on those configurations.
func registerDevicesFromConfig(handler DeviceHandler, autoEnumCfg []map[string]interface{}) error {
	var instanceCfg []*config.DeviceConfig

	// get any instance configurations from plugin-defined enumeration function
	for _, enumCfg := range autoEnumCfg {
		deviceEnum, err := handler.EnumerateDevices(enumCfg)
		if err != nil {
			Logger.Errorf("Error enumerating devices with %+v: %v", enumCfg, err)
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
	devices := makeDevices(instanceCfg, protoCfg, handler)

	for _, device := range devices {
		deviceMap[device.GUID()] = device
	}
	return nil
}

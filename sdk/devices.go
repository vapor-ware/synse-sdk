package sdk

import (
	"time"

	"github.com/vapor-ware/synse-server-grpc/go"
)

const (
	// the directory which contains the device configurations.
	// FIXME - this is currently relative to the binary.. should be configurable?
	configDir = "config"
)

var deviceMap = make(map[string]*Device)

// Device describes a single configured device for the plugin.
type Device struct {
	Prototype *PrototypeConfig
	Instance  *DeviceConfig
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

// UID gets the id for the Device.
func (d *Device) UID() string {
	protocolComp := d.Handler.GetProtocolIdentifiers(d.Data())
	return newUID(d.Protocol(), d.Type(), d.Model(), protocolComp)
}

// IDString generates a globally unique ID string by creating a composite
// string from the rack, board, and device UID.
func (d *Device) IDString() string {
	return makeIDString(d.Location().Rack, d.Location().Board, d.UID())
}

// Output gets the list of configured reading outputs for the Device.
func (d *Device) Output() []DeviceOutput {
	return d.Prototype.Output
}

// Location gets the configured location of the Device.
func (d *Device) Location() DeviceLocation {
	return d.Instance.Location
}

// Data gets the plugin-specific data for the device. This is left as a map
// of string to string (how it is read from the config YAML) and is left to
// the plugin itself to parse further.
func (d *Device) Data() map[string]string {
	return d.Instance.Data
}

// toMetainfoResponse converts the Device into its corresponding
// MetainfoResponse representation.
func (d *Device) toMetainfoResponse() *synse.MetainfoResponse {

	location := d.Location()

	var output []*synse.MetaOutput
	for _, out := range d.Output() {
		mo := out.toMetaOutput()
		output = append(output, mo)
	}

	return &synse.MetainfoResponse{
		Timestamp:    time.Now().String(),
		Uid:          d.UID(),
		Type:         d.Type(),
		Model:        d.Model(),
		Manufacturer: d.Manufacturer(),
		Protocol:     d.Protocol(),
		Info:         d.Data()["info"],
		Comment:      d.Data()["comment"],
		Location:     location.toMetaLocation(),
		Output:       output,
	}
}

func registerDevicesFromConfig(handler DeviceHandler) error {

	var instanceCfg []*DeviceConfig

	// get any instance configurations from plugin-defined enumeration function
	for _, enumCfg := range Config.AutoEnumerate {
		deviceEnum, err := handler.EnumerateDevices(enumCfg)
		if err != nil {
			Logger.Errorf("Error enumerating devices with %+v: %v", enumCfg, err)
		} else {
			instanceCfg = append(instanceCfg, deviceEnum...)
		}
	}

	// get any instance configurations from YAML
	deviceCfg, err := parseDeviceConfig(configDir)
	if err != nil {
		return err
	}
	instanceCfg = append(instanceCfg, deviceCfg...)

	// get the prototype configurations from YAML
	protoCfg, err := parsePrototypeConfig(configDir)
	if err != nil {
		return err
	}

	// make the composite device records
	devices := makeDevices(instanceCfg, protoCfg, handler)

	for _, device := range devices {
		deviceMap[device.IDString()] = device
	}
	return nil
}

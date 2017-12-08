package sdk

import (
	"crypto/md5"
	"fmt"
	"io"
	"time"

	"github.com/vapor-ware/synse-server-grpc/go"
)

// Reading describes a single device reading with a timestamp.
type Reading struct {
	Timestamp string
	Type      string
	Value     string
}

// ReadResource is used to associate a set of Readings with a known device,
// which is specified by its uid string.
//
// Since a single device can provide more than one reading (e.g. a humidity
// device could provide humidity and temperature data, or an LED could provide
// on/off state, color, etc.) a ReadResource will allow multiple readings to
// be associated with the device. Note that a ReadResource corresponds to a
// single pass of the read loop.
type ReadResource struct {
	Device  string
	Board   string
	Rack    string
	Reading []*Reading
}

// IDString returns a compound string that can identify the resource by its
// rack, board, and device. This ID should be globally unique. It simply follows
// the pattern {rack}-{board}-{device}.
func (r *ReadResource) IDString() string {
	return makeIDString(r.Rack, r.Board, r.Device)
}

// WriteResource describes a single write transaction.
type WriteResource struct {
	transaction *TransactionState
	device      string
	board       string
	rack        string
	data        *synse.WriteData
}

// IDString returns a compound string that can identify the resource by its
// rack, board, and device. This ID should be globally unique. It simply follows
// the pattern {rack}-{board}-{device}.
func (r *WriteResource) IDString() string {
	return makeIDString(r.rack, r.board, r.device)
}

// WriteData is an SDK alias for the Synse gRPC WriteData. This is done to
// make writing new plugins easier.
type WriteData synse.WriteData

// toGRPC converts the SDK WriteData to the Synse gRPC WriteData.
func (w *WriteData) toGRPC() *synse.WriteData {
	return &synse.WriteData{
		Raw:    w.Raw,
		Action: w.Action,
	}
}

// writeDataFromGRPC takes the Synse gRPC WriteData and converts it to the
// SDK WriteData.
func writeDataFromGRPC(data *synse.WriteData) *WriteData {
	return &WriteData{
		Raw:    data.Raw,
		Action: data.Action,
	}
}

// newUID creates a new unique identifier for a device. The device id is
// deterministic because it is created as a hash of various components that
// make up the device's configuration. By definition, each device will have
// a (slightly) different configuration (otherwise they would just be the same
// devices).
//
// These device IDs are not guaranteed to be globally unique, but they should
// be unique to the board they reside on.
func newUID(protocol, deviceType, model, protoComp string) string {
	h := md5.New()
	io.WriteString(h, protocol)
	io.WriteString(h, deviceType)
	io.WriteString(h, model)
	io.WriteString(h, protoComp)

	return fmt.Sprintf("%x", h.Sum(nil))
}

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

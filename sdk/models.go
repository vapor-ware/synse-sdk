package sdk


import (
	synse "github.com/vapor-ware/synse-server-grpc/go"

	"crypto/md5"
	"io"
	"fmt"
	"time"
)


// SDK model for a reading off of a device. The timestamp corresponds to the
// time which the reading was taken. The value represents the reading itself.
type Reading struct {
	Timestamp string
	Value     string
}

// A ReadResource is an SDK-internal model which is used to associate a set
// of Readings with a known device. The device is specified by its UID string.
// Since a single device can provide more than one reading (e.g. a humidity
// device could provide humidity and temperature data, or an LED could provide
// on/off state, color, etc.) a ReadResource will allow multiple readings to
// be associated with the device. Note that a ReadResource corresponds to a
// single pass of the read loop.
type ReadResource struct {
	Device  string
	Reading []Reading
}

//
type WriteResource struct {
	transaction *TransactionState
	device      string
	data        []string
}




// TODO - figure out how to generate a deterministic device id here.
//     should the device id be globally unique? (e.g. include the rack
//     and board) or unique within the scope of boards?
//     perhaps within the SDK it should be globally unique (lookups easier)
//     OR we could have UID (for device within board scope) and GID (for
//     device within global scope) which would include the rack/board


// Deterministic device IDs allow the specification of devices to become the
// identity of devices. That is to say, if you have two devices of type T on
// board B of rack R, they would be differentiated by their configuration. if
// device-1 operated on channel C1 and device-2 operated on channel C2, then
// compounding all of this information should yield a unique identifier for
// the device which does not need to be persisted. Because the identifying bits
// come from configuration, as long as the same configuration is provided, the
// physical devices should always map to the same virtual ids.
//
// devices themselves may not have a globally unique id, but all devices should
// be unique on a given board.
func NewUID(protocol, deviceType, model, protoComp string) string {
	h := md5.New()
	io.WriteString(h, protocol)
	io.WriteString(h, deviceType)
	io.WriteString(h, model)
	io.WriteString(h, protoComp)

	return fmt.Sprintf("%x", h.Sum(nil))
}




//
type Device struct {
	Prototype PrototypeConfig
	Instance DeviceConfig

	Handler DeviceHandler
}

//
func (d *Device) Type() string {
	return d.Prototype.Type
}

//
func (d *Device) Model() string {
	return d.Prototype.Model
}

//
func (d *Device) Manufacturer() string {
	return d.Prototype.Manufacturer
}

//
func (d *Device) Protocol() string {
	return d.Prototype.Protocol
}

//
func (d *Device) Uid() string {
	protocolComp := d.Handler.GetProtocolIdentifiers(d.Data())
	return NewUID(d.Protocol(), d.Type(), d.Model(), protocolComp)
}

//
func (d *Device) Output() []DeviceOutput {
	return d.Prototype.Output
}

//
func (d *Device) Location() DeviceLocation {
	return d.Instance.Location
}

//
func (d *Device) Data() map[string]string {
	return d.Instance.Data
}

//
func (d *Device) ToMetainfoResponse() *synse.MetainfoResponse {

	location := d.Location()

	var output []*synse.MetaOutput
	for _, out := range d.Output() {
		mo := out.ToMetaOutput()
		output = append(output, mo)
	}

	return &synse.MetainfoResponse{
		Timestamp: time.Now().String(),
		Uid: d.Uid(),
		Type: d.Type(),
		Model: d.Model(),
		Manufacturer: d.Manufacturer(),
		Protocol: d.Protocol(),
		Info: d.Data()["info"],
		Comment: d.Data()["comment"],
		Location: location.ToMetalLocation(),
		Output: output,
	}
}
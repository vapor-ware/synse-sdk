package sdk

import (
	"fmt"

	"github.com/vapor-ware/synse-server-grpc/go"
)

// Reading describes a single device reading with a timestamp. The timestamp
// should be formatted with the RFC3339Nano layout.
type Reading struct {
	Timestamp string
	Type      string
	Info      string
	Unit      Unit
	Value     interface{}
}

// NewReading creates a new instance of a Reading. This is the recommended method
// for creating new readings. It uses the current time (time.Now) to fill in the
// Timestamp field, formatted with the RFC3339Nano layout.
func NewReading(output *Output, value interface{}) *Reading {
	return &Reading{
		Timestamp: GetCurrentTime(),
		Type:      output.Type(),
		Info:      output.Info,
		Unit:      output.Unit,
		Value:     value,
	}
}

// encode translates the SDK Reading type to the corresponding gRPC Reading type.
func (reading *Reading) encode() *synse.Reading {
	r := synse.Reading{
		Timestamp: reading.Timestamp,
		Type:      reading.Type,
		Info:      reading.Info,
		Unit:      reading.Unit.encode(),
	}

	switch t := reading.Value.(type) {
	case string:
		r.Value = &synse.Reading_StringValue{t}
	case bool:
		r.Value = &synse.Reading_BoolValue{t}
	case float64:
		r.Value = &synse.Reading_Float64Value{t}
	case float32:
		r.Value = &synse.Reading_Float32Value{t}
	case int64:
		r.Value = &synse.Reading_Int64Value{t}
	case int32:
		r.Value = &synse.Reading_Int32Value{t}
	case int16:
		r.Value = &synse.Reading_Int32Value{int32(t)}
	case int8:
		r.Value = &synse.Reading_Int32Value{int32(t)}
	case int:
		r.Value = &synse.Reading_Int64Value{int64(t)}
	case []byte:
		r.Value = &synse.Reading_BytesValue{t}
	case uint64:
		r.Value = &synse.Reading_Uint64Value{t}
	case uint32:
		r.Value = &synse.Reading_Uint32Value{t}
	case uint16:
		r.Value = &synse.Reading_Uint32Value{uint32(t)}
	case uint8:
		r.Value = &synse.Reading_Uint32Value{uint32(t)}
	case uint:
		r.Value = &synse.Reading_Uint64Value{uint64(t)}
	case nil:
		r.Value = nil
	default:
		// If the reading type isn't one of the above, panic. The plugin should
		// terminate. This is indicative of the plugin doing something wrong.
		panic(fmt.Sprintf("unsupported reading value type: %s", t))
	}

	return &r
}

// ReadContext provides the context for a device reading. This context
// identifies the device being read and associates it with a set of readings
// at a given time.
//
// A single device can provide more than one reading (e.g. a humidity sensor
// could provide both a humidity and temperature reading). To accommodate, the
// ReadContext allows for multiple readings to be associated with the device.
// Note that the collection of readings in a single ReadContext would correspond
// to a single Read request.
type ReadContext struct {
	Device  string
	Board   string
	Rack    string
	Reading []*Reading
}

// NewReadContext creates a new instance of a ReadContext from the given
// device and corresponding readings.
func NewReadContext(device *Device, readings []*Reading) (*ReadContext, error) {
	return &ReadContext{
		Device:  device.ID(),
		Board:   device.Location.Board,
		Rack:    device.Location.Rack,
		Reading: readings,
	}, nil
}

// ID returns a compound string that can identify the resource by its
// rack, board, and device. This ID should be globally unique. It simply follows
// the pattern {rack}-{board}-{device}.
func (ctx *ReadContext) ID() string {
	return makeIDString(ctx.Rack, ctx.Board, ctx.Device)
}

// WriteContext describes a single write transaction.
type WriteContext struct {
	transaction *transaction
	device      string
	board       string
	rack        string
	data        *synse.WriteData
}

// ID returns a compound string that can identify the resource by its
// rack, board, and device. This ID should be globally unique. It simply follows
// the pattern {rack}-{board}-{device}.
func (ctx *WriteContext) ID() string {
	return makeIDString(ctx.rack, ctx.board, ctx.device)
}

// WriteData is an SDK alias for the Synse gRPC WriteData. This is done to
// make writing new plugins easier.
type WriteData synse.WriteData

// encode translates the WriteData to a corresponding gRPC WriteData.
func (w *WriteData) encode() *synse.WriteData {
	return &synse.WriteData{
		Data:   w.Data,
		Action: w.Action,
	}
}

// decodeWriteData decodes the gRPC WriteData to the SDK WriteData.
func decodeWriteData(data *synse.WriteData) *WriteData {
	return &WriteData{
		Data:   data.Data,
		Action: data.Action,
	}
}

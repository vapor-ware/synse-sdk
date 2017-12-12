package sdk

import (
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
type ReadContext struct {
	Device  string
	Board   string
	Rack    string
	Reading []*Reading
}

// ID returns a compound string that can identify the resource by its
// rack, board, and device. This ID should be globally unique. It simply follows
// the pattern {rack}-{board}-{device}.
func (ctx *ReadContext) ID() string {
	return makeIDString(ctx.Rack, ctx.Board, ctx.Device)
}

// WriteResource describes a single write transaction.
type WriteContext struct {
	transaction *Transaction
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

// FIXME -- should rename to 'encode'
// toGRPC converts the SDK WriteData to the Synse gRPC WriteData.
func (w *WriteData) toGRPC() *synse.WriteData {
	return &synse.WriteData{
		Raw:    w.Raw,
		Action: w.Action,
	}
}

// FIXME -- rename to 'decode'
// writeDataFromGRPC takes the Synse gRPC WriteData and converts it to the
// SDK WriteData.
func writeDataFromGRPC(data *synse.WriteData) *WriteData {
	return &WriteData{
		Raw:    data.Raw,
		Action: data.Action,
	}
}

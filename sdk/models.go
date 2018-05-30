package sdk

import (
	"github.com/vapor-ware/synse-server-grpc/go"
)

// Reading describes a single device reading with a timestamp. The timestamp
// should be formatted with the RFC3339Nano layout.
type Reading struct {
	Timestamp string
	Type      string
	Value     string
}

// NewReading creates a new instance of a Reading. It uses the current time
// (time.Now) to fill in the Timestamp field, formatted with the RFC3339Nano
// layout. This is the preferred method for creating new Reading instances.
func NewReading(readingType, readingValue string) *Reading {
	return &Reading{
		Timestamp: GetCurrentTime(),
		Type:      readingType,
		Value:     readingValue,
	}
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
	rack, err := device.Location.Rack.Get()
	if err != nil {
		return nil, err
	}
	board, err := device.Location.Board.Get()
	if err != nil {
		return nil, err
	}

	return &ReadContext{
		Device:  device.ID(),
		Board:   board,
		Rack:    rack,
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

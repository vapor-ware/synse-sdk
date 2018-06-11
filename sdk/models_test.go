package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-server-grpc/go"
)

// TestReadContext_ID tests successfully generating the ID of the ReadContext.
func TestReadContext_ID(t *testing.T) {
	ctx := ReadContext{
		Device: "123",
		Board:  "456",
		Rack:   "789",
	}

	assert.Equal(t, "789-456-123", ctx.ID())
}

// TestWriteContext_ID tests successfully generating the ID of the WriteContext.
func TestWriteContext_ID(t *testing.T) {
	ctx := WriteContext{
		device: "123",
		board:  "456",
		rack:   "789",
	}

	assert.Equal(t, "789-456-123", ctx.ID())
}

// TestWriteData_encode tests encoding the SDK WriteData into the Synse gRPC
// WriteData model.
func TestWriteData_encode(t *testing.T) {
	expected := &synse.WriteData{
		Data:   []byte{0, 1, 2},
		Action: "test",
	}

	wd := WriteData{
		Data:   []byte{0, 1, 2},
		Action: "test",
	}

	actual := wd.encode()

	assert.Equal(t, expected.Action, actual.Action)
	assert.Equal(t, len(expected.Data), len(actual.Data))
	for i := 0; i < len(expected.Data); i++ {
		assert.Equal(t, expected.Data[i], actual.Data[i])
	}
}

// TestDecodeWriteData tests decoding a Synse gRPC WriteData into the SDK
// WriteData model.
func TestDecodeWriteData(t *testing.T) {
	expected := &WriteData{
		Data:   []byte{3, 2, 1},
		Action: "test",
	}

	wd := &synse.WriteData{
		Data:   []byte{3, 2, 1},
		Action: "test",
	}

	actual := decodeWriteData(wd)

	assert.Equal(t, expected.Action, actual.Action)
	assert.Equal(t, len(expected.Data), len(actual.Data))
	for i := 0; i < len(expected.Data); i++ {
		assert.Equal(t, expected.Data[i], actual.Data[i])
	}
}

// TestNewReading tests creating a new Reading.
func TestNewReading(t *testing.T) {
	output := &Output{
		OutputType: OutputType{
			Name: "test",
			Unit: Unit{
				Name:   "abc",
				Symbol: "A",
			},
		},
		Info: "foobar",
	}

	reading := NewReading(output, 42)
	assert.Equal(t, "test", reading.Type)
	assert.Equal(t, "foobar", reading.Info)
	assert.Equal(t, "abc", reading.Unit.Name)
	assert.Equal(t, "A", reading.Unit.Symbol)
	assert.Equal(t, 42, reading.Value)
}

// TestNewReadContext tests creating a new ReadContext.
func TestNewReadContext(t *testing.T) {
	device := &Device{
		Location: &Location{
			Rack:  "rack",
			Board: "board",
		},
	}
	ctx := NewReadContext(device, []*Reading{{Type: "test", Value: 2}})

	assert.Equal(t, "d41d8cd98f00b204e9800998ecf8427e", ctx.Device)
	assert.Equal(t, "board", ctx.Board)
	assert.Equal(t, "rack", ctx.Rack)
	assert.Equal(t, 1, len(ctx.Reading))
}

// TestReading_encode_string tests encoding a Reading when the value is a string.
func TestReading_encode_string(t *testing.T) {
	reading := Reading{
		Type:  "test",
		Value: "foo",
	}
	out := reading.encode()
	assert.Equal(t, "test", out.Type)
	assert.Equal(t, "foo", out.GetStringValue())
}

// TestReading_encode_bool tests encoding a Reading when the value is a bool.
func TestReading_encode_bool(t *testing.T) {
	reading := Reading{
		Type:  "test",
		Value: true,
	}
	out := reading.encode()
	assert.Equal(t, "test", out.Type)
	assert.Equal(t, true, out.GetBoolValue())
}

// TestReading_encode_float64 tests encoding a Reading when the value is a float64.
func TestReading_encode_float64(t *testing.T) {
	reading := Reading{
		Type:  "test",
		Value: float64(7),
	}
	out := reading.encode()
	assert.Equal(t, "test", out.Type)
	assert.Equal(t, float64(7), out.GetFloat64Value())
}

// TestReading_encode_float32 tests encoding a Reading when the value is a float32.
func TestReading_encode_float32(t *testing.T) {
	reading := Reading{
		Type:  "test",
		Value: float32(7),
	}
	out := reading.encode()
	assert.Equal(t, "test", out.Type)
	assert.Equal(t, float32(7), out.GetFloat32Value())
}

// TestReading_encode_int64 tests encoding a Reading when the value is an int64.
func TestReading_encode_int64(t *testing.T) {
	reading := Reading{
		Type:  "test",
		Value: int64(7),
	}
	out := reading.encode()
	assert.Equal(t, "test", out.Type)
	assert.Equal(t, int64(7), out.GetInt64Value())
}

// TestReading_encode_int32 tests encoding a Reading when the value is an int32.
func TestReading_encode_int32(t *testing.T) {
	reading := Reading{
		Type:  "test",
		Value: int32(7),
	}
	out := reading.encode()
	assert.Equal(t, "test", out.Type)
	assert.Equal(t, int32(7), out.GetInt32Value())
}

// TestReading_encode_int16 tests encoding a Reading when the value is an int16.
func TestReading_encode_int16(t *testing.T) {
	reading := Reading{
		Type:  "test",
		Value: int16(7),
	}
	out := reading.encode()
	assert.Equal(t, "test", out.Type)
	assert.Equal(t, int32(7), out.GetInt32Value())
}

// TestReading_encode_int8 tests encoding a Reading when the value is an int8.
func TestReading_encode_int8(t *testing.T) {
	reading := Reading{
		Type:  "test",
		Value: int8(7),
	}
	out := reading.encode()
	assert.Equal(t, "test", out.Type)
	assert.Equal(t, int32(7), out.GetInt32Value())
}

// TestReading_encode_int tests encoding a Reading when the value is an int.
func TestReading_encode_int(t *testing.T) {
	reading := Reading{
		Type:  "test",
		Value: int(7),
	}
	out := reading.encode()
	assert.Equal(t, "test", out.Type)
	assert.Equal(t, int64(7), out.GetInt64Value())
}

// TestReading_encode_uint64 tests encoding a Reading when the value is a uint64.
func TestReading_encode_uint64(t *testing.T) {
	reading := Reading{
		Type:  "test",
		Value: uint64(7),
	}
	out := reading.encode()
	assert.Equal(t, "test", out.Type)
	assert.Equal(t, uint64(7), out.GetUint64Value())
}

// TestReading_encode_uint32 tests encoding a Reading when the value is a uint32.
func TestReading_encode_uint32(t *testing.T) {
	reading := Reading{
		Type:  "test",
		Value: uint32(7),
	}
	out := reading.encode()
	assert.Equal(t, "test", out.Type)
	assert.Equal(t, uint32(7), out.GetUint32Value())
}

// TestReading_encode_uint16 tests encoding a Reading when the value is a uint16.
func TestReading_encode_uint16(t *testing.T) {
	reading := Reading{
		Type:  "test",
		Value: uint16(7),
	}
	out := reading.encode()
	assert.Equal(t, "test", out.Type)
	assert.Equal(t, uint32(7), out.GetUint32Value())
}

// TestReading_encode_uint8 tests encoding a Reading when the value is a uint8.
func TestReading_encode_uint8(t *testing.T) {
	reading := Reading{
		Type:  "test",
		Value: uint8(7),
	}
	out := reading.encode()
	assert.Equal(t, "test", out.Type)
	assert.Equal(t, uint32(7), out.GetUint32Value())
}

// TestReading_encode_uint tests encoding a Reading when the value is a uint.
func TestReading_encode_uint(t *testing.T) {
	reading := Reading{
		Type:  "test",
		Value: uint(7),
	}
	out := reading.encode()
	assert.Equal(t, "test", out.Type)
	assert.Equal(t, uint64(7), out.GetUint64Value())
}

// TestReading_encode_bytes tests encoding a Reading when the value is a slice of bytes.
func TestReading_encode_bytes(t *testing.T) {
	reading := Reading{
		Type:  "test",
		Value: []byte("test"),
	}
	out := reading.encode()
	assert.Equal(t, "test", out.Type)
	assert.Equal(t, []byte("test"), out.GetBytesValue())
}

// TestReading_encode_nil tests encoding a Reading when the value is nil.
func TestReading_encode_nil(t *testing.T) {
	reading := Reading{
		Type:  "test",
		Value: nil,
	}
	out := reading.encode()
	assert.Equal(t, "test", out.Type)
	assert.Equal(t, nil, out.GetValue())
}

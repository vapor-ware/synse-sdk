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

// TestNewReading tests creating a new Reading from the NewReading constructor.
func TestNewReading(t *testing.T) {
	r := NewReading("test", "value")

	assert.IsType(t, Reading{}, *r)
	assert.NotEqual(t, "", r.Timestamp)
	assert.Equal(t, "test", r.Type)
	assert.Equal(t, "value", r.Value)
}

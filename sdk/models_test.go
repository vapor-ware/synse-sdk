package sdk

import (
	"testing"

	"github.com/vapor-ware/synse-server-grpc/go"
)

func TestReadContext_ID(t *testing.T) {
	ctx := ReadContext{
		Device: "123",
		Board:  "456",
		Rack:   "789",
	}

	id := ctx.ID()
	if id != "789-456-123" {
		t.Errorf("ReadContext.ID() -> unexpected result: %s", id)
	}
}

func TestWriteContext_ID(t *testing.T) {
	ctx := WriteContext{
		device: "123",
		board:  "456",
		rack:   "789",
	}

	id := ctx.ID()
	if id != "789-456-123" {
		t.Errorf("WriteContext.ID() -> unexpected result: %s", id)
	}
}

func TestWriteData_encode(t *testing.T) {
	expected := &synse.WriteData{
		Raw:    [][]byte{{0, 1, 2}},
		Action: "test",
	}

	wd := WriteData{
		Raw:    [][]byte{{0, 1, 2}},
		Action: "test",
	}

	actual := wd.encode()

	if len(actual.Raw) != len(expected.Raw) {
		t.Error("WriteData Raw length mismatch")
	}

	for i := 0; i < len(actual.Raw); i++ {
		for j := 0; j < len(actual.Raw[i]); j++ {
			if actual.Raw[i][j] != expected.Raw[i][j] {
				t.Error("WriteData Raw value mismatch")
			}
		}
	}

	if actual.Action != expected.Action {
		t.Error("WriteData Action mismatch.")
	}
}

func TestDecodeWriteData(t *testing.T) {
	expected := &WriteData{
		Raw:    [][]byte{{3, 2, 1}},
		Action: "test",
	}

	wd := &synse.WriteData{
		Raw:    [][]byte{{3, 2, 1}},
		Action: "test",
	}

	actual := decodeWriteData(wd)

	if len(actual.Raw) != len(expected.Raw) {
		t.Error("WriteData Raw length mismatch")
	}

	for i := 0; i < len(actual.Raw); i++ {
		for j := 0; j < len(actual.Raw[i]); j++ {
			if actual.Raw[i][j] != expected.Raw[i][j] {
				t.Error("WriteData Raw value mismatch")
			}
		}
	}

	if actual.Action != expected.Action {
		t.Error("WriteData Action mismatch.")
	}
}

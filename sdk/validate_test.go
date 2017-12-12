package sdk

import (
	"testing"

	"github.com/vapor-ware/synse-server-grpc/go"
)

func TestValidateReadRequest(t *testing.T) {
	// everything is there
	request := &synse.ReadRequest{Device: "device", Board: "board", Rack: "rack"}
	err := validateReadRequest(request)
	if err != nil {
		t.Errorf("validateReadRequest(%q) -> unexpected error: %q", request, err)
	}

	// missing device
	request = &synse.ReadRequest{Board: "board", Rack: "rack"}
	err = validateReadRequest(request)
	if err == nil {
		t.Error("validateReadRequest() -> expected error but got nil")
	}

	// missing board
	request = &synse.ReadRequest{Device: "device", Rack: "rack"}
	err = validateReadRequest(request)
	if err == nil {
		t.Error("validateReadRequest() -> expected error but got nil")
	}

	// missing rack
	request = &synse.ReadRequest{Device: "device", Board: "board"}
	err = validateReadRequest(request)
	if err == nil {
		t.Error("validateReadRequest() -> expected error but got nil")
	}
}

func TestValidateWriteRequest(t *testing.T) {
	// everything is there
	request := &synse.WriteRequest{Device: "device", Board: "board", Rack: "rack"}
	err := validateWriteRequest(request)
	if err != nil {
		t.Errorf("validateWriteRequest(%q) -> unexpected error: %q", request, err)
	}

	// missing device
	request = &synse.WriteRequest{Board: "board", Rack: "rack"}
	err = validateWriteRequest(request)
	if err == nil {
		t.Error("validateWriteRequest() -> expected error but got nil")
	}

	// missing board
	request = &synse.WriteRequest{Device: "device", Rack: "rack"}
	err = validateWriteRequest(request)
	if err == nil {
		t.Error("validateWriteRequest() -> expected error but got nil")
	}

	// missing rack
	request = &synse.WriteRequest{Device: "device", Board: "board"}
	err = validateWriteRequest(request)
	if err == nil {
		t.Error("validateWriteRequest() -> expected error but got nil")
	}
}
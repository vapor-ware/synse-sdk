package sdk

import (
	"testing"

	"github.com/vapor-ware/synse-server-grpc/go"
)

// ===== Test Data =====

// FIXME - this is used elsewhere for testing. what is the best way of sharing
// test data? not sure what a "good" way of doing that in Go is. for now leaving
// it here, but maybe we need a separate "test_utils" file or something?

type testDeviceHandler struct{}

func (h *testDeviceHandler) GetProtocolIdentifiers(in map[string]string) string {
	return ""
}

func (h *testDeviceHandler) EnumerateDevices(map[string]interface{}) ([]*DeviceConfig, error) {
	return nil, nil
}


type testPluginHandler struct{}

func (h *testPluginHandler) Read(dev *Device) (*ReadContext, error) {
	return nil, nil
}

func (h *testPluginHandler) Write(dev *Device, data *WriteData) error {
	return nil
}

// ===== Test Cases =====

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

func TestValidateHandlers(t *testing.T) {
	// handlers are ok
	h := &Handlers{
		Plugin: &testPluginHandler{},
		Device: &testDeviceHandler{},
	}
	err := validateHandlers(h)
	if err != nil {
		t.Errorf("validateHandlers(%v) -> unexpected error: %q", h, err)
	}

	// plugin handler nil
	h = &Handlers{
		Device: &testDeviceHandler{},
	}
	err = validateHandlers(h)
	if err == nil {
		t.Errorf("validateHandlers(%v) -> expected error but got nil", h)
	}

	// device handler nil
	h = &Handlers{
		Plugin: &testPluginHandler{},
	}
	err = validateHandlers(h)
	if err == nil {
		t.Errorf("validateHandlers(%v) -> expected error but got nil", h)
	}

	// both handlers nil
	h = &Handlers{}
	err = validateHandlers(h)
	if err == nil {
		t.Errorf("validateHandlers(%v) -> expected error but got nil", h)
	}
}
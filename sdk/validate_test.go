package sdk

import (
	"testing"

	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-server-grpc/go"
)

// ===== Test Data =====

func testDeviceIdentifier(in map[string]string) string                            { return "" }
func testDeviceEnumerator(map[string]interface{}) ([]*config.DeviceConfig, error) { return nil, nil }

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
	// handlers is ok
	h := &Handlers{
		DeviceIdentifier: testDeviceIdentifier,
		DeviceEnumerator: testDeviceEnumerator,
	}
	err := validateHandlers(h)
	if err != nil {
		t.Errorf("validateHandlers(%v) -> unexpected error: %q", h, err)
	}

	// handlers are ok
	h = &Handlers{
		DeviceIdentifier: testDeviceIdentifier,
	}
	err = validateHandlers(h)
	if err != nil {
		t.Errorf("validateHandlers(%v) -> unexpected error: %q", h, err)
	}

	// device identifier nil
	h = &Handlers{
		DeviceEnumerator: testDeviceEnumerator,
	}
	err = validateHandlers(h)
	if err == nil {
		t.Errorf("validateHandlers(%v) -> expected error but got nil", h)
	}

	// all device handlers nil
	h = &Handlers{}
	err = validateHandlers(h)
	if err == nil {
		t.Errorf("validateHandlers(%v) -> expected error but got nil", h)
	}
}

func TestValidateForRead_1(t *testing.T) {
	// the given ID is not in the device map
	deviceMap = make(map[string]*Device)

	err := validateForRead("foo")
	if err == nil {
		t.Errorf("validateForRead -> expected error but got nil")
	}
}

func TestValidateForRead_2(t *testing.T) {
	// the given device is not readable
	deviceMap = make(map[string]*Device)
	deviceMap["abc"] = &Device{
		Handler: &DeviceHandler{},
	}

	err := validateForRead("abc")
	if err == nil {
		t.Errorf("validateForRead -> expected error but got nil")
	}
}

func TestValidateForRead_3(t *testing.T) {
	// the given device is readable
	deviceMap = make(map[string]*Device)
	deviceMap["abc"] = &Device{
		Handler: &DeviceHandler{
			Read: func(d *Device) ([]*Reading, error) { return nil, nil },
		},
	}

	err := validateForRead("abc")
	if err != nil {
		t.Errorf("validateForRead -> unexpected error: %v", err)
	}
}

func TestValidateForWrite_1(t *testing.T) {
	// the given ID is not in the device map
	deviceMap = make(map[string]*Device)

	err := validateForWrite("foo")
	if err == nil {
		t.Errorf("validateForWrite -> expected error but got nil")
	}
}

func TestValidateForWrite_2(t *testing.T) {
	// the given device is not writable
	deviceMap = make(map[string]*Device)
	deviceMap["abc"] = &Device{
		Handler: &DeviceHandler{},
	}

	err := validateForWrite("abc")
	if err == nil {
		t.Errorf("validateForWrite -> expected error but got nil")
	}
}

func TestValidateForWrite_3(t *testing.T) {
	// the given device is writable
	deviceMap = make(map[string]*Device)
	deviceMap["abc"] = &Device{
		Handler: &DeviceHandler{
			Write: func(d *Device, data *WriteData) error { return nil },
		},
	}

	err := validateForWrite("abc")
	if err != nil {
		t.Errorf("validateForWrite -> unexpected error: %v", err)
	}
}

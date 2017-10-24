package sdk

import (
	"testing"
)


var testInst1 = DeviceConfig{
	Version: "1.0",
	Type: "test-device",
	Model: "td-1",
	Location: DeviceLocation{
		Rack: "rack-1",
		Board: "board-1",
	},
}

var testInst2 = DeviceConfig{
	Version: "1.0",
	Type: "test-device",
	Model: "td-1",
	Location: DeviceLocation{
		Rack: "rack-1",
		Board: "board-2",
	},
}

var testProto1 = PrototypeConfig{
	Version: "1.0",
	Type: "test-device",
	Model: "td-1",
	Manufacturer: "vaporio",
	Protocol: "test",
}

var testProto2 = PrototypeConfig{
	Version: "1.0",
	Type: "test-device",
	Model: "td-3",
	Manufacturer: "vaporio",
	Protocol: "test",
}


func TestMakeDevices(t *testing.T) {
	inst := []DeviceConfig{testInst1, testInst2}
	proto := []PrototypeConfig{testProto1}

	devices := makeDevices(inst, proto, &TestHandler{})

	if len(devices) != 2 {
		t.Error("Expected two instances to match the prototype.")
	}
}

func TestMakeDevices2(t *testing.T) {
	inst := []DeviceConfig{testInst1, testInst2}
	proto := []PrototypeConfig{testProto2}

	devices := makeDevices(inst, proto, &TestHandler{})

	if len(devices) != 0 {
		t.Error("Expected no instances to match the prototype.")
	}
}

func TestMakeDevices3(t *testing.T) {
	inst := []DeviceConfig{testInst1}
	proto := []PrototypeConfig{testProto1, testProto2}

	devices := makeDevices(inst, proto, &TestHandler{})

	if len(devices) != 1 {
		t.Error("Expected one instance to match the prototypes.")
	}
}

func TestMakeDevices4(t *testing.T) {
	inst := []DeviceConfig{testInst1, testInst2}
	proto := []PrototypeConfig{}

	devices := makeDevices(inst, proto, &TestHandler{})

	if len(devices) != 0 {
		t.Error("Expected no matches - no prototypes defined.")
	}
}

func TestMakeDevices5(t *testing.T) {
	inst := []DeviceConfig{}
	proto := []PrototypeConfig{testProto1, testProto2}

	devices := makeDevices(inst, proto, &TestHandler{})

	if len(devices) != 0 {
		t.Error("Expected no matches - no instances defined.")
	}
}

package sdk

import (
	"testing"
	"os"
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
	inst := []*DeviceConfig{&testInst1, &testInst2}
	proto := []*PrototypeConfig{&testProto1}

	devices := makeDevices(inst, proto, &TestHandler{})

	if len(devices) != 2 {
		t.Error("Expected two instances to match the prototype.")
	}
}

func TestMakeDevices2(t *testing.T) {
	inst := []*DeviceConfig{&testInst1, &testInst2}
	proto := []*PrototypeConfig{&testProto2}

	devices := makeDevices(inst, proto, &TestHandler{})

	if len(devices) != 0 {
		t.Error("Expected no instances to match the prototype.")
	}
}

func TestMakeDevices3(t *testing.T) {
	inst := []*DeviceConfig{&testInst1}
	proto := []*PrototypeConfig{&testProto1, &testProto2}

	devices := makeDevices(inst, proto, &TestHandler{})

	if len(devices) != 1 {
		t.Error("Expected one instance to match the prototypes.")
	}
}

func TestMakeDevices4(t *testing.T) {
	inst := []*DeviceConfig{&testInst1, &testInst2}
	proto := []*PrototypeConfig{}

	devices := makeDevices(inst, proto, &TestHandler{})

	if len(devices) != 0 {
		t.Error("Expected no matches - no prototypes defined.")
	}
}

func TestMakeDevices5(t *testing.T) {
	inst := []*DeviceConfig{}
	proto := []*PrototypeConfig{&testProto1, &testProto2}

	devices := makeDevices(inst, proto, &TestHandler{})

	if len(devices) != 0 {
		t.Error("Expected no matches - no instances defined.")
	}
}


// setup the socket when the socket path does not exist.
func TestSetupSocket(t *testing.T) {
	_ = os.RemoveAll(sockPath)

	_, err := os.Stat(sockPath)
	if !os.IsNotExist(err) {
		t.Errorf("Expected path to not exist, got error: %v", err)
	}

	sock, err := setupSocket("test")

	if err != nil {
		t.Error(err)
	}

	if sock != "/synse/procs/test.sock" {
		t.Errorf("Unexpected socket path returned: %v", sock)
	}

	_, err = os.Stat(sockPath)
	if err != nil {
		t.Errorf("Error when checking socket path: %v", err)
	}
}

// setup the socket when the path and socket already exist.
func TestSetupSocket2(t *testing.T) {
	_ = os.MkdirAll("/synse/procs", os.ModePerm)

	filename := "/synse/procs/test.sock"
	_, err := os.Create(filename)
	if err != nil {
		t.Error(err)
	}

	_, err = os.Stat(filename)
	if err != nil {
		t.Errorf("Expected file to exist, but does not.")
	}

	sock, err := setupSocket("test")

	if err != nil {
		t.Error(err)
	}

	if sock != filename {
		t.Errorf("Unexpected socket path returned: %v", sock)
	}

	_, err = os.Stat(filename)
	if !os.IsNotExist(err) {
		t.Error("Socket should no longer exist, but still does.")
	}
}

func TestMakeIdString(t *testing.T) {
	matrix := map[string][]string{
		"rack-board-device": {"rack", "board", "device"},
		"123-456-789": {"123", "456", "789"},
		"abc-def-ghi": {"abc", "def", "ghi"},
		"1234567890abcdefghi-1-2": {"1234567890abcdefghi", "1", "2"},
		"------_____-+=+=&8^": {"-----", "_____", "+=+=&8^"},
	}

	for expected, test := range matrix {
		actual := makeIDString(test[0], test[1], test[2])
		if expected != actual {
			t.Errorf("Failed to make expected id string (%v): %v", expected, actual)
		}
	}
}
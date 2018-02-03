package sdk

import (
	"os"
	"testing"

	"github.com/vapor-ware/synse-sdk/sdk/config"
)

func TestMakeDevices(t *testing.T) {
	inst := []*config.DeviceConfig{&testDeviceConfig1, &testDeviceConfig2}
	proto := []*config.PrototypeConfig{&testPrototypeConfig1}

	devices, err := makeDevices(inst, proto, &testPlugin)
	if err != nil {
		t.Error(err)
	}
	if len(devices) != 2 {
		t.Error("expected 2 instances to match the prototype")
	}
}

func TestMakeDevices2(t *testing.T) {
	inst := []*config.DeviceConfig{&testDeviceConfig1, &testDeviceConfig2}
	proto := []*config.PrototypeConfig{&testPrototypeConfig2}

	devices, err := makeDevices(inst, proto, &testPlugin)
	if err != nil {
		t.Error(err)
	}
	if len(devices) != 0 {
		t.Error("expected 0 instances to match the prototype")
	}
}

func TestMakeDevices3(t *testing.T) {
	inst := []*config.DeviceConfig{&testDeviceConfig1}
	proto := []*config.PrototypeConfig{&testPrototypeConfig1, &testPrototypeConfig2}

	devices, err := makeDevices(inst, proto, &testPlugin)
	if err != nil {
		t.Error(err)
	}
	if len(devices) != 1 {
		t.Error("expected 1 instance to match the prototypes")
	}
}

func TestMakeDevices4(t *testing.T) {
	inst := []*config.DeviceConfig{&testDeviceConfig1, &testDeviceConfig2}
	var proto []*config.PrototypeConfig

	devices, err := makeDevices(inst, proto, &testPlugin)
	if err != nil {
		t.Error(err)
	}
	if len(devices) != 0 {
		t.Error("expected 0 matches (no prototypes defined)")
	}
}

func TestMakeDevices5(t *testing.T) {
	var inst []*config.DeviceConfig
	proto := []*config.PrototypeConfig{&testPrototypeConfig1, &testPrototypeConfig2}

	devices, err := makeDevices(inst, proto, &testPlugin)
	if err != nil {
		t.Error(err)
	}
	if len(devices) != 0 {
		t.Error("expected 0 matches (no instances defined)")
	}
}

// setup the socket when the socket path does not exist.
func TestSetupSocket(t *testing.T) {
	_ = os.RemoveAll(sockPath)

	_, err := os.Stat(sockPath)
	if !os.IsNotExist(err) {
		t.Errorf("expected path to not exist, got error: %v", err)
	}

	sock, err := setupSocket("test.sock")
	if err != nil {
		t.Error(err)
	}

	if sock != "/tmp/synse/procs/test.sock" {
		t.Errorf("unexpected socket path returned: %v", sock)
	}

	_, err = os.Stat(sockPath)
	if err != nil {
		t.Errorf("error when checking socket path: %v", err)
	}
}

// setup the socket when the path and socket already exist.
func TestSetupSocket2(t *testing.T) {
	_ = os.MkdirAll("/tmp/synse/procs", os.ModePerm)

	filename := "/tmp/synse/procs/test.sock"
	_, err := os.Create(filename)
	if err != nil {
		t.Error(err)
	}

	_, err = os.Stat(filename)
	if err != nil {
		t.Errorf("expected file to exist, but does not")
	}

	sock, err := setupSocket("test.sock")
	if err != nil {
		t.Error(err)
	}

	if sock != filename {
		t.Errorf("unexpected socket path returned: %v", sock)
	}

	_, err = os.Stat(filename)
	if !os.IsNotExist(err) {
		t.Error("socket should no longer exist, but still does")
	}
}

var makeIDStringTestTable = []struct {
	rack   string
	board  string
	device string
	out    string
}{
	{
		rack:   "rack",
		board:  "board",
		device: "device",
		out:    "rack-board-device",
	},
	{
		rack:   "123",
		board:  "456",
		device: "789",
		out:    "123-456-789",
	},
	{
		rack:   "abc",
		board:  "def",
		device: "ghi",
		out:    "abc-def-ghi",
	},
	{
		rack:   "1234567890abcdefghi",
		board:  "1",
		device: "2",
		out:    "1234567890abcdefghi-1-2",
	},
	{
		rack:   "-----",
		board:  "_____",
		device: "+=+=&8^",
		out:    "------_____-+=+=&8^",
	},
}

func TestMakeIDString(t *testing.T) {
	for _, tc := range makeIDStringTestTable {
		r := makeIDString(tc.rack, tc.board, tc.device)
		if r != tc.out {
			t.Errorf("makeIDString(%s, %s, %s) => %s, want %q", tc.rack, tc.board, tc.device, r, tc.out)
		}
	}
}

var newUIDTestTable = []struct {
	p   string
	d   string
	m   string
	c   string
	out string
}{
	{
		p:   "test-protocol",
		d:   "test-device",
		m:   "test-model",
		c:   "test-comp",
		out: "732bb43a825b8330e6d50a6722a8e1f0",
	},
	{
		p:   "i2c",
		d:   "thermistor",
		m:   "max116",
		c:   "1",
		out: "019de8ff9de6aba9ddb9ebb6d5f5b5e0",
	},
	{
		p:   "",
		d:   "",
		m:   "",
		c:   "",
		out: "d41d8cd98f00b204e9800998ecf8427e",
	},
	{
		p:   "?",
		d:   "!",
		m:   "%",
		c:   "$",
		out: "65722f8565fb36c7a6da67bae4ee1f2d",
	},
}

func TestNewUID(t *testing.T) {
	for _, tc := range newUIDTestTable {
		r := newUID(tc.p, tc.d, tc.m, tc.c)
		if r != tc.out {
			t.Errorf("newUID(%s, %s, %s, %s) => %s, want %s", tc.p, tc.d, tc.m, tc.c, r, tc.out)
		}
	}
}

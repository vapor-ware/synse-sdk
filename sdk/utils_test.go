package sdk

import (
	"os"
	"testing"
	"io/ioutil"
)


func TestDevicesFromConfig(t *testing.T) {
	devices, err := DevicesFromConfig("some/nonexistant/dir", &TestHandler{})

	if devices != nil {
		t.Error("Expecting error - devices should be nil.")
	}

	if _, ok := err.(*os.PathError); !ok {
		t.Error("Expected path error.")
	}
}

func TestDevicesFromConfig2(t *testing.T) {
	os.MkdirAll("tmp/proto", os.ModePerm)
	defer func () {
		os.RemoveAll("tmp")
	}()

	devices, err := DevicesFromConfig("tmp", &TestHandler{})

	if devices != nil {
		t.Error("Expecting error - devices should be nil.")
	}

	if _, ok := err.(*os.PathError); !ok {
		t.Error("Expected path error.")
	}
}

func TestDevicesFromConfig3(t *testing.T) {
	os.MkdirAll("tmp/proto", os.ModePerm)
	os.MkdirAll("tmp/device", os.ModePerm)
	defer func () {
		os.RemoveAll("tmp")
	}()

	devices, err := DevicesFromConfig("tmp", &TestHandler{})
	if err != nil {
		t.Error("Failed to create devices from empty config.")
	}
	if len(devices) != 0 {
		t.Error("Created devices from no configuration.")
	}
}

func TestDevicesFromConfig4(t *testing.T) {
	os.MkdirAll("tmp/proto", os.ModePerm)
	os.MkdirAll("tmp/device", os.ModePerm)
	defer func () {
		os.RemoveAll("tmp")
	}()

	protoCfg := `version: 1.0
type: emulated-temperature
model: emul8-temp
manufacturer: vaporio
protocol: emulator
output:
  - type: temperature
    unit:
      name: celsius
      symbol: C
    precision: 2
    range:
      min: 0
      max: 100`

	err := ioutil.WriteFile("tmp/proto/test.yaml", []byte(protoCfg), 0644)
	if err != nil {
		t.Error("Failed to write test data to file.")
	}

	deviceCfg := `version: 1.0
type: emulated-temperature
model: emul8-temp

locations:
  unknown:
    rack: unknown
    board: unknown

devices:
  - id: 1
    location: unknown
    comment: first emulated temperature device
    info: CEC temp 1
  - id: 2
    location: unknown
    comment: second emulated temperature device
    info: CEC temp 2
  - id: 3
    location: unknown
    comment: third emulated temperature device
    info: CEC temp 3`

	err = ioutil.WriteFile("tmp/device/test.yaml", []byte(deviceCfg), 0644)
	if err != nil {
		t.Error("Failed to write test data to file.")
	}

	devices, err := DevicesFromConfig("tmp", &TestHandler{})
	if err != nil {
		t.Error("Failed to create devices from config.")
	}
	if len(devices) != 3 {
		t.Error("Created incorrect number of devices.")
	}
}
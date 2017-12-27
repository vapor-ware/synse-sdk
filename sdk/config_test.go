package sdk

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"gopkg.in/yaml.v2"
)

// no prototype directory is present
func TestParsePrototypeConfig(t *testing.T) {
	devices, err := parsePrototypeConfig("some/nonexistant/dir")

	if devices != nil {
		t.Error("Expecting error - devices should be nil.")
	}

	if _, ok := err.(*os.PathError); !ok {
		t.Errorf("Expected os.PathError but was %v", reflect.TypeOf(err))
	}
}

// FIXME - test fails with 'mkdir tmp/proto: permission denied'
// bad permissions on the config proto directory
//func TestParsePrototypeConfig2(t *testing.T) {
//	err := os.MkdirAll("tmp/proto", 0555)
//	if err != nil {
//		t.Error(err)
//	}
//	defer func() {
//		err = os.RemoveAll("tmp")
//		if err != nil {
//			t.Error(err)
//		}
//	}()
//
//	devices, err := parsePrototypeConfig("tmp")
//
//	if devices != nil {
//		t.Error("Expecting error - devices should be nil.")
//	}
//
//	if _, ok := err.(*os.PathError); !ok {
//		t.Errorf("Expected os.PathError but was %v", reflect.TypeOf(err))
//	}
//}

// bad permissions on the config file in the proto directory
func TestParsePrototypeConfig3(t *testing.T) {
	err := os.MkdirAll("tmp/proto", os.ModePerm)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll("tmp")
		if err != nil {
			t.Error(err)
		}
	}()

	// this string represents invalid YAML - it should cause failure when
	// we try to unmarshall.
	cfgFile := `--~+::\n:-`

	err = ioutil.WriteFile("tmp/proto/test_config.yaml", []byte(cfgFile), 0333)
	if err != nil {
		t.Error("Failed to write test data to file.")
	}

	devices, err := parsePrototypeConfig("tmp")

	if devices != nil {
		t.Error("Expecting error - devices should be nil.")
	}

	if _, ok := err.(*os.PathError); !ok {
		t.Errorf("Expected os.PathError but was %v", reflect.TypeOf(err))
	}
}

// invalid YAML contents
func TestParsePrototypeConfig4(t *testing.T) {
	err := os.MkdirAll("tmp/proto", os.ModePerm)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll("tmp")
		if err != nil {
			t.Error(err)
		}
	}()

	// this string represents invalid YAML - it should cause failure when
	// we try to unmarshall.
	cfgFile := `--~+::\n:-`

	err = ioutil.WriteFile("tmp/proto/test_config.yaml", []byte(cfgFile), 0644)
	if err != nil {
		t.Error("Failed to write test data to file.")
	}

	devices, err := parsePrototypeConfig("tmp")

	if devices != nil {
		t.Error("Expecting error - devices should be nil.")
	}

	if _, ok := err.(*yaml.TypeError); !ok {
		t.Errorf("Expected yaml.TypeError but was %v", reflect.TypeOf(err))
	}
}

// the positive case - there should be no errors here
func TestParsePrototypeConfig5(t *testing.T) {
	err := os.MkdirAll("tmp/proto", os.ModePerm)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll("tmp")
		if err != nil {
			t.Error(err)
		}
	}()

	cfgFile := `version: 1.0
type: emulated-temperature
model: emul8-temp
manufacturer: vaporio
protocol: emulator
output:
  - type: temperature
    data_type: float
    unit:
      name: celsius
      symbol: C
    precision: 2
    range:
      min: 0
      max: 100`

	err = ioutil.WriteFile("tmp/proto/test_config.yaml", []byte(cfgFile), 0644)
	if err != nil {
		t.Error("Failed to write test data to file.")
	}

	devices, err := parsePrototypeConfig("tmp")

	if err != nil {
		t.Errorf("Not expecting error, but got %v", err)
	}

	if len(devices) != 1 {
		t.Errorf("Expecting only one device proto config, but got %v", len(devices))
	}
}

// no device config directory is present.
func TestParseDeviceConfig(t *testing.T) {
	devices, err := parseDeviceConfig("some/nonexistant/dir")

	if devices != nil {
		t.Error("Expecting error - devices should be nil.")
	}

	if _, ok := err.(*os.PathError); !ok {
		t.Errorf("Expected os.PathError but was %v", reflect.TypeOf(err))
	}
}

// FIXME - test fails with 'mkdir tmp/device: permission denied'
// bad permissions on the device config directory
//func TestParseDeviceConfig2(t *testing.T) {
//	err := os.MkdirAll("tmp/device", os.ModePerm)
//	if err != nil {
//		t.Error(err)
//	}
//	defer func() {
//		os.RemoveAll("tmp")
//	}()
//
//	devices, err := parseDeviceConfig("tmp")
//
//	if devices != nil {
//		t.Error("Expecting error - devices should be nil.")
//	}
//
//	if _, ok := err.(*os.PathError); !ok {
//		t.Errorf("Expected os.PathError but was %v", reflect.TypeOf(err))
//	}
//}

// bad permissions on the device config files.
func TestParseDeviceConfig3(t *testing.T) {
	err := os.MkdirAll("tmp/device", os.ModePerm)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll("tmp")
		if err != nil {
			t.Error(err)
		}
	}()

	// this string represents invalid YAML - it should cause failure when
	// we try to unmarshall.
	cfgFile := `--~+::\n:-`

	err = ioutil.WriteFile("tmp/device/test_config.yaml", []byte(cfgFile), 0333)
	if err != nil {
		t.Error("Failed to write test data to file.")
	}

	devices, err := parseDeviceConfig("tmp")

	if devices != nil {
		t.Error("Expecting error - devices should be nil.")
	}

	if _, ok := err.(*os.PathError); !ok {
		t.Errorf("Expected os.PathError but was %v", reflect.TypeOf(err))
	}
}

// invalid YAML contents
func TestParseDeviceConfig4(t *testing.T) {
	err := os.MkdirAll("tmp/device", os.ModePerm)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll("tmp")
		if err != nil {
			t.Error(err)
		}
	}()

	// this string represents invalid YAML - it should cause failure when
	// we try to unmarshall.
	cfgFile := `--~+::\n:-`

	err = ioutil.WriteFile("tmp/device/test_config.yaml", []byte(cfgFile), 0644)
	if err != nil {
		t.Error("Failed to write test data to file.")
	}

	devices, err := parseDeviceConfig("tmp")

	if devices != nil {
		t.Error("Expecting error - devices should be nil.")
	}

	if _, ok := err.(*yaml.TypeError); !ok {
		t.Errorf("Expected yaml.TypeError but was %v", reflect.TypeOf(err))
	}
}

// the positive case -- there should be no errors here.
func TestParseDeviceConfig5(t *testing.T) {
	err := os.MkdirAll("tmp/device", os.ModePerm)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll("tmp")
		if err != nil {
			t.Error(err)
		}
	}()

	cfgFile := `version: 1.0
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

	err = ioutil.WriteFile("tmp/device/test_config.yaml", []byte(cfgFile), 0644)
	if err != nil {
		t.Error("Failed to write test data to file.")
	}

	devices, err := parseDeviceConfig("tmp")

	if err != nil {
		t.Errorf("Not expecting error, but got %v", err)
	}

	if len(devices) != 3 {
		t.Errorf("Expecting only one device proto config, but got %v", len(devices))
	}
}

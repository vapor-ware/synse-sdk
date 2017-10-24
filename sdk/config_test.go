package sdk

import (
	"io/ioutil"
	"os"
	"testing"
	"gopkg.in/yaml.v2"
	"reflect"
)


func TestPluginConfig_FromFile(t *testing.T) {
	c := PluginConfig{}
	err := c.FromFile("path/does/not/exist")

	if _, ok := err.(*os.PathError); !ok {
		t.Error("Expected os path error.")
	}

}


func TestPluginConfig_FromFile2(t *testing.T) {
	os.MkdirAll("tmp", os.ModePerm)
	defer func() {
		os.RemoveAll("tmp")
	}()

	cfgFile := `--~+::\n:-`
	err := ioutil.WriteFile("tmp/test_config.yaml", []byte(cfgFile), 0644)
	if err != nil {
		t.Error("Failed to write test data to file.")
	}

	c := PluginConfig{}
	err = c.FromFile("tmp/test_config.yaml")

	if _, ok := err.(*yaml.TypeError); !ok {
		t.Error("Expected yaml type error.")
	}
}


func TestPluginConfig_FromFile3(t *testing.T) {
	os.MkdirAll("tmp", os.ModePerm)
	defer func() {
		os.RemoveAll("tmp")
	}()

	cfgFile := `name: test-plugin
version: 1.0
debug: false
settings:
  loop_delay: 10
  read:
    buffer_size: 50
  write:
    buffer_size: 50
    per_loop: 10
  transaction:
    ttl: 60
auto_enumerate:
  - name: enum1
    ip: 10.10.10.10
  - name: enum2
    ip: 11.11.11.11
context:
  key1: value1
  key2:
    subkey: subvalue`

	err := ioutil.WriteFile("tmp/test_config.yaml", []byte(cfgFile), 0644)
	if err != nil {
		t.Error("Failed to write test data to file.")
	}

	c := PluginConfig{}
	err = c.FromFile("tmp/test_config.yaml")
	if err != nil {
		t.Error("Parsing from file failed.")
	}

	if c.Name != "test-plugin" {
		t.Error("Unexpected value in the Name field.")
	}

	if c.Version != "1.0" {
		t.Error("Unexpected value in the Version field.")
	}

	if c.Debug != false {
		t.Error("Unexpected value in the Debug field.")
	}

	if c.Settings.LoopDelay != 10 {
		t.Error("Unexpected value in the Settings LoopDelay field.")
	}

	if c.Settings.Read.BufferSize != 50 {
		t.Error("Unexpected value in the Settings Read BufferSize field.")
	}

	if c.Settings.Write.BufferSize != 50 {
		t.Error("Unexpected value in the Settings Write BufferSize field.")
	}

	if c.Settings.Write.PerLoop != 10 {
		t.Error("Unexpected value in the Settings Write PerLoop field.")
	}

	if c.Settings.Transaction.TTL != 60 {
		t.Error("Unexpected value in the Settings Transaction TTL field.")
	}

	if len(c.AutoEnumerate) != 2 {
		t.Error("Unexpected value length in the AutoEnumerate field.")
	}

	if c.Context["key1"] != "value1" {
		t.Error("Unexpected value in the Context field.")
	}

	if c.Context["key2"].(map[interface{}]interface{})["subkey"] != "subvalue" {
		t.Error("Unexpected value in the Context field.")
	}
}

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

// bad permissions on the config proto directory
func TestParsePrototypeConfig2(t *testing.T) {
	os.MkdirAll("tmp/proto", 0555)
	defer func() {
		os.RemoveAll("tmp")
	}()

	devices, err := parsePrototypeConfig("tmp")

	if devices != nil {
		t.Error("Expecting error - devices should be nil.")
	}

	if _, ok := err.(*os.PathError); !ok {
		t.Errorf("Expected os.PathError but was %v", reflect.TypeOf(err))
	}
}


// bad permissions on the config file in the proto directory
func TestParsePrototypeConfig3(t *testing.T) {
	os.MkdirAll("tmp/proto", os.ModePerm)
	defer func() {
		os.RemoveAll("tmp")
	}()

	cfgFile := `--~+::\n:-`

	err := ioutil.WriteFile("tmp/proto/test_config.yaml", []byte(cfgFile), 0333)
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
	os.MkdirAll("tmp/proto", os.ModePerm)
	defer func() {
		os.RemoveAll("tmp")
	}()

	cfgFile := `--~+::\n:-`

	err := ioutil.WriteFile("tmp/proto/test_config.yaml", []byte(cfgFile), 0644)
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
	os.MkdirAll("tmp/proto", os.ModePerm)
	defer func() {
		os.RemoveAll("tmp")
	}()

	cfgFile := `version: 1.0
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

	err := ioutil.WriteFile("tmp/proto/test_config.yaml", []byte(cfgFile), 0644)
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

// bad permissions on the device config directory
func TestParseDeviceConfig2(t *testing.T) {
	os.MkdirAll("tmp/device", 0555)
	defer func() {
		os.RemoveAll("tmp")
	}()

	devices, err := parseDeviceConfig("tmp")

	if devices != nil {
		t.Error("Expecting error - devices should be nil.")
	}

	if _, ok := err.(*os.PathError); !ok {
		t.Errorf("Expected os.PathError but was %v", reflect.TypeOf(err))
	}
}

// bad permissions on the device config files.
func TestParseDeviceConfig3(t *testing.T) {
	os.MkdirAll("tmp/device", os.ModePerm)
	defer func() {
		os.RemoveAll("tmp")
	}()

	cfgFile := `--~+::\n:-`

	err := ioutil.WriteFile("tmp/device/test_config.yaml", []byte(cfgFile), 0333)
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
	os.MkdirAll("tmp/device", os.ModePerm)
	defer func() {
		os.RemoveAll("tmp")
	}()

	cfgFile := `--~+::\n:-`

	err := ioutil.WriteFile("tmp/device/test_config.yaml", []byte(cfgFile), 0644)
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
	os.MkdirAll("tmp/device", os.ModePerm)
	defer func() {
		os.RemoveAll("tmp")
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

	err := ioutil.WriteFile("tmp/device/test_config.yaml", []byte(cfgFile), 0644)
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

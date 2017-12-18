package sdk

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"gopkg.in/yaml.v2"
)

// the config path does not exist, FromFile should fail.
func TestPluginConfig_FromFile(t *testing.T) {
	c := PluginConfig{}
	err := c.FromFile("path/does/not/exist")

	if _, ok := err.(*os.PathError); !ok {
		t.Error("Expected os path error.")
	}

}

// the config file contains invalid YAML
func TestPluginConfig_FromFile2(t *testing.T) {
	err := os.MkdirAll("tmp", os.ModePerm)
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
	err = ioutil.WriteFile("tmp/test_config.yaml", []byte(cfgFile), 0644)
	if err != nil {
		t.Error("Failed to write test data to file.")
	}

	c := PluginConfig{}
	err = c.FromFile("tmp/test_config.yaml")

	if _, ok := err.(*yaml.TypeError); !ok {
		t.Error("Expected yaml type error.")
	}
}

// positive test - the config should load from file with no errors.
func TestPluginConfig_FromFile3(t *testing.T) {
	err := os.MkdirAll("tmp", os.ModePerm)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll("tmp")
		if err != nil {
			t.Error(err)
		}
	}()

	cfgFile := `name: test-plugin
version: 1.0
debug: false
socket:
  network: unix
  address: /vapor/test.sock
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

	err = ioutil.WriteFile("tmp/test_config.yaml", []byte(cfgFile), 0644)
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

	if c.Debug {
		t.Error("Unexpected value in the Debug field.")
	}

	if c.Socket.Network != "unix" {
		t.Error("Unexpected value in the Socket Network field.")
	}

	if c.Socket.Address != "/vapor/test.sock" {
		t.Error("Unexpected value in the Socket Address field.")
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

// merge PluginConfig when the required "name" is not specified.
func TestPluginConfig_merge(t *testing.T) {
	base := PluginConfig{}
	toMerge := PluginConfig{
		Version: "1.0",
		Socket: PluginConfigSocket{
			Network: "tcp",
			Address: ":666",
		},
	}

	err := base.merge(&toMerge)
	if err == nil {
		t.Error("merge should fail due to missing required fields.")
	}
}

// merge PluginConfig when the required "version" is not specified.
func TestPluginConfig_merge2(t *testing.T) {
	base := PluginConfig{}
	toMerge := PluginConfig{
		Name: "test-plugin",
		Socket: PluginConfigSocket{
			Network: "tcp",
			Address: ":666",
		},
	}

	err := base.merge(&toMerge)
	if err == nil {
		t.Error("merge should fail due to missing required fields.")
	}
}

// merge with all configurations specified. in this case, the
// original config struct should have the same values as the one
// that is being merged in
func TestPluginConfig_merge3(t *testing.T) {
	base := PluginConfig{
		Name:    "initial",
		Version: "1.0",
		Debug:   false,
		Settings: PluginConfigSettings{
			LoopDelay: 10,
			Read: PluginConfigSettingsRead{
				BufferSize: 200,
			},
			Write: PluginConfigSettingsWrite{
				BufferSize: 200,
				PerLoop:    10,
			},
			Transaction: PluginConfigSettingsTransaction{
				TTL: 350,
			},
		},
		Socket: PluginConfigSocket{
			Network: "unix",
			Address: "test.sock",
		},
		AutoEnumerate: []map[string]interface{}{
			{"test-key": "test-value"},
		},
		Context: map[string]interface{}{
			"ctx-key": "ctx-value",
		},
	}

	toMerge := PluginConfig{
		Name:    "new",
		Version: "2.0",
		Debug:   true,
		Settings: PluginConfigSettings{
			LoopDelay: 15,
			Read: PluginConfigSettingsRead{
				BufferSize: 300,
			},
			Write: PluginConfigSettingsWrite{
				BufferSize: 300,
				PerLoop:    5,
			},
			Transaction: PluginConfigSettingsTransaction{
				TTL: 500,
			},
		},
		Socket: PluginConfigSocket{
			Network: "tcp",
			Address: ":50051",
		},
		AutoEnumerate: []map[string]interface{}{
			{"new-key": "new-value"},
		},
		Context: map[string]interface{}{
			"new-ctx-key": "new-ctx-value",
		},
	}

	err := base.merge(&toMerge)
	if err != nil {
		t.Error(err)
	}

	if !configEqual(base, toMerge) {
		t.Errorf("configs expected to be equal after merge: %#v but wanted %#v", base, toMerge)
	}
}

func configEqual(c1, c2 PluginConfig) bool {
	if c1.Name != c2.Name {
		return false
	}
	if c1.Version != c2.Version {
		return false
	}
	if c1.Debug != c2.Debug {
		return false
	}
	if c1.Settings.LoopDelay != c2.Settings.LoopDelay {
		return false
	}
	if c1.Settings.Read != c2.Settings.Read {
		return false
	}
	if c1.Settings.Write != c2.Settings.Write {
		return false
	}
	if c1.Settings.Transaction != c2.Settings.Transaction {
		return false
	}
	if c1.Socket != c2.Socket {
		return false
	}
	if len(c1.AutoEnumerate) != len(c2.AutoEnumerate) {
		return false
	}
	for i := 0; i < len(c1.AutoEnumerate); i++ {
		if !reflect.DeepEqual(c1.AutoEnumerate[i], c2.AutoEnumerate[i]) {
			return false
		}
	}
	return reflect.DeepEqual(c1.Context, c2.Context)
}

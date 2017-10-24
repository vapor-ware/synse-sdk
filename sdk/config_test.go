package sdk

import (
	"io/ioutil"
	"os"
	"testing"
	"gopkg.in/yaml.v2"
	"fmt"
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

	fmt.Printf("context: %+v\n", c.Context)

	if c.Context["key2"].(map[interface{}]interface{})["subkey"] != "subvalue" {
		t.Error("Unexpected value in the Context field.")
	}
}

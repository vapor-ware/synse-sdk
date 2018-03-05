// sdktest.go provides utilities for testing the SDK.

package sdk

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/vapor-ware/synse-sdk/sdk/config"
)

// == Plugin Stubs ==

func testDeviceIdentifier(in map[string]string) string                               { return "" }
func testDeviceEnumerator(in map[string]interface{}) ([]*config.DeviceConfig, error) { return nil, nil }

var testHandlers = Handlers{
	DeviceIdentifier: testDeviceIdentifier,
	DeviceEnumerator: testDeviceEnumerator,
}

var testDevHandlers = []*DeviceHandler{
	&testDeviceHandler1,
	&testDeviceHandler2,
}

var testPlugin = Plugin{
	handlers:       &testHandlers,
	deviceHandlers: testDevHandlers,
}

var testDeviceHandler = DeviceHandler{
	Read:  func(device *Device) ([]*Reading, error) { return nil, nil },
	Write: func(device *Device, data *WriteData) error { return nil },
}

// == Utility Functions ==

func makeDeviceConfig() *config.DeviceConfig {
	location := config.Location{
		Rack:  "TestRack",
		Board: "TestBoard",
	}
	location.Validate()
	return &config.DeviceConfig{
		Version:  "1.0",
		Type:     "TestDevice",
		Model:    "TestModel",
		Location: location,
		Data:     map[string]string{"testKey": "testValue"},
	}
}

func makePrototypeConfig() *config.PrototypeConfig {
	return &config.PrototypeConfig{
		Version:      "1.0",
		Type:         "TestDevice",
		Model:        "TestModel",
		Manufacturer: "TestManufacturer",
		Protocol:     "TestProtocol",
		Output: []config.DeviceOutput{{
			Type:     "TestType",
			DataType: "string",
			Unit: &config.Unit{
				Name:   "TestName",
				Symbol: "TestSymbol",
			},
			Precision: 3,
			Range: &config.Range{
				Min: 1,
				Max: 5,
			},
		}},
	}
}

func makeTestDevice() *Device {
	protoConfig := makePrototypeConfig()
	deviceConfig := makeDeviceConfig()

	return &Device{
		pconfig:      protoConfig,
		dconfig:      deviceConfig,
		Type:         protoConfig.Type,
		Model:        protoConfig.Model,
		Manufacturer: protoConfig.Manufacturer,
		Protocol:     protoConfig.Protocol,
		Output:       protoConfig.Output,
		Location:     deviceConfig.Location,
		Data:         deviceConfig.Data,

		Handler:    &DeviceHandler{},
		Identifier: testDeviceIdentifier,
	}
}

func writeConfigFile(path, config string) error {
	_, err := os.Stat(filepath.Dir(path))
	if os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(path), os.ModePerm)
	}
	return ioutil.WriteFile(path, []byte(config), 0644)
}

// == Test Data ==

var testDeviceConfig1 = config.DeviceConfig{
	Version: "1.0",
	Type:    "test-device",
	Model:   "td-1",
	Location: config.Location{
		Rack:  "rack-1",
		Board: "board-1",
	},
}

var testDeviceConfig2 = config.DeviceConfig{
	Version: "1.0",
	Type:    "test-device",
	Model:   "td-1",
	Location: config.Location{
		Rack:  "rack-1",
		Board: "board-2",
	},
}

var testPrototypeConfig1 = config.PrototypeConfig{
	Version:      "1.0",
	Type:         "test-device",
	Model:        "td-1",
	Manufacturer: "vaporio",
	Protocol:     "test",
}

var testPrototypeConfig2 = config.PrototypeConfig{
	Version:      "1.0",
	Type:         "test-device",
	Model:        "td-3",
	Manufacturer: "vaporio",
	Protocol:     "test",
}

var testDeviceHandler1 = DeviceHandler{
	Type:  "test-device",
	Model: "td-1",
}

var testDeviceHandler2 = DeviceHandler{
	Type:  "test-device",
	Model: "td-3",
}

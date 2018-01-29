package sdk

import (
	"testing"

	"github.com/vapor-ware/synse-sdk/sdk/config"
)

// ===== Test Data =====

func makeTestPrototypeConfig() *config.PrototypeConfig {
	return &config.PrototypeConfig{
		Version:      "1",
		Type:         "TestDevice",
		Model:        "TestModel",
		Manufacturer: "TestManufacturer",
		Protocol:     "TestProtocol",
		Output: []config.DeviceOutput{{
			Type: "TestType",
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

func makeTestDeviceConfig() *config.DeviceConfig {
	return &config.DeviceConfig{
		Version: "1",
		Type:    "TestDevice",
		Model:   "TestModel",
		Location: config.Location{
			Rack:  "TestRack",
			Board: "TestBoard",
		},
		Data: map[string]string{"testKey": "testValue"},
	}
}

var protoConfig = makeTestPrototypeConfig()
var deviceConfig = makeTestDeviceConfig()

var testDevice = Device{
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

var deviceHandler = DeviceHandler{
	Read:  func(device *Device) ([]*Reading, error) { return nil, nil },
	Write: func(device *Device, data *WriteData) error { return nil },
}

// ===== Test Cases =====

func TestDevice_Type(t *testing.T) {
	if testDevice.Type != protoConfig.Type {
		t.Error("device Type does not match prototype config")
	}
}

func TestDevice_Model(t *testing.T) {
	if testDevice.Model != protoConfig.Model {
		t.Error("device Model does not match prototype config")
	}
}

func TestDevice_Manufacturer(t *testing.T) {
	if testDevice.Manufacturer != protoConfig.Manufacturer {
		t.Error("device Manufacturer does not match prototype config")
	}
}

func TestDevice_Protocol(t *testing.T) {
	if testDevice.Protocol != protoConfig.Protocol {
		t.Error("device Protocol does not match prototype config")
	}
}

func TestDevice_ID(t *testing.T) {
	if testDevice.ID() != "664f6cfa51c9bef163682bd2a766613b" {
		t.Errorf("device ID %q does not match expected ID", testDevice.ID())
	}
}

func TestDevice_GUID(t *testing.T) {
	if testDevice.GUID() != "TestRack-TestBoard-664f6cfa51c9bef163682bd2a766613b" {
		t.Errorf("device GUID %q does not match expected GUID", testDevice.GUID())
	}
}

func TestDevice_Output(t *testing.T) {
	if len(testDevice.Output) != len(protoConfig.Output) {
		t.Error("device Output length does not match expected")
	}
	for i := 0; i < len(testDevice.Output); i++ {
		if testDevice.Output[i] != protoConfig.Output[i] {
			t.Error("device Output does not match prototype config")
		}
	}
}

func TestDevice_Location(t *testing.T) {
	if testDevice.Location != deviceConfig.Location {
		t.Error("device Location does not match instance config")
	}
}

func TestDevice_Data(t *testing.T) {
	if len(testDevice.Data) != len(deviceConfig.Data) {
		t.Error("device Data length does not match expected")
	}
	for k, v := range testDevice.Data {
		if deviceConfig.Data[k] != v {
			t.Error("device Data key/value mismatch")
		}
	}
}

func TestEncodeDevice(t *testing.T) {
	encoded := testDevice.encode()

	if encoded.Uid != testDevice.ID() {
		t.Error("Device.encode() -> Uid incorrect")
	}

	if encoded.Type != testDevice.Type {
		t.Error("Device.encode() -> Type incorrect")
	}

	if encoded.Model != testDevice.Model {
		t.Error("Device.encode() -> Model incorrect")
	}

	if encoded.Manufacturer != testDevice.Manufacturer {
		t.Error("Device.encode() -> Manufacturer incorrect")
	}

	if encoded.Protocol != testDevice.Protocol {
		t.Error("Device.encode() -> Protocol incorrect")
	}

	if encoded.Info != "" {
		t.Error("Device.encode() -> Info incorrect")
	}

	if encoded.Comment != "" {
		t.Error("Device.encode() -> Comment incorrect")
	}

	if encoded.Location.Rack != testDevice.Location.Rack {
		t.Error("Device.encode() -> Location.Rack incorrect")
	}

	if encoded.Location.Board != testDevice.Location.Board {
		t.Error("Device.encode() -> Location.Board incorrect")
	}
}

func TestNewDevice(t *testing.T) {
	p := Plugin{
		handlers: &Handlers{
			DeviceIdentifier: testDeviceIdentifier,
		},
	}

	d, err := NewDevice(protoConfig, deviceConfig, &deviceHandler, &p)
	if err != nil {
		t.Error(err)
	}

	if d.Type != protoConfig.Type {
		t.Errorf("device Type does not match expected")
	}

	if d.Model != protoConfig.Model {
		t.Errorf("device Model does not match expected")
	}

	if d.Manufacturer != protoConfig.Manufacturer {
		t.Errorf("device Manufacturer does not match expected")
	}

	if d.Protocol != protoConfig.Protocol {
		t.Errorf("device Protocol does not match expected")
	}

	if len(d.Output) != len(protoConfig.Output) {
		t.Errorf("device Output length does not match expected")
	}

	if d.Location != deviceConfig.Location {
		t.Errorf("device Location does not match expected")
	}

	if len(d.Data) != len(deviceConfig.Data) {
		t.Errorf("device Data length does not match expected")
	}

	if d.Handler != &deviceHandler {
		t.Errorf("device Handler does not match expected")
	}

	if d.dconfig != deviceConfig {
		t.Errorf("device instance config does not match expected")
	}

	if d.pconfig != protoConfig {
		t.Errorf("device prototype config does not match expected")
	}
}

func TestNewDevice2(t *testing.T) {
	p := Plugin{
		handlers: &Handlers{},
	}

	_, err := NewDevice(protoConfig, deviceConfig, &deviceHandler, &p)
	if err == nil {
		t.Error("NewDevice -> expected validation error, but got no error")
	}
}

func TestNewDevice3(t *testing.T) {
	p := Plugin{
		handlers: &Handlers{
			DeviceIdentifier: testDeviceIdentifier,
		},
	}

	dCfg := makeTestDeviceConfig()
	dCfg.Type = "foo"

	_, err := NewDevice(protoConfig, dCfg, &deviceHandler, &p)
	if err == nil {
		t.Error("NewDevice -> expected validation error, but got no error")
	}
}

func TestNewDevice4(t *testing.T) {
	p := Plugin{
		handlers: &Handlers{
			DeviceIdentifier: testDeviceIdentifier,
		},
	}

	dCfg := makeTestDeviceConfig()
	dCfg.Model = "foo"

	_, err := NewDevice(protoConfig, dCfg, &deviceHandler, &p)
	if err == nil {
		t.Error("NewDevice -> expected validation error, but got no error")
	}
}

//func makeDeviceConfig() error {
//	err := os.MkdirAll("config/device", os.ModePerm)
//	if err != nil {
//		return err
//	}
//	cfgFile := `version: 1.0
//type: emulated-temperature
//model: emul8-temp
//
//locations:
//  unknown:
//    rack: unknown
//    board: unknown
//
//devices:
//  - id: 1
//    location: unknown
//    comment: first emulated temperature device
//    info: CEC temp 1`
//
//	return ioutil.WriteFile("config/device/test_config.yaml", []byte(cfgFile), 0644)
//}
//
//func makeProtoConfig() error {
//	err := os.MkdirAll("config/proto", os.ModePerm)
//	if err != nil {
//		return err
//	}
//	cfgFile := `version: 1.0
//type: emulated-temperature
//model: emul8-temp
//manufacturer: vaporio
//protocol: emulator
//output:
//  - type: temperature
//    data_type: float
//    unit:
//      name: celsius
//      symbol: C
//    precision: 2
//    range:
//      min: 0
//      max: 100`
//
//	return ioutil.WriteFile("config/proto/test_config.yaml", []byte(cfgFile), 0644)
//}
//
//type devicesTestHandler struct{}
//
//func (h *devicesTestHandler) GetProtocolIdentifiers(data map[string]string) string {
//	return data["id"]
//}
//
//func (h *devicesTestHandler) EnumerateDevices(cfg map[string]interface{}) ([]*config.DeviceConfig, error) {
//	dc := config.DeviceConfig{
//		Version: "1.0",
//		Type:    "emulated-temperature",
//		Model:   "emul8-temp",
//		Location: config.Location{
//			Rack:  "unknown",
//			Board: "unknown",
//		},
//		Data: map[string]string{
//			"id":      cfg["id"].(string),
//			"comment": "auto-enumerated",
//		},
//	}
//	return []*config.DeviceConfig{&dc}, nil
//}

// FIXME -- theses tests are doing a bad thing! removing 'config' dir.
// now that we have a 'config' package here, it will delete that. right
// now "config" is hardcoded as the path for device/proto configs. that is
// set to change in the next batch of work (e.g. upcoming PR) so instead of
// dealing with it here, just disable the tests for the time being.

//func TestRegisterDevicesFromConfig(t *testing.T) {
//	err := makeProtoConfig()
//	if err != nil {
//		t.Error(err)
//	}
//	err = makeDeviceConfig()
//	if err != nil {
//		t.Error(err)
//	}
//	defer func() {
//		err = os.RemoveAll("config")
//		if err != nil {
//			t.Error(err)
//		}
//		// reset the device map
//		deviceMap = make(map[string]*Device)
//	}()
//
//	startLen := len(deviceMap)
//
//	err = registerDevicesFromConfig(&devicesTestHandler{}, []map[string]interface{}{})
//	if err != nil {
//		t.Errorf("unexpected error when registering devices from config: %v", err)
//	}
//
//	if len(deviceMap) != startLen+1 {
//		t.Errorf("expected 1 device to be added to device map, %v added instead", len(deviceMap)-startLen)
//	}
//}
//
//// no device instance configurations
//func TestRegisterDevicesFromConfig2(t *testing.T) {
//	err := makeProtoConfig()
//	if err != nil {
//		t.Error(err)
//	}
//	defer func() {
//		err = os.RemoveAll("config")
//		if err != nil {
//			t.Error(err)
//		}
//		// reset the device map
//		deviceMap = make(map[string]*Device)
//	}()
//
//	startLen := len(deviceMap)
//
//	err = registerDevicesFromConfig(&devicesTestHandler{}, []map[string]interface{}{})
//	if err == nil {
//		t.Errorf("expected error for missing device instance config, but got none")
//	}
//
//	if startLen != len(deviceMap) {
//		t.Error("deviceMap size changed when nothing should have been added")
//	}
//}
//
//// no device prototype configurations
//func TestRegisterDevicesFromConfig3(t *testing.T) {
//	err := makeDeviceConfig()
//	if err != nil {
//		t.Error(err)
//	}
//	defer func() {
//		err = os.RemoveAll("config")
//		if err != nil {
//			t.Error(err)
//		}
//		// reset the device map
//		deviceMap = make(map[string]*Device)
//	}()
//
//	startLen := len(deviceMap)
//
//	err = registerDevicesFromConfig(&devicesTestHandler{}, []map[string]interface{}{})
//	if err == nil {
//		t.Errorf("expected error for missing device prototype config, but got none")
//	}
//
//	if startLen != len(deviceMap) {
//		t.Error("deviceMap size changed when nothing should have been added")
//	}
//}
//
//// test with auto-enumeration
//func TestRegisterDevicesFromConfig4(t *testing.T) {
//	autoEnum := []map[string]interface{}{
//		{"id": "2"},
//	}
//	err := makeProtoConfig()
//	if err != nil {
//		t.Error(err)
//	}
//	err = makeDeviceConfig()
//	if err != nil {
//		t.Error(err)
//	}
//	defer func() {
//		err = os.RemoveAll("config")
//		if err != nil {
//			t.Error(err)
//		}
//		// reset the device map
//		deviceMap = make(map[string]*Device)
//	}()
//
//	startLen := len(deviceMap)
//
//	err = registerDevicesFromConfig(&devicesTestHandler{}, autoEnum)
//	if err != nil {
//		t.Errorf("unexpected error when registering devices from config: %v", err)
//	}
//
//	if len(deviceMap) != startLen+2 {
//		t.Errorf("expected 2 devices to be added to device map, %v added instead", len(deviceMap)-startLen)
//	}
//}

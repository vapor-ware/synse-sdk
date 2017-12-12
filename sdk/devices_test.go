package sdk

import (
	"testing"
)

// ===== Test Data =====

var protoConfig = PrototypeConfig{
	Version:      "1",
	Type:         "TestDevice",
	Model:        "TestModel",
	Manufacturer: "TestManufacturer",
	Protocol:     "TestProtocol",
	Output: []DeviceOutput{{
		Type: "TestType",
		Unit: &OutputUnit{
			Name:   "TestName",
			Symbol: "TestSymbol",
		},
		Precision: 3,
		Range: &OutputRange{
			Min: 1,
			Max: 5,
		},
	}},
}

var deviceConfig = DeviceConfig{
	Version: "1",
	Type:    "TestDevice",
	Model:   "TestModel",
	Location: DeviceLocation{
		Rack:  "TestRack",
		Board: "TestBoard",
	},
	Data: map[string]string{"testKey": "testValue"},
}


var testDevice = Device{
	Prototype: &protoConfig,
	Instance:  &deviceConfig,
	Handler:   &testDeviceHandler{},
}

// ===== Test Cases =====

func TestDevice_Type(t *testing.T) {
	if testDevice.Type() != protoConfig.Type {
		t.Error("device Type does not match prototype config")
	}
}

func TestDevice_Model(t *testing.T) {
	if testDevice.Model() != protoConfig.Model {
		t.Error("device Model does not match prototype config")
	}
}

func TestDevice_Manufacturer(t *testing.T) {
	if testDevice.Manufacturer() != protoConfig.Manufacturer {
		t.Error("device Manufacturer does not match prototype config")
	}
}

func TestDevice_Protocol(t *testing.T) {
	if testDevice.Protocol() != protoConfig.Protocol {
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
	if len(testDevice.Output()) != len(protoConfig.Output) {
		t.Error("device Output length does not match expected")
	}
	for i := 0; i < len(testDevice.Output()); i++ {
		if testDevice.Output()[i] != protoConfig.Output[i] {
			t.Error("device Output does nto match prototype config")
		}
	}
}

func TestDevice_Location(t *testing.T) {
	if testDevice.Location() != deviceConfig.Location {
		t.Error("device Location does not match instance config")
	}
}

func TestDevice_Data(t *testing.T) {
	if len(testDevice.Data()) != len(deviceConfig.Data) {
		t.Error("device Data length does not match expected")
	}
	for k, v := range testDevice.Data() {
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

	if encoded.Type != testDevice.Type() {
		t.Error("Device.encode() -> Type incorrect")
	}

	if encoded.Model != testDevice.Model() {
		t.Error("Device.encode() -> Model incorrect")
	}

	if encoded.Manufacturer != testDevice.Manufacturer() {
		t.Error("Device.encode() -> Manufacturer incorrect")
	}

	if encoded.Protocol != testDevice.Protocol() {
		t.Error("Device.encode() -> Protocol incorrect")
	}

	if encoded.Info != "" {
		t.Error("Device.encode() -> Info incorrect")
	}

	if encoded.Comment != "" {
		t.Error("Device.encode() -> Comment incorrect")
	}

	if encoded.Location.Rack != testDevice.Location().Rack {
		t.Error("Device.encode() -> Location.Rack incorrect")
	}

	if encoded.Location.Board != testDevice.Location().Board {
		t.Error("Device.encode() -> Location.Board incorrect")
	}
}

package sdk

import (
	"testing"
)

func TestDevice_Type(t *testing.T) {
	testDevice := makeTestDevice()

	if testDevice.Type != testDevice.pconfig.Type {
		t.Error("device Type does not match prototype config")
	}

	if testDevice.Model != testDevice.pconfig.Model {
		t.Error("device Model does not match prototype config")
	}

	if testDevice.Manufacturer != testDevice.pconfig.Manufacturer {
		t.Error("device Manufacturer does not match prototype config")
	}

	if testDevice.Protocol != testDevice.pconfig.Protocol {
		t.Error("device Protocol does not match prototype config")
	}

	if testDevice.ID() != "664f6cfa51c9bef163682bd2a766613b" {
		t.Errorf("device ID %q does not match expected ID", testDevice.ID())
	}

	if testDevice.GUID() != "TestRack-TestBoard-664f6cfa51c9bef163682bd2a766613b" {
		t.Errorf("device GUID %q does not match expected GUID", testDevice.GUID())
	}

	if len(testDevice.Output) != len(testDevice.pconfig.Output) {
		t.Error("device Output length does not match expected")
	}
	for i := 0; i < len(testDevice.Output); i++ {
		if testDevice.Output[i] != testDevice.pconfig.Output[i] {
			t.Error("device Output does not match prototype config")
		}
	}

	if testDevice.Location != testDevice.dconfig.Location {
		t.Error("device Location does not match instance config")
	}

	if len(testDevice.Data) != len(testDevice.dconfig.Data) {
		t.Error("device Data length does not match expected")
	}
	for k, v := range testDevice.Data {
		if testDevice.dconfig.Data[k] != v {
			t.Error("device Data key/value mismatch")
		}
	}
}

func TestEncodeDevice(t *testing.T) {
	testDevice := makeTestDevice()
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

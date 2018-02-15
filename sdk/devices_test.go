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

func TestNewDevice(t *testing.T) {
	// Create Handlers.
	handlers, err := NewHandlers(testDeviceIdentifier, nil)
	if err != nil {
		t.Errorf("TestNewDevice expected no error, got: %v", err)
	}

	// Initialize Plugin with handlers.
	p := Plugin{
		handlers: handlers,
	}

	protoConfig := makePrototypeConfig()
	deviceConfig := makeDeviceConfig()

	d, err := NewDevice(protoConfig, deviceConfig, &testDeviceHandler, &p)
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

	if d.Handler != &testDeviceHandler {
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

	protoConfig := makePrototypeConfig()
	deviceConfig := makeDeviceConfig()

	_, err := NewDevice(protoConfig, deviceConfig, &testDeviceHandler, &p)
	if err == nil {
		t.Error("NewDevice -> expected validation error, but got no error")
	}
}

func TestNewDevice3(t *testing.T) {
	// Create handlers.
	handlers, err := NewHandlers(testDeviceIdentifier, nil)
	if err != nil {
		t.Errorf("TestNewDevice3 expected no error, got: %v", err)
	}
	// Initialize plugin with handlers.
	p := Plugin{
		handlers: handlers,
	}

	protoConfig := makePrototypeConfig()
	deviceConfig := makeDeviceConfig()
	deviceConfig.Type = "foo"

	_, err = NewDevice(protoConfig, deviceConfig, &testDeviceHandler, &p)
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

	protoConfig := makePrototypeConfig()
	deviceConfig := makeDeviceConfig()
	deviceConfig.Type = "foo"

	_, err := NewDevice(protoConfig, deviceConfig, &testDeviceHandler, &p)
	if err == nil {
		t.Error("NewDevice -> expected validation error, but got no error")
	}
}

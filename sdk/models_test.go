package sdk

import (
	"testing"

	synse "github.com/vapor-ware/synse-server-grpc/go"
)

var pConfig = PrototypeConfig{
	Version: "1",
	Type: "TestDevice",
	Model: "TestModel",
	Manufacturer: "TestManufacturer",
	Protocol: "TestProtocol",
	Output: []DeviceOutput{{
		Type: "TestType",
		Unit: &OutputUnit{
			Name: "TestName",
			Symbol: "TestSymbol",
		},
		Precision: 3,
		Range: &OutputRange{
			Min: 1,
			Max: 5,
		},
	}},
}

var dConfig = DeviceConfig{
	Version: "1",
	Type: "TestDevice",
	Model: "TestModel",
	Location: DeviceLocation{
		Rack: "TestRack",
		Board: "TestBoard",
	},
	Data: map[string]string{"testKey": "testValue"},
}

type TestHandler struct {}

func (h *TestHandler) GetProtocolIdentifiers(in map[string]string) string {
	return ""
}


func TestWriteData_ToGRPC(t *testing.T) {
	expected := &synse.WriteData{
		Raw: [][]byte{{0, 1, 2}},
		Action: "test",
	}

	wd := WriteData{
		Raw: [][]byte{{0, 1, 2}},
		Action: "test",
	}

	actual := wd.toGRPC()

	if len(actual.Raw) != len(expected.Raw) {
		t.Error("WriteData Raw length mismatch.")
	}

	for i := 0; i < len(actual.Raw); i++ {
		for j := 0; j < len(actual.Raw[i]); j++ {
			if actual.Raw[i][j] != expected.Raw[i][j] {
				t.Error("WriteData Raw value mismatch")
			}
		}
	}

	if actual.Action != expected.Action {
		t.Error("WriteData Action mismatch.")
	}
}

func TestWriteDataFromGRPC(t *testing.T) {
	expected := &WriteData{
		Raw: [][]byte{{3, 2, 1}},
		Action: "test",
	}

	wd := &synse.WriteData{
		Raw: [][]byte{{3, 2, 1}},
		Action: "test",
	}

	actual := writeDataFromGRPC(wd)

	if len(actual.Raw) != len(expected.Raw) {
		t.Error("WriteData Raw length mismatch.")
	}

	for i := 0; i < len(actual.Raw); i++ {
		for j := 0; j < len(actual.Raw[i]); j++ {
			if actual.Raw[i][j] != expected.Raw[i][j] {
				t.Error("WriteData Raw value mismatch")
			}
		}
	}

	if actual.Action != expected.Action {
		t.Error("WriteData Action mismatch.")
	}
}


func TestNewUID(t *testing.T) {
	hash1 := newUID("", "", "", "")
	hash2 := newUID("", "", "", "")
	hash3 := newUID("a", "b", "c", "d")
	hash4 := newUID("a", "b", "c", "d")

	if hash1 != hash2 {
		t.Errorf("Empty value hashes do not match: %v %v", hash1, hash2)
	}

	if hash3 != hash4 {
		t.Errorf("Hashes do not match, but are expected to: %v %v", hash3, hash4)
	}

}


func TestDevice_Type(t *testing.T) {
	d := Device{
		Prototype: pConfig,
		Instance: dConfig,
		Handler: &TestHandler{},
	}

	if d.Type() != pConfig.Type {
		t.Error("Device Type does not match prototype config.")
	}
}


func TestDevice_Model(t *testing.T) {
	d := Device{
		Prototype: pConfig,
		Instance: dConfig,
		Handler: &TestHandler{},
	}

	if d.Model() != pConfig.Model {
		t.Error("Device Model does not match prototype config.")
	}
}


func TestDevice_Manufacturer(t *testing.T) {
	d := Device{
		Prototype: pConfig,
		Instance: dConfig,
		Handler: &TestHandler{},
	}

	if d.Manufacturer() != pConfig.Manufacturer {
		t.Error("Device Manufacturer does not match prototype config.")
	}
}


func TestDevice_Protocol(t *testing.T) {
	d := Device{
		Prototype: pConfig,
		Instance: dConfig,
		Handler: &TestHandler{},
	}

	if d.Protocol() != pConfig.Protocol {
		t.Error("Device Protocol does not match prototype config.")
	}
}


func TestDevice_UID(t *testing.T) {
	d := Device{
		Prototype: pConfig,
		Instance: dConfig,
		Handler: &TestHandler{},
	}

	if d.UID() != "664f6cfa51c9bef163682bd2a766613b" {
		t.Error("Device Uid does not generated uid.")
	}
}


func TestDevice_Output(t *testing.T) {
	d := Device{
		Prototype: pConfig,
		Instance: dConfig,
		Handler: &TestHandler{},
	}

	if len(d.Output()) != len(pConfig.Output) {
		t.Error("Device Output length mismatch.")
	}

	for i := 0; i < len(d.Output()); i++ {
		if d.Output()[i] != pConfig.Output[i] {
			t.Error("Device Output does not match prototype config.")
		}
	}
}


func TestDevice_Location(t *testing.T) {
	d := Device{
		Prototype: pConfig,
		Instance: dConfig,
		Handler: &TestHandler{},
	}

	if d.Location() != dConfig.Location {
		t.Error("Device Location does not match instance config.")
	}
}


func TestDevice_Data(t *testing.T) {
	d := Device{
		Prototype: pConfig,
		Instance: dConfig,
		Handler: &TestHandler{},
	}

	if len(d.Data()) != len(dConfig.Data) {
		t.Error("Device Data length mismatch.")
	}

	for k, v := range d.Data() {
		if dConfig.Data[k] != v {
			t.Error("Device Data key/value mismatch.")
		}
	}
}


func TestDevice_ToMetainfoResponse(t *testing.T) {
	d := Device{
		Prototype: pConfig,
		Instance: dConfig,
		Handler: &TestHandler{},
	}

	meta := d.toMetainfoResponse()

	if meta.Uid != d.UID() {
		t.Error("MetainfoResponse Uid incorrect.")
	}

	if meta.Type != d.Type() {
		t.Error("MetainfoResponse Type incorrect.")
	}

	if meta.Model != d.Model() {
		t.Error("MetainfoResponse Model incorrect.")
	}

	if meta.Manufacturer != d.Manufacturer() {
		t.Error("MetainfoResponse Manufacturer incorrect.")
	}

	if meta.Protocol != d.Protocol() {
		t.Error("MetainfoResponse Protocol incorrect.")
	}

	if meta.Info != "" {
		t.Error("MetainfoResponse Info incorrect.")
	}

	if meta.Comment != "" {
		t.Error("MetainfoResponse Comment incorrect.")
	}

	if meta.Location.Rack != d.Location().Rack {
		t.Error("MetainfoResponse Location Rack incorrect.")
	}

	if meta.Location.Board != d.Location().Board {
		t.Error("MetainfoResponse Location Board incorrect.")
	}
}

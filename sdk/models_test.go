package sdk

//import (
//	"testing"
//
//	synse "github.com/vapor-ware/synse-server-grpc/go"
//)
//
//var pConfig = PrototypeConfig{
//	Version:      "1",
//	Type:         "TestDevice",
//	Model:        "TestModel",
//	Manufacturer: "TestManufacturer",
//	Protocol:     "TestProtocol",
//	Output: []DeviceOutput{{
//		Type: "TestType",
//		Unit: &OutputUnit{
//			Name:   "TestName",
//			Symbol: "TestSymbol",
//		},
//		Precision: 3,
//		Range: &OutputRange{
//			Min: 1,
//			Max: 5,
//		},
//	}},
//}
//
//var dConfig = DeviceConfig{
//	Version: "1",
//	Type:    "TestDevice",
//	Model:   "TestModel",
//	Location: DeviceLocation{
//		Rack:  "TestRack",
//		Board: "TestBoard",
//	},
//	Data: map[string]string{"testKey": "testValue"},
//}
//
//type TestHandler struct{}
//
//func (h *TestHandler) GetProtocolIdentifiers(in map[string]string) string {
//	return ""
//}
//
//func (h *TestHandler) EnumerateDevices(map[string]interface{}) ([]*DeviceConfig, error) {
//	return nil, nil
//}
//
//func TestWriteData_ToGRPC(t *testing.T) {
//	expected := &synse.WriteData{
//		Raw:    [][]byte{{0, 1, 2}},
//		Action: "test",
//	}
//
//	wd := WriteData{
//		Raw:    [][]byte{{0, 1, 2}},
//		Action: "test",
//	}
//
//	actual := wd.encode()
//
//	if len(actual.Raw) != len(expected.Raw) {
//		t.Error("WriteData Raw length mismatch.")
//	}
//
//	for i := 0; i < len(actual.Raw); i++ {
//		for j := 0; j < len(actual.Raw[i]); j++ {
//			if actual.Raw[i][j] != expected.Raw[i][j] {
//				t.Error("WriteData Raw value mismatch")
//			}
//		}
//	}
//
//	if actual.Action != expected.Action {
//		t.Error("WriteData Action mismatch.")
//	}
//}
//
//func TestWriteDataFromGRPC(t *testing.T) {
//	expected := &WriteData{
//		Raw:    [][]byte{{3, 2, 1}},
//		Action: "test",
//	}
//
//	wd := &synse.WriteData{
//		Raw:    [][]byte{{3, 2, 1}},
//		Action: "test",
//	}
//
//	actual := writeDataFromGRPC(wd)
//
//	if len(actual.Raw) != len(expected.Raw) {
//		t.Error("WriteData Raw length mismatch.")
//	}
//
//	for i := 0; i < len(actual.Raw); i++ {
//		for j := 0; j < len(actual.Raw[i]); j++ {
//			if actual.Raw[i][j] != expected.Raw[i][j] {
//				t.Error("WriteData Raw value mismatch")
//			}
//		}
//	}
//
//	if actual.Action != expected.Action {
//		t.Error("WriteData Action mismatch.")
//	}
//}
//
//func TestNewUID(t *testing.T) {
//	hash1 := newUID("", "", "", "")
//	hash2 := newUID("", "", "", "")
//	hash3 := newUID("a", "b", "c", "d")
//	hash4 := newUID("a", "b", "c", "d")
//
//	if hash1 != hash2 {
//		t.Errorf("Empty value hashes do not match: %v %v", hash1, hash2)
//	}
//
//	if hash3 != hash4 {
//		t.Errorf("Hashes do not match, but are expected to: %v %v", hash3, hash4)
//	}
//
//}
//
//
//func TestReadResource_IdString(t *testing.T) {
//	r := ReadResource{
//		Device: "123",
//		Board:  "456",
//		Rack:   "789",
//	}
//
//	id := r.IDString()
//	if id != "789-456-123" {
//		t.Error("Unexpected IDString generated for ReadResource.")
//	}
//}
//
//func TestWriteResource_IdString(t *testing.T) {
//	r := WriteResource{
//		nil,
//		"abc",
//		"def",
//		"ghi",
//		nil,
//	}
//
//	id := r.IDString()
//	if id != "ghi-def-abc" {
//		t.Error("Unexpected IDString generated for WriteResource.")
//	}
//}

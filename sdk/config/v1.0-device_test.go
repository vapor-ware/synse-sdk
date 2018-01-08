package config

import (
	"testing"
)

func TestV1DeviceConfigHandlerProcessProtoConfig(t *testing.T) {
	data := `--~+::\n:-`

	handler := v1DeviceConfigHandler{}

	_, err := handler.processPrototypeConfig([]byte(data))
	if err == nil {
		t.Error("expected error: invalid YAML provided")
	}
}

func TestV1DeviceConfigHandlerProcessProtoConfig2(t *testing.T) {
	data := `version: 1.0`

	handler := v1DeviceConfigHandler{}

	_, err := handler.processPrototypeConfig([]byte(data))
	if err == nil {
		t.Error("expected error: no 'prototypes' field present in config")
	}
}

func TestV1DeviceConfigHandlerProcessProtoConfig3(t *testing.T) {
	data := `version: 1.0
prototypes:
  - type: airflow
    model: air8884
    manufacturer: vaporio
    protocol: emulator
    output:
      - type: airflow
        data_type: float
        unit:
          name: cubic feet per minute
          symbol: CFM
        precision: 2
        range:
          min: 0
          max: 1000
  - type: temperature
    model: temp2010
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

	handler := v1DeviceConfigHandler{}

	c, err := handler.processPrototypeConfig([]byte(data))
	if err != nil {
		t.Error(err)
	}

	if len(c) != 2 {
		t.Errorf("Expecting 2 prototypes, but got %v", len(c))
	}
}

func TestV1DeviceConfigHandlerProcessDeviceConfig(t *testing.T) {
	data := `--~+::\n:-`

	handler := v1DeviceConfigHandler{}

	_, err := handler.processDeviceConfig([]byte(data))
	if err == nil {
		t.Error("expected error: invalid YAML provided")
	}
}

func TestV1DeviceConfigHandlerProcessDeviceConfig2(t *testing.T) {
	data := `version: 1.0
locations:
  r1b1:
    rack: rack-1
    board: board-1
devices:
  - type: airflow
    model: air8884
    instances:
      - id: 1
        comment: first emulated airflow device`

	handler := v1DeviceConfigHandler{}

	_, err := handler.processDeviceConfig([]byte(data))
	if err == nil {
		t.Error("expected error: no 'location' field present in config")
	}
}

func TestV1DeviceConfigHandlerProcessDeviceConfig3(t *testing.T) {
	data := `version: 1.0
locations:
  r1b1:
    rack: rack-1
    board: board-1
  r1b2:
    rack: rack-1
    board: board-2
  r2b1:
    rack: rack-2
    board: board-1
devices:
  - type: airflow
    model: air8884
    instances:
      - id: 1
        location: r1b1
        comment: first emulated airflow device
      - id: 2
        location: r1b2
        comment: second emulated airflow device
      - id: 3
        location: r2b1
        comment: third emulated airflow device`

	handler := v1DeviceConfigHandler{}

	c, err := handler.processDeviceConfig([]byte(data))
	if err != nil {
		t.Error(err)
	}

	if len(c) != 3 {
		t.Errorf("Expecting 3 device configs, but got %v", len(c))
	}
}

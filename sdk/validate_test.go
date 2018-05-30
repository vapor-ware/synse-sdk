package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-server-grpc/go"
)

// TestValidateReadRequest tests validating a Read request successfully.
func TestValidateReadRequest(t *testing.T) {
	request := &synse.DeviceFilter{
		Device: "device",
		Board:  "board",
		Rack:   "rack",
	}

	err := validateReadRequest(request)
	assert.NoError(t, err)
}

// TestValidateReadRequestErr tests validating a Read request when the validation
// should fail and cause an error.
func TestValidateReadRequestErr(t *testing.T) {
	var cases = []synse.DeviceFilter{
		{
			// missing device
			Board: "board",
			Rack:  "rack",
		},
		{
			// missing board
			Device: "device",
			Rack:   "rack",
		},
		{
			// missing rack
			Device: "device",
			Board:  "board",
		},
		{
			// missing all
		},
	}

	for _, testCase := range cases {
		err := validateReadRequest(&testCase)
		assert.Error(t, err)
	}
}

// TestValidateWriteRequest tests validating a Write request successfully.
func TestValidateWriteRequest(t *testing.T) {
	request := &synse.WriteInfo{
		DeviceFilter: &synse.DeviceFilter{
			Device: "device",
			Board:  "board",
			Rack:   "rack",
		},
	}

	err := validateWriteRequest(request)
	assert.NoError(t, err)
}

// TestValidateWriteRequestErr tests validating a Write request when the validation
// should fail and cause an error.
func TestValidateWriteRequestErr(t *testing.T) {
	var cases = []synse.WriteInfo{
		{
			// missing device
			DeviceFilter: &synse.DeviceFilter{
				Board: "board",
				Rack:  "rack",
			},
		},
		{
			// missing board
			DeviceFilter: &synse.DeviceFilter{
				Device: "device",
				Rack:   "rack",
			},
		},
		{
			// missing rack
			DeviceFilter: &synse.DeviceFilter{
				Device: "device",
				Board:  "board",
			},
		},
		{
			// missing all
			DeviceFilter: &synse.DeviceFilter{},
		},
	}

	for _, testCase := range cases {
		err := validateWriteRequest(&testCase)
		assert.Error(t, err)
	}
}

// TestValidateForRead_1 tests validating a device for read, when the specified
// device is not in the device map.
func TestValidateForRead_1(t *testing.T) {
	deviceMap = make(map[string]*Device)

	err := validateForRead("foo")
	assert.Error(t, err)
}

// TestValidateForRead_2 tests validating a device for read, when no read handler
// is defined for the device.
func TestValidateForRead_2(t *testing.T) {
	deviceMap = make(map[string]*Device)
	deviceMap["abc"] = &Device{
		Handler: &DeviceHandler{},
	}

	err := validateForRead("abc")
	assert.Error(t, err)
}

// TestValidateForRead_3 tests validating a device for read, when it does exist
// in the device map and does have a read handler defined.
func TestValidateForRead_3(t *testing.T) {
	deviceMap = make(map[string]*Device)
	deviceMap["abc"] = &Device{
		Handler: &DeviceHandler{
			Read: func(d *Device) ([]*Reading, error) { return nil, nil },
		},
	}

	err := validateForRead("abc")
	assert.NoError(t, err)
}

// TestValidateForWrite_1 tests validating a device for write, when the specified
// device is not in the device map.
func TestValidateForWrite_1(t *testing.T) {
	deviceMap = make(map[string]*Device)

	err := validateForWrite("foo")
	assert.Error(t, err)
}

// TestValidateForWrite_2 tests validating a device for write, when no write handler
// is defined for the device.
func TestValidateForWrite_2(t *testing.T) {
	deviceMap = make(map[string]*Device)
	deviceMap["abc"] = &Device{
		Handler: &DeviceHandler{},
	}

	err := validateForWrite("abc")
	assert.Error(t, err)
}

// TestValidateForWrite_3 tests validating a device for write, when it does exist
// in the device map and does have a write handler defined.
func TestValidateForWrite_3(t *testing.T) {
	deviceMap = make(map[string]*Device)
	deviceMap["abc"] = &Device{
		Handler: &DeviceHandler{
			Write: func(d *Device, data *WriteData) error { return nil },
		},
	}

	err := validateForWrite("abc")
	assert.NoError(t, err)
}

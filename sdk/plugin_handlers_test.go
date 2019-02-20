package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testData = map[string]interface{}{
	"foo":  "bar",
	"baz":  1,
	"bool": true,
}

// Test_defaultDynamicDeviceRegistration tests the default dynamic device registration
// functionality.
func Test_defaultDynamicDeviceRegistration(t *testing.T) {
	devices, err := defaultDynamicDeviceRegistration(testData)
	assert.NoError(t, err)
	assert.Empty(t, devices)
}

// Test_defaultDynamicDeviceConfigRegistration tests the default dynamic device config
// functionality.
func Test_defaultDynamicDeviceConfigRegistration(t *testing.T) {
	cfgs, err := defaultDynamicDeviceConfigRegistration(testData)
	assert.NoError(t, err)
	assert.Empty(t, cfgs)
}

// Test_defaultDeviceDataValidator tests the default device data validator functionality.
func Test_defaultDeviceDataValidator(t *testing.T) {
	err := defaultDeviceDataValidator(testData)
	assert.NoError(t, err)
}

// Test_defaultDeviceIdentifier tests the default device identifier functionality.
func Test_defaultDeviceIdentifier(t *testing.T) {
	idComponent := defaultDeviceIdentifier(testData)
	assert.Equal(t, "1truebar", idComponent)
}

// Test_defaultDeviceIdentifier2 tests the default device identifier functionality with
// more complex data types.
func Test_defaultDeviceIdentifier2(t *testing.T) {
	data := map[string]interface{}{
		"foo":  "bar",
		"list": []string{"a", "b", "c"},
		"map": map[string]int{
			"foo": 1,
			"bar": 2,
			"abc": 3,
			"def": 4,
		},
		"a": 3.23,
		"z": false,
		"b": -4,
	}

	expected := "3.23-4bar[a b c]false" // map value should not make it in
	for i := 0; i < 20; i++ {
		idComponent := defaultDeviceIdentifier(data)
		assert.Equal(t, expected, idComponent)
	}
}

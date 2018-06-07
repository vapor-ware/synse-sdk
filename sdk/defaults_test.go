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

// Test_defaultDeviceIdentifier tests the default device identifier functionality.
func Test_defaultDeviceIdentifier(t *testing.T) {
	idComponent := defaultDeviceIdentifier(testData)
	assert.Equal(t, "1truebar", idComponent)
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

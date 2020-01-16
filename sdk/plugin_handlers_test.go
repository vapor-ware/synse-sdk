// Synse SDK
// Copyright (c) 2019-2020 Vapor IO
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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

func TestNewDefaultPluginHandlers(t *testing.T) {
	handlers := NewDefaultPluginHandlers()
	assert.NotNil(t, handlers.DeviceIdentifier)
	assert.NotNil(t, handlers.DynamicRegistrar)
	assert.NotNil(t, handlers.DynamicConfigRegistrar)
	assert.NotNil(t, handlers.DeviceDataValidator)
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

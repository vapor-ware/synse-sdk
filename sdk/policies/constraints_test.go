package policies

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test_constraintPluginConfigNecessity tests the PluginConfigNecessity constraint.
func Test_constraintPluginConfigNecessity(t *testing.T) {
	var testTable = []struct {
		desc     string
		policies []ConfigPolicy
		hasErr   bool
	}{
		{
			desc:     "no policies - should not fail",
			policies: []ConfigPolicy{},
			hasErr:   false,
		},
		{
			desc:     "no PluginConfig policies - should not fail",
			policies: []ConfigPolicy{DeviceConfigOptional, DeviceConfigRequired},
			hasErr:   false,
		},
		{
			desc:     "one PluginConfig policy - should not fail",
			policies: []ConfigPolicy{PluginConfigOptional},
			hasErr:   false,
		},
		{
			desc:     "conflicting PluginConfig policies - should fail",
			policies: []ConfigPolicy{PluginConfigOptional, PluginConfigRequired},
			hasErr:   true,
		},
	}

	for _, testCase := range testTable {
		err := constraintPluginConfigNecessity(testCase.policies)
		if testCase.hasErr {
			assert.Error(t, err, testCase.desc)
		} else {
			assert.NoError(t, err, testCase.desc)
		}
	}
}

// Test_constraintDeviceConfigNecessity tests the DeviceConfigNecessity constraint.
func Test_constraintDeviceConfigNecessity(t *testing.T) {
	var testTable = []struct {
		desc     string
		policies []ConfigPolicy
		hasErr   bool
	}{
		{
			desc:     "no policies - should not fail",
			policies: []ConfigPolicy{},
			hasErr:   false,
		},
		{
			desc:     "no DeviceConfig policies - should not fail",
			policies: []ConfigPolicy{PluginConfigRequired, PluginConfigOptional},
			hasErr:   false,
		},
		{
			desc:     "one DeviceConfig policy - should not fail",
			policies: []ConfigPolicy{DeviceConfigRequired},
			hasErr:   false,
		},
		{
			desc:     "conflicting DeviceConfig policies - should fail",
			policies: []ConfigPolicy{DeviceConfigRequired, DeviceConfigOptional},
			hasErr:   true,
		},
	}

	for _, testCase := range testTable {
		err := constraintDeviceConfigNecessity(testCase.policies)
		if testCase.hasErr {
			assert.Error(t, err, testCase.desc)
		} else {
			assert.NoError(t, err, testCase.desc)
		}
	}
}

// TestCheckConstraints tests checking constraints against lists of ConfigPolicies
func TestCheckConstraints(t *testing.T) {
	var testTable = []struct {
		desc     string
		policies []ConfigPolicy
		errCount int
	}{
		{
			desc:     "no policies - should not fail",
			policies: []ConfigPolicy{},
			errCount: 0,
		},
		{
			desc:     "one DeviceConfig policy - should not fail",
			policies: []ConfigPolicy{DeviceConfigOptional},
			errCount: 0,
		},
		{
			desc:     "one PluginConfig policy - should not fail",
			policies: []ConfigPolicy{PluginConfigRequired},
			errCount: 0,
		},
		{
			desc:     "one DeviceConfig and PluginConfig policy - should not fail",
			policies: []ConfigPolicy{DeviceConfigRequired, PluginConfigOptional},
			errCount: 0,
		},
		{
			desc:     "conflicting DeviceConfig policies - should fail",
			policies: []ConfigPolicy{DeviceConfigRequired, DeviceConfigOptional, PluginConfigOptional},
			errCount: 1,
		},
		{
			desc:     "conflicting PluginConfig policies - should fail",
			policies: []ConfigPolicy{PluginConfigOptional, PluginConfigRequired, DeviceConfigOptional},
			errCount: 1,
		},
		{
			desc:     "conflicting DeviceConfig and PluginConfig policies - should fail",
			policies: []ConfigPolicy{DeviceConfigRequired, DeviceConfigOptional, PluginConfigOptional, PluginConfigRequired},
			errCount: 2,
		},
	}

	for _, testCase := range testTable {
		multiErr := CheckConstraints(testCase.policies)
		assert.Equal(t, testCase.errCount, len(multiErr.Errors), testCase.desc)
	}
}

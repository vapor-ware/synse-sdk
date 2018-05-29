package policies

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConfigPolicy_String tests getting the string name for ConfigPolicy instances.
func TestConfigPolicy_String(t *testing.T) {
	var testTable = []struct {
		desc     string
		policy   ConfigPolicy
		expected string
	}{
		{
			desc:     "String for PluginConfigRequired",
			policy:   PluginConfigRequired,
			expected: "PluginConfigRequired",
		},
		{
			desc:     "String for PluginConfigOptional",
			policy:   PluginConfigOptional,
			expected: "PluginConfigOptional",
		},
		{
			desc:     "String for DeviceConfigRequired",
			policy:   DeviceConfigRequired,
			expected: "DeviceConfigRequired",
		},
		{
			desc:     "String for DeviceConfigOptional",
			policy:   DeviceConfigOptional,
			expected: "DeviceConfigOptional",
		},
		{
			desc:     "String for custom policy",
			policy:   ConfigPolicy(8),
			expected: "unknown",
		},
	}

	for _, testCase := range testTable {
		actual := testCase.policy.String()
		assert.Equal(t, testCase.expected, actual, testCase.desc)
	}
}

// TestPolicyManager_GetDeviceConfigPolicy tests getting the device config
// policy from the global PolicyManager.
func TestPolicyManager_GetDeviceConfigPolicy(t *testing.T) {

	// Get the device config policy when none is set - this should give the
	// default.
	assert.Empty(t, PolicyManager.deviceConfigPolicy)
	policy := PolicyManager.GetDeviceConfigPolicy()
	assert.Equal(t, DeviceConfigRequired, policy)

	// Get the device config policy when Optional is set.
	PolicyManager.deviceConfigPolicy = DeviceConfigOptional
	policy = PolicyManager.GetDeviceConfigPolicy()
	assert.Equal(t, DeviceConfigOptional, policy)

	// Get the device config policy when Required is set.
	PolicyManager.deviceConfigPolicy = DeviceConfigRequired
	policy = PolicyManager.GetDeviceConfigPolicy()
	assert.Equal(t, DeviceConfigRequired, policy)
}

// TestPolicyManager_GetPluginConfigPolicy tests getting the plugin config
// policy from the global PolicyManager.
func TestPolicyManager_GetPluginConfigPolicy(t *testing.T) {

	// Get the plugin config policy when none is set - this should give the
	// default.
	assert.Empty(t, PolicyManager.pluginConfigPolicy)
	policy := PolicyManager.GetPluginConfigPolicy()
	assert.Equal(t, PluginConfigOptional, policy)

	// Get the plugin config policy when Optional is set.
	PolicyManager.pluginConfigPolicy = PluginConfigOptional
	policy = PolicyManager.GetPluginConfigPolicy()
	assert.Equal(t, PluginConfigOptional, policy)

	// Get the plugin config policy when Required is set.
	PolicyManager.pluginConfigPolicy = PluginConfigRequired
	policy = PolicyManager.GetPluginConfigPolicy()
	assert.Equal(t, PluginConfigRequired, policy)
}

// TestSet tests setting the global config policies.
func TestSet(t *testing.T) {
	var testTable = []struct{
		desc string
		policies []ConfigPolicy
		expectedPluginPolicy ConfigPolicy
		expectedDevicePolicy ConfigPolicy
	}{
		{
			desc: "No policies set",
			policies: []ConfigPolicy{},
			expectedPluginPolicy: 0,
			expectedDevicePolicy: 0,
		},
		{
			desc: "One device config policy set",
			policies: []ConfigPolicy{
				DeviceConfigOptional,
			},
			expectedPluginPolicy: 0,
			expectedDevicePolicy: DeviceConfigOptional,
		},
		{
			desc: "One plugin config policy set",
			policies: []ConfigPolicy{
				PluginConfigOptional,
			},
			expectedPluginPolicy: PluginConfigOptional,
			expectedDevicePolicy: 0,
		},
		{
			desc: "Two device config policies set",
			policies: []ConfigPolicy{
				DeviceConfigRequired,
				DeviceConfigOptional,
			},
			expectedPluginPolicy: 0,
			expectedDevicePolicy: DeviceConfigOptional,
		},
		{
			desc: "Two plugin config policies set",
			policies: []ConfigPolicy{
				PluginConfigRequired,
				PluginConfigOptional,
			},
			expectedPluginPolicy: PluginConfigOptional,
			expectedDevicePolicy: 0,
		},
		{
			desc: "One of each policy",
			policies: []ConfigPolicy{
				PluginConfigRequired,
				DeviceConfigOptional,
			},
			expectedPluginPolicy: PluginConfigRequired,
			expectedDevicePolicy: DeviceConfigOptional,
		},
		{
			desc: "Two of each policy",
			policies: []ConfigPolicy{
				DeviceConfigRequired,
				DeviceConfigOptional,
				PluginConfigRequired,
				PluginConfigOptional,
			},
			expectedPluginPolicy: PluginConfigOptional,
			expectedDevicePolicy: DeviceConfigOptional,
		},
	}

	for _, testCase := range testTable {
		// reset the policy manager every time
		PolicyManager = policyManager{}

		Set(testCase.policies)
		assert.Equal(t, testCase.expectedDevicePolicy, PolicyManager.deviceConfigPolicy, testCase.desc)
		assert.Equal(t, testCase.expectedPluginPolicy, PolicyManager.pluginConfigPolicy, testCase.desc)
	}
}

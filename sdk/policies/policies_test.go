package policies

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// resetPolicyManager is a helper that should be run after every test
// to reset test state.
func resetPolicyManager() {
	policyManager = manager{}
}

// TestConfigPolicy_String tests getting the string name for ConfigPolicy instances.
func TestConfigPolicy_String(t *testing.T) {
	defer resetPolicyManager()

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

// TestGetDeviceConfigPolicy tests getting the device config
// policy from the global policy manager.
func TestGetDeviceConfigPolicy(t *testing.T) {
	defer resetPolicyManager()

	// Get the device config policy when none is set - this should give the default.
	assert.Empty(t, policyManager.deviceConfigPolicy)
	policy := GetDeviceConfigPolicy()
	assert.Equal(t, DeviceConfigRequired, policy)

	// Get the device config policy when Optional is set.
	policyManager.deviceConfigPolicy = DeviceConfigOptional
	policy = GetDeviceConfigPolicy()
	assert.Equal(t, DeviceConfigOptional, policy)

	// Get the device config policy when Required is set.
	policyManager.deviceConfigPolicy = DeviceConfigRequired
	policy = GetDeviceConfigPolicy()
	assert.Equal(t, DeviceConfigRequired, policy)
}

// TestGetPluginConfigPolicy tests getting the plugin config
// policy from the global policy manager.
func TestGetPluginConfigPolicy(t *testing.T) {
	defer resetPolicyManager()

	// Get the plugin config policy when none is set - this should give the default.
	assert.Empty(t, policyManager.pluginConfigPolicy)
	policy := GetPluginConfigPolicy()
	assert.Equal(t, PluginConfigOptional, policy)

	// Get the plugin config policy when Optional is set.
	policyManager.pluginConfigPolicy = PluginConfigOptional
	policy = GetPluginConfigPolicy()
	assert.Equal(t, PluginConfigOptional, policy)

	// Get the plugin config policy when Required is set.
	policyManager.pluginConfigPolicy = PluginConfigRequired
	policy = GetPluginConfigPolicy()
	assert.Equal(t, PluginConfigRequired, policy)
}

// TestManager_Set tests setting the global config policies.
func TestManager_Set(t *testing.T) {
	defer resetPolicyManager()

	var testTable = []struct {
		desc                 string
		policies             []ConfigPolicy
		expectedPluginPolicy ConfigPolicy
		expectedDevicePolicy ConfigPolicy
	}{
		{
			desc:                 "No policies set",
			policies:             []ConfigPolicy{},
			expectedPluginPolicy: NoPolicy,
			expectedDevicePolicy: NoPolicy,
		},
		{
			desc: "One device config policy set",
			policies: []ConfigPolicy{
				DeviceConfigOptional,
			},
			expectedPluginPolicy: NoPolicy,
			expectedDevicePolicy: DeviceConfigOptional,
		},
		{
			desc: "One plugin config policy set",
			policies: []ConfigPolicy{
				PluginConfigOptional,
			},
			expectedPluginPolicy: PluginConfigOptional,
			expectedDevicePolicy: NoPolicy,
		},
		{
			desc: "Two device config policies set",
			policies: []ConfigPolicy{
				DeviceConfigRequired,
				DeviceConfigOptional,
			},
			expectedPluginPolicy: NoPolicy,
			expectedDevicePolicy: DeviceConfigOptional,
		},
		{
			desc: "Two plugin config policies set",
			policies: []ConfigPolicy{
				PluginConfigRequired,
				PluginConfigOptional,
			},
			expectedPluginPolicy: PluginConfigOptional,
			expectedDevicePolicy: NoPolicy,
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
		policyManager = manager{}

		policyManager.Set(testCase.policies)
		assert.Equal(t, testCase.expectedDevicePolicy, policyManager.deviceConfigPolicy, testCase.desc)
		assert.Equal(t, testCase.expectedPluginPolicy, policyManager.pluginConfigPolicy, testCase.desc)
	}
}

// TestApplyOk tests applying policies to the manager with no error.
func TestApplyOk(t *testing.T) {
	defer resetPolicyManager()

	policies := []ConfigPolicy{DeviceConfigOptional, PluginConfigRequired}

	assert.Equal(t, NoPolicy, policyManager.pluginConfigPolicy)
	assert.Equal(t, NoPolicy, policyManager.deviceConfigPolicy)

	err := Apply(policies)
	assert.NoError(t, err)

	assert.Equal(t, PluginConfigRequired, policyManager.pluginConfigPolicy)
	assert.Equal(t, DeviceConfigOptional, policyManager.deviceConfigPolicy)
}

// TestApplyOk2 tests applying policies to the manager with no error, when
// no policies are specified.
func TestApplyOk2(t *testing.T) {
	defer resetPolicyManager()

	assert.Equal(t, NoPolicy, policyManager.pluginConfigPolicy)
	assert.Equal(t, NoPolicy, policyManager.deviceConfigPolicy)

	err := Apply([]ConfigPolicy{})
	assert.NoError(t, err)

	assert.Equal(t, NoPolicy, policyManager.pluginConfigPolicy)
	assert.Equal(t, NoPolicy, policyManager.deviceConfigPolicy)
}

// TestApplyError tests applying policies to the manager with error.
func TestApplyError(t *testing.T) {
	defer resetPolicyManager()

	policies := []ConfigPolicy{DeviceConfigOptional, PluginConfigRequired, DeviceConfigRequired}

	assert.Equal(t, NoPolicy, policyManager.pluginConfigPolicy)
	assert.Equal(t, NoPolicy, policyManager.deviceConfigPolicy)

	err := Apply(policies)
	assert.Error(t, err)

	assert.Equal(t, NoPolicy, policyManager.pluginConfigPolicy)
	assert.Equal(t, NoPolicy, policyManager.deviceConfigPolicy)
}
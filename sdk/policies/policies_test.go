package policies

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// resetPolicyManager is a helper that should be run after every test
// to reset test state.
func resetPolicyManager() {
	defaultManager = manager{}
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
	assert.Empty(t, defaultManager.deviceConfigPolicy)
	policy := GetDeviceConfigPolicy()
	assert.Equal(t, DeviceConfigRequired, policy)

	// Get the device config policy when Optional is set.
	defaultManager.deviceConfigPolicy = DeviceConfigOptional
	policy = GetDeviceConfigPolicy()
	assert.Equal(t, DeviceConfigOptional, policy)

	// Get the device config policy when Required is set.
	defaultManager.deviceConfigPolicy = DeviceConfigRequired
	policy = GetDeviceConfigPolicy()
	assert.Equal(t, DeviceConfigRequired, policy)

	// Reset the device config policy and add the policy to the
	// tracked policies. It should now find it from there.
	defaultManager.deviceConfigPolicy = NoPolicy
	defaultManager.policies = []ConfigPolicy{DeviceConfigOptional}
	policy = GetDeviceConfigPolicy()
	assert.Equal(t, DeviceConfigOptional, policy)
}

// TestGetPluginConfigPolicy tests getting the plugin config
// policy from the global policy manager.
func TestGetPluginConfigPolicy(t *testing.T) {
	defer resetPolicyManager()

	// Get the plugin config policy when none is set - this should give the default.
	assert.Empty(t, defaultManager.pluginConfigPolicy)
	policy := GetPluginConfigPolicy()
	assert.Equal(t, PluginConfigOptional, policy)

	// Get the plugin config policy when Optional is set.
	defaultManager.pluginConfigPolicy = PluginConfigOptional
	policy = GetPluginConfigPolicy()
	assert.Equal(t, PluginConfigOptional, policy)

	// Get the plugin config policy when Required is set.
	defaultManager.pluginConfigPolicy = PluginConfigRequired
	policy = GetPluginConfigPolicy()
	assert.Equal(t, PluginConfigRequired, policy)

	// Reset the plugin config policy and add the policy to the
	// tracked policies. It should now find it from there.
	defaultManager.pluginConfigPolicy = NoPolicy
	defaultManager.policies = []ConfigPolicy{PluginConfigOptional}
	policy = GetPluginConfigPolicy()
	assert.Equal(t, PluginConfigOptional, policy)
}

// TestSet tests adding multiple policies to the manager.
func TestSet(t *testing.T) {
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
		// reset manager state
		defaultManager = manager{}

		Set(testCase.policies)
		assert.Equal(t, len(testCase.policies), len(defaultManager.policies), testCase.desc)
		// Setting should not change the internal state for the plugin/device policies
		assert.Equal(t, NoPolicy, defaultManager.deviceConfigPolicy, testCase.desc)
		assert.Equal(t, NoPolicy, defaultManager.pluginConfigPolicy, testCase.desc)
	}
}

// TestAdd tests adding policies to the manager.
func TestAdd(t *testing.T) {
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
		// reset manager state
		defaultManager = manager{}

		for _, p := range testCase.policies {
			Add(p)
		}

		assert.Equal(t, len(testCase.policies), len(defaultManager.policies), testCase.desc)
		// Setting should not change the internal state for the plugin/device policies
		assert.Equal(t, NoPolicy, defaultManager.deviceConfigPolicy, testCase.desc)
		assert.Equal(t, NoPolicy, defaultManager.pluginConfigPolicy, testCase.desc)
	}
}

// TestCheckOk tests checking policies with no error.
func TestCheckOk(t *testing.T) {
	defer resetPolicyManager()

	policies := []ConfigPolicy{DeviceConfigOptional, PluginConfigRequired}

	assert.Equal(t, NoPolicy, defaultManager.pluginConfigPolicy)
	assert.Equal(t, NoPolicy, defaultManager.deviceConfigPolicy)

	defaultManager.Set(policies)
	err := Check()
	assert.NoError(t, err)

	// State should not change on check
	assert.Equal(t, NoPolicy, defaultManager.pluginConfigPolicy)
	assert.Equal(t, NoPolicy, defaultManager.deviceConfigPolicy)
}

// TestCheckOk2 tests checking policies with no error, when no policies are specified.
func TestCheckOk2(t *testing.T) {
	defer resetPolicyManager()

	assert.Equal(t, NoPolicy, defaultManager.pluginConfigPolicy)
	assert.Equal(t, NoPolicy, defaultManager.deviceConfigPolicy)

	err := Check()
	assert.NoError(t, err)

	// State should not change on check
	assert.Equal(t, NoPolicy, defaultManager.pluginConfigPolicy)
	assert.Equal(t, NoPolicy, defaultManager.deviceConfigPolicy)
}

// TestCheckError tests checking policies resulting in error.
func TestCheckError(t *testing.T) {
	defer resetPolicyManager()

	policies := []ConfigPolicy{DeviceConfigOptional, PluginConfigRequired, DeviceConfigRequired}

	assert.Equal(t, NoPolicy, defaultManager.pluginConfigPolicy)
	assert.Equal(t, NoPolicy, defaultManager.deviceConfigPolicy)

	defaultManager.Set(policies)
	err := Check()
	assert.Error(t, err)

	// State should not change on check
	assert.Equal(t, NoPolicy, defaultManager.pluginConfigPolicy)
	assert.Equal(t, NoPolicy, defaultManager.deviceConfigPolicy)
}

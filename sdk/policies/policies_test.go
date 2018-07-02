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
			desc:     "String for PluginConfigFileRequired",
			policy:   PluginConfigFileRequired,
			expected: "PluginConfigFileRequired",
		},
		{
			desc:     "String for PluginConfigFileOptional",
			policy:   PluginConfigFileOptional,
			expected: "PluginConfigFileOptional",
		},
		{
			desc:     "String for PluginConfigFileProhibited",
			policy:   PluginConfigFileProhibited,
			expected: "PluginConfigFileProhibited",
		},
		{
			desc:     "String for DeviceConfigFileRequired",
			policy:   DeviceConfigFileRequired,
			expected: "DeviceConfigFileRequired",
		},
		{
			desc:     "String for DeviceConfigFileOptional",
			policy:   DeviceConfigFileOptional,
			expected: "DeviceConfigFileOptional",
		},
		{
			desc:     "String for DeviceConfigFileProhibited",
			policy:   DeviceConfigFileProhibited,
			expected: "DeviceConfigFileProhibited",
		},
		{
			desc:     "String for DeviceConfigDynamicOptional",
			policy:   DeviceConfigDynamicOptional,
			expected: "DeviceConfigDynamicOptional",
		},
		{
			desc:     "String for DeviceConfigDynamicRequired",
			policy:   DeviceConfigDynamicRequired,
			expected: "DeviceConfigDynamicRequired",
		},
		{
			desc:     "String for DeviceConfigDynamicProhibited",
			policy:   DeviceConfigDynamicProhibited,
			expected: "DeviceConfigDynamicProhibited",
		},
		{
			desc:     "String for TypeConfigFileOptional",
			policy:   TypeConfigFileOptional,
			expected: "TypeConfigFileOptional",
		},
		{
			desc:     "String for TypeConfigFileRequired",
			policy:   TypeConfigFileRequired,
			expected: "TypeConfigFileRequired",
		},
		{
			desc:     "String for TypeConfigFileProhibited",
			policy:   TypeConfigFileProhibited,
			expected: "TypeConfigFileProhibited",
		},
		{
			desc:     "String for custom policy",
			policy:   ConfigPolicy(17),
			expected: "unknown",
		},
	}

	for _, testCase := range testTable {
		actual := testCase.policy.String()
		assert.Equal(t, testCase.expected, actual, testCase.desc)
	}
}

// TestGetDeviceConfigFilePolicy tests getting the device config
// policy from the global policy manager.
func TestGetDeviceConfigFilePolicy(t *testing.T) {
	defer resetPolicyManager()

	// Get the device config file policy when none is set - this should give the default.
	assert.Empty(t, defaultManager.deviceConfigFilePolicy)
	policy := GetDeviceConfigFilePolicy()
	assert.Equal(t, DeviceConfigFileRequired, policy)

	// Get the device config file policy when Optional is set.
	defaultManager.deviceConfigFilePolicy = DeviceConfigFileOptional
	policy = GetDeviceConfigFilePolicy()
	assert.Equal(t, DeviceConfigFileOptional, policy)

	// Get the device config file policy when Required is set.
	defaultManager.deviceConfigFilePolicy = DeviceConfigFileRequired
	policy = GetDeviceConfigFilePolicy()
	assert.Equal(t, DeviceConfigFileRequired, policy)

	// Get the device config file policy when Prohibited is set.
	defaultManager.deviceConfigFilePolicy = DeviceConfigFileProhibited
	policy = GetDeviceConfigFilePolicy()
	assert.Equal(t, DeviceConfigFileProhibited, policy)

	// Reset the device config file policy and add the policy to the
	// tracked policies. It should now find it from there.
	defaultManager.deviceConfigFilePolicy = NoPolicy
	defaultManager.policies = []ConfigPolicy{DeviceConfigFileOptional}
	policy = GetDeviceConfigFilePolicy()
	assert.Equal(t, DeviceConfigFileOptional, policy)
}

// TestGetPluginConfigFilePolicy tests getting the plugin config
// policy from the global policy manager.
func TestGetPluginConfigFilePolicy(t *testing.T) {
	defer resetPolicyManager()

	// Get the plugin config file policy when none is set - this should give the default.
	assert.Empty(t, defaultManager.pluginConfigFilePolicy)
	policy := GetPluginConfigFilePolicy()
	assert.Equal(t, PluginConfigFileOptional, policy)

	// Get the plugin config file policy when Optional is set.
	defaultManager.pluginConfigFilePolicy = PluginConfigFileOptional
	policy = GetPluginConfigFilePolicy()
	assert.Equal(t, PluginConfigFileOptional, policy)

	// Get the plugin config file policy when Required is set.
	defaultManager.pluginConfigFilePolicy = PluginConfigFileRequired
	policy = GetPluginConfigFilePolicy()
	assert.Equal(t, PluginConfigFileRequired, policy)

	// Get the plugin config file policy when Prohibited is set.
	defaultManager.pluginConfigFilePolicy = PluginConfigFileProhibited
	policy = GetPluginConfigFilePolicy()
	assert.Equal(t, PluginConfigFileProhibited, policy)

	// Reset the plugin config file policy and add the policy to the
	// tracked policies. It should now find it from there.
	defaultManager.pluginConfigFilePolicy = NoPolicy
	defaultManager.policies = []ConfigPolicy{PluginConfigFileOptional}
	policy = GetPluginConfigFilePolicy()
	assert.Equal(t, PluginConfigFileOptional, policy)
}

// TestGetDeviceConfigDynamicPolicy tests getting the dynamic device config
// policy from the global policy manager.
func TestGetDeviceConfigDynamicPolicy(t *testing.T) {
	defer resetPolicyManager()

	// Get the dynamic device config policy when none is set - this should give the default.
	assert.Empty(t, defaultManager.deviceConfigDynamicPolicy)
	policy := GetDeviceConfigDynamicPolicy()
	assert.Equal(t, DeviceConfigDynamicOptional, policy)

	// Get the dynamic device config policy when Optional is set.
	defaultManager.deviceConfigDynamicPolicy = DeviceConfigDynamicOptional
	policy = GetDeviceConfigDynamicPolicy()
	assert.Equal(t, DeviceConfigDynamicOptional, policy)

	// Get the dynamic device config policy when Required is set.
	defaultManager.deviceConfigDynamicPolicy = DeviceConfigDynamicRequired
	policy = GetDeviceConfigDynamicPolicy()
	assert.Equal(t, DeviceConfigDynamicRequired, policy)

	// Get the dynamic device config policy when Prohibited is set.
	defaultManager.deviceConfigDynamicPolicy = DeviceConfigDynamicProhibited
	policy = GetDeviceConfigDynamicPolicy()
	assert.Equal(t, DeviceConfigDynamicProhibited, policy)

	// Reset the dynamic device config policy and add the policy to the
	// tracked policies. It should now find it from there.
	defaultManager.deviceConfigDynamicPolicy = NoPolicy
	defaultManager.policies = []ConfigPolicy{DeviceConfigDynamicOptional}
	policy = GetDeviceConfigDynamicPolicy()
	assert.Equal(t, DeviceConfigDynamicOptional, policy)
}

// TestGetTypeConfigFilePolicy tests getting the output type config
// policy from the global policy manager.
func TestGetTypeConfigFilePolicy(t *testing.T) {
	defer resetPolicyManager()

	// Get the output type config file policy when none is set - this should give the default.
	assert.Empty(t, defaultManager.typeConfigFilePolicy)
	policy := GetTypeConfigFilePolicy()
	assert.Equal(t, TypeConfigFileOptional, policy)

	// Get the output type config file policy when Optional is set.
	defaultManager.typeConfigFilePolicy = TypeConfigFileOptional
	policy = GetTypeConfigFilePolicy()
	assert.Equal(t, TypeConfigFileOptional, policy)

	// Get the output type config file policy when Required is set.
	defaultManager.typeConfigFilePolicy = TypeConfigFileRequired
	policy = GetTypeConfigFilePolicy()
	assert.Equal(t, TypeConfigFileRequired, policy)

	// Get the output type config file policy when Prohibited is set.
	defaultManager.typeConfigFilePolicy = TypeConfigFileProhibited
	policy = GetTypeConfigFilePolicy()
	assert.Equal(t, TypeConfigFileProhibited, policy)

	// Reset the output type config file policy and add the policy to the
	// tracked policies. It should now find it from there.
	defaultManager.typeConfigFilePolicy = NoPolicy
	defaultManager.policies = []ConfigPolicy{TypeConfigFileOptional}
	policy = GetTypeConfigFilePolicy()
	assert.Equal(t, TypeConfigFileOptional, policy)
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
				DeviceConfigFileOptional,
			},
			expectedPluginPolicy: NoPolicy,
			expectedDevicePolicy: DeviceConfigFileOptional,
		},
		{
			desc: "One plugin config policy set",
			policies: []ConfigPolicy{
				PluginConfigFileOptional,
			},
			expectedPluginPolicy: PluginConfigFileOptional,
			expectedDevicePolicy: NoPolicy,
		},
		{
			desc: "Two device config policies set",
			policies: []ConfigPolicy{
				DeviceConfigFileRequired,
				DeviceConfigFileOptional,
			},
			expectedPluginPolicy: NoPolicy,
			expectedDevicePolicy: DeviceConfigFileOptional,
		},
		{
			desc: "Two plugin config policies set",
			policies: []ConfigPolicy{
				PluginConfigFileRequired,
				PluginConfigFileOptional,
			},
			expectedPluginPolicy: PluginConfigFileOptional,
			expectedDevicePolicy: NoPolicy,
		},
		{
			desc: "One of each policy",
			policies: []ConfigPolicy{
				PluginConfigFileRequired,
				DeviceConfigFileOptional,
			},
			expectedPluginPolicy: PluginConfigFileRequired,
			expectedDevicePolicy: DeviceConfigFileOptional,
		},
		{
			desc: "Two of each policy",
			policies: []ConfigPolicy{
				DeviceConfigFileRequired,
				DeviceConfigFileOptional,
				PluginConfigFileRequired,
				PluginConfigFileOptional,
			},
			expectedPluginPolicy: PluginConfigFileOptional,
			expectedDevicePolicy: DeviceConfigFileOptional,
		},
	}

	for _, testCase := range testTable {
		// reset manager state
		defaultManager = manager{}

		Set(testCase.policies)
		assert.Equal(t, len(testCase.policies), len(defaultManager.policies), testCase.desc)
		// Setting should not change the internal state for the plugin/device policies
		assert.Equal(t, NoPolicy, defaultManager.deviceConfigFilePolicy, testCase.desc)
		assert.Equal(t, NoPolicy, defaultManager.pluginConfigFilePolicy, testCase.desc)
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
				DeviceConfigFileOptional,
			},
			expectedPluginPolicy: NoPolicy,
			expectedDevicePolicy: DeviceConfigFileOptional,
		},
		{
			desc: "One plugin config policy set",
			policies: []ConfigPolicy{
				PluginConfigFileOptional,
			},
			expectedPluginPolicy: PluginConfigFileOptional,
			expectedDevicePolicy: NoPolicy,
		},
		{
			desc: "Two device config policies set",
			policies: []ConfigPolicy{
				DeviceConfigFileRequired,
				DeviceConfigFileOptional,
			},
			expectedPluginPolicy: NoPolicy,
			expectedDevicePolicy: DeviceConfigFileOptional,
		},
		{
			desc: "Two plugin config policies set",
			policies: []ConfigPolicy{
				PluginConfigFileRequired,
				PluginConfigFileOptional,
			},
			expectedPluginPolicy: PluginConfigFileOptional,
			expectedDevicePolicy: NoPolicy,
		},
		{
			desc: "One of each policy",
			policies: []ConfigPolicy{
				PluginConfigFileRequired,
				DeviceConfigFileOptional,
			},
			expectedPluginPolicy: PluginConfigFileRequired,
			expectedDevicePolicy: DeviceConfigFileOptional,
		},
		{
			desc: "Two of each policy",
			policies: []ConfigPolicy{
				DeviceConfigFileRequired,
				DeviceConfigFileOptional,
				PluginConfigFileRequired,
				PluginConfigFileOptional,
			},
			expectedPluginPolicy: PluginConfigFileOptional,
			expectedDevicePolicy: DeviceConfigFileOptional,
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
		assert.Equal(t, NoPolicy, defaultManager.deviceConfigFilePolicy, testCase.desc)
		assert.Equal(t, NoPolicy, defaultManager.pluginConfigFilePolicy, testCase.desc)
	}
}

// TestCheckOk tests checking policies with no error.
func TestCheckOk(t *testing.T) {
	defer resetPolicyManager()

	policies := []ConfigPolicy{
		DeviceConfigFileOptional,
		PluginConfigFileRequired,
		TypeConfigFileProhibited,
		DeviceConfigDynamicOptional,
	}

	assert.Equal(t, NoPolicy, defaultManager.pluginConfigFilePolicy)
	assert.Equal(t, NoPolicy, defaultManager.deviceConfigFilePolicy)
	assert.Equal(t, NoPolicy, defaultManager.deviceConfigDynamicPolicy)
	assert.Equal(t, NoPolicy, defaultManager.typeConfigFilePolicy)

	defaultManager.Set(policies)
	err := Check()
	assert.NoError(t, err)

	// State should not change on check
	assert.Equal(t, NoPolicy, defaultManager.pluginConfigFilePolicy)
	assert.Equal(t, NoPolicy, defaultManager.deviceConfigFilePolicy)
	assert.Equal(t, NoPolicy, defaultManager.deviceConfigDynamicPolicy)
	assert.Equal(t, NoPolicy, defaultManager.typeConfigFilePolicy)
}

// TestCheckOk2 tests checking policies with no error, when no policies are specified.
func TestCheckOk2(t *testing.T) {
	defer resetPolicyManager()

	assert.Equal(t, NoPolicy, defaultManager.pluginConfigFilePolicy)
	assert.Equal(t, NoPolicy, defaultManager.deviceConfigFilePolicy)
	assert.Equal(t, NoPolicy, defaultManager.deviceConfigDynamicPolicy)
	assert.Equal(t, NoPolicy, defaultManager.typeConfigFilePolicy)

	err := Check()
	assert.NoError(t, err)

	// State should not change on check
	assert.Equal(t, NoPolicy, defaultManager.pluginConfigFilePolicy)
	assert.Equal(t, NoPolicy, defaultManager.deviceConfigFilePolicy)
	assert.Equal(t, NoPolicy, defaultManager.deviceConfigDynamicPolicy)
	assert.Equal(t, NoPolicy, defaultManager.typeConfigFilePolicy)
}

// TestCheckError tests checking policies resulting in error.
func TestCheckError(t *testing.T) {
	defer resetPolicyManager()

	policies := []ConfigPolicy{
		DeviceConfigFileOptional,
		PluginConfigFileRequired,
		DeviceConfigFileRequired,
		DeviceConfigDynamicOptional,
	}

	assert.Equal(t, NoPolicy, defaultManager.pluginConfigFilePolicy)
	assert.Equal(t, NoPolicy, defaultManager.deviceConfigFilePolicy)
	assert.Equal(t, NoPolicy, defaultManager.deviceConfigDynamicPolicy)
	assert.Equal(t, NoPolicy, defaultManager.typeConfigFilePolicy)

	defaultManager.Set(policies)
	err := Check()
	assert.Error(t, err)

	// State should not change on check
	assert.Equal(t, NoPolicy, defaultManager.pluginConfigFilePolicy)
	assert.Equal(t, NoPolicy, defaultManager.deviceConfigFilePolicy)
	assert.Equal(t, NoPolicy, defaultManager.deviceConfigDynamicPolicy)
	assert.Equal(t, NoPolicy, defaultManager.typeConfigFilePolicy)
}

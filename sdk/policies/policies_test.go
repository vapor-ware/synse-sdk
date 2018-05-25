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

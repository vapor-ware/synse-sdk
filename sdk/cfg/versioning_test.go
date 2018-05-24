package cfg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewSchemeVersion_Ok tests creating a new SchemeVersion with no errors.
func TestNewSchemeVersion_Ok(t *testing.T) {
	var testTable = []struct {
		desc     string
		in       string
		expected SchemeVersion
	}{
		{
			desc:     "Version with only major specified",
			in:       "1",
			expected: SchemeVersion{1, 0},
		},
		{
			desc:     "Version with major and 0-valued minor",
			in:       "1.0",
			expected: SchemeVersion{1, 0},
		},
		{
			desc:     "Version with 0-valued major and minor",
			in:       "0.1",
			expected: SchemeVersion{0, 1},
		},
		{
			desc:     "Version with non-0 major and minor",
			in:       "2.5",
			expected: SchemeVersion{2, 5},
		},
		{
			desc:     "Version with large major/minor",
			in:       "12345.12345",
			expected: SchemeVersion{12345, 12345},
		},
		{
			desc:     "Version with double zero major",
			in:       "00.1",
			expected: SchemeVersion{0, 1},
		},
		{
			desc:     "Version with double zero minor",
			in:       "1.00",
			expected: SchemeVersion{1, 0},
		},
	}

	for _, testCase := range testTable {
		sv, err := NewSchemeVersion(testCase.in)
		assert.NoError(t, err, testCase.desc)
		assert.NotNil(t, sv, testCase.desc)
		assert.IsType(t, &SchemeVersion{}, sv, testCase.desc)
		assert.Equal(t, testCase.expected.Major, sv.Major, testCase.desc)
		assert.Equal(t, testCase.expected.Minor, sv.Minor, testCase.desc)
	}
}

// TestNewSchemeVersion_Error tests creating a new SchemeVersion with errors.
func TestNewSchemeVersion_Error(t *testing.T) {
	var testTable = []struct {
		desc string
		in   string
	}{
		{
			desc: "Empty string used as version",
			in:   "",
		},
		{
			desc: "Invalid major version, no minor (not an int)",
			in:   "xyz",
		},
		{
			desc: "Invalid major version (not an int)",
			in:   "xyz.0",
		},
		{
			desc: "Invalid minor version (not an int)",
			in:   "1.xyz",
		},
		{
			desc: "Invalid major and minor versions (not an int)",
			in:   "xyz.xyz",
		},
		{
			desc: "Extra version number components",
			in:   "1.2.3.4",
		},
	}

	for _, testCase := range testTable {
		sv, err := NewSchemeVersion(testCase.in)
		assert.Nil(t, sv, testCase.desc)
		assert.Error(t, err, testCase.desc)
	}
}

// TestSchemeVersion_String tests converting a SchemeVersion to a string
func TestSchemeVersion_String(t *testing.T) {
	var testTable = []struct {
		scheme   SchemeVersion
		expected string
	}{
		{
			scheme:   SchemeVersion{0, 1},
			expected: "0.1",
		},
		{
			scheme:   SchemeVersion{1, 0},
			expected: "1.0",
		},
		{
			scheme:   SchemeVersion{1, 1},
			expected: "1.1",
		},
		{
			scheme:   SchemeVersion{1234, 4321},
			expected: "1234.4321",
		},
	}

	for _, testCase := range testTable {
		actual := testCase.scheme.String()
		assert.Equal(t, testCase.expected, actual)
	}
}

// TestSchemeVersion_IsEqual test equality of SchemeVersions
func TestSchemeVersion_IsEqual(t *testing.T) {
	var testTable = []struct {
		scheme1 *SchemeVersion
		scheme2 *SchemeVersion
		equal   bool
	}{
		{
			scheme1: &SchemeVersion{1, 0},
			scheme2: &SchemeVersion{1, 0},
			equal:   true,
		},
		{
			scheme1: &SchemeVersion{0, 1},
			scheme2: &SchemeVersion{0, 1},
			equal:   true,
		},
		{
			scheme1: &SchemeVersion{4, 51},
			scheme2: &SchemeVersion{4, 51},
			equal:   true,
		},
		{
			scheme1: &SchemeVersion{1, 0},
			scheme2: &SchemeVersion{2, 0},
			equal:   false,
		},
		{
			scheme1: &SchemeVersion{1, 1},
			scheme2: &SchemeVersion{1, 2},
			equal:   false,
		},
	}

	for _, testCase := range testTable {
		actual := testCase.scheme1.IsEqual(testCase.scheme2)
		assert.Equal(t, testCase.equal, actual)
	}
}

// TestSchemeVersion_IsLessThan tests if one SchemeVersion is less than another
func TestSchemeVersion_IsLessThan(t *testing.T) {
	var testTable = []struct {
		scheme1  *SchemeVersion
		scheme2  *SchemeVersion
		lessThan bool
	}{
		{
			scheme1:  &SchemeVersion{1, 0},
			scheme2:  &SchemeVersion{1, 0},
			lessThan: false,
		},
		{
			scheme1:  &SchemeVersion{0, 1},
			scheme2:  &SchemeVersion{0, 1},
			lessThan: false,
		},
		{
			scheme1:  &SchemeVersion{4, 51},
			scheme2:  &SchemeVersion{4, 51},
			lessThan: false,
		},
		{
			scheme1:  &SchemeVersion{1, 0},
			scheme2:  &SchemeVersion{2, 0},
			lessThan: true,
		},
		{
			scheme1:  &SchemeVersion{1, 1},
			scheme2:  &SchemeVersion{1, 2},
			lessThan: true,
		},
		{
			scheme1:  &SchemeVersion{1, 2},
			scheme2:  &SchemeVersion{1, 1},
			lessThan: false,
		},
	}

	for _, testCase := range testTable {
		actual := testCase.scheme1.IsLessThan(testCase.scheme2)
		assert.Equal(t, testCase.lessThan, actual)
	}
}

// TestSchemeVersion_IsGreaterOrEqualTo tests if one SchemeVersion is greater than
// or qual to another
func TestSchemeVersion_IsGreaterOrEqualTo(t *testing.T) {
	var testTable = []struct {
		scheme1 *SchemeVersion
		scheme2 *SchemeVersion
		gte     bool
	}{
		{
			scheme1: &SchemeVersion{1, 0},
			scheme2: &SchemeVersion{1, 0},
			gte:     true,
		},
		{
			scheme1: &SchemeVersion{0, 1},
			scheme2: &SchemeVersion{0, 1},
			gte:     true,
		},
		{
			scheme1: &SchemeVersion{4, 51},
			scheme2: &SchemeVersion{4, 51},
			gte:     true,
		},
		{
			scheme1: &SchemeVersion{1, 0},
			scheme2: &SchemeVersion{2, 0},
			gte:     false,
		},
		{
			scheme1: &SchemeVersion{1, 1},
			scheme2: &SchemeVersion{1, 2},
			gte:     false,
		},
		{
			scheme1: &SchemeVersion{1, 2},
			scheme2: &SchemeVersion{1, 1},
			gte:     true,
		},
		{
			scheme1: &SchemeVersion{2, 1},
			scheme2: &SchemeVersion{1, 2},
			gte:     true,
		},
	}

	for _, testCase := range testTable {
		actual := testCase.scheme1.IsGreaterOrEqualTo(testCase.scheme2)
		assert.Equal(t, testCase.gte, actual)
	}
}

// TestConfigVersion_GetSchemeVersion_Ok tests getting the scheme version from a ConfigVersion
func TestConfigVersion_GetSchemeVersion_Ok(t *testing.T) {
	var testTable = []struct {
		desc    string
		version string
		scheme  SchemeVersion
	}{
		{
			desc:    "Version with only major specified",
			version: "1",
			scheme:  SchemeVersion{1, 0},
		},
		{
			desc:    "Version with major and 0-valued minor",
			version: "1.0",
			scheme:  SchemeVersion{1, 0},
		},
		{
			desc:    "Version with 0-valued major and minor",
			version: "0.1",
			scheme:  SchemeVersion{0, 1},
		},
		{
			desc:    "Version with non-0 major and minor",
			version: "2.5",
			scheme:  SchemeVersion{2, 5},
		},
		{
			desc:    "Version with large major/minor",
			version: "12345.12345",
			scheme:  SchemeVersion{12345, 12345},
		},
		{
			desc:    "Version with double zero major",
			version: "00.1",
			scheme:  SchemeVersion{0, 1},
		},
		{
			desc:    "Version with double zero minor",
			version: "1.00",
			scheme:  SchemeVersion{1, 0},
		},
	}

	for _, testCase := range testTable {
		cfgVer := ConfigVersion{Version: testCase.version}
		sv, err := cfgVer.GetSchemeVersion()
		assert.NoError(t, err, testCase.desc)
		assert.Equal(t, testCase.scheme.Major, sv.Major, testCase.desc)
		assert.Equal(t, testCase.scheme.Minor, sv.Minor, testCase.desc)
	}
}

// TestConfigVersion_GetSchemeVersion_Error tests getting the scheme version from a ConfigVersion
// which results in error
func TestConfigVersion_GetSchemeVersion_Error(t *testing.T) {
	var testTable = []struct {
		desc    string
		version string
	}{
		{
			desc:    "Empty string used as version",
			version: "",
		},
		{
			desc:    "Invalid major version, no minor (not an int)",
			version: "xyz",
		},
		{
			desc:    "Invalid major version (not an int)",
			version: "xyz.0",
		},
		{
			desc:    "Invalid minor version (not an int)",
			version: "1.xyz",
		},
		{
			desc:    "Invalid major and minor versions (not an int)",
			version: "xyz.xyz",
		},
		{
			desc:    "Extra version number components",
			version: "1.2.3.4",
		},
	}

	for _, testCase := range testTable {
		cfgVer := ConfigVersion{Version: testCase.version}
		sv, err := cfgVer.GetSchemeVersion()
		assert.Error(t, err, testCase.desc)
		assert.Nil(t, sv, testCase.desc)
	}
}

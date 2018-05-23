package cfg

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDeviceConfig_Validate_Ok tests validating a DeviceConfig with no errors.
func TestDeviceConfig_Validate_Ok(t *testing.T) {
	var testTable = []struct {
		desc   string
		config DeviceConfig
	}{
		{
			desc: "DeviceConfig has valid version",
			config: DeviceConfig{
				Version: &ConfigVersion{Version: "1.0"},
			},
		},
		{
			desc: "DeviceConfig has valid version and location",
			config: DeviceConfig{
				Version:   &ConfigVersion{Version: "1.0"},
				Locations: []*Location{{Name: "test", Rack: &LocationData{Name: "test"}, Board: &LocationData{Name: "test"}}},
			},
		},
		{
			desc: "DeviceConfig has valid version, location, and DeviceKind",
			config: DeviceConfig{
				Version:   &ConfigVersion{Version: "1.0"},
				Locations: []*Location{{Name: "test", Rack: &LocationData{Name: "test"}, Board: &LocationData{Name: "test"}}},
				Devices:   []*DeviceKind{{Name: "test"}},
			},
		},
		{
			desc: "DeviceConfig has valid version, invalid Locations (Locations not validated here)",
			config: DeviceConfig{
				Version:   &ConfigVersion{Version: "1.0"},
				Locations: []*Location{{Name: ""}},
			},
		},
		{
			desc: "DeviceConfig has valid version and locations, invalid DeviceKinds (DeviceKinds not validated here)",
			config: DeviceConfig{
				Version:   &ConfigVersion{Version: "1.0"},
				Locations: []*Location{{Name: "test", Rack: &LocationData{Name: "test"}, Board: &LocationData{Name: "test"}}},
				Devices:   []*DeviceKind{{Name: ""}},
			},
		},
	}

	for _, testCase := range testTable {
		err := testCase.config.Validate()
		assert.NoError(t, err, testCase.desc)
	}
}

// TestDeviceConfig_Validate_Error tests validating a DeviceConfig with errors.
func TestDeviceConfig_Validate_Error(t *testing.T) {
	var testTable = []struct {
		desc   string
		config DeviceConfig
	}{
		{
			desc: "DeviceConfig has invalid version",
			config: DeviceConfig{
				Version: &ConfigVersion{Version: "abc"},
			},
		},
	}

	for _, testCase := range testTable {
		err := testCase.config.Validate()
		assert.Error(t, err, testCase.desc)
	}
}

// TestLocation_Validate_Ok tests validating a Location with no errors.
func TestLocation_Validate_Ok(t *testing.T) {
	var testTable = []struct {
		desc     string
		location Location
	}{
		{
			desc: "Valid Location instance",
			location: Location{
				Name:  "test",
				Rack:  &LocationData{Name: "test"},
				Board: &LocationData{Name: "test"},
			},
		},
	}

	for _, testCase := range testTable {
		err := testCase.location.Validate()
		assert.NoError(t, err, testCase.desc)
	}
}

// TestLocation_Validate_Error tests validating a Location with errors.
func TestLocation_Validate_Error(t *testing.T) {
	var testTable = []struct {
		desc     string
		location Location
	}{
		{
			desc: "Location requires a name, but has none",
			location: Location{
				Rack:  &LocationData{Name: "test"},
				Board: &LocationData{Name: "test"},
			},
		},
		{
			desc: "Location requires a rack, but has none",
			location: Location{
				Name:  "test",
				Board: &LocationData{Name: "test"},
			},
		},
		{
			desc: "Location requires a board, but has none",
			location: Location{
				Name: "test",
				Rack: &LocationData{Name: "test"},
			},
		},
		{
			desc: "Location both rack and board, has neither",
			location: Location{
				Name: "test",
			},
		},
		{
			desc: "Location has an invalid rack",
			location: Location{
				Rack:  &LocationData{Name: ""},
				Board: &LocationData{Name: "test"},
			},
		},
		{
			desc: "Location has an invalid board",
			location: Location{
				Rack:  &LocationData{Name: "test"},
				Board: &LocationData{Name: ""},
			},
		},
	}

	for _, testCase := range testTable {
		err := testCase.location.Validate()
		assert.Error(t, err, testCase.desc)
	}
}

// TestLocationData_Validate_Ok tests validating a LocationData with no errors.
func TestLocationData_Validate_Ok(t *testing.T) {
	var testTable = []struct {
		desc         string
		locationData LocationData
	}{
		{
			desc:         "LocationData has a valid name",
			locationData: LocationData{Name: "test"},
		},
		{
			desc:         "LocationData has a valid fromEnv",
			locationData: LocationData{FromEnv: "TEST_ENV"},
		},
		{
			desc:         "LocationData has both name and fromEnv",
			locationData: LocationData{Name: "foo", FromEnv: "TEST_ENV"},
		},
	}

	err := os.Setenv("TEST_ENV", "test")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err := os.Unsetenv("TEST_ENV")
		if err != nil {
			t.Error(err)
		}
	}()

	for _, testCase := range testTable {
		err := testCase.locationData.Validate()
		assert.NoError(t, err, testCase.desc)
	}
}

// TestLocationData_Validate_Error tests validating a LocationData with errors.
func TestLocationData_Validate_Error(t *testing.T) {
	var testTable = []struct {
		desc         string
		locationData LocationData
	}{
		{
			desc:         "LocationData has no fromEnv or name",
			locationData: LocationData{},
		},
		{
			desc:         "LocationData fromEnv does not resolve",
			locationData: LocationData{FromEnv: "FOO_BAR_BAZ"},
		},
	}

	for _, testCase := range testTable {
		err := testCase.locationData.Validate()
		assert.Error(t, err, testCase.desc)
	}
}

// TestLocationData_Get_Ok tests getting the locational data with no errors.
func TestLocationData_Get_Ok(t *testing.T) {
	var testTable = []struct {
		desc         string
		locationData LocationData
		expected     string
	}{
		{
			desc:         "LocationData has a valid name",
			locationData: LocationData{Name: "test"},
			expected:     "test",
		},
		{
			desc:         "LocationData has a valid fromEnv",
			locationData: LocationData{FromEnv: "TEST_ENV"},
			expected:     "foo",
		},
		{
			desc:         "LocationData has both name and fromEnv (should take name)",
			locationData: LocationData{Name: "test", FromEnv: "TEST_ENV"},
			expected:     "test",
		},
		{
			desc:         "LocationData has no fromEnv or name",
			locationData: LocationData{},
			expected:     "",
		},
	}

	err := os.Setenv("TEST_ENV", "foo")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err := os.Unsetenv("TEST_ENV")
		if err != nil {
			t.Error(err)
		}
	}()

	for _, testCase := range testTable {
		actual, err := testCase.locationData.Get()
		assert.NoError(t, err, testCase.desc)
		assert.Equal(t, testCase.expected, actual, testCase.desc)
	}
}

// TestLocationData_Get_Error tests getting the location data with errors.
func TestLocationData_Get_Error(t *testing.T) {
	var testTable = []struct {
		desc         string
		locationData LocationData
	}{
		{
			desc:         "LocationData fromEnv does not resolve",
			locationData: LocationData{FromEnv: "FOO_BAR_BAZ"},
		},
	}

	for _, testCase := range testTable {
		actual, err := testCase.locationData.Get()
		assert.Error(t, err, testCase.desc)
		assert.Equal(t, "", actual, testCase.desc)
	}
}

// TestDeviceKind_Validate_Ok tests validating a DeviceKind with no errors.
func TestDeviceKind_Validate_Ok(t *testing.T) {
	var testTable = []struct {
		desc string
		kind DeviceKind
	}{
		{
			desc: "DeviceKind has a valid name",
			kind: DeviceKind{
				Name: "test",
			},
		},
		{
			desc: "DeviceKind has a valid name and instances",
			kind: DeviceKind{
				Name:      "test",
				Instances: []*DeviceInstance{{Location: "test"}},
			},
		},
		{
			desc: "DeviceKind has a valid name, instances, and outputs",
			kind: DeviceKind{
				Name:      "test",
				Instances: []*DeviceInstance{{Location: "test"}},
				Outputs:   []*DeviceOutput{{Type: "test"}},
			},
		},
		{
			desc: "DeviceKind has valid name, invalid instances (DeviceInstance not validated here)",
			kind: DeviceKind{
				Name:      "test",
				Instances: []*DeviceInstance{{Location: ""}},
			},
		},
		{
			desc: "DeviceKind has valid name, invalid outputs (DeviceOutputs not validated here)",
			kind: DeviceKind{
				Name:    "test",
				Outputs: []*DeviceOutput{{Type: ""}},
			},
		},
	}

	for _, testCase := range testTable {
		err := testCase.kind.Validate()
		assert.NoError(t, err, testCase.desc)
	}
}

// TestDeviceKind_Validate_Error tests validating a DeviceKind with errors.
func TestDeviceKind_Validate_Error(t *testing.T) {
	var testTable = []struct {
		desc string
		kind DeviceKind
	}{
		{
			desc: "DeviceKind has no name specified",
			kind: DeviceKind{},
		},
	}

	for _, testCase := range testTable {
		err := testCase.kind.Validate()
		assert.Error(t, err, testCase.desc)
	}
}

// TestDeviceInstance_Validate_Ok tests validating a DeviceInstance with no errors.
func TestDeviceInstance_Validate_Ok(t *testing.T) {
	var testTable = []struct {
		desc     string
		instance DeviceInstance
	}{
		{
			desc: "DeviceInstance has a valid location",
			instance: DeviceInstance{
				Location: "test",
			},
		},
		{
			desc: "DeviceInstance has a valid location and outputs",
			instance: DeviceInstance{
				Location: "test",
				Outputs:  []*DeviceOutput{{Type: "test"}},
			},
		},
		{
			desc: "DeviceInstance has valid location, invalid outputs (DeviceOutputs not validated here)",
			instance: DeviceInstance{
				Location: "test",
				Outputs:  []*DeviceOutput{{Type: ""}},
			},
		},
	}

	for _, testCase := range testTable {
		err := testCase.instance.Validate()
		assert.NoError(t, err, testCase.desc)
	}
}

// TestDeviceInstance_Validate_Error tests validating a DeviceInstance with errors.
func TestDeviceInstance_Validate_Error(t *testing.T) {
	var testTable = []struct {
		desc     string
		instance DeviceInstance
	}{
		{
			desc: "DeviceInstance has a no location",
			instance: DeviceInstance{
				Location: "",
			},
		},
	}

	for _, testCase := range testTable {
		err := testCase.instance.Validate()
		assert.Error(t, err, testCase.desc)
	}
}

// TestDeviceOutput_Validate_Ok tests validating a DeviceOutput with no errors.
func TestDeviceOutput_Validate_Ok(t *testing.T) {
	var testTable = []struct {
		desc   string
		output DeviceOutput
	}{
		{
			desc: "DeviceOutput has valid type",
			output: DeviceOutput{
				Type: "test",
			},
		},
	}

	for _, testCase := range testTable {
		err := testCase.output.Validate()
		assert.NoError(t, err, testCase.desc)
	}
}

// TestDeviceOutput_Validate_Error tests validating a DeviceOutput with errors.
func TestDeviceOutput_Validate_Error(t *testing.T) {
	var testTable = []struct {
		desc   string
		output DeviceOutput
	}{
		{
			desc: "DeviceOutput has no type",
			output: DeviceOutput{
				Type: "",
			},
		},
	}

	for _, testCase := range testTable {
		err := testCase.output.Validate()
		assert.Error(t, err, testCase.desc)
	}
}

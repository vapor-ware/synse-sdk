package config

import (
	"os"
	"testing"

	"fmt"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
)

// TestNewDeviceConfig tests initializing a new DeviceConfig.
func TestNewDeviceConfig(t *testing.T) {
	cfg := NewDeviceConfig()

	assert.IsType(t, &DeviceConfig{}, cfg)
	assert.Equal(t, currentDeviceSchemeVersion, cfg.Version)
	assert.Equal(t, 0, len(cfg.Locations))
	assert.Equal(t, 0, len(cfg.Devices))
}

// TestDeviceConfig_GetLocation_Ok tests getting locations from a DeviceConfig successfully.
func TestDeviceConfig_GetLocation_Ok(t *testing.T) {
	var testTable = []struct {
		desc     string
		location string
		config   DeviceConfig
	}{
		{
			desc:     "DeviceConfig has single location",
			location: "test",
			config: DeviceConfig{
				Locations: []*Location{
					{Name: "test", Rack: &LocationData{Name: "test"}, Board: &LocationData{Name: "test"}},
				},
			},
		},
		{
			desc:     "DeviceConfig has multiple locations",
			location: "test",
			config: DeviceConfig{
				Locations: []*Location{
					{Name: "test", Rack: &LocationData{Name: "test"}, Board: &LocationData{Name: "test"}},
					{Name: "foo", Rack: &LocationData{Name: "foo"}, Board: &LocationData{Name: "foo"}},
					{Name: "bar", Rack: &LocationData{Name: "bar"}, Board: &LocationData{Name: "bar"}},
				},
			},
		},
		{
			desc:     "DeviceConfig has multiple locations",
			location: "bar",
			config: DeviceConfig{
				Locations: []*Location{
					{Name: "test", Rack: &LocationData{Name: "test"}, Board: &LocationData{Name: "test"}},
					{Name: "foo", Rack: &LocationData{Name: "foo"}, Board: &LocationData{Name: "foo"}},
					{Name: "bar", Rack: &LocationData{Name: "bar"}, Board: &LocationData{Name: "bar"}},
				},
			},
		},
	}

	for _, testCase := range testTable {
		l, err := testCase.config.GetLocation(testCase.location)
		assert.NoError(t, err, testCase.desc)
		assert.NotNil(t, l, testCase.desc)
		assert.Equal(t, testCase.location, l.Name, testCase.desc)
	}
}

// TestDeviceConfig_GetLocation_Err tests getting locations from a DeviceConfig unsuccessfully.
func TestDeviceConfig_GetLocation_Err(t *testing.T) {
	var testTable = []struct {
		desc     string
		location string
		config   DeviceConfig
	}{
		{
			desc:     "DeviceConfig has no locations defined",
			location: "test",
			config: DeviceConfig{
				Locations: []*Location{},
			},
		},
		{
			desc:     "Specified name does not match any location",
			location: "baz",
			config: DeviceConfig{
				Locations: []*Location{
					{Name: "test", Rack: &LocationData{Name: "test"}, Board: &LocationData{Name: "test"}},
					{Name: "foo", Rack: &LocationData{Name: "foo"}, Board: &LocationData{Name: "foo"}},
					{Name: "bar", Rack: &LocationData{Name: "bar"}, Board: &LocationData{Name: "bar"}},
				},
			},
		},
	}

	for _, testCase := range testTable {
		l, err := testCase.config.GetLocation(testCase.location)
		assert.Error(t, err, testCase.desc)
		assert.Nil(t, l, testCase.desc)
	}
}

// TestDeviceConfig_Validate_Ok tests validating a DeviceConfig with no errors.
func TestDeviceConfig_Validate_Ok(t *testing.T) {
	var testTable = []struct {
		desc   string
		config DeviceConfig
	}{
		{
			desc: "DeviceConfig has valid version",
			config: DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
			},
		},
		{
			desc: "DeviceConfig has valid version and location",
			config: DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations:     []*Location{{Name: "test", Rack: &LocationData{Name: "test"}, Board: &LocationData{Name: "test"}}},
			},
		},
		{
			desc: "DeviceConfig has valid version, location, and DeviceKind",
			config: DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations:     []*Location{{Name: "test", Rack: &LocationData{Name: "test"}, Board: &LocationData{Name: "test"}}},
				Devices:       []*DeviceKind{{Name: "test"}},
			},
		},
		{
			desc: "DeviceConfig has valid version, invalid Locations (Locations not validated here)",
			config: DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations:     []*Location{{Name: ""}},
			},
		},
		{
			desc: "DeviceConfig has valid version and locations, invalid DeviceKinds (DeviceKinds not validated here)",
			config: DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations:     []*Location{{Name: "test", Rack: &LocationData{Name: "test"}, Board: &LocationData{Name: "test"}}},
				Devices:       []*DeviceKind{{Name: ""}},
			},
		},
	}

	for _, testCase := range testTable {
		merr := errors.NewMultiError("test")

		testCase.config.Validate(merr)
		assert.NoError(t, merr.Err(), testCase.desc)
	}
}

// TestDeviceConfig_Validate_Error tests validating a DeviceConfig with errors.
func TestDeviceConfig_Validate_Error(t *testing.T) {
	var testTable = []struct {
		desc     string
		errCount int
		config   DeviceConfig
	}{
		{
			desc:     "DeviceConfig has invalid version",
			errCount: 1,
			config: DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "abc"},
			},
		},
	}

	for _, testCase := range testTable {
		merr := errors.NewMultiError("test")

		testCase.config.Validate(merr)
		assert.Error(t, merr.Err(), testCase.desc)
		assert.Equal(t, testCase.errCount, len(merr.Errors), merr.Error())
	}
}

// TestLocation_Equals tests whether a location equals another one.
func TestLocation_Equals(t *testing.T) {
	testLoc := Location{
		Name:  "test",
		Rack:  &LocationData{Name: "rack"},
		Board: &LocationData{Name: "board"},
	}

	var testTable = []struct {
		desc      string
		isEqual   bool
		location1 *Location
		location2 *Location
	}{
		{
			desc:      "pointer to the same Location",
			isEqual:   true,
			location1: &testLoc,
			location2: &testLoc,
		},
		{
			desc:      "different location instances, same data (empty)",
			isEqual:   true,
			location1: &Location{},
			location2: &Location{},
		},
		{
			desc:      "different location instances, same data",
			isEqual:   true,
			location1: &testLoc,
			location2: &Location{
				Name:  "test",
				Rack:  &LocationData{Name: "rack"},
				Board: &LocationData{Name: "board"},
			},
		},
		{
			desc:      "different locations, different name",
			isEqual:   false,
			location1: &testLoc,
			location2: &Location{
				Name:  "foo",
				Rack:  &LocationData{Name: "rack"},
				Board: &LocationData{Name: "board"},
			},
		},
		{
			desc:      "different locations, different rack",
			isEqual:   false,
			location1: &testLoc,
			location2: &Location{
				Name:  "test",
				Rack:  &LocationData{Name: "foo"},
				Board: &LocationData{Name: "board"},
			},
		},
		{
			desc:      "different locations, different board",
			isEqual:   false,
			location1: &testLoc,
			location2: &Location{
				Name:  "test",
				Rack:  &LocationData{Name: "rack"},
				Board: &LocationData{Name: "foo"},
			},
		},
	}

	for _, testCase := range testTable {
		equals := testCase.location1.Equals(testCase.location2)
		assert.Equal(t, testCase.isEqual, equals, testCase.desc)
	}
}

// TestLocationData_Equals tests whether two LocationData instances are equal.
func TestLocationData_Equals(t *testing.T) {
	testLoc := LocationData{
		Name: "foo",
	}

	var testTable = []struct {
		desc    string
		isEqual bool
		loc1    *LocationData
		loc2    *LocationData
	}{
		{
			desc:    "pointer to the same LocationData",
			isEqual: true,
			loc1:    &testLoc,
			loc2:    &testLoc,
		},
		{
			desc:    "different LocationData, same data (empty)",
			isEqual: true,
			loc1:    &LocationData{},
			loc2:    &LocationData{},
		},
		{
			desc:    "different LocationData, same data (name)",
			isEqual: true,
			loc1: &LocationData{
				Name: "test",
			},
			loc2: &LocationData{
				Name: "test",
			},
		},
		{
			desc:    "different LocationData, same data (from env)",
			isEqual: true,
			loc1: &LocationData{
				FromEnv: "HOSTNAME",
			},
			loc2: &LocationData{
				FromEnv: "HOSTNAME",
			},
		},
		{
			desc:    "different LocationData, different data (name)",
			isEqual: false,
			loc1: &LocationData{
				Name: "foo",
			},
			loc2: &LocationData{
				Name: "bar",
			},
		},
		{
			desc:    "different LocationData, different data (from env)",
			isEqual: false,
			loc1: &LocationData{
				FromEnv: "NODENAME",
			},
			loc2: &LocationData{
				FromEnv: "HOSTNAME",
			},
		},
		{
			desc:    "different LocationData, different data (mixed)",
			isEqual: false,
			loc1: &LocationData{
				Name: "test",
			},
			loc2: &LocationData{
				FromEnv: "HOSTNAME",
			},
		},
		{
			desc:    "different LocationData, different data (empty)",
			isEqual: false,
			loc1:    &LocationData{},
			loc2: &LocationData{
				FromEnv: "HOSTNAME",
			},
		},
	}

	for _, testCase := range testTable {
		equals := testCase.loc1.Equals(testCase.loc2)
		assert.Equal(t, testCase.isEqual, equals, testCase.desc)
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
		merr := errors.NewMultiError("test")

		testCase.location.Validate(merr)
		assert.NoError(t, merr.Err(), testCase.desc)
	}
}

// TestLocation_Validate_Error tests validating a Location with errors.
func TestLocation_Validate_Error(t *testing.T) {
	var testTable = []struct {
		desc     string
		errCount int
		location Location
	}{
		{
			desc:     "Location requires a name, but has none",
			errCount: 1,
			location: Location{
				Rack:  &LocationData{Name: "test"},
				Board: &LocationData{Name: "test"},
			},
		},
		{
			desc:     "Location requires a rack, but has none",
			errCount: 1,
			location: Location{
				Name:  "test",
				Board: &LocationData{Name: "test"},
			},
		},
		{
			desc:     "Location requires a board, but has none",
			errCount: 1,
			location: Location{
				Name: "test",
				Rack: &LocationData{Name: "test"},
			},
		},
		{
			desc:     "Location both rack and board, has neither",
			errCount: 2,
			location: Location{
				Name: "test",
			},
		},
		{
			desc:     "Location has an invalid rack",
			errCount: 1,
			location: Location{
				Name:  "test",
				Rack:  &LocationData{Name: ""},
				Board: &LocationData{Name: "test"},
			},
		},
		{
			desc:     "Location has an invalid board",
			errCount: 1,
			location: Location{
				Name:  "test",
				Rack:  &LocationData{Name: "test"},
				Board: &LocationData{Name: ""},
			},
		},
		{
			desc:     "Location missing all fields",
			errCount: 3,
			location: Location{},
		},
	}

	for _, testCase := range testTable {
		merr := errors.NewMultiError("test")

		testCase.location.Validate(merr)
		assert.Error(t, merr.Err(), testCase.desc)
		assert.Equal(t, testCase.errCount, len(merr.Errors), merr.Error())
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
		merr := errors.NewMultiError("test")

		testCase.locationData.Validate(merr)
		assert.NoError(t, merr.Err(), testCase.desc)
	}
}

// TestLocationData_Validate_Error tests validating a LocationData with errors.
func TestLocationData_Validate_Error(t *testing.T) {
	var testTable = []struct {
		desc         string
		errCount     int
		locationData LocationData
	}{
		{
			desc:         "LocationData has no fromEnv or name",
			errCount:     2,
			locationData: LocationData{},
		},
		{
			desc:         "LocationData fromEnv does not resolve",
			errCount:     2,
			locationData: LocationData{FromEnv: "FOO_BAR_BAZ"},
		},
	}

	for _, testCase := range testTable {
		merr := errors.NewMultiError("test")

		testCase.locationData.Validate(merr)
		assert.Error(t, merr.Err(), testCase.desc)
		assert.Equal(t, testCase.errCount, len(merr.Errors), merr.Error())
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
		merr := errors.NewMultiError("test")

		testCase.kind.Validate(merr)
		assert.NoError(t, merr.Err(), testCase.desc)
	}
}

// TestDeviceKind_Validate_Error tests validating a DeviceKind with errors.
func TestDeviceKind_Validate_Error(t *testing.T) {
	var testTable = []struct {
		desc     string
		errCount int
		kind     DeviceKind
	}{
		{
			desc:     "DeviceKind has no name specified",
			errCount: 1,
			kind:     DeviceKind{},
		},
	}

	for _, testCase := range testTable {
		merr := errors.NewMultiError("test")

		testCase.kind.Validate(merr)
		assert.Error(t, merr.Err(), testCase.desc)
		assert.Equal(t, testCase.errCount, len(merr.Errors), merr.Error())
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
		merr := errors.NewMultiError("test")

		testCase.instance.Validate(merr)
		assert.NoError(t, merr.Err(), testCase.desc)
	}
}

// TestDeviceInstance_Validate_Error tests validating a DeviceInstance with errors.
func TestDeviceInstance_Validate_Error(t *testing.T) {
	var testTable = []struct {
		desc     string
		errCount int
		instance DeviceInstance
	}{
		{
			desc:     "DeviceInstance has a no location",
			errCount: 1,
			instance: DeviceInstance{
				Location: "",
			},
		},
	}

	for _, testCase := range testTable {
		merr := errors.NewMultiError("test")

		testCase.instance.Validate(merr)
		assert.Error(t, merr.Err(), testCase.desc)
		assert.Equal(t, testCase.errCount, len(merr.Errors), merr.Error())
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
		merr := errors.NewMultiError("test")

		testCase.output.Validate(merr)
		assert.NoError(t, merr.Err(), testCase.desc)
	}
}

// TestDeviceOutput_Validate_Error tests validating a DeviceOutput with errors.
func TestDeviceOutput_Validate_Error(t *testing.T) {
	var testTable = []struct {
		desc     string
		errCount int
		output   DeviceOutput
	}{
		{
			desc:     "DeviceOutput has no type",
			errCount: 1,
			output: DeviceOutput{
				Type: "",
			},
		},
	}

	for _, testCase := range testTable {
		merr := errors.NewMultiError("test")

		testCase.output.Validate(merr)
		assert.Error(t, merr.Err(), testCase.desc)
		assert.Equal(t, testCase.errCount, len(merr.Errors), merr.Error())
	}
}

// TestDeviceConfig_ValidateDeviceConfigDataOk tests validating config data when there
// are no errors.
func TestDeviceConfig_ValidateDeviceConfigDataOk(t *testing.T) {
	var testTable = []struct {
		desc   string
		config DeviceConfig
	}{
		{
			desc: "no data field in the device config",
			config: DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations:     []*Location{},
				Devices:       []*DeviceKind{},
			},
		},
		{
			desc: "data in the device kind output",
			config: DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations:     []*Location{},
				Devices: []*DeviceKind{
					{
						Outputs: []*DeviceOutput{
							{
								Data: map[string]interface{}{
									"foo": "bar",
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "data in the device instance",
			config: DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations:     []*Location{},
				Devices: []*DeviceKind{
					{
						Instances: []*DeviceInstance{
							{
								Data: map[string]interface{}{
									"foo": "bar",
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "data in the device instance output",
			config: DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations:     []*Location{},
				Devices: []*DeviceKind{
					{
						Instances: []*DeviceInstance{
							{
								Outputs: []*DeviceOutput{
									{
										Data: map[string]interface{}{
											"foo": "bar",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// define the validator function. it does nothing and returns nil,
	// so we should never get an error.
	var validator = func(_ map[string]interface{}) error {
		return nil
	}

	for _, testCase := range testTable {
		err := testCase.config.ValidateDeviceConfigData(validator)
		assert.NoError(t, err.Err(), testCase.desc)
	}
}

// TestDeviceConfig_ValidateDeviceConfigDataError tests validating config data when there
// are errors.
func TestDeviceConfig_ValidateDeviceConfigDataError(t *testing.T) {
	var testTable = []struct {
		desc   string
		config DeviceConfig
	}{
		{
			desc: "data in the device kind output",
			config: DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations:     []*Location{},
				Devices: []*DeviceKind{
					{
						Outputs: []*DeviceOutput{
							{
								Data: map[string]interface{}{
									"foo": "bar",
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "data in the device instance",
			config: DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations:     []*Location{},
				Devices: []*DeviceKind{
					{
						Instances: []*DeviceInstance{
							{
								Data: map[string]interface{}{
									"foo": "bar",
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "data in the device instance output",
			config: DeviceConfig{
				SchemeVersion: SchemeVersion{Version: "1.0"},
				Locations:     []*Location{},
				Devices: []*DeviceKind{
					{
						Instances: []*DeviceInstance{
							{
								Outputs: []*DeviceOutput{
									{
										Data: map[string]interface{}{
											"foo": "bar",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// define the validator function. it does nothing and returns an error, so
	// all things should fail validation.
	var validator = func(_ map[string]interface{}) error {
		return fmt.Errorf("test error")
	}

	for _, testCase := range testTable {
		err := testCase.config.ValidateDeviceConfigData(validator)
		assert.Error(t, err.Err(), testCase.desc)
	}
}

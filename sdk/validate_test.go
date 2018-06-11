package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-server-grpc/go"
)

//
// Test Structures / Type Definitions
//

// simpleTestConfig is a simple struct that fulfils both the ConfigBase
// and ConfigComponent interface. It is used to test simple validation
// cases.
type simpleTestConfig struct {
	SchemeVersion

	TestField string `addedIn:"1.0" deprecatedIn:"1.5" removedIn:"2.0"`

	BadAddedInTag      string `addedIn:"xyz.1"`
	BadDeprecatedInTag string `deprecatedIn:"xyz.1"`
	BadRemovedInTag    string `removedIn:"xyz.1"`
}

func (config simpleTestConfig) Validate(multiErr *errors.MultiError) {
	_, err := config.GetVersion()
	if err != nil {
		multiErr.Add(errors.NewValidationError(multiErr.Context["source"], err.Error()))
	}

	if config.TestField == "" {
		multiErr.Add(errors.NewFieldRequiredError(multiErr.Context["source"], "testField"))
	}
}

// complexTestConfig is a simple struct that fulfils both the ConfigBase
// and ConfigComponent interface. It is used to test complex validation
// cases where there are nested and grouped components.
type complexTestConfig struct {
	SchemeVersion

	Foo      bool               `addedIn:"1.0" removedIn:"1.5"`
	Simples  []simpleTestConfig `addedIn:"1.5"`
	FloatVal float64            `addedIn:"1.0" deprecatedIn:"3.0"`
	IntVal   int32              `addedIn:"1.0"`
	UintVal  uint8              `addedIn:"1.0" deprecatedIn:"2.0" removedIn:"3.0"`

	RootUser *nestedStruct   `addedIn:"1.0"`
	Users    []*nestedStruct `addedIn:"1.5"`
}

func (config complexTestConfig) Validate(multiErr *errors.MultiError) {
	_, err := config.GetVersion()
	if err != nil {
		multiErr.Add(errors.NewValidationError(multiErr.Context["source"], err.Error()))
	}

	if config.UintVal > 8 {
		multiErr.Add(errors.NewInvalidValueError(
			multiErr.Context["source"],
			"uintVal",
			"less than or equal to 8",
		))
	}
}

type nestedStruct struct {
	User string `addedIn:"1.0"`
	Pass string `addedIn:"1.0"`
}

func (n nestedStruct) Validate(multiError *errors.MultiError) {
	if n.User == "" {
		multiError.Add(errors.NewFieldRequiredError(multiError.Context["source"], "user"))
	}
}

//
// Helper Functions
//

// checkValidationCleanup checks to make sure that the validator cleaned up
// after a Validation run. This prevents previous run state from persisting
// to the next run.
func checkValidationCleanup(t *testing.T) {
	assert.Nil(t, Validator.context)
	assert.Nil(t, Validator.errors)
	assert.Nil(t, Validator.version)
}

//
// Test Cases
//

// TestSchemeValidator_Validate_Simple_Ok tests validating a simple struct where everything is ok.
func TestSchemeValidator_Validate_Simple_Ok(t *testing.T) {
	toValidate := &ConfigContext{
		Source: "<simple test config>",
		Config: &simpleTestConfig{
			SchemeVersion: SchemeVersion{Version: "1.0"},
			TestField:     "foo",
		},
	}

	err := Validator.Validate(toValidate)
	assert.NoError(t, err.Err())

	// check that validation cleanup was successful
	checkValidationCleanup(t)
}

// TestSchemeValidator_Validate_Simple_UnsupportedVersion tests validating a simple struct where
// the SchemeVersion of the struct is less than that of a specified field.
func TestSchemeValidator_Validate_Simple_UnsupportedVersion(t *testing.T) {
	toValidate := &ConfigContext{
		Source: "<simple test config>",
		Config: &simpleTestConfig{
			SchemeVersion: SchemeVersion{Version: "0.5"}, // config version less than addedIn tag for both fields
			TestField:     "foo",
		},
	}

	err := Validator.Validate(toValidate)
	assert.Error(t, err.Err())
	assert.Equal(t, 2, len(err.Errors), err.Error())

	// check that validation cleanup was successful
	checkValidationCleanup(t)
}

// TestSchemeValidator_Validate_Simple_DeprecatedVersion1 tests validating a simple struct where
// the SchemeVersion of the struct is equal to the deprecatedIn tag of a field.
func TestSchemeValidator_Validate_Simple_DeprecatedVersion1(t *testing.T) {
	toValidate := &ConfigContext{
		Source: "<simple test config>",
		Config: &simpleTestConfig{
			SchemeVersion: SchemeVersion{Version: "1.5"}, // equal to deprecatedIn tag for TestField
			TestField:     "foo",
		},
	}

	err := Validator.Validate(toValidate)
	assert.NoError(t, err.Err()) // deprecated logs warning, no error

	// check that validation cleanup was successful
	checkValidationCleanup(t)
}

// TestSchemeValidator_Validate_Simple_DeprecatedVersion2 tests validating a simple struct where
// the SchemeVersion of the struct is greater than the deprecatedIn tag of a field.
func TestSchemeValidator_Validate_Simple_DeprecatedVersion2(t *testing.T) {
	toValidate := &ConfigContext{
		Source: "<simple test config>",
		Config: &simpleTestConfig{
			SchemeVersion: SchemeVersion{Version: "1.8"}, // greater than deprecatedIn tag for TestField
			TestField:     "foo",
		},
	}

	err := Validator.Validate(toValidate)
	assert.NoError(t, err.Err()) // deprecated logs warning, no error

	// check that validation cleanup was successful
	checkValidationCleanup(t)
}

// TestSchemeValidator_Validate_Simple_RemovedVersion1 tests validating a simple struct where
// the SchemeVersion of the struct is equal to the removedIn tag of a field.
func TestSchemeValidator_Validate_Simple_RemovedVersion1(t *testing.T) {
	toValidate := &ConfigContext{
		Source: "<simple test config>",
		Config: &simpleTestConfig{
			SchemeVersion: SchemeVersion{Version: "2.0"}, // equal to removedIn tag for TestField
			TestField:     "foo",
		},
	}

	err := Validator.Validate(toValidate)
	assert.Error(t, err.Err())
	assert.Equal(t, 1, len(err.Errors), err.Error())

	// check that validation cleanup was successful
	checkValidationCleanup(t)
}

// TestSchemeValidator_Validate_Simple_RemovedVersion2 tests validating a simple struct where
// the SchemeVersion of the struct is greater than the removedIn tag of a field.
func TestSchemeValidator_Validate_Simple_RemovedVersion2(t *testing.T) {
	toValidate := &ConfigContext{
		Source: "<simple test config>",
		Config: &simpleTestConfig{
			SchemeVersion: SchemeVersion{Version: "2.1"}, // greater than removedIn tag for TestField
			TestField:     "foo",
		},
	}

	err := Validator.Validate(toValidate)
	assert.Error(t, err.Err())
	assert.Equal(t, 1, len(err.Errors), err.Error())

	// check that validation cleanup was successful
	checkValidationCleanup(t)
}

// TestSchemeValidator_Validate_Error tests validating a simple struct where
// the versions resolve, but the Validate function fails for one field.
func TestSchemeValidator_Validate_Error(t *testing.T) {
	toValidate := &ConfigContext{
		Source: "<simple test config>",
		Config: &simpleTestConfig{
			SchemeVersion: SchemeVersion{Version: "1.0"},
			TestField:     "", // TestField required, will fail Validate()
		},
	}

	err := Validator.Validate(toValidate)
	assert.Error(t, err.Err())
	assert.Equal(t, 1, len(err.Errors), err.Error())

	// check that validation cleanup was successful
	checkValidationCleanup(t)
}

// TestSchemeValidator_Validate_Error2 tests validating a simple struct where
// the version is out of bounds and the Validate function fails for a field.
// In this case, the field that is required but not specified is the one out
// of bounds of the version. If a field is not set, we won't validate its version,
// so it should only result in one of those errors being captured.
func TestSchemeValidator_Validate_Error2(t *testing.T) {
	toValidate := &ConfigContext{
		Source: "<simple test config>",
		Config: &simpleTestConfig{
			SchemeVersion: SchemeVersion{Version: "3.0"},
			TestField:     "", // TestField required, will fail Validate()
		},
	}

	err := Validator.Validate(toValidate)
	assert.Error(t, err.Err())
	assert.Equal(t, 1, len(err.Errors), err.Error())

	// check that validation cleanup was successful
	checkValidationCleanup(t)
}

// TestSchemeValidator_Validate_Simple_BadConfigScheme tests validating a simple struct where
// the SchemeVersion specified a bad scheme version.
func TestSchemeValidator_Validate_Simple_BadConfigScheme(t *testing.T) {
	toValidate := &ConfigContext{
		Source: "<simple test config>",
		Config: &simpleTestConfig{
			SchemeVersion: SchemeVersion{Version: "xyz.xyz"}, // bad scheme version
			TestField:     "foo",
		},
	}

	err := Validator.Validate(toValidate)
	assert.Error(t, err.Err())
	assert.Equal(t, 1, len(err.Errors), err.Error())

	// check that validation cleanup was successful
	checkValidationCleanup(t)
}

// TestSchemeValidator_Validate_Simple_BadAddedInTag tests validating a simple struct where
// a field specifies a bad "addedIn" tag.
func TestSchemeValidator_Validate_Simple_BadAddedInTag(t *testing.T) {
	toValidate := &ConfigContext{
		Source: "<simple test config>",
		Config: &simpleTestConfig{
			SchemeVersion: SchemeVersion{Version: "1.0"},
			TestField:     "foo",
			BadAddedInTag: "bar", // this field has a bad addedIn tag in the struct definition
		},
	}

	err := Validator.Validate(toValidate)
	assert.Error(t, err.Err())
	assert.Equal(t, 1, len(err.Errors), err.Error())

	// check that validation cleanup was successful
	checkValidationCleanup(t)
}

// TestSchemeValidator_Validate_Simple_BadDeprecatedInTag tests validating a simple struct where
// a field specifies a bad "deprecatedIn" tag.
func TestSchemeValidator_Validate_Simple_BadDeprecatedInTag(t *testing.T) {
	toValidate := &ConfigContext{
		Source: "<simple test config>",
		Config: &simpleTestConfig{
			SchemeVersion:      SchemeVersion{Version: "1.0"},
			TestField:          "foo",
			BadDeprecatedInTag: "bar", // this field has a bad deprecatedIn tag in the struct definition
		},
	}

	err := Validator.Validate(toValidate)
	assert.Error(t, err.Err())
	assert.Equal(t, 1, len(err.Errors), err.Error())

	// check that validation cleanup was successful
	checkValidationCleanup(t)
}

// TestSchemeValidator_Validate_Simple_BadRemovedInTag tests validating a simple struct where
// a field specifies a bad "removedIn" tag.
func TestSchemeValidator_Validate_Simple_BadRemovedInTag(t *testing.T) {
	toValidate := &ConfigContext{
		Source: "<simple test config>",
		Config: &simpleTestConfig{
			SchemeVersion:   SchemeVersion{Version: "1.0"},
			TestField:       "foo",
			BadRemovedInTag: "bar", // this field has a bad removedIn tag in the struct definition
		},
	}

	err := Validator.Validate(toValidate)
	assert.Error(t, err.Err())
	assert.Equal(t, 1, len(err.Errors), err.Error())

	// check that validation cleanup was successful
	checkValidationCleanup(t)
}

// TestSchemeValidator_validate passes a bad value to the validate function, so we expect
// it to fail.
func TestSchemeValidator_validate(t *testing.T) {
	defer Validator.clearState()

	// Validate expects either an interface, pointer, or struct. If an interface
	// or pointer, it should ultimately resolve down to a struct. Here we will
	// give a pointer to a slice.
	x := []string{"foo"}

	// Since all of the setup is done in Validate (the exported function), we need to
	// ensure we have all the pieces we need here manually.
	multiErr := errors.NewMultiError("<validate test>")
	Validator.errors = multiErr
	Validator.context = NewConfigContext("test", &simpleTestConfig{})

	Validator.validate(x)

	assert.Error(t, multiErr.Err())
	assert.Equal(t, 1, len(multiErr.Errors), multiErr.Error())
}

// TestSchemeValidator_Validate_Complex_Ok tests validating a complex struct where everything is ok.
func TestSchemeValidator_Validate_Complex_Ok(t *testing.T) {
	toValidate := &ConfigContext{
		Source: "<complex test config>",
		Config: &complexTestConfig{
			SchemeVersion: SchemeVersion{Version: "1.0"},
			Foo:           true,
			FloatVal:      20,
			IntVal:        3,
			UintVal:       2,
			RootUser: &nestedStruct{
				User: "admin",
				Pass: "admin",
			},
		},
	}

	err := Validator.Validate(toValidate)
	assert.NoError(t, err.Err())

	// check that validation cleanup was successful
	checkValidationCleanup(t)
}

// TestSchemeValidator_Validate_Complex_Ok2 tests validating a complex struct where everything is ok,
// but some values are specified as the zero value for that type.
func TestSchemeValidator_Validate_Complex_Ok2(t *testing.T) {
	toValidate := &ConfigContext{
		Source: "<complex test config>",
		Config: &complexTestConfig{
			SchemeVersion: SchemeVersion{Version: "1.0"},
			Foo:           false,
			FloatVal:      0,
			IntVal:        0,
			UintVal:       0,
			RootUser: &nestedStruct{
				User: "admin",
				Pass: "admin",
			},
		},
	}

	err := Validator.Validate(toValidate)
	assert.NoError(t, err.Err())

	// check that validation cleanup was successful
	checkValidationCleanup(t)
}

// TestSchemeValidator_Validate_Complex_Error tests validating a complex struct where
// there are errors due to SchemeVersion mismatches. In this case, it is for fields
// specified in a version before they are supported.
func TestSchemeValidator_Validate_Complex_Error(t *testing.T) {
	toValidate := &ConfigContext{
		Source: "<complex test config>",
		Config: &complexTestConfig{
			SchemeVersion: SchemeVersion{Version: "1.0"},
			Simples: []simpleTestConfig{
				{
					SchemeVersion: SchemeVersion{Version: "1.0"},
					TestField:     "foo",
				},
			},
			Foo:      true,
			FloatVal: 20,
			IntVal:   3,
			UintVal:  2,
			RootUser: &nestedStruct{
				User: "admin",
				Pass: "admin",
			},
			Users: []*nestedStruct{
				{
					User: "other",
					Pass: "foobar",
				},
			},
		},
	}

	err := Validator.Validate(toValidate)
	assert.Error(t, err.Err())
	assert.Equal(t, 2, len(err.Errors), err.Error())

	// check that validation cleanup was successful
	checkValidationCleanup(t)
}

// TestSchemeValidator_Validate_Complex_Error2 tests validating a complex struct where
// there are errors due to SchemeVersion mismatches. In this case, it is for fields
// specified in a version after they were removed.
func TestSchemeValidator_Validate_Complex_Error2(t *testing.T) {
	toValidate := &ConfigContext{
		Source: "<complex test config>",
		Config: &complexTestConfig{
			SchemeVersion: SchemeVersion{Version: "3.0"},
			Simples: []simpleTestConfig{
				{
					SchemeVersion: SchemeVersion{Version: "3.0"},
					TestField:     "foo",
				},
			},
			Foo:      true,
			FloatVal: 20,
			IntVal:   3,
			UintVal:  2,
			RootUser: &nestedStruct{
				User: "admin",
				Pass: "admin",
			},
			Users: []*nestedStruct{
				{
					User: "other",
					Pass: "foobar",
				},
			},
		},
	}

	err := Validator.Validate(toValidate)
	assert.Error(t, err.Err())
	assert.Equal(t, 3, len(err.Errors), err.Error())

	// check that validation cleanup was successful
	checkValidationCleanup(t)
}

// TestSchemeValidator_Validate_Complex_Error3 tests validating a complex struct where
// there are errors due to SchemeVersion mismatches. In this case, it is for ConfigComponent
// validation errors.
func TestSchemeValidator_Validate_Complex_Error3(t *testing.T) {
	toValidate := &ConfigContext{
		Source: "<complex test config>",
		Config: &complexTestConfig{
			SchemeVersion: SchemeVersion{Version: "1.5"},
			Simples: []simpleTestConfig{
				{
					SchemeVersion: SchemeVersion{Version: "1.0"},
					TestField:     "", // error #1
				},
			},
			FloatVal: 20,
			IntVal:   3,
			UintVal:  9, // error #2
			RootUser: &nestedStruct{
				User: "", // error #3
				Pass: "admin",
			},
			Users: []*nestedStruct{
				{
					User: "", // error #4
					Pass: "foobar",
				},
			},
		},
	}

	err := Validator.Validate(toValidate)
	assert.Error(t, err.Err())
	assert.Equal(t, 4, len(err.Errors), err.Error())

	// check that validation cleanup was successful
	checkValidationCleanup(t)
}

// TestValidateReadRequest tests validating a Read request successfully.
func TestValidateReadRequest(t *testing.T) {
	request := &synse.DeviceFilter{
		Device: "device",
		Board:  "board",
		Rack:   "rack",
	}

	err := validateDeviceFilter(request)
	assert.NoError(t, err)
}

// TestValidateReadRequestErr tests validating a Read request when the validation
// should fail and cause an error.
func TestValidateReadRequestErr(t *testing.T) {
	var cases = []synse.DeviceFilter{
		{
			// missing device
			Board: "board",
			Rack:  "rack",
		},
		{
			// missing board
			Device: "device",
			Rack:   "rack",
		},
		{
			// missing rack
			Device: "device",
			Board:  "board",
		},
		{
			// missing all
		},
	}

	for _, testCase := range cases {
		err := validateDeviceFilter(&testCase)
		assert.Error(t, err)
	}
}

// TestValidateWriteRequest tests validating a Write request successfully.
func TestValidateWriteRequest(t *testing.T) {
	request := &synse.WriteInfo{
		DeviceFilter: &synse.DeviceFilter{
			Device: "device",
			Board:  "board",
			Rack:   "rack",
		},
	}

	err := validateWriteInfo(request)
	assert.NoError(t, err)
}

// TestValidateWriteRequestErr tests validating a Write request when the validation
// should fail and cause an error.
func TestValidateWriteRequestErr(t *testing.T) {
	var cases = []synse.WriteInfo{
		{
			// missing device
			DeviceFilter: &synse.DeviceFilter{
				Board: "board",
				Rack:  "rack",
			},
		},
		{
			// missing board
			DeviceFilter: &synse.DeviceFilter{
				Device: "device",
				Rack:   "rack",
			},
		},
		{
			// missing rack
			DeviceFilter: &synse.DeviceFilter{
				Device: "device",
				Board:  "board",
			},
		},
		{
			// missing all
			DeviceFilter: &synse.DeviceFilter{},
		},
	}

	for _, testCase := range cases {
		err := validateWriteInfo(&testCase)
		assert.Error(t, err)
	}
}

// TestValidateForRead_1 tests validating a device for read, when the specified
// device is not in the device map.
func TestValidateForRead_1(t *testing.T) {
	defer resetContext()

	err := validateForRead("foo")
	assert.Error(t, err)
}

// TestValidateForRead_2 tests validating a device for read, when no read handler
// is defined for the device.
func TestValidateForRead_2(t *testing.T) {
	defer resetContext()

	ctx.devices["abc"] = &Device{
		Handler: &DeviceHandler{},
	}

	err := validateForRead("abc")
	assert.Error(t, err)
}

// TestValidateForRead_3 tests validating a device for read, when it does exist
// in the device map and does have a read handler defined.
func TestValidateForRead_3(t *testing.T) {
	defer resetContext()

	ctx.devices["abc"] = &Device{
		Handler: &DeviceHandler{
			Read: func(d *Device) ([]*Reading, error) { return nil, nil },
		},
	}

	err := validateForRead("abc")
	assert.NoError(t, err)
}

// TestValidateForWrite_1 tests validating a device for write, when the specified
// device is not in the device map.
func TestValidateForWrite_1(t *testing.T) {
	defer resetContext()

	err := validateForWrite("foo")
	assert.Error(t, err)
}

// TestValidateForWrite_2 tests validating a device for write, when no write handler
// is defined for the device.
func TestValidateForWrite_2(t *testing.T) {
	defer resetContext()

	ctx.devices["abc"] = &Device{
		Handler: &DeviceHandler{},
	}

	err := validateForWrite("abc")
	assert.Error(t, err)
}

// TestValidateForWrite_3 tests validating a device for write, when it does exist
// in the device map and does have a write handler defined.
func TestValidateForWrite_3(t *testing.T) {
	defer resetContext()

	ctx.devices["abc"] = &Device{
		Handler: &DeviceHandler{
			Write: func(d *Device, data *WriteData) error { return nil },
		},
	}

	err := validateForWrite("abc")
	assert.NoError(t, err)
}

package oldconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-server-grpc/go"
)

// TestOutputType_Type tests getting the type of the reading
// from the namespaced OutputType name.
func TestOutputType_Type(t *testing.T) {
	var testTable = []struct {
		name     string
		expected string
	}{
		{
			name:     "foo",
			expected: "foo",
		},
		{
			name:     "foo.bar",
			expected: "bar",
		},
		{
			name:     "test.device.sample.temperature",
			expected: "temperature",
		},
	}

	for _, tc := range testTable {
		readingType := OutputType{Name: tc.name}
		assert.Equal(t, tc.expected, readingType.Type())
	}
}

// TestOutputType_Validate_Ok tests validating the OutputType when there are no errors.
func TestOutputType_Validate_Ok(t *testing.T) {
	var testTable = []struct {
		desc   string
		output OutputType
	}{
		{
			desc: "Valid OutputType instance",
			output: OutputType{
				Name: "test",
			},
		},
	}

	for _, testCase := range testTable {
		merr := errors.NewMultiError("test")

		testCase.output.Validate(merr)
		assert.NoError(t, merr.Err(), testCase.desc)
	}
}

// TestOutputType_Validate_Error tests validating the OutputType when there are errors.
func TestOutputType_Validate_Error(t *testing.T) {
	var testTable = []struct {
		desc     string
		errCount int
		output   OutputType
	}{
		{
			desc:     "OutputType requires name, but has none",
			errCount: 1,
			output:   OutputType{},
		},
		{
			desc:     "OutputType has an invalid scaling factor",
			errCount: 1,
			output: OutputType{
				Name:          "test",
				ScalingFactor: "invalid factor",
			},
		},
		{
			desc:     "OutputType has an invalid scaling factor and no name",
			errCount: 2,
			output: OutputType{
				ScalingFactor: "invalid factor",
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

// TestOutputType_GetScalingFactor_Ok tests getting the scaling factor from the OutputType successfully.
func TestOutputType_GetScalingFactor_Ok(t *testing.T) {
	var testTable = []struct {
		desc     string
		output   OutputType
		expected float64
	}{
		{
			desc:     "no scaling factor set, get default",
			output:   OutputType{},
			expected: 1,
		},
		{
			desc: "scaling factor is positive integer",
			output: OutputType{
				ScalingFactor: "2",
			},
			expected: 2,
		},
		{
			desc: "scaling factor is positive integer with sign",
			output: OutputType{
				ScalingFactor: "+2",
			},
			expected: 2,
		},
		{
			desc: "scaling factor is negative integer",
			output: OutputType{
				ScalingFactor: "-2",
			},
			expected: -2,
		},
		{
			desc: "scaling factor is positive float",
			output: OutputType{
				ScalingFactor: "2.4",
			},
			expected: 2.4,
		},
		{
			desc: "scaling factor is positive float with sign",
			output: OutputType{
				ScalingFactor: "+2.4",
			},
			expected: 2.4,
		},
		{
			desc: "scaling factor is negative float",
			output: OutputType{
				ScalingFactor: "-2.4",
			},
			expected: -2.4,
		},
		{
			desc: "scaling factor is float with positive exp",
			output: OutputType{
				ScalingFactor: "2.4e2",
			},
			expected: 240,
		},
		{
			desc: "scaling factor is float with negative exp",
			output: OutputType{
				ScalingFactor: "2.4e-2",
			},
			expected: 0.024,
		},
		{
			desc: "scaling factor is negative int with positive exp",
			output: OutputType{
				ScalingFactor: "-3E2",
			},
			expected: -300,
		},
		{
			desc: "scaling factor is negative int with negative exp",
			output: OutputType{
				ScalingFactor: "-3e-3",
			},
			expected: -0.003,
		},
		{
			desc: "scaling factor is decimal with no leading zero",
			output: OutputType{
				ScalingFactor: ".3e2",
			},
			expected: 30,
		},
		{
			desc: "scaling factor is decimal with no leading zero and sign",
			output: OutputType{
				ScalingFactor: "+.3e2",
			},
			expected: 30,
		},
	}

	for _, testCase := range testTable {
		sf, err := testCase.output.GetScalingFactor()
		assert.NoError(t, err, testCase.desc)
		assert.Equal(t, testCase.expected, sf, testCase.desc)

	}
}

// TestOutputType_GetScalingFactor_Error tests getting the scaling factor from the OutputType successfully.
func TestOutputType_GetScalingFactor_Error(t *testing.T) {
	var testTable = []struct {
		desc   string
		output OutputType
	}{
		{
			desc: "invalid format",
			output: OutputType{
				ScalingFactor: "+ 0.0 E 3",
			},
		},
		{
			desc: "invalid string",
			output: OutputType{
				ScalingFactor: "foobar",
			},
		},
		{
			desc: "additional decimal",
			output: OutputType{
				ScalingFactor: "+0.124.2e4",
			},
		},
	}

	for _, testCase := range testTable {
		sf, err := testCase.output.GetScalingFactor()
		assert.Error(t, err, testCase.desc)
		assert.Zero(t, sf, testCase.desc)
	}
}

// TestOutputType_Apply tests applying the output transformations to the output value.
func TestOutputType_Apply(t *testing.T) {
	var testTable = []struct {
		desc     string
		output   OutputType
		value    interface{}
		expected interface{}
	}{
		{
			desc: "multiply factor by value 0",
			output: OutputType{
				ScalingFactor: "3",
			},
			value:    0,
			expected: 0,
		},
		{
			desc: "multiply factor by value 1",
			output: OutputType{
				ScalingFactor: "3",
			},
			value:    1,
			expected: 3,
		},
		{
			desc: "multiply value by 0.5 factor",
			output: OutputType{
				ScalingFactor: "0.5",
			},
			value:    float64(1),
			expected: 0.5,
		},
		{
			desc: "value is a float64, factor is > 1",
			output: OutputType{
				ScalingFactor: "2",
			},
			value:    float64(3),
			expected: float64(6),
		},
		{
			desc: "value is a float64, factor is < 1",
			output: OutputType{
				ScalingFactor: "0.5",
			},
			value:    float64(3),
			expected: float64(1.5),
		},
		{
			desc: "value is a float32, factor is > 1",
			output: OutputType{
				ScalingFactor: "2",
			},
			value:    float32(3),
			expected: float32(6),
		},
		{
			desc: "value is a float32, factor is < 1",
			output: OutputType{
				ScalingFactor: "0.5",
			},
			value:    float32(3),
			expected: float32(1.5),
		},
		{
			desc: "value is a int64, factor is > 1",
			output: OutputType{
				ScalingFactor: "2",
			},
			value:    int64(3),
			expected: int64(6),
		},
		{
			desc: "value is a int64, factor is < 1",
			output: OutputType{
				ScalingFactor: "0.5",
			},
			value:    int64(3),
			expected: int64(1),
		},
		{
			desc: "value is a int32, factor is > 1",
			output: OutputType{
				ScalingFactor: "2",
			},
			value:    int32(3),
			expected: int32(6),
		},
		{
			desc: "value is a int32, factor is < 1",
			output: OutputType{
				ScalingFactor: "0.5",
			},
			value:    int32(3),
			expected: int32(1),
		},
		{
			desc: "value is a int16, factor is > 1",
			output: OutputType{
				ScalingFactor: "2",
			},
			value:    int16(3),
			expected: int16(6),
		},
		{
			desc: "value is a int16, factor is < 1",
			output: OutputType{
				ScalingFactor: "0.5",
			},
			value:    int16(3),
			expected: int16(1),
		},
		{
			desc: "value is a int8, factor is > 1",
			output: OutputType{
				ScalingFactor: "2",
			},
			value:    int8(3),
			expected: int8(6),
		},
		{
			desc: "value is a int8, factor is < 1",
			output: OutputType{
				ScalingFactor: "0.5",
			},
			value:    int8(3),
			expected: int8(1),
		},
		{
			desc: "value is a int, factor is > 1",
			output: OutputType{
				ScalingFactor: "2",
			},
			value:    int(3),
			expected: int(6),
		},
		{
			desc: "value is a int, factor is < 1",
			output: OutputType{
				ScalingFactor: "0.5",
			},
			value:    int(3),
			expected: int(1),
		},
		{
			desc: "value is a uint64, factor is > 1",
			output: OutputType{
				ScalingFactor: "2",
			},
			value:    uint64(3),
			expected: uint64(6),
		},
		{
			desc: "value is a uint64, factor is < 1",
			output: OutputType{
				ScalingFactor: "0.5",
			},
			value:    uint64(3),
			expected: uint64(1),
		},
		{
			desc: "value is a uint32, factor is > 1",
			output: OutputType{
				ScalingFactor: "2",
			},
			value:    uint32(3),
			expected: uint32(6),
		},
		{
			desc: "value is a uint32, factor is < 1",
			output: OutputType{
				ScalingFactor: "0.5",
			},
			value:    uint32(3),
			expected: uint32(1),
		},
		{
			desc: "value is a uint16, factor is > 1",
			output: OutputType{
				ScalingFactor: "2",
			},
			value:    uint32(3),
			expected: uint32(6),
		},
		{
			desc: "value is a uint16, factor is < 1",
			output: OutputType{
				ScalingFactor: "0.5",
			},
			value:    uint16(3),
			expected: uint16(1),
		},
		{
			desc: "value is a uint8, factor is > 1",
			output: OutputType{
				ScalingFactor: "2",
			},
			value:    uint8(3),
			expected: uint8(6),
		},
		{
			desc: "value is a uint8, factor is < 1",
			output: OutputType{
				ScalingFactor: "0.5",
			},
			value:    uint8(3),
			expected: uint8(1),
		},
		{
			desc: "value is a uint, factor is > 1",
			output: OutputType{
				ScalingFactor: "2",
			},
			value:    uint(3),
			expected: uint(6),
		},
		{
			desc: "value is a uint, factor is < 1",
			output: OutputType{
				ScalingFactor: "0.5",
			},
			value:    uint(3),
			expected: uint(1),
		},
	}

	for _, testCase := range testTable {
		actual := testCase.output.Apply(testCase.value)
		assert.Equal(t, testCase.expected, actual, testCase.desc)
	}
}

// TestOutputType_Apply_Error tests applying when the scaling factor is invalid.
func TestOutputType_Apply_Error(t *testing.T) {
	output := OutputType{
		ScalingFactor: "foobar",
	}
	actual := output.Apply(2)
	assert.Equal(t, 2, actual, "on parse failure, nothing changes")
}

// TestUnit_Validate tests validating the Unit. There is nothing to validate here,
// so it should all validate successfully.
func TestUnit_Validate(t *testing.T) {
	merr := errors.NewMultiError("test")
	unit := Unit{}
	unit.Validate(merr)
	assert.NoError(t, merr.Err())
}

// TestUnit_Encode tests encoding the Unit to the gRPC message.
func TestUnit_Encode(t *testing.T) {
	var testTable = []struct {
		desc    string
		unit    Unit
		message synse.Unit
	}{
		{
			desc:    "empty unit",
			unit:    Unit{},
			message: synse.Unit{},
		},
		{
			desc: "name and symbol specified",
			unit: Unit{
				Name:   "foo",
				Symbol: "bar",
			},
			message: synse.Unit{
				Name:   "foo",
				Symbol: "bar",
			},
		},
		{
			desc: "only name specified",
			unit: Unit{
				Name: "foo",
			},
			message: synse.Unit{
				Name: "foo",
			},
		},
		{
			desc: "only symbol specified",
			unit: Unit{
				Symbol: "bar",
			},
			message: synse.Unit{
				Symbol: "bar",
			},
		},
	}

	for _, testCase := range testTable {
		actual := testCase.unit.Encode()
		assert.Equal(t, testCase.unit.Name, actual.GetName())
		assert.Equal(t, testCase.unit.Symbol, actual.GetSymbol())
	}
}

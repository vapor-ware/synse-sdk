// Synse SDK
// Copyright (c) 2019 Vapor IO
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package output

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- test outputs ---
var (
	// An output used for testing where the output does not
	// have any unit or system.
	testOutput1 = Output{
		Name: "output1",
		Type: "testoutput",
	}

	// An output used for testing where the output has a
	// system agnostic unit.
	testOutput2 = Output{
		Name: "output2",
		Type: "testoutput",
		Units: map[SystemOfMeasure]*Unit{
			NONE: {
				Name:   "foo",
				Symbol: "FOO",
				System: string(NONE),
			},
		},
	}

	// An output used for testing where the output has units
	// for different systems.
	testOutput3 = Output{
		Name:      "output3",
		Type:      "testoutput",
		Precision: 3,
		Units: map[SystemOfMeasure]*Unit{
			METRIC: {
				Name:   "met",
				Symbol: "MET",
				System: string(METRIC),
			},
			IMPERIAL: {
				Name:   "imp",
				Symbol: "IMP",
				System: string(IMPERIAL),
			},
		},
		Converters: map[SystemOfMeasure]func(value interface{}, to SystemOfMeasure) (interface{}, error){
			METRIC: func(value interface{}, to SystemOfMeasure) (i interface{}, e error) {
				switch to {
				case METRIC:
					return value, nil
				case IMPERIAL:
					return value.(int) * 2, nil
				default:
					return nil, fmt.Errorf("default err")
				}
			},
			IMPERIAL: func(value interface{}, to SystemOfMeasure) (i interface{}, e error) {
				switch to {
				case METRIC:
					return value.(int) / 2, nil
				case IMPERIAL:
					return value, nil
				default:
					return nil, fmt.Errorf("default err")
				}
			},
		},
	}
)

// --- test cases ---

func TestOutput_FromImperial_output1(t *testing.T) {
	r := testOutput1.FromImperial(2)

	assert.Equal(t, NONE, r.System)
	assert.Equal(t, "testoutput", r.Type)
	assert.Equal(t, "", r.Info)
	assert.Equal(t, 2, r.Value)
	assert.Nil(t, r.Unit)
	assert.NotEmpty(t, r.Timestamp)
}

func TestOutput_FromImperial_output2(t *testing.T) {
	r := testOutput2.FromImperial(2)

	assert.Equal(t, NONE, r.System)
	assert.Equal(t, "testoutput", r.Type)
	assert.Equal(t, "", r.Info)
	assert.Equal(t, 2, r.Value)
	assert.Equal(t, "foo", r.Unit.Name)
	assert.Equal(t, "FOO", r.Unit.Symbol)
	assert.NotEmpty(t, r.Timestamp)
}

func TestOutput_FromImperial_output3(t *testing.T) {
	r := testOutput3.FromImperial(2)

	assert.Equal(t, IMPERIAL, r.System)
	assert.Equal(t, "testoutput", r.Type)
	assert.Equal(t, "", r.Info)
	assert.Equal(t, 2, r.Value)
	assert.Equal(t, "imp", r.Unit.Name)
	assert.Equal(t, "IMP", r.Unit.Symbol)
	assert.NotEmpty(t, r.Timestamp)
}

func TestOutput_FromMetric_output1(t *testing.T) {
	r := testOutput1.FromMetric(2)

	assert.Equal(t, NONE, r.System)
	assert.Equal(t, "testoutput", r.Type)
	assert.Equal(t, "", r.Info)
	assert.Equal(t, 2, r.Value)
	assert.Nil(t, r.Unit)
	assert.NotEmpty(t, r.Timestamp)
}

func TestOutput_FromMetric_output2(t *testing.T) {
	r := testOutput2.FromMetric(2)

	assert.Equal(t, NONE, r.System)
	assert.Equal(t, "testoutput", r.Type)
	assert.Equal(t, "", r.Info)
	assert.Equal(t, 2, r.Value)
	assert.Equal(t, "foo", r.Unit.Name)
	assert.Equal(t, "FOO", r.Unit.Symbol)
	assert.NotEmpty(t, r.Timestamp)
}

func TestOutput_FromMetric_output3(t *testing.T) {
	r := testOutput3.FromMetric(2)

	assert.Equal(t, METRIC, r.System)
	assert.Equal(t, "testoutput", r.Type)
	assert.Equal(t, "", r.Info)
	assert.Equal(t, 2, r.Value)
	assert.Equal(t, "met", r.Unit.Name)
	assert.Equal(t, "MET", r.Unit.Symbol)
	assert.NotEmpty(t, r.Timestamp)
}

func TestOutput_From_output1(t *testing.T) {
	r := testOutput1.From(2)

	assert.Equal(t, NONE, r.System)
	assert.Equal(t, "testoutput", r.Type)
	assert.Equal(t, "", r.Info)
	assert.Equal(t, 2, r.Value)
	assert.Nil(t, r.Unit)
	assert.NotEmpty(t, r.Timestamp)
}

func TestOutput_From_output2(t *testing.T) {
	r := testOutput2.From(2)

	assert.Equal(t, NONE, r.System)
	assert.Equal(t, "testoutput", r.Type)
	assert.Equal(t, "", r.Info)
	assert.Equal(t, 2, r.Value)
	assert.Equal(t, "foo", r.Unit.Name)
	assert.Equal(t, "FOO", r.Unit.Symbol)
	assert.NotEmpty(t, r.Timestamp)
}

func TestOutput_From_output3(t *testing.T) {
	r := testOutput3.From(2)

	assert.Equal(t, NONE, r.System)
	assert.Equal(t, "testoutput", r.Type)
	assert.Equal(t, "", r.Info)
	assert.Equal(t, 2, r.Value)
	assert.Nil(t, r.Unit)
	assert.NotEmpty(t, r.Timestamp)
}

func TestOutput_Convert_noUnits(t *testing.T) {
	// Test converting with no units, this should not do anything to the
	// input value.
	val, err := testOutput1.Convert(4, METRIC, IMPERIAL)
	assert.NoError(t, err)
	assert.Equal(t, 4, val)

	val, err = testOutput1.Convert(4.435, IMPERIAL, METRIC)
	assert.NoError(t, err)
	assert.Equal(t, 4.435, val)
}

func TestOutput_Convert_systemAgnostic(t *testing.T) {
	// Test converting with system agnostic units. This should do nothing to
	// input value.
	val, err := testOutput2.Convert(4, METRIC, IMPERIAL)
	assert.NoError(t, err)
	assert.Equal(t, 4, val)

	val, err = testOutput2.Convert(4.435, IMPERIAL, METRIC)
	assert.NoError(t, err)
	assert.Equal(t, 4.435, val)
}

func TestOutput_Convert_convertersNotDefined(t *testing.T) {
	// Test converting when the output is defined incorrectly and does not
	// specify the needed converter function.
	output := Output{
		Name: "invalid",
		Units: map[SystemOfMeasure]*Unit{
			METRIC:   {Name: "foo"},
			IMPERIAL: {Name: "bar"},
		},
		// no converters defined
	}

	val, err := output.Convert(4, METRIC, IMPERIAL)
	assert.Error(t, err)
	assert.Nil(t, val)
}

func TestOutput_Convert_converterNotExist(t *testing.T) {
	// Test converting when the specified converter doesn't exist.
	val, err := testOutput3.Convert(4, SystemOfMeasure("foo"), IMPERIAL)
	assert.Error(t, err)
	assert.Nil(t, val)
}

func TestOutput_Convert_convertOk(t *testing.T) {
	// Test converting with no error.
	val, err := testOutput3.Convert(4, METRIC, IMPERIAL)
	assert.NoError(t, err)
	assert.Equal(t, 8, val)

	val, err = testOutput3.Convert(4, IMPERIAL, METRIC)
	assert.NoError(t, err)
	assert.Equal(t, 2, val)

	val, err = testOutput3.Convert(4, IMPERIAL, IMPERIAL)
	assert.NoError(t, err)
	assert.Equal(t, 4, val)

	val, err = testOutput3.Convert(4, METRIC, METRIC)
	assert.NoError(t, err)
	assert.Equal(t, 4, val)
}

func TestOutput_Convert_convertErr(t *testing.T) {
	// Test converting when the converter errors.
	output := Output{
		Name: "invalid",
		Units: map[SystemOfMeasure]*Unit{
			METRIC:   {Name: "foo"},
			IMPERIAL: {Name: "bar"},
		},
		Converters: map[SystemOfMeasure]func(value interface{}, to SystemOfMeasure) (interface{}, error){
			METRIC: func(value interface{}, to SystemOfMeasure) (i interface{}, e error) {
				return nil, fmt.Errorf("test error")
			},
			IMPERIAL: func(value interface{}, to SystemOfMeasure) (i interface{}, e error) {
				return nil, fmt.Errorf("test error")
			},
		},
	}

	val, err := output.Convert(4, IMPERIAL, METRIC)
	assert.Error(t, err)
	assert.Nil(t, val)

	val, err = output.Convert(4, METRIC, IMPERIAL)
	assert.Error(t, err)
	assert.Nil(t, val)
}

func TestOutput_GetUnit_noUnits(t *testing.T) {
	// Test getting a unit from an output which does not have a unit.
	unit, err := testOutput1.GetUnit(IMPERIAL)
	assert.NoError(t, err)
	assert.Nil(t, unit)

	unit, err = testOutput1.GetUnit(METRIC)
	assert.NoError(t, err)
	assert.Nil(t, unit)

	unit, err = testOutput1.GetUnit(NONE)
	assert.NoError(t, err)
	assert.Nil(t, unit)
}

func TestOutput_GetUnit_systemAgnosticUnit(t *testing.T) {
	// Test getting a unit from an output which specifies a system agnostic unit.
	unit, err := testOutput2.GetUnit(IMPERIAL)
	assert.NoError(t, err)
	assert.NotNil(t, unit)
	assert.Equal(t, string(NONE), unit.System)

	unit, err = testOutput2.GetUnit(METRIC)
	assert.NoError(t, err)
	assert.NotNil(t, unit)
	assert.Equal(t, string(NONE), unit.System)

	unit, err = testOutput2.GetUnit(NONE)
	assert.NoError(t, err)
	assert.NotNil(t, unit)
	assert.Equal(t, string(NONE), unit.System)
}

func TestOutput_GetUnit_unitNotExist(t *testing.T) {
	// Test getting a unit which does not exist for the output.
	unit, err := testOutput3.GetUnit(NONE)
	assert.Error(t, err)
	assert.Nil(t, unit)
}

func TestOutput_GetUnit_unitExists(t *testing.T) {
	// Test getting a unit which does exist for the output.
	unit, err := testOutput3.GetUnit(IMPERIAL)
	assert.NoError(t, err)
	assert.NotNil(t, unit)
	assert.Equal(t, string(IMPERIAL), unit.System)

	unit, err = testOutput3.GetUnit(METRIC)
	assert.NoError(t, err)
	assert.NotNil(t, unit)
	assert.Equal(t, string(METRIC), unit.System)
}

func TestUnit_Encode(t *testing.T) {
	u := Unit{
		Name:   "fahrenheit",
		Symbol: "F",
		System: "imperial",
	}

	encoded := u.Encode()
	assert.Equal(t, u.Name, encoded.Name)
	assert.Equal(t, u.Symbol, encoded.Symbol)
	assert.Equal(t, u.System, encoded.System)
}

func TestUnit_Encode_empty(t *testing.T) {
	u := Unit{}

	encoded := u.Encode()
	assert.Empty(t, encoded.Name)
	assert.Empty(t, encoded.Symbol)
	assert.Empty(t, encoded.System)
}

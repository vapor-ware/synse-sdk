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

//
// For test output definitions, see output_test.go
//

func TestReading_To_noUnit(t *testing.T) {
	r := Reading{
		output: &testOutput1,
		System: NONE,
		Value:  "abc",
	}

	reading, err := r.To(METRIC)
	assert.NoError(t, err)
	assert.Equal(t, "abc", reading.Value)
	assert.Equal(t, NONE, reading.System)
	assert.Nil(t, reading.Unit)
}

func TestReading_To_systemAgnostic(t *testing.T) {
	r := Reading{
		output: &testOutput2,
		System: NONE,
		Value:  "abc",
	}

	reading, err := r.To(METRIC)
	assert.NoError(t, err)
	assert.Equal(t, "abc", reading.Value)
	assert.Equal(t, NONE, reading.System)
	assert.Equal(t, "foo", reading.Unit.Name)
	assert.Equal(t, "FOO", reading.Unit.Symbol)
}

func TestReading_To_badConversion(t *testing.T) {
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

	r := Reading{
		output: &output,
		System: METRIC,
		Value:  "abc",
	}

	reading, err := r.To(IMPERIAL)
	assert.Error(t, err)
	assert.Nil(t, reading)
}

func TestReading_To_badUnit(t *testing.T) {
	output := Output{
		Name: "invalid",
		Units: map[SystemOfMeasure]*Unit{
			METRIC: {Name: "foo"},
		},
		Converters: map[SystemOfMeasure]func(value interface{}, to SystemOfMeasure) (interface{}, error){
			METRIC: func(value interface{}, to SystemOfMeasure) (i interface{}, e error) {
				return value, nil
			},
			IMPERIAL: func(value interface{}, to SystemOfMeasure) (i interface{}, e error) {
				return value, nil
			},
		},
	}

	r := Reading{
		output: &output,
		System: METRIC,
		Value:  "abc",
	}

	reading, err := r.To(IMPERIAL)
	assert.Error(t, err)
	assert.Nil(t, reading)
}

func TestReading_To(t *testing.T) {
	r := Reading{
		output: &testOutput3,
		System: METRIC,
		Value:  4,
	}

	reading, err := r.To(IMPERIAL)
	assert.NoError(t, err)
	assert.Equal(t, 8, reading.Value)
	assert.Equal(t, IMPERIAL, reading.System)
	assert.Equal(t, "imp", reading.Unit.Name)
	assert.Equal(t, "IMP", reading.Unit.Symbol)
}

func TestReading_To_noSystemSpecified(t *testing.T) {
	// Do not specify a system, defaults to METRIC
	r := Reading{
		output: &testOutput3,
		System: METRIC,
		Value:  4,
	}

	reading, err := r.To("")
	assert.NoError(t, err)
	assert.Equal(t, 4, reading.Value)
	assert.Equal(t, METRIC, reading.System)
	assert.Equal(t, "met", reading.Unit.Name)
	assert.Equal(t, "MET", reading.Unit.Symbol)
}

func TestReading_Encode(t *testing.T) {
	cases := []struct {
		value interface{}
	}{
		{"abc"},
		{[]byte("abc")},
		{true},
		{float64(3.1)},
		{float32(3.1)},
		{int64(3)},
		{int32(3)},
		{int16(3)},
		{int8(3)},
		{int(3)},
		{uint64(3)},
		{uint32(3)},
		{uint16(3)},
		{uint8(3)},
		{uint(3)},
		{nil},
	}

	for _, c := range cases {
		r := Reading{
			Timestamp: "now",
			Type:      "testtype",
			Info:      "foo",
			Value:     c.value,
		}

		encoded := r.Encode()
		// todo: check the value.. since its part of the gRPC oneOf, its hard
		//  to get at the actual value..
		assert.Equal(t, "now", encoded.Timestamp)
		assert.Equal(t, "testtype", encoded.Type)
		assert.Equal(t, "", encoded.Unit.System)
		assert.Equal(t, "", encoded.Unit.Name)
		assert.Equal(t, "", encoded.Unit.Symbol)
	}
}

func TestReading_Encode2(t *testing.T) {
	// encode with a unit specified
	r := Reading{
		Timestamp: "now",
		Type:      "testtype",
		Info:      "foo",
		Value:     123,
		Unit: &Unit{
			Name:   "unit",
			Symbol: "u",
			System: string(NONE),
		},
	}

	encoded := r.Encode()
	// todo: check the value.. since its part of the gRPC oneOf, its hard
	//  to get at the actual value..
	assert.Equal(t, "now", encoded.Timestamp)
	assert.Equal(t, "testtype", encoded.Type)
	assert.Equal(t, "none", encoded.Unit.System)
	assert.Equal(t, "unit", encoded.Unit.Name)
	assert.Equal(t, "u", encoded.Unit.Symbol)
}

func TestReading_Encode_error(t *testing.T) {
	cases := []struct {
		value interface{}
	}{
		{map[string]int{"foo": 1}},
		{[]int{1, 2}},
		{struct{}{}},
	}

	for _, c := range cases {
		r := Reading{
			Timestamp: "now",
			Type:      "testtype",
			Info:      "foo",
			Value:     c.value,
		}

		assert.Panics(t, func() {
			r.Encode()
		})
	}
}

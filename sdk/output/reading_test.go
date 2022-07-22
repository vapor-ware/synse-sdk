// Synse SDK
// Copyright (c) 2017-2022 Vapor IO
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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReading_GetOutput(t *testing.T) {
	o := &Output{
		Name: "test",
	}
	r := Reading{
		output: o,
	}

	assert.Equal(t, o, r.GetOutput())
}

func TestReading_GetOutput_noOutput(t *testing.T) {
	r := Reading{}
	assert.Nil(t, r.GetOutput())
}

func TestReading_WithContext_noContext(t *testing.T) {
	r := Reading{}
	r.WithContext(map[string]string{})

	assert.Empty(t, r.Context)
}

func TestReading_WithContext_newContext(t *testing.T) {
	r := Reading{}
	r.WithContext(map[string]string{"foo": "bar"})

	assert.Equal(t, map[string]string{"foo": "bar"}, r.Context)
}

func TestReading_WithContext_noOverride(t *testing.T) {
	r := Reading{
		Context: map[string]string{"abc": "def"},
	}
	r.WithContext(map[string]string{"123": "456"})

	assert.Equal(t, map[string]string{"abc": "def", "123": "456"}, r.Context)
}

func TestReading_WithContext_withOverride(t *testing.T) {
	r := Reading{
		Context: map[string]string{"abc": "def"},
	}

	r.WithContext(map[string]string{"abc": "456"})

	assert.Equal(t, map[string]string{"abc": "456"}, r.Context)
}

func TestReading_Scale(t *testing.T) {
	cases := []struct {
		value    interface{}
		scale    float64
		expected interface{}
	}{
		{
			value:    2,
			scale:    1,
			expected: int(2),
		},
		{
			value:    10,
			scale:    2,
			expected: float64(20),
		},
		{
			value:    10,
			scale:    0.5,
			expected: float64(5),
		},
		{
			value:    100000,
			scale:    0.0001,
			expected: float64(10),
		},
	}

	for i, c := range cases {
		r := Reading{
			Value: c.value,
		}
		err := r.Scale(c.scale)
		assert.NoError(t, err, "test case: %d", i)
		assert.Equal(t, c.expected, r.Value, "test case: %d", i)
	}
}

func TestReading_Scale_err(t *testing.T) {
	cases := []struct {
		value interface{}
		scale float64
	}{
		{
			value: 2,
			scale: 0,
		},
		{
			value: "foobar",
			scale: 2,
		},
		{
			value: []int{1, 2},
			scale: 3,
		},
		{
			value: Reading{},
			scale: 3,
		},
		{
			value: &Reading{},
			scale: 3,
		},
	}

	for i, c := range cases {
		r := Reading{
			Value: c.value,
		}
		err := r.Scale(c.scale)
		assert.Error(t, err, "test case: %d", i)
	}
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
			Context:   map[string]string{"foo": "bar"},
			Value:     c.value,
		}

		encoded := r.Encode()
		// todo: check the value.. since its part of the gRPC oneOf, its hard
		//  to get at the actual value..
		assert.Equal(t, "now", encoded.Timestamp)
		assert.Equal(t, "testtype", encoded.Type)
		assert.Equal(t, "", encoded.Unit.Name)
		assert.Equal(t, "", encoded.Unit.Symbol)
		assert.Equal(t, map[string]string{"foo": "bar"}, encoded.Context)
	}
}

func TestReading_Encode2(t *testing.T) {
	// encode with a unit specified
	r := Reading{
		Timestamp: "now",
		Type:      "testtype",
		Context:   map[string]string{"foo": "bar"},
		Value:     123,
		Unit: &Unit{
			Name:   "unit",
			Symbol: "u",
		},
	}

	encoded := r.Encode()
	// todo: check the value.. since its part of the gRPC oneOf, its hard
	//  to get at the actual value..
	assert.Equal(t, "now", encoded.Timestamp)
	assert.Equal(t, "testtype", encoded.Type)
	assert.Equal(t, "unit", encoded.Unit.Name)
	assert.Equal(t, "u", encoded.Unit.Symbol)
	assert.Equal(t, map[string]string{"foo": "bar"}, encoded.Context)
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
			Context:   map[string]string{"foo": "bar"},
			Value:     c.value,
		}

		assert.Panics(t, func() {
			r.Encode()
		})
	}
}

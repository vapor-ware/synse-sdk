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

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertToFloat64(t *testing.T) {
	cases := []struct {
		value    interface{}
		expected float64
	}{
		{
			value:    float64(3.14),
			expected: float64(3.14),
		},
		{
			value:    float32(3.14),
			expected: float64(3.14),
		},
		{
			value:    int64(3),
			expected: float64(3),
		},
		{
			value:    int32(3),
			expected: float64(3),
		},
		{
			value:    int16(3),
			expected: float64(3),
		},
		{
			value:    int8(3),
			expected: float64(3),
		},
		{
			value:    int(3),
			expected: float64(3),
		},
		{
			value:    uint64(3),
			expected: float64(3),
		},
		{
			value:    uint32(3),
			expected: float64(3),
		},
		{
			value:    uint16(3),
			expected: float64(3),
		},
		{
			value:    uint8(3),
			expected: float64(3),
		},
		{
			value:    uint(3),
			expected: float64(3),
		},
		{
			value:    "3",
			expected: float64(3),
		},
		{
			value:    "3.14",
			expected: float64(3.14),
		},
	}

	for i, c := range cases {
		converted, err := ConvertToFloat64(c.value)
		assert.NoError(t, err, "case: %d", i)
		assert.InDelta(t, c.expected, converted, 0.000001, "case: %d", i)
	}
}

func TestConvertToFloat64_err(t *testing.T) {
	cases := []struct {
		value interface{}
	}{
		{struct{}{}},
		{nil},
		{[]int{}},
		{map[int]float64{}},
		{"not-a-float"},
	}

	for i, c := range cases {
		_, err := ConvertToFloat64(c.value)
		assert.Error(t, err, "case: %d", i)
	}
}

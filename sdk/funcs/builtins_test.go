// Synse SDK
// Copyright (c) 2019-2020 Vapor IO
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

package funcs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBuiltins(t *testing.T) {
	fns := GetBuiltins()
	assert.NotEmpty(t, fns)
}

func TestFtoC_Fn(t *testing.T) {
	cases := []struct {
		f float64
		c float64
	}{
		{f: -459.67, c: -273.15},
		{f: -50, c: -45.56},
		{f: -40, c: -40.00},
		{f: -30, c: -34.44},
		{f: -20, c: -28.89},
		{f: -10, c: -23.33},
		{f: 0, c: -17.78},
		{f: 10, c: -12.22},
		{f: 20, c: -6.67},
		{f: 30, c: -1.11},
		{f: 32, c: 0},
		{f: 40, c: 4.44},
		{f: 50, c: 10.00},
		{f: 60, c: 15.56},
		{f: 70, c: 21.11},
		{f: 80, c: 26.67},
		{f: 90, c: 32.22},
		{f: 100, c: 37.78},
		{f: 110, c: 43.33},
		{f: 120, c: 48.89},
		{f: 130, c: 54.44},
		{f: 140, c: 60.00},
		{f: 150, c: 65.56},
		{f: 160, c: 71.11},
		{f: 170, c: 76.67},
		{f: 180, c: 82.22},
		{f: 190, c: 87.78},
		{f: 200, c: 93.33},
		{f: 212, c: 100},
		{f: 300, c: 148.89},
		{f: 400, c: 204.44},
		{f: 500, c: 260.00},
		{f: 600, c: 315.56},
		{f: 700, c: 371.11},
		{f: 800, c: 426.67},
		{f: 900, c: 482.22},
		{f: 1000, c: 537.78},
	}

	for _, c := range cases {
		val, err := FtoC.Fn(c.f)
		assert.NoError(t, err)
		assert.InDelta(t, c.c, val, 0.01)
	}
}

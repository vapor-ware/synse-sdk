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
	"testing"

	"github.com/stretchr/testify/assert"
)

//
// See:
// - https://www.rapidtables.com/convert/temperature/celsius-to-fahrenheit-chart.html
// - https://www.rapidtables.com/convert/temperature/how-fahrenheit-to-celsius.html
//

func Test_Temperature_fromCelsius_toFahrenheit(t *testing.T) {
	cases := []struct {
		c float64
		f float64
	}{
		{c: -50, f: -58.0},
		{c: -40, f: -40.0},
		{c: -30, f: -22.0},
		{c: -20, f: -4.0},
		{c: -10, f: 14.0},
		{c: -9, f: 15.8},
		{c: -8, f: 17.6},
		{c: -7, f: 19.4},
		{c: -6, f: 21.2},
		{c: -5, f: 23.0},
		{c: -4, f: 24.8},
		{c: -3, f: 26.6},
		{c: -2, f: 28.4},
		{c: -1, f: 30.2},
		{c: 0, f: 32.0},
		{c: 1, f: 33.8},
		{c: 2, f: 35.6},
		{c: 3, f: 37.4},
		{c: 4, f: 39.2},
		{c: 5, f: 41.0},
		{c: 6, f: 42.8},
		{c: 7, f: 44.6},
		{c: 8, f: 46.4},
		{c: 9, f: 48.2},
		{c: 10, f: 50.0},
		{c: 20, f: 68.0},
		{c: 30, f: 86.0},
		{c: 40, f: 104.0},
		{c: 50, f: 122.0},
		{c: 60, f: 140.0},
		{c: 70, f: 158.0},
		{c: 80, f: 176.0},
		{c: 90, f: 194.0},
		{c: 100, f: 212.0},
		{c: 200, f: 392.0},
		{c: 300, f: 572.0},
		{c: 400, f: 752.0},
		{c: 500, f: 932.0},
		{c: 600, f: 1112.0},
		{c: 700, f: 1292.0},
		{c: 800, f: 1472.0},
		{c: 900, f: 1652.0},
		{c: 1000, f: 1832.0},
	}

	for _, c := range cases {
		val, err := fromCelsius(c.c, IMPERIAL)
		assert.NoError(t, err)
		assert.InDelta(t, c.f, val, 0.01)
	}
}

func Test_Temperature_fromCelsius_toCelsius(t *testing.T) {
	cases := []struct {
		c float64
	}{
		{-50},
		{-25},
		{-10},
		{-1.53},
		{-1},
		{0},
		{1},
		{1.53},
		{10},
		{25},
		{50},
	}

	for _, c := range cases {
		val, err := fromCelsius(c.c, METRIC)
		assert.NoError(t, err)
		assert.Equal(t, c.c, val)
	}
}

func Test_Temperature_fromCelsius_toUnknown(t *testing.T) {
	val, err := fromCelsius(20, SystemOfMeasure("unknown"))
	assert.Error(t, err)
	assert.Nil(t, val)
}

func Test_Temperature_fromCelsius_invalidValue(t *testing.T) {
	val, err := fromCelsius(struct{}{}, IMPERIAL)
	assert.Error(t, err)
	assert.Nil(t, val)
}

func Test_Temperature_fromFahrenheit_toCelsius(t *testing.T) {
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
		val, err := fromFahrenheit(c.f, METRIC)
		assert.NoError(t, err)
		assert.InDelta(t, c.c, val, 0.01)
	}
}

func Test_Temperature_fromFahrenheit_toFahrenheit(t *testing.T) {
	cases := []struct {
		f float64
	}{
		{-50},
		{-25},
		{-10},
		{-1.53},
		{-1},
		{0},
		{1},
		{1.53},
		{10},
		{25},
		{50},
	}

	for _, c := range cases {
		val, err := fromFahrenheit(c.f, IMPERIAL)
		assert.NoError(t, err)
		assert.Equal(t, c.f, val)
	}
}

func Test_Temperature_fromFahrenheit_toUnknown(t *testing.T) {
	val, err := fromFahrenheit(20, SystemOfMeasure("unknown"))
	assert.Error(t, err)
	assert.Nil(t, val)
}

func Test_Temperature_fromFahrenheit_invalidValue(t *testing.T) {
	val, err := fromFahrenheit(struct{}{}, METRIC)
	assert.Error(t, err)
	assert.Nil(t, val)
}

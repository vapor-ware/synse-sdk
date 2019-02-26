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
// - https://www.metric-conversions.org/speed/meters-per-second-to-miles-per-hour.htm#metricConversionTable
// - https://www.metric-conversions.org/speed/miles-per-hour-to-meters-per-second.htm#metricConversionTable
//

func Test_Speed_fromMetersPerSecond_toMilesPerHour(t *testing.T) {
	cases := []struct {
		ms  float64
		mph float64
	}{
		{ms: -10, mph: -22.369},
		{ms: -9, mph: -20.132},
		{ms: -8, mph: -17.895},
		{ms: -7, mph: -15.659},
		{ms: -6, mph: -13.422},
		{ms: -5, mph: -11.185},
		{ms: -4, mph: -8.948},
		{ms: -3, mph: -6.711},
		{ms: -2, mph: -4.474},
		{ms: -1, mph: -2.237},
		{ms: 0, mph: 0.0},
		{ms: 1, mph: 2.237},
		{ms: 2, mph: 4.474},
		{ms: 3, mph: 6.711},
		{ms: 4, mph: 8.948},
		{ms: 5, mph: 11.185},
		{ms: 6, mph: 13.422},
		{ms: 7, mph: 15.659},
		{ms: 8, mph: 17.895},
		{ms: 9, mph: 20.132},
		{ms: 10, mph: 22.369},
	}

	for _, c := range cases {
		val, err := fromMetersPerSecond(c.ms, IMPERIAL)
		assert.NoError(t, err)
		assert.InDelta(t, c.mph, val, 0.01)
	}
}

func Test_Speed_fromMetersPerSecond_toMetersPerSecond(t *testing.T) {
	cases := []struct {
		ms float64
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
		val, err := fromMetersPerSecond(c.ms, METRIC)
		assert.NoError(t, err)
		assert.Equal(t, c.ms, val)
	}
}

func Test_Speed_fromMetersPerSecond_toUnknown(t *testing.T) {
	val, err := fromMetersPerSecond(20, SystemOfMeasure("unknown"))
	assert.Error(t, err)
	assert.Nil(t, val)
}

func Test_Speed_fromMetersPerSecond_invalidValue(t *testing.T) {
	val, err := fromMetersPerSecond(struct{}{}, IMPERIAL)
	assert.Error(t, err)
	assert.Nil(t, val)
}

func Test_Speed_fromMilesPerHour_toMetersPerSecond(t *testing.T) {
	cases := []struct {
		mph float64
		ms  float64
	}{

		{mph: -10, ms: -4.470},
		{mph: -9, ms: -4.023},
		{mph: -8, ms: -3.576},
		{mph: -7, ms: -3.129},
		{mph: -6, ms: -2.682},
		{mph: -5, ms: -2.235},
		{mph: -4, ms: -1.788},
		{mph: -3, ms: -1.341},
		{mph: -2, ms: -0.894},
		{mph: -1, ms: -0.447},
		{mph: 0, ms: 0},
		{mph: 1, ms: 0.447},
		{mph: 2, ms: 0.894},
		{mph: 3, ms: 1.341},
		{mph: 4, ms: 1.788},
		{mph: 5, ms: 2.235},
		{mph: 6, ms: 2.682},
		{mph: 7, ms: 3.129},
		{mph: 8, ms: 3.576},
		{mph: 9, ms: 4.023},
		{mph: 10, ms: 4.470},
	}

	for _, c := range cases {
		val, err := fromMilesPerHour(c.mph, METRIC)
		assert.NoError(t, err)
		assert.InDelta(t, c.ms, val, 0.01)
	}
}

func Test_Speed_fromMilesPerHour_toMilesPerHour(t *testing.T) {
	cases := []struct {
		mph float64
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
		val, err := fromMilesPerHour(c.mph, IMPERIAL)
		assert.NoError(t, err)
		assert.Equal(t, c.mph, val)
	}
}

func Test_Speed_fromMilesPerHour_toUnknown(t *testing.T) {
	val, err := fromMilesPerHour(20, SystemOfMeasure("unknown"))
	assert.Error(t, err)
	assert.Nil(t, val)
}

func Test_Speed_fromMilesPerHour_invalidValue(t *testing.T) {
	val, err := fromMilesPerHour(struct{}{}, METRIC)
	assert.Error(t, err)
	assert.Nil(t, val)
}

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
// - https://www.metric-conversions.org/pressure/pascals-to-pound-force-per-square-inch.htm#metricConversionTable
// - https://www.metric-conversions.org/pressure/pound-force-per-square-inch-to-pascals.htm#metricConversionTable
//

func Test_Pressure_fromPascal_toPoundForcePerSquareInch(t *testing.T) {
	cases := []struct {
		pa  float64
		psi float64
	}{
		{pa: -10, psi: -0.0014504},
		{pa: -9, psi: -0.0013053},
		{pa: -8, psi: -0.0011603},
		{pa: -7, psi: -0.0010153},
		{pa: -6, psi: -0.00087023},
		{pa: -5, psi: -0.00072519},
		{pa: -4, psi: -0.00058015},
		{pa: -3, psi: -0.00043511},
		{pa: -2, psi: -0.00029008},
		{pa: -1, psi: -0.00014504},
		{pa: 0, psi: 0.0},
		{pa: 1, psi: 0.00014504},
		{pa: 2, psi: 0.00029008},
		{pa: 3, psi: 0.00043511},
		{pa: 4, psi: 0.00058015},
		{pa: 5, psi: 0.00072519},
		{pa: 6, psi: 0.00087023},
		{pa: 7, psi: 0.0010153},
		{pa: 8, psi: 0.0011603},
		{pa: 9, psi: 0.0013053},
		{pa: 10, psi: 0.0014504},
		{pa: 100, psi: 0.014504},
		{pa: 1000, psi: 0.14504},
	}

	for _, c := range cases {
		val, err := fromPascal(c.pa, IMPERIAL)
		assert.NoError(t, err)
		assert.InDelta(t, c.psi, val, 0.00001)
	}
}

func Test_Pressure_fromMetersPerSecond_toPascal(t *testing.T) {
	cases := []struct {
		pa float64
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
		val, err := fromPascal(c.pa, METRIC)
		assert.NoError(t, err)
		assert.Equal(t, c.pa, val)
	}
}

func Test_Pressure_fromMetersPerSecond_toUnknown(t *testing.T) {
	val, err := fromPascal(20, SystemOfMeasure("unknown"))
	assert.Error(t, err)
	assert.Nil(t, val)
}

func Test_Pressure_fromMetersPerSecond_invalidValue(t *testing.T) {
	val, err := fromPascal(struct{}{}, IMPERIAL)
	assert.Error(t, err)
	assert.Nil(t, val)
}

func Test_Pressure_fromPoundForcePerSquareInch_toPascal(t *testing.T) {
	cases := []struct {
		psi float64
		pa  float64
	}{

		{psi: -10, pa: -68948},
		{psi: -9, pa: -62053},
		{psi: -8, pa: -55158},
		{psi: -7, pa: -48263},
		{psi: -6, pa: -41369},
		{psi: -5, pa: -34474},
		{psi: -4, pa: -27579},
		{psi: -3, pa: -20684},
		{psi: -2, pa: -13790},
		{psi: -1, pa: -6894.8},
		{psi: 0, pa: 0},
		{psi: 1, pa: 6894.8},
		{psi: 2, pa: 13790},
		{psi: 3, pa: 20684},
		{psi: 4, pa: 27579},
		{psi: 5, pa: 34474},
		{psi: 6, pa: 41369},
		{psi: 7, pa: 48263},
		{psi: 8, pa: 55158},
		{psi: 9, pa: 62053},
		{psi: 10, pa: 68948},
	}

	for _, c := range cases {
		val, err := fromPoundForcePerSquareInch(c.psi, METRIC)
		assert.NoError(t, err)
		assert.InDelta(t, c.pa, val, 1)
	}
}

func Test_Pressure_fromPoundForcePerSquareInch_toPoundForcePerSquareInch(t *testing.T) {
	cases := []struct {
		psi float64
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
		val, err := fromPoundForcePerSquareInch(c.psi, IMPERIAL)
		assert.NoError(t, err)
		assert.Equal(t, c.psi, val)
	}
}

func Test_Pressure_fromPoundForcePerSquareInch_toUnknown(t *testing.T) {
	val, err := fromPoundForcePerSquareInch(20, SystemOfMeasure("unknown"))
	assert.Error(t, err)
	assert.Nil(t, val)
}

func Test_Pressure_fromPoundForcePerSquareInch_invalidValue(t *testing.T) {
	val, err := fromPoundForcePerSquareInch(struct{}{}, METRIC)
	assert.Error(t, err)
	assert.Nil(t, val)
}

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

func TestOutput_MakeReading(t *testing.T) {
	o := Output{
		Name:      "test-output",
		Precision: 2,
		Type:      "test",
		Unit: &Unit{
			Name:   "test",
			Symbol: "t",
		},
	}

	r := o.MakeReading(3)

	assert.Equal(t, &o, r.output)
	assert.Equal(t, 3, r.Value)
	assert.Equal(t, o.Type, r.Type)
	assert.Equal(t, o.Unit.Symbol, r.Unit.Symbol)
	assert.Equal(t, o.Unit.Name, r.Unit.Name)
	assert.Empty(t, r.Info)
	assert.NotEmpty(t, r.Timestamp)
}

func TestOutput_MakeReading_noUnit(t *testing.T) {
	o := Output{
		Name:      "test-output",
		Precision: 2,
		Type:      "test",
	}

	r := o.MakeReading(3)

	assert.Equal(t, &o, r.output)
	assert.Equal(t, 3, r.Value)
	assert.Equal(t, o.Type, r.Type)
	assert.Nil(t, r.Unit)
	assert.Empty(t, r.Info)
	assert.NotEmpty(t, r.Timestamp)
}

func TestUnit_Encode(t *testing.T) {
	u := Unit{
		Name:   "fahrenheit",
		Symbol: "F",
	}

	encoded := u.Encode()
	assert.Equal(t, u.Name, encoded.Name)
	assert.Equal(t, u.Symbol, encoded.Symbol)
}

func TestUnit_Encode_empty(t *testing.T) {
	u := Unit{}

	encoded := u.Encode()
	assert.Empty(t, encoded.Name)
	assert.Empty(t, encoded.Symbol)
}

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

func TestGet_notExists(t *testing.T) {
	o := Get("nonexistent-output-name")
	assert.Nil(t, o)
}

func TestGet_exists(t *testing.T) {
	// The output here is a built-in, so it should always exist.
	o := Get("temperature")
	assert.NotNil(t, o)
	assert.Equal(t, "temperature", o.Name)
}

func TestRegister_noOutputs(t *testing.T) {
	// Copy the map and reset it once we're done so we don't
	// pollute it for other tests.
	var registeredCopy = map[string]*Output{}
	for k, v := range registeredOutputs {
		registeredCopy[k] = v
	}
	defer func() {
		registeredOutputs = registeredCopy
	}()

	initLen := len(registeredOutputs)

	err := Register()
	assert.NoError(t, err)
	assert.Len(t, registeredOutputs, initLen)
}

func TestRegister_oneOutput(t *testing.T) {
	// Copy the map and reset it once we're done so we don't
	// pollute it for other tests.
	var registeredCopy = map[string]*Output{}
	for k, v := range registeredOutputs {
		registeredCopy[k] = v
	}
	defer func() {
		registeredOutputs = registeredCopy
	}()

	initLen := len(registeredOutputs)

	err := Register(&Output{
		Name: "test-output-1",
	})
	assert.NoError(t, err)
	assert.Len(t, registeredOutputs, initLen+1)
}

func TestRegister_conflict(t *testing.T) {
	// Copy the map and reset it once we're done so we don't
	// pollute it for other tests.
	var registeredCopy = map[string]*Output{}
	for k, v := range registeredOutputs {
		registeredCopy[k] = v
	}
	defer func() {
		registeredOutputs = registeredCopy
	}()

	initLen := len(registeredOutputs)

	err := Register(&Output{
		Name: "temperature", // same name as a built-in, should conflict
	})
	assert.Error(t, err)
	assert.Len(t, registeredOutputs, initLen)
}

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

func TestOutput_Encode(t *testing.T) {
	o := Output{
		Name:      "test-output",
		Precision: 2,
		Type:      "test",
		Unit: &Unit{
			Name:   "test",
			Symbol: "t",
		},
	}

	e := o.Encode()

	assert.Equal(t, "test-output", e.Name)
	assert.Equal(t, "test", e.Type)
	assert.Equal(t, int32(2), e.Precision)
	assert.Equal(t, "test", e.Unit.Name)
	assert.Equal(t, "t", e.Unit.Symbol)
}

func TestOutput_Encode_noUnit(t *testing.T) {
	o := Output{
		Name:      "test-output",
		Precision: 2,
		Type:      "test",
	}

	e := o.Encode()

	assert.Equal(t, "test-output", e.Name)
	assert.Equal(t, "test", e.Type)
	assert.Equal(t, int32(2), e.Precision)
	assert.Nil(t, e.Unit)
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

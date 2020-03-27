// Synse SDK
// Copyright (c) 2017-2020 Vapor IO
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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet_notExists(t *testing.T) {
	f := Get("nonexistent")
	assert.Nil(t, f)
}

func TestGet_exists(t *testing.T) {
	// The func here is built-in so it should always exist.
	f := Get("FtoC")
	assert.NotNil(t, f)
	assert.Equal(t, "FtoC", f.Name)
}

func TestRegister_noOutputs(t *testing.T) {
	// Copy the map and reset it once we're done so we don't
	// pollute it for other tests.
	var registeredCopy = map[string]*Func{}
	for k, v := range registeredFuncs {
		registeredCopy[k] = v
	}
	defer func() {
		registeredFuncs = registeredCopy
	}()

	initLen := len(registeredFuncs)

	err := Register()
	assert.NoError(t, err)
	assert.Len(t, registeredFuncs, initLen)
}

func TestRegister_oneOutput(t *testing.T) {
	// Copy the map and reset it once we're done so we don't
	// pollute it for other tests.
	var registeredCopy = map[string]*Func{}
	for k, v := range registeredFuncs {
		registeredCopy[k] = v
	}
	defer func() {
		registeredFuncs = registeredCopy
	}()

	initLen := len(registeredFuncs)

	err := Register(&Func{
		Name: "test-func-1",
	})
	assert.NoError(t, err)
	assert.Len(t, registeredFuncs, initLen+1)
}

func TestRegister_conflict(t *testing.T) {
	// Copy the map and reset it once we're done so we don't
	// pollute it for other tests.
	var registeredCopy = map[string]*Func{}
	for k, v := range registeredFuncs {
		registeredCopy[k] = v
	}
	defer func() {
		registeredFuncs = registeredCopy
	}()

	initLen := len(registeredFuncs)

	err := Register(&Func{
		Name: "FtoC", // same name as a built-in, should conflict
	})
	assert.Error(t, err)
	assert.Len(t, registeredFuncs, initLen)
}

func TestFunc_Call_ok(t *testing.T) {
	fn := Func{
		Name: "test",
		Fn: func(value interface{}) (i interface{}, e error) {
			return value, nil
		},
	}

	val, err := fn.Call(1)
	assert.NoError(t, err)
	assert.Equal(t, 1, val.(int))
}

func TestFunc_Call_err(t *testing.T) {
	fn := Func{
		Name: "test",
		Fn: func(value interface{}) (i interface{}, e error) {
			return nil, fmt.Errorf("test error")
		},
	}

	val, err := fn.Call(1)
	assert.Error(t, err)
	assert.Nil(t, val)
}

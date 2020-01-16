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
	"fmt"

	"github.com/vapor-ware/synse-sdk/sdk/errors"
)

var registeredFuncs map[string]*Func

func init() {
	registeredFuncs = make(map[string]*Func)
	for _, f := range GetBuiltins() {
		registeredFuncs[f.Name] = f
	}
}

// Get gets a Func by its name. If a func with the specified name
// is not found, nil is returned.
func Get(name string) *Func {
	return registeredFuncs[name]
}

// Register registers new funcs to the tracked funcs.
func Register(funcs ...*Func) error {
	multiErr := errors.NewMultiError("func registration")

	for _, f := range funcs {
		if _, exists := registeredFuncs[f.Name]; exists {
			multiErr.Add(fmt.Errorf("conflict: Func with name '%s' already exists", f.Name))
			continue
		}
		registeredFuncs[f.Name] = f
	}
	return multiErr.Err()
}

// Func is a function that can be applied to a device reading.
type Func struct {
	// Name is the name of the function. This is how it is identified
	// and referenced.
	Name string

	// Fn is the function which will be called on the reading value.
	Fn func(value interface{}) (interface{}, error)
}

// Call calls the function defined for the Func.
func (fn *Func) Call(value interface{}) (interface{}, error) {
	return fn.Fn(value)
}

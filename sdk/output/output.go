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

package output

import (
	"fmt"

	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/utils"
	synse "github.com/vapor-ware/synse-server-grpc/go"
)

var registeredOutputs map[string]*Output

func init() {
	registeredOutputs = make(map[string]*Output)
	for _, o := range GetBuiltins() {
		registeredOutputs[o.Name] = o
	}
}

// Get gets an Output by name. If an output with the specified name
// is not found, nil is returned.
func Get(name string) *Output {
	return registeredOutputs[name]
}

// Register registers new outputs to the tracked outputs.
func Register(output ...*Output) error {
	multiErr := errors.NewMultiError("output registration")

	for _, o := range output {
		if _, exists := registeredOutputs[o.Name]; exists {
			multiErr.Add(fmt.Errorf("conflict: output with name '%s' already exists", o.Name))
			continue
		}
		registeredOutputs[o.Name] = o
	}
	return multiErr.Err()
}

// Output defines the output information associated with a device reading.
type Output struct {
	// Name is the name of the output. Output names should be unique, as
	// outputs can be referenced by name.
	Name string

	// Precision is the precision that the reading will take with this output,
	// e.g. the number of decimal places to round to. This is only used when
	// the reading value is a float.
	Precision int

	// Type is the type that an output will assign to a reading which is
	// generated from it.
	Type string

	// The unit applied to all readings for the output.
	Unit *Unit
}

// MakeReading makes a new Reading for the provided value, applying the pertinent
// output data to the reading.
func (output *Output) MakeReading(value interface{}) (reading *Reading, err error) {
	// Check value type.
	// Bytes and byte arrays may not serialize well, so force
	// the caller to define a type that will by erroring out.
	switch value.(type) {
	case byte:
		return nil, fmt.Errorf("MakeReading byte value is not directly serializable")
	case []byte:
		return nil, fmt.Errorf("MakeReading []byte value is not directly serialzable")
	}

	return &Reading{
		Timestamp: utils.GetCurrentTime(),
		Type:      output.Type,
		Unit:      output.Unit,
		Value:     value,
		output:    output,
	}, nil
}

// Encode translates the Output to its corresponding gRPC message.
func (output *Output) Encode() *synse.V3DeviceOutput {
	var unit *synse.V3OutputUnit
	if output.Unit != nil {
		unit = output.Unit.Encode()
	}

	return &synse.V3DeviceOutput{
		Name:      output.Name,
		Type:      output.Type,
		Precision: int32(output.Precision),
		Unit:      unit,
	}
}

// Unit is the unit of measure for a device reading.
type Unit struct {
	// Name is the full name of the unit.
	Name string

	// Symbol is the symbolic representation of the unit.
	Symbol string
}

// Encode translates the Unit to its corresponding gRPC message.
func (unit *Unit) Encode() *synse.V3OutputUnit {
	return &synse.V3OutputUnit{
		Name:   unit.Name,
		Symbol: unit.Symbol,
	}
}

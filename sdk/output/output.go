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
	"fmt"

	"github.com/vapor-ware/synse-sdk/sdk/utils"
	synse "github.com/vapor-ware/synse-server-grpc/go"
)

const (
	NONE     SystemOfMeasure = "none"
	IMPERIAL SystemOfMeasure = "imperial"
	METRIC   SystemOfMeasure = "metric"
)

type SystemOfMeasure string

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

	// Units defines all possible units that this output can provide. A reading
	// derived from this output can only use a single unit. An output's unit
	// may differ based on the system of measure being used.
	//
	// If the units map is empty, the output is considered unit-less (e.g.
	// "state" has no unit). If the units map defines the NONE system of
	// measure, the output is considered system-agnostic and will use that
	// unit for all systems.
	Units map[SystemOfMeasure]*Unit

	// Converters is a map which defines the unit conversion functions supported
	// by the Output. The map key is the starting system of measure, e.g. which
	// system we are converting from. Each function takes the value to convert
	// and the system to convert to. It is the responsibility of the function
	// to handle all system conversions.
	Converters map[SystemOfMeasure]func(value interface{}, to SystemOfMeasure) (interface{}, error)
}

// Convert converts a reading value from one system of measure to another, using
// the conversion functions defined for the Output.
func (output *Output) Convert(value interface{}, from, to SystemOfMeasure) (interface{}, error) {
	// The output is unit-less, there is nothing to convert.
	if len(output.Units) == 0 {
		// todo: log
		return value, nil
	}

	// The output uses a system-agnostic unit, so the unit is the same for
	// all systems. Nothing to convert.
	if _, exists := output.Units[NONE]; exists {
		// todo: log
		return value, nil
	}

	// The output has units, but no converters are defined. In this case we must
	// error as we do not know how to convert between the defined units.
	if len(output.Converters) == 0 {
		// fixme: better err handling
		return nil, fmt.Errorf("unable to convert")
	}

	// Get the converter for the specified system. If it doesn't exist, we can't
	// convert.
	converter, exists := output.Converters[from]
	if !exists {
		// fixme: better err handling
		return nil, fmt.Errorf("no converter defiend for %v", from)
	}

	// Run the converter.
	return converter(value, to)
}

// GetUnit gets the output unit for the specified system, if it is defined. If there
// are no units defined, nil is returned. If a system-agnostic unit is defined, it is
// returned. If no unit is defined for the specified
func (output *Output) GetUnit(system SystemOfMeasure) (*Unit, error) {
	// If the output is unit-less, return nil.
	if len(output.Units) == 0 {
		return nil, nil
	}

	// If the system-agnostic unit is specified for the unit, return it.
	if unit, exists := output.Units[NONE]; exists {
		return unit, nil
	}

	// Otherwise get the unit for the specified system. If it does not exist,
	// return an error.
	unit, exists := output.Units[system]
	if !exists {
		// fixme: better err handling
		return nil, fmt.Errorf("unit does not exist for system %v", system)
	}
	return unit, nil
}

// FromImperial creates a new Reading from a value in the Outputs imperial
// unit. For example:
//
//     Temperature.FromImperial(30)
//
// create a Reading which represents 30 degrees Fahrenheit, since Fahrenheit
// is the imperial unit for Temperature.
func (output *Output) FromImperial(value interface{}) *Reading {
	var unit *Unit
	var system = NONE
	if len(output.Units) > 0 {
		systemless, exists := output.Units[NONE]
		if exists {
			unit = systemless
		} else {
			system = IMPERIAL
			unit = output.Units[IMPERIAL]
		}
	}

	return &Reading{
		Timestamp: utils.GetCurrentTime(),
		Type:      output.Type,
		Unit:      unit,
		Value:     value,
		System:    system,
		output:    output,
	}
}

// FromMetric creates a new Reading from a value in the Outputs metric
// unit. For example:
//
//     Temperature.FromMetric(30)
//
// create a Reading which represents 30 degrees Celsius, since Celsius
// is the metric unit for Temperature.
func (output *Output) FromMetric(value interface{}) *Reading {
	var unit *Unit
	var system = NONE
	if len(output.Units) > 0 {
		systemless, exists := output.Units[NONE]
		if exists {
			unit = systemless
		} else {
			system = METRIC
			unit = output.Units[METRIC]
		}
	}

	return &Reading{
		Timestamp: utils.GetCurrentTime(),
		Type:      output.Type,
		Unit:      unit,
		Value:     value,
		System:    system,
		output:    output,
	}
}

// From creates a system-less new Reading for a value. If the Output defines
// a system-agnostic unit, it will be used. Otherwise, no unit is given to the
// Reading.
func (output *Output) From(value interface{}) *Reading {
	var unit *Unit
	if len(output.Units) > 0 {
		systemless, exists := output.Units[NONE]
		if exists {
			unit = systemless
		}
	}

	return &Reading{
		Timestamp: utils.GetCurrentTime(),
		Type:      output.Type,
		Unit:      unit,
		Value:     value,
		System:    NONE,
		output:    output,
	}
}

// Unit is the unit of measure for a device reading.
type Unit struct {
	// Name is the full name of the unit.
	Name string

	// Symbol is the symbolic representation of the unit.
	Symbol string

	// System is the system of measure which the unit belongs to.
	System string
}

// Encode translates the Unit to its corresponding gRPC message.
func (unit *Unit) Encode() *synse.V3OutputUnit {
	return &synse.V3OutputUnit{
		Name:   unit.Name,
		Symbol: unit.Symbol,
		System: unit.System,
	}
}

// fixme: dropping this here (previously from type.go) for reference for future work

//// GetScalingFactor gets the scaling factor for the reading type.
//func (outputType *OutputType) GetScalingFactor() (float64, error) {
//	if outputType.ScalingFactor == "" {
//		return 1, nil
//	}
//	return strconv.ParseFloat(outputType.ScalingFactor, 64)
//}

//// applyScalingFactor multiplies the raw reading value (the value parameter) by the output
//// scaling factor and returns the scaled reading.
//func (outputType *OutputType) applyScalingFactor(value interface{}) interface{} {
//	scalingFactor, err := outputType.GetScalingFactor()
//	if err != nil {
//		log.Errorf(
//			"[type] Unable to get scaling factor for outputType %+v, error: %v",
//			outputType, err.Error())
//		return value // TODO: Return the error.
//	}
//
//	// If the scaling factor is 0, log a warning and just return the original value.
//	if scalingFactor == 0 {
//		log.WithField("value", value).Warn(
//			"[type] got scaling factor of 0; will not apply",
//		)
//		return value
//	}
//
//	// If the scaling factor is 1, there is nothing to do. Return the value.
//	if scalingFactor == 1 {
//		return value
//	}
//
//	// Otherwise, the scaling factor is non-zero and not 1, so it will
//	// need to be applied.
//	f, err := utils.ConvertToFloat64(value)
//	if err != nil {
//		log.Errorf("[type] Unable to apply scaling factor %v to value %v of type %T", scalingFactor, value, value)
//		// TODO: Return the error.
//	}
//	return f * scalingFactor
//}

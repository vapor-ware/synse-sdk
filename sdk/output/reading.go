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

package output

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk/utils"
	synse "github.com/vapor-ware/synse-server-grpc/go"
)

// Reading describes a single device reading at a given time. The timestamp
// for a reading is represented using the RFC3339 layout.
type Reading struct {
	// Timestamp is the RFC3339-formatted time at which the reading was taken.
	Timestamp string

	// Type is the type of the reading, as defined by the Reading's output.
	Type string

	// Unit describes the unit of measure for the reading.
	Unit *Unit

	// Value is the reading value itself.
	Value interface{}

	// Context provides an arbitrary key-value mapping which can be used to
	// provide contextual information about the reading. This is not required
	// but can be useful if a device provides multiple readings from the
	// same output, or readings which are meaningless on their own.
	Context map[string]string

	// output is the Output used to render and format the reading.
	output *Output
}

// GetOutput gets the associated output for a Reading.
func (reading *Reading) GetOutput() *Output {
	return reading.output
}

// WithContext adds a context to the reading. This is useful when creating a
// reading from an output and you wish to in-line the setting of the reading
// context, e.g.
//
//    SomeOutput.MakeReading(3).WithContext(map[string]string{"source": "foo"})
//
// This will merge the provided context with any existing context. If there
// is a conflict, the context value provided here will override the pre-existing
// value.
func (reading *Reading) WithContext(ctx map[string]string) *Reading {
	if reading.Context == nil {
		reading.Context = make(map[string]string)
	}
	for k, v := range ctx {
		reading.Context[k] = v
	}
	return reading
}

// Scale multiplies the given scaling factor to the Reading value and updates
// the Value with the new scaled value.
//
// The scaling factor is defined in the device config, so this is applied by
// the SDK by the scheduler upon receiving the reading. The scaling factor is
// applied after any other transformation functions (see the sdk/funcs package)
// have been applied.
func (reading *Reading) Scale(factor float64) error {
	// If the scaling factor is 0, log a warning, but do nothing. The SDK explicitly
	// prohibits scaling factors of 0 to prevent all values from being zeroed out.

	// The SDK explicitly prohibits scaling factors of 0 to prevent all values from
	// being zeroed out.
	if factor == 0 {
		log.Error("[reading] invalid scaling factor - will not apply value 0")
		return fmt.Errorf("cannot have scaling factor of 0")
	}

	// If the scaling factor is 1, there is nothing for us to do.
	if factor == 1 {
		return nil
	}

	// Otherwise, calculate the new scaled value by multiplying the scaling factor.
	v, err := utils.ConvertToFloat64(reading.Value)
	if err != nil {
		log.WithFields(log.Fields{
			"value": reading.Value,
		}).Error("[reading] error converting reading value to float64")
		return err
	}
	reading.Value = v * factor
	return nil
}

// Encode translates the Reading to its corresponding gRPC message.
func (reading *Reading) Encode() *synse.V3Reading {
	var unit = &Unit{}
	if reading.Unit != nil {
		unit = reading.Unit
	}

	r := synse.V3Reading{
		Timestamp: reading.Timestamp,
		Type:      reading.Type,
		Context:   reading.Context,
		Unit:      unit.Encode(),
	}

	switch t := reading.Value.(type) {
	case string:
		r.Value = &synse.V3Reading_StringValue{StringValue: t}
	case bool:
		r.Value = &synse.V3Reading_BoolValue{BoolValue: t}
	case float64:
		r.Value = &synse.V3Reading_Float64Value{Float64Value: t}
	case float32:
		r.Value = &synse.V3Reading_Float32Value{Float32Value: t}
	case int64:
		r.Value = &synse.V3Reading_Int64Value{Int64Value: t}
	case int32:
		r.Value = &synse.V3Reading_Int32Value{Int32Value: t}
	case int16:
		r.Value = &synse.V3Reading_Int32Value{Int32Value: int32(t)}
	case int8:
		r.Value = &synse.V3Reading_Int32Value{Int32Value: int32(t)}
	case int:
		r.Value = &synse.V3Reading_Int64Value{Int64Value: int64(t)}
	case []byte:
		r.Value = &synse.V3Reading_BytesValue{BytesValue: t}
	case uint64:
		r.Value = &synse.V3Reading_Uint64Value{Uint64Value: t}
	case uint32:
		r.Value = &synse.V3Reading_Uint32Value{Uint32Value: t}
	case uint16:
		r.Value = &synse.V3Reading_Uint32Value{Uint32Value: uint32(t)}
	case uint8:
		r.Value = &synse.V3Reading_Uint32Value{Uint32Value: uint32(t)}
	case uint:
		r.Value = &synse.V3Reading_Uint64Value{Uint64Value: uint64(t)}
	case nil:
		r.Value = nil
	default:
		// If the reading type isn't one of the above, panic. The plugin should
		// terminate. This is indicative of the plugin is providing bad data.
		panic(fmt.Sprintf("unsupported reading value type: %s", t))
	}
	return &r
}

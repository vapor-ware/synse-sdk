package sdk

import (
	"strconv"
	"strings"

	"github.com/vapor-ware/synse-server-grpc/go"
)

var (
	outputTypeMap map[string]*ReadingType
)

func init() {
	outputTypeMap = map[string]*ReadingType{}
}

/*
TODO:
- function(s) for proper type casting
- function(s) for applying scaling factor
- function for parsing scaling factor
- function for applying precision
- functionality for reading in from YAML
*/

// ReadingType provides information about the type of a device reading.
type ReadingType struct {
	// Name is the name of the reading type. Each reading type
	// should have a unique name. Names can be namespaced with
	// '.' as the delimiter.
	Name string

	// DataType is the type of the reading data. This should be one of:
	// float64, float32, float (alias for float64),
	// int64, int32, int, uint64, uint32, bool, string, bytes
	DataType string

	// Precision is the number of decimal places to round to.
	// This is only used when the type is a float-type.
	Precision int

	// Unit is the unit of measure for the reading.
	Unit Unit

	// ScalingFactor is an optional value by which to scale the
	// reading. This is useful when a device returns reading data
	// that must be scaled.
	//
	// This value should resolve to a numeric. By default, it will
	// have a value of 1. Negative values and fractional values are
	// supported. This can be the value itself, e.g. "0.01", or
	// a mathematical representation of the value, e.g. "1e-2".
	ScalingFactor string
}

// Type gets the type of the reading. This is encoded in the ReadingType
// name. If the ReadingType is namespaced, this will be the last element
// of the namespace. If it is not namespaced, it will be the name itself.
func (readingType *ReadingType) Type() string {
	if strings.Contains(readingType.Name, ".") {
		nameSpace := strings.Split(readingType.Name, ".")
		return nameSpace[len(nameSpace)-1]
	}
	return readingType.Name
}

// GetScalingFactor gets the scaling factor for the reading type.
func (readingType *ReadingType) GetScalingFactor() (float64, error) {
	return strconv.ParseFloat(readingType.ScalingFactor, 64)
}

// Unit is the unit of measure for a device reading.
type Unit struct {
	// Name is the full name of the unit.
	Name string

	// Symbol is the symbolic representation of the unit.
	Symbol string
}

// encode translates the SDK Unit type to the corresponding gRPC Unit type.
func (unit *Unit) encode() *synse.Unit {
	return &synse.Unit{
		Name:   unit.Name,
		Symbol: unit.Symbol,
	}
}

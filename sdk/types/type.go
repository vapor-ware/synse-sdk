package types

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
	Unit ReadingUnit

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

// ReadingUnit is the unit of measure for a device reading.
type ReadingUnit struct {
	// Name is the full name of the unit.
	Name string

	// Symbol is the symbolic representation of the unit.
	Symbol string
}

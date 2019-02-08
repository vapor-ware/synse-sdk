package sdk

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-server-grpc/go"
)

// OutputType provides information about the output of a device reading.
type OutputType struct {

	// Version is the version of the configuration scheme.
	Version int `yaml:"version,omitempty"`

	// Name is the name of the output type. Each output type
	// should have a unique name. Names can be namespaced with
	// '.' as the delimiter.
	Name string `yaml:"name,omitempty" addedIn:"1.0"`

	// Precision is the number of decimal places to round to.
	// This is only used when the type is a float-type.
	Precision int `yaml:"precision,omitempty" addedIn:"1.0"`

	// Unit is the unit of measure for the reading.
	Unit Unit `yaml:"unit,omitempty" addedIn:"1.0"`

	// ScalingFactor is an optional value by which to scale the
	// reading. This is useful when a device returns reading data
	// that must be scaled.
	//
	// This value should resolve to a numeric. By default, it will
	// have a value of 1. Negative values and fractional values are
	// supported. This can be the value itself, e.g. "0.01", or
	// a mathematical representation of the value, e.g. "1e-2".
	ScalingFactor string `yaml:"scalingFactor,omitempty" addedIn:"1.0"`

	// Conversion is an arbitrary string that allows the sdk to
	// perform a conversion. Initially only the string englishToMetricTemperature
	// will be supported for temperature sensors.
	// This field is not in the Output message, therefore the grpc client never sees this.
	Conversion string `yaml:"conversion,omitempty" addedIn:"1.2"`
}

// JSON encodes the config as JSON. This can be useful for logging and debugging.
func (outputType *OutputType) JSON() (string, error) {
	bytes, err := json.Marshal(outputType)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Validate validates that the OutputType has no configuration errors.
func (outputType OutputType) Validate(multiErr *errors.MultiError) {
	if outputType.Name == "" {
		multiErr.Add(errors.NewFieldRequiredError(multiErr.Context["source"], "outputType.name"))
	}

	// Try parsing the scaling factor to validate it is a correctly specified
	// duration string.
	_, err := outputType.GetScalingFactor()
	if err != nil {
		multiErr.Add(errors.NewValidationError(multiErr.Context["source"], err.Error()))
	}
}

// GetVersion fulfills the VersionedConfig interface. It just returns the version
// of the config.
func (outputType *OutputType) GetVersion() int {
	return outputType.Version
}

// Type gets the type of the reading. This is encoded in the OutputType
// name. If the OutputType is namespaced, this will be the last element
// of the namespace. If it is not namespaced, it will be the name itself.
func (outputType *OutputType) Type() string {
	if strings.Contains(outputType.Name, ".") {
		nameSpace := strings.Split(outputType.Name, ".")
		return nameSpace[len(nameSpace)-1]
	}
	return outputType.Name
}

// GetScalingFactor gets the scaling factor for the reading type.
func (outputType *OutputType) GetScalingFactor() (float64, error) {
	if outputType.ScalingFactor == "" {
		return 1, nil
	}
	return strconv.ParseFloat(outputType.ScalingFactor, 64)
}

// applyScalingFactor multiplies the raw reading value (the value parameter) by the output
// scaling factor and returns the scaled reading.
func (outputType *OutputType) applyScalingFactor(value interface{}) interface{} {
	scalingFactor, err := outputType.GetScalingFactor()
	if err != nil {
		log.Errorf(
			"[type] Unable to get scaling factor for outputType %+v, error: %v",
			outputType, err.Error())
		return value // TODO: Return the error.
	}

	// If the scaling factor is 0, log a warning and just return the original value.
	if scalingFactor == 0 {
		log.WithField("value", value).Warn(
			"[type] got scaling factor of 0; will not apply",
		)
		return value
	}

	// If the scaling factor is 1, there is nothing to do. Return the value.
	if scalingFactor == 1 {
		return value
	}

	// Otherwise, the scaling factor is non-zero and not 1, so it will
	// need to be applied.
	f, err := ConvertToFloat64(value)
	if err != nil {
		log.Errorf("[type] Unable to apply scaling factor %v to value %v of type %T", scalingFactor, value, value)
		// TODO: Return the error.
	}
	return f * scalingFactor
}

// applyConversion applies the conversion based on the output conversion string and
// the scaled reading.
// For now this is pretty limited, but it's a place to start.
func (outputType *OutputType) applyConversion(value interface{}) (result interface{}, err error) {
	// Only one supported conversion string for now.
	switch outputType.Conversion {
	case "": // Nothing to do.
		return value, nil
	case "englishToMetricTemperature":
		var f float64
		f, err = ConvertToFloat64(value)
		if err != nil {
			return
		}
		c := (f - 32.0) * 5.0 / 9.0
		return c, nil
	default:
		return nil, fmt.Errorf("Unknown conversion %v", outputType.Conversion)
	}
}

// Apply applies the transformations specified by the OutputType to
// a reading value. These transformations are (in the order that they
// are applied): multiply scaling factor.
//
// Precision is not applied at this level, but will instead be applied
// in Synse server before the corresponding reading is returned to the
// user.
func (outputType *OutputType) Apply(value interface{}) interface{} {

	value = outputType.applyScalingFactor(value)

	value, err := outputType.applyConversion(value)
	if err != nil {
		log.Errorf("Unable to apply conversion: %v, error %v", outputType.Conversion, err)
		// TODO: Return the error.
	}
	return value
}

// Unit is the unit of measure for a device reading.
type Unit struct {
	// Name is the full name of the unit.
	Name string `yaml:"name,omitempty" addedIn:"1.0"`

	// Symbol is the symbolic representation of the unit.
	Symbol string `yaml:"symbol,omitempty" addedIn:"1.0"`
}

// Validate validates that the Unit has no configuration errors.
func (unit Unit) Validate(multiErr *errors.MultiError) {
	// nothing to validate
}

// encode translates the SDK Unit type to the corresponding gRPC Unit type.
func (unit *Unit) encode() *synse.Unit {
	return &synse.Unit{
		Name:   unit.Name,
		Symbol: unit.Symbol,
	}
}

package config

import (
	"strconv"
	"strings"

	"fmt"

	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
	"github.com/vapor-ware/synse-sdk/sdk/policies"
	"github.com/vapor-ware/synse-server-grpc/go"
)

// ProcessOutputTypeConfig searches for, reads, and validates the output type
// configuration from file. Its behavior will vary depending on the output type
// config policy that is set. If output type config is processed successfully,
// the found output type configs are returned.
func ProcessOutputTypeConfig() ([]*OutputType, error) { // nolint: gocyclo
	// Get the plugin's policy for output type config files.
	outputTypeFilePolicy := policies.GetTypeConfigFilePolicy()
	logger.Debugf("output type config file policy: %s", outputTypeFilePolicy.String())

	// Now, try getting the output type config(s) from file.
	outputTypeCtxs, err := GetOutputTypeConfigsFromFile()

	// If the error is not a "config not found" error, then we will return it.
	_, notFoundErr := err.(*errors.ConfigsNotFound)
	if !notFoundErr {
		return nil, err
	}

	switch outputTypeFilePolicy {
	case policies.TypeConfigFileRequired:
		if err != nil {
			return nil, errors.NewPolicyViolationError(
				outputTypeFilePolicy.String(),
				fmt.Sprintf("output type config file(s) required, but not found: %v", err),
			)
		}

	case policies.TypeConfigFileOptional:
		if err != nil {
			outputTypeCtxs = []*Context{}
			logger.Debug("no type configuration config files found")
		}

	case policies.TypeConfigFileProhibited:
		// If the output type config file is prohibited, we will log a warning
		// if a file is found, but we will ultimately not fail. Instead, we
		// will just pass along an empty config.
		if err == nil && len(outputTypeCtxs) > 0 {
			logger.Warn(
				"output type config file(s) found, but its use is prohibited via policy. " +
					"the output type config files will be ignored.",
			)
			outputTypeCtxs = []*Context{}
		}

	default:
		return nil, errors.NewPolicyViolationError(
			outputTypeFilePolicy.String(),
			"unsupported output type config file policy",
		)
	}

	var outputs []*OutputType

	// Validate the plugin config
	for _, outputTypeCtx := range outputTypeCtxs {
		multiErr := Validator.Validate(outputTypeCtx)
		if multiErr.HasErrors() {
			return nil, multiErr
		}
		cfg := outputTypeCtx.Config.(*OutputType)
		outputs = append(outputs, cfg)
	}
	return outputs, nil
}

// OutputType provides information about the output of a device reading.
type OutputType struct {

	// SchemeVersion is the version of the configuration scheme.
	SchemeVersion `yaml:",inline"`

	// Name is the name of the output type. Each reading type
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

// Apply applies the transformations specified by the OutputType to
// a reading value. These transformations are (in the order that they
// are applied): multiply scaling factor.
//
// Precision is not applied at this level, but will instead be applied
// in Synse Server before the corresponding reading is returned to the
// user.
func (outputType *OutputType) Apply(value interface{}) interface{} { // nolint: gocyclo
	scalingFactor, err := outputType.GetScalingFactor()
	if err != nil {
		return value
	}

	// Do not permit a scaling factor of 0.
	if scalingFactor != 0 {
		switch t := value.(type) {
		case float64:
			value = t * scalingFactor
		case float32:
			value = float32(float64(t) * scalingFactor)
		case int64:
			value = int64(float64(t) * scalingFactor)
		case int32:
			value = int32(float64(t) * scalingFactor)
		case int16:
			value = int16(float64(t) * scalingFactor)
		case int8:
			value = int8(float64(t) * scalingFactor)
		case int:
			value = int(float64(t) * scalingFactor)
		case uint64:
			value = uint64(float64(t) * scalingFactor)
		case uint32:
			value = uint32(float64(t) * scalingFactor)
		case uint16:
			value = uint16(float64(t) * scalingFactor)
		case uint8:
			value = uint8(float64(t) * scalingFactor)
		case uint:
			value = uint(float64(t) * scalingFactor)
		}
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
	// nothing to validate here -- neither are required.
}

// Encode translates the SDK Unit type to the corresponding gRPC Unit type.
func (unit *Unit) Encode() *synse.Unit {
	return &synse.Unit{
		Name:   unit.Name,
		Symbol: unit.Symbol,
	}
}

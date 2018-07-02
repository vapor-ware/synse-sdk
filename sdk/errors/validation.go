package errors

import (
	"fmt"
)

// ValidationError is used as a general error during config validation.
type ValidationError struct {
	// source is the source of the configuration which caused the validation error.
	source string

	// msg is the error message.
	msg string
}

// NewValidationError returns a new instance of a ValidationError.
func NewValidationError(source, msg string) *ValidationError {
	return &ValidationError{
		source: source,
		msg:    msg,
	}
}

// Error returns the error string and fulfils the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validating config %s: %s", e.source, e.msg)
}

// FieldNotSupported is an error returned when a configuration scheme version is
// less than the "addedIn" scheme version for a field.
type FieldNotSupported struct {
	// source is the source of the configuration which caused the validation error.
	source string

	// field is the field which is not supported
	field string

	// fieldVersion is the version that the field was added in
	fieldVersion string

	// configVersion is the scheme version of the config
	configVersion string
}

// NewFieldNotSupportedError returns a new instance of a FieldNotSupported error.
func NewFieldNotSupportedError(source, field, fieldVersion, configVersion string) *FieldNotSupported {
	return &FieldNotSupported{
		source:        source,
		field:         field,
		fieldVersion:  fieldVersion,
		configVersion: configVersion,
	}
}

// Error returns the error string and fulfils the error interface.
func (e *FieldNotSupported) Error() string {
	return fmt.Sprintf(
		"validating config %s: field '%s' not supported in v%s (field added in v%s)",
		e.source, e.field, e.configVersion, e.fieldVersion,
	)
}

// FieldRemoved is an error returned when a configuration scheme version is greater
// than or equal to the "removedIn" scheme version for a field.
type FieldRemoved struct {
	// source is the source of the configuration which caused the validation error.
	source string

	// field is the field which is not supported
	field string

	// fieldVersion is the version that the field was added in
	fieldVersion string

	// configVersion is the scheme version of the config
	configVersion string
}

// NewFieldRemovedError returns a new instance of a FieldRemoved error.
func NewFieldRemovedError(source, field, fieldVersion, configVersion string) *FieldRemoved {
	return &FieldRemoved{
		source:        source,
		field:         field,
		fieldVersion:  fieldVersion,
		configVersion: configVersion,
	}
}

// Error returns the error string and fulfils the error interface.
func (e *FieldRemoved) Error() string {
	return fmt.Sprintf(
		"validating config %s: field '%s' not supported in v%s (field removed in v%s)",
		e.source, e.field, e.configVersion, e.fieldVersion,
	)
}

// FieldRequired is an error returned when a configuration is being validated and
// a field is not filled, but it is required.
type FieldRequired struct {
	// source is the source of the configuration which caused the validation error.
	source string

	// field is the field which is not supported
	field string
}

// NewFieldRequiredError returns a new instance of a FieldRequired error.
func NewFieldRequiredError(source, field string) *FieldRequired {
	return &FieldRequired{
		source: source,
		field:  field,
	}
}

func (e *FieldRequired) Error() string {
	return fmt.Sprintf(
		"validating config %s: missing required field '%s'",
		e.source, e.field,
	)
}

// InvalidValue is an error returned when a configuration is being validated and
// a field does not contain the expected data.
type InvalidValue struct {
	// source is the source of the configuration which caused the validation error.
	source string

	// field is the field which is not supported
	field string

	// needs is a string that specifies what the field needs
	needs string
}

// NewInvalidValueError returns a new instance of an InvalidValue error.
func NewInvalidValueError(source, field, needs string) *InvalidValue {
	return &InvalidValue{
		source: source,
		field:  field,
		needs:  needs,
	}
}

func (e *InvalidValue) Error() string {
	return fmt.Sprintf(
		"validating config %s: invalid value for field '%s'. must be %s",
		e.source, e.field, e.needs,
	)
}

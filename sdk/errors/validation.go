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
	return fmt.Sprintf("validating %s: %s", e.source, e.msg)
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
		"validating %s: field '%s' not supported in v%s (field added in v%s)",
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
		"validating %s: field '%s' not supported in v%s (field removed in v%s)",
		e.source, e.field, e.configVersion, e.fieldVersion,
	)
}

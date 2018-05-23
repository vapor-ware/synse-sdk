package cfg

import (
	"fmt"
	"reflect"

	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
)

/*
TODO:
---------------
- maintain context of what is being validated?
	- e.g., for errors, we want to be able to say that X type of config is invalid in file Y
- use sdk validation errors


- We could make this scheme validation logic its own project. I haven't really
  come across anything like this, and I think that its a pretty simple and decent
  solution for what it tries to do.. definitely easier than managing external scheme
  files or managing multiple versions of different structs, etc.
*/

// SchemeValidator is used to validate the scheme of a config.
type SchemeValidator struct {
	Version *SchemeVersion

	errors *errors.MultiError
}

// NewSchemeValidator creates a new instance of a SchemeValidator for the
// specified SchemeVersion.
func NewSchemeValidator(version *SchemeVersion) *SchemeValidator {
	return &SchemeValidator{
		Version: version,
		errors:  errors.NewMultiError("scheme validation"),
	}
}

// ValidateConfig validates a struct that holds configuration information. The
// validation works by search all fields and nested fields for the "addedIn",
// "deprecatedIn", and "removedIn" tags. It compares the versions specified in
// those tags with the scheme version of the configuration itself.
//
// Validation will result in errors if a field has a value and the version of the
// config scheme is out of bounds with the tags. A version could be out of bounds
// if it is less than the "addedIn" tag, or greater than or equal to the "removedIn"
// tag.
//
// Validation will log a warning if a field has a value and the version of the
// config scheme is greater than or equal to the "deprecatedIn" flag.
func (validator *SchemeValidator) ValidateConfig(config interface{}) error {
	val := reflect.ValueOf(config)

	if val.Kind() == reflect.Int || val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// ValidateConfig can only be called on a struct representing a configuration
	// component.
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("config validation: only accepts structs, but got %s", val.Kind())
	}
	validator.walk(val)

	return validator.errors.Err()
}

// walk is in intermediary step in config validation that will attempt to
// walk down into any nested fields/collections.
func (validator *SchemeValidator) walk(v reflect.Value) {
	switch v.Kind() {
	case reflect.Struct:
		validator.walkStructFields(v)

		// If the struct implements the ConfigComponent interface, validate the struct.
		ifaceType := reflect.TypeOf((*ConfigComponent)(nil)).Elem()
		if v.Type().Implements(ifaceType) {
			err := v.Interface().(ConfigComponent).Validate()
			if err != nil {
				validator.errors.Add(err)
			}
		}

	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			validator.walk(v.Index(i))
		}
	}
}

// walkStructFields goes through all of the fields of a struct and validates
// the fields. Only exported fields are validated.
//
// If the field is a nested struct or a collection of nested structs, it will
// be validated as well.
func (validator *SchemeValidator) walkStructFields(v reflect.Value) {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := v.Field(i)
		structField := t.Field(i)

		// Ignore unexported fields
		if structField.PkgPath != "" {
			continue
		}

		// First, validate the field of the struct
		validator.validateField(field, structField)

		// Try to walk through this field. If it is any nested type,
		// it will go through and validate, otherwise it will return
		// with no error.
		validator.walk(field)
	}
}

// validateField validates that a field of a struct is valid for the config's
// version scheme.
func (validator *SchemeValidator) validateField(field reflect.Value, structField reflect.StructField) { // nolint: gocyclo
	version := validator.Version

	// We should only care about validation if the field is set.
	if !validator.isEmptyValue(field) {
		// Check the "addedIn" tag
		tag := structField.Tag.Get(tagAddedIn)
		if tag != "" {
			addedInScheme, err := NewSchemeVersion(tag)
			if err != nil {
				validator.errors.Add(err)
			}
			if version.IsLessThan(addedInScheme) {
				validator.errors.Add(fmt.Errorf("field not supported: '%v'. added in: %v, config version: %v", structField.Name, addedInScheme.String(), version.String()))
			}
		}

		// Check the "deprecatedIn" tag
		tag = structField.Tag.Get(tagDeprecatedIn)
		if tag != "" {
			deprecatedInScheme, err := NewSchemeVersion(tag)
			if err != nil {
				validator.errors.Add(err)
			}
			if version.IsGreaterOrEqualTo(deprecatedInScheme) {
				logger.Warnf(
					"config field '%s' was deprecated in scheme version %s (current config scheme: %s)",
					structField.Name, deprecatedInScheme.String(), version.String(),
				)
			}
		}

		// Check the "removedIn" tag
		tag = structField.Tag.Get(tagRemovedIn)
		if tag != "" {
			removedInScheme, err := NewSchemeVersion(tag)
			if err != nil {
				validator.errors.Add(err)
			}
			if version.IsGreaterOrEqualTo(removedInScheme) {
				validator.errors.Add(fmt.Errorf("field not supported: '%v'. removed in: %v, config version: %v", structField.Name, removedInScheme.String(), version.String()))
			}
		}
	}
}

// isEmptyValue checks if a value is its empty type.
func (validator *SchemeValidator) isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	default:
		logger.Warn("No case for empty value check: %v", v.Kind())
	}
	return false
}

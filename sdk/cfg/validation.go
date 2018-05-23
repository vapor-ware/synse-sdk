package cfg

import (
	"fmt"
	"reflect"

	"github.com/vapor-ware/synse-sdk/sdk/logger"
)

/*
TODO:
---------------
- maintain context of what is being validated?
	- e.g., for errors, we want to be able to say that X type of config is invalid in file Y
- use sdk validation errors
*/

// SchemeValidator is used to validate the scheme of a config.
type SchemeValidator struct {
	Version *SchemeVersion
}

// NewSchemeValidator creates a new instance of a SchemeValidator for the
// specified SchemeVersion.
func NewSchemeValidator(version *SchemeVersion) *SchemeValidator {
	return &SchemeValidator{
		Version: version,
	}
}

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
	return validator.walk(val)
}

func (validator *SchemeValidator) walk(v reflect.Value) error {
	switch v.Kind() {
	case reflect.Struct:
		err := validator.walkStructFields(v)
		if err != nil {
			return err
		}

	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			err := validator.walk(v.Index(i))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (validator *SchemeValidator) walkStructFields(v reflect.Value) error {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := v.Field(i)
		structField := t.Field(i)

		// Ignore unexported fields
		if structField.PkgPath != "" {
			continue
		}

		// First, validate the field of the struct
		err := validator.validateField(field, structField)
		if err != nil {
			return err
		}

		// Try to walk through this field. If it is any nested type,
		// it will go through and validate, otherwise it will return
		// with no error.
		err = validator.walk(field)
		if err != nil {
			return err
		}
	}
	return nil
}

func (validator *SchemeValidator) validateField(field reflect.Value, structField reflect.StructField) error {
	version := validator.Version

	// We should only care about validation if the field is set.
	if !validator.isEmptyValue(field) {
		// Check the "addedIn" tag
		tag := structField.Tag.Get(tagAddedIn)
		if tag != "" {
			addedInScheme, err := NewSchemeVersion(tag)
			if err != nil {
				return err
			}
			if version.IsLessThan(addedInScheme) {
				return fmt.Errorf("field not supported: '%v'. added in: %v, config version: %v", structField.Name, addedInScheme.String(), version.String())
			}
		}

		// Check the "deprecatedIn" tag
		tag = structField.Tag.Get(tagDeprecatedIn)
		if tag != "" {
			deprecatedInScheme, err := NewSchemeVersion(tag)
			if err != nil {
				return err
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
				return err
			}
			if version.IsGreaterOrEqualTo(removedInScheme) {
				return fmt.Errorf("field not supported: '%v'. removed in: %v, config version: %v", structField.Name, removedInScheme.String(), version.String())
			}
		}
	}
	return nil
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

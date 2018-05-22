package cfg

import (
	"fmt"
	"reflect"

	"github.com/vapor-ware/synse-sdk/sdk/logger"
)

/*
TODO:
---------------
- implement util for annotating struct fields with version info
- implement something to walk a struct and:
	- get the config version.
		- if no version, use the latest version
	- validate that the version is supported
	- go through all fields/sub-fields of the config struct
		- check if a value is present
			- if so, get the version tags, we should support: addedIn, deprecatedIn, removedIn
				- check if the scheme version is out of bounds for any of
				  those constraints. If so, log/error.
			- if not, move on
*/

func ValidateFieldsForVersion(version *SchemeVersion, config interface{}) error {
	cfg := reflect.ValueOf(config)

	if cfg.Kind() != reflect.Ptr {
		return fmt.Errorf("field validation: config struct must specified as a pointer")
	}

	v := cfg.Elem()
	t := v.Type()
	return walkStructFields(version, v, t)

}

func walkStructFields(version *SchemeVersion, v reflect.Value, t reflect.Type) error {
	for i := 0; i < t.NumField(); i++ {
		field := v.Field(i)
		typeField := t.Field(i)

		// --- Break this out into its own function -------

		// We should only care about validation if the field is set.
		if !isEmptyValue(field) {
			// Check the "addedIn" tag
			tag := typeField.Tag.Get(tagAddedIn)
			if tag != "" {
				addedInScheme, err := NewSchemeVersion(tag)
				if err != nil {
					return err
				}
				// FIXME - perhaps an IsLessThan() function would be better
				if version.Compare(addedInScheme) == LessThan {
					return fmt.Errorf("field not supported: '%v'. added in: %v, config version: %v", typeField.Name, addedInScheme.String(), version.String())
				}
			}

			// Check the "deprecatedIn" tag
			tag = typeField.Tag.Get(tagDeprecatedIn)
			if tag != "" {
				deprecatedInScheme, err := NewSchemeVersion(tag)
				if err != nil {
					return err
				}
				// FIXME - perhaps an IsGreaterOrEqaul() function would be better
				cmp := version.Compare(deprecatedInScheme)
				if cmp == GreaterThan || cmp == EqualTo {
					// FIXME - this should be a warning. need to figure out how best to return
					// all errors/all warnings for validation methods
					return fmt.Errorf("field deprecated: '%v'. deprecated in: %v, config version: %v", typeField.Name, deprecatedInScheme.String(), version.String())
				}
			}

			// Check the "removedIn" tag
			tag = typeField.Tag.Get(tagRemovedIn)
			if tag != "" {
				removedInScheme, err := NewSchemeVersion(tag)
				if err != nil {
					return err
				}
				// FIXME - perhaps an IsGreaterOrEqaul() function would be better
				cmp := version.Compare(removedInScheme)
				if cmp == GreaterThan || cmp == EqualTo {
					// FIXME - this should be a warning. need to figure out how best to return
					// all errors/all warnings for validation methods
					return fmt.Errorf("field not supported: '%v'. removed in: %v, config version: %v", typeField.Name, removedInScheme.String(), version.String())
				}
			}
		}

		// ------- END break --------

		// For any nested types, go through and validate those as well.
		switch field.Kind() {
		case reflect.Struct:
			// TODO
		case reflect.Slice:
			// TODO
		case reflect.Ptr:
			// TODO
		}

	}
	return nil
}

func isEmptyValue(v reflect.Value) bool {
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

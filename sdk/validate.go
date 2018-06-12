package sdk

import (
	"fmt"
	"reflect"

	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
	"github.com/vapor-ware/synse-server-grpc/go"
)

// validator is the global schemeValidator that is used to validate plugin
// configuration files.
var validator = &schemeValidator{}

// schemeValidator is used to validate the scheme of a config.
type schemeValidator struct {
	// context is the ConfigContext, which references the configuration
	// currently being validated.
	context *ConfigContext

	// errors is the collection of errors that are found when validating.
	errors *errors.MultiError

	// version is the version of the scheme to validate configs with. This
	// is taken from the configuration being validated.
	version *ConfigVersion
}

// Validate validates a struct that holds configuration information. There are
// two kinds of validation that occur: field version validation, data validation.
//
// Field version validation is where we go through the config struct, and all
// nested config structs, and look for the "addedIn", "deprecatedIn", and "removedIn"
// tags. It compares the versions specified by those tags with the scheme version
// of the configuration itself. This validation will result in errors if the field
// has a value and the version of the config scheme is out of bounds (e.g. smaller
// than the addedIn value, or greater than or equal to the removedIn value).
// Warnings are logged if the config scheme is greater than or equal to the
// deprecatedIn tag.
//
// Data validation is where the Validate() method is called a struct which implements
// the ConfigComponent interface. Each struct should define its own validation. The
// validation here is typically checking to make sure required values exist, or that
// values are correct and can be parsed correctly.
//
// This function takes a Context, which provides both the SchemeVersion to
// validate against, and the config to validate. The "source" from the context is
// used to attribute to the errors in the event that any are found.
func (validator *schemeValidator) Validate(context *ConfigContext) *errors.MultiError {
	// Once we're done validating, we'll want to clear the state from this validation.
	defer validator.clearState()

	// Before we start validating, apply the state to the validator.
	validator.errors = errors.NewMultiError(context.Source)

	version, err := context.Config.GetVersion()
	if err != nil {
		validator.errors.Add(errors.NewValidationError(context.Source, err.Error()))
		return validator.errors
	}
	validator.context = context
	validator.version = version

	// We will also want to add to the MultiError context to specify the
	// source of the config, e.g. which file this came from. This info
	// can be used by the Validate() functions of ConfigComponents to
	// generate descriptive errors.
	validator.errors.Context["source"] = context.Source

	// Once we're done validating, clear the state from this validation.
	defer validator.clearState()

	// Now, validate the configuration provided by the context.
	validator.validate(context.Config)

	// Return validation errors, if any were found.
	return validator.errors
}

// clearState clears the state tracked for a single validation run.
func (validator *schemeValidator) clearState() {
	validator.context = nil
	validator.errors = nil
	validator.version = nil
}

// validate is the entry point for validation.
func (validator *schemeValidator) validate(config interface{}) {
	val := reflect.ValueOf(config)

	if val.Kind() == reflect.Interface || val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// ValidateConfig can only be called on a struct representing a
	// configuration component.
	if val.Kind() != reflect.Struct {
		validator.errors.Add(errors.NewValidationError(
			validator.context.Source,
			fmt.Sprintf("config validation: only accepts structs, but got %s", val.Kind()),
		))
		// Since we shouldn't be validating against anything else, there is
		// no point in doing anything more here.
		return
	}
	validator.walk(val)
}

// walk is in intermediary step in config validation that will attempt to
// walk down into any nested fields/collections.
func (validator *schemeValidator) walk(v reflect.Value) {
	switch v.Kind() {
	case reflect.Struct:
		validator.walkStructFields(v)

		// If the struct implements the ConfigComponent interface, validate the struct.
		ifaceType := reflect.TypeOf(new(ConfigComponent)).Elem()
		if v.Type().Implements(ifaceType) {
			v.Interface().(ConfigComponent).Validate(validator.errors)
		}

	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			validator.walk(v.Index(i))
		}

	case reflect.Ptr, reflect.Interface:
		validator.walk(v.Elem())
	}
}

// walkStructFields goes through all of the fields of a struct and validates
// the fields. Only exported fields are validated.
//
// If the field is a nested struct or a collection of nested structs, it will
// be validated as well.
func (validator *schemeValidator) walkStructFields(v reflect.Value) {
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
func (validator *schemeValidator) validateField(field reflect.Value, structField reflect.StructField) { // nolint: gocyclo
	version := validator.version

	// We should only care about validation if the field is set.
	if !validator.isEmptyValue(field) {
		// Check the "addedIn" tag
		tag := structField.Tag.Get(tagAddedIn)
		if tag != "" {
			addedInScheme, err := NewVersion(tag)
			if err != nil {
				validator.errors.Add(err)
			} else {
				if version.IsLessThan(addedInScheme) {
					validator.errors.Add(errors.NewFieldNotSupportedError(
						validator.context.Source,
						structField.Name,
						addedInScheme.String(),
						version.String(),
					))
				}
			}
		}

		// Check the "deprecatedIn" tag
		tag = structField.Tag.Get(tagDeprecatedIn)
		if tag != "" {
			deprecatedInScheme, err := NewVersion(tag)
			if err != nil {
				validator.errors.Add(err)
			} else {
				if version.IsGreaterOrEqualTo(deprecatedInScheme) {
					logger.Warnf(
						"config field '%s' was deprecated in scheme version %s (current config scheme: %s)",
						structField.Name, deprecatedInScheme.String(), version.String(),
					)
				}
			}
		}

		// Check the "removedIn" tag
		tag = structField.Tag.Get(tagRemovedIn)
		if tag != "" {
			removedInScheme, err := NewVersion(tag)
			if err != nil {
				validator.errors.Add(err)
			} else {
				if version.IsGreaterOrEqualTo(removedInScheme) {
					validator.errors.Add(errors.NewFieldRemovedError(
						validator.context.Source,
						structField.Name,
						removedInScheme.String(),
						version.String(),
					))
				}
			}
		}
	}
}

// isEmptyValue checks if a value is its empty type.
func (validator *schemeValidator) isEmptyValue(v reflect.Value) bool { // nolint: gocyclo
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
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			isEmpty := validator.isEmptyValue(v.Field(i))
			if !isEmpty {
				return false
			}
		}
		// If we get here, then all fields of the struct are empty,
		// so the struct is empty.
		return true
	default:
		logger.Warnf("No case for empty value check: %v", v.Kind())
	}
	return false
}

// validateDeviceFilter checks to make sure that a DeviceFilter has all of the
// fields populated that we need in order to process it as a valid request.
func validateDeviceFilter(request *synse.DeviceFilter) error {
	if request.GetDevice() == "" {
		return errors.InvalidArgumentErr("no device UID supplied in request")
	}
	if request.GetBoard() == "" {
		return errors.InvalidArgumentErr("no board supplied in request")
	}
	if request.GetRack() == "" {
		return errors.InvalidArgumentErr("no rack supplied in request")
	}
	return nil
}

// validateWriteInfo checks to make sure that a WriteInfo has all of the
// fields populated that we need in order to process it as a valid request.
func validateWriteInfo(request *synse.WriteInfo) error {
	return validateDeviceFilter(request.DeviceFilter)
}

// validateForRead validates that a device with the given device ID is readable.
func validateForRead(deviceID string) error {
	device := ctx.devices[deviceID]
	if device == nil {
		return fmt.Errorf("no device found with ID %s", deviceID)
	}

	if !device.IsReadable() {
		return fmt.Errorf("reading not enabled for device %s (no read handler)", deviceID)
	}
	return nil
}

// validateForWrite validates that a device with the given device ID is writable.
func validateForWrite(deviceID string) error {
	device := ctx.devices[deviceID]
	if device == nil {
		return fmt.Errorf("no device found with ID %s", deviceID)
	}

	if !device.IsWritable() {
		return fmt.Errorf("writing not enabled for device %s (no write handler)", deviceID)
	}
	return nil
}

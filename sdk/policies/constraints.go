package policies

import (
	"fmt"

	"github.com/vapor-ware/synse-sdk/sdk/errors"
)

type constraint func([]ConfigPolicy) error

var constraints = []constraint{
	constraintDeviceConfigNecessity,
	constraintPluginConfigNecessity,
}

// CheckConstraints checks the given slice of ConfigPolicies for constraint
// violations. All constraints are checked and all violations returned.
func CheckConstraints(policies []ConfigPolicy) *errors.MultiError {
	multiErr := errors.NewMultiError("ConfigPolicy Constraints")
	for _, constr := range constraints {
		err := constr(policies)
		if err != nil {
			multiErr.Add(err)
		}
	}
	return multiErr
}

// constraintPluginConfigNecessity checks that both PluginConfigRequired and
// PluginConfigOptional are not specified.
func constraintPluginConfigNecessity(policies []ConfigPolicy) error {
	var (
		hasReq, hasOpt bool
	)
	for _, policy := range policies {
		switch policy {
		case PluginConfigRequired:
			hasReq = true
		case PluginConfigOptional:
			hasOpt = true
		}
	}
	if hasReq && hasOpt {
		// FIXME: custom config policy error?
		return fmt.Errorf("both PluginConfigRequired and PluginConfigOptional are specified, but are mutually exclusive")
	}
	return nil
}

// constraintDeviceConfigNecessity checks that both DeviceConfigRequired and
// DeviceConfigOptional are not specified.
func constraintDeviceConfigNecessity(policies []ConfigPolicy) error {
	var (
		hasReq, hasOpt bool
	)
	for _, policy := range policies {
		switch policy {
		case DeviceConfigRequired:
			hasReq = true
		case DeviceConfigOptional:
			hasOpt = true
		}
	}
	if hasReq && hasOpt {
		// FIXME: custom config policy error?
		return fmt.Errorf("both DeviceConfigRequired and DeviceConfigOptional are specified, but are mutually exclusive")
	}
	return nil
}

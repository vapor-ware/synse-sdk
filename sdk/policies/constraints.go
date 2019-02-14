package policies

import (
	"fmt"
	"strings"

	"github.com/vapor-ware/synse-sdk/sdk/errors"
)

type constraint func([]ConfigPolicy) error

var constraints = []constraint{
	oneOrNoneOf(PluginConfigFileOptional, PluginConfigFileRequired, PluginConfigFileProhibited),
	oneOrNoneOf(DeviceConfigFileOptional, DeviceConfigFileRequired, DeviceConfigFileProhibited),
	oneOrNoneOf(DeviceConfigDynamicOptional, DeviceConfigDynamicRequired, DeviceConfigDynamicProhibited),
}

// checkConstraints checks the given slice of ConfigPolicies for constraint
// violations. All constraints are checked and all violations returned.
func checkConstraints(policies []ConfigPolicy) *errors.MultiError {
	multiErr := errors.NewMultiError("ConfigPolicy Constraints")
	for _, constr := range constraints {
		err := constr(policies)
		if err != nil {
			multiErr.Add(err)
		}
	}
	return multiErr
}

// oneOrNoneOf creates a constraint on the given set of policies where either
// none of the policies should be present, or only one of the policies should
// be present.
func oneOrNoneOf(policies ...ConfigPolicy) constraint {
	return func(p []ConfigPolicy) error {
		var has []ConfigPolicy
		for _, toCheck := range policies {
			for _, policy := range p {
				if toCheck == policy {
					has = append(has, policy)
					break
				}
			}
		}
		if len(has) > 1 {
			var names []string
			for _, policy := range has {
				names = append(names, policy.String())
			}
			return errors.NewPolicyViolationError(
				strings.Join(names, ","),
				fmt.Sprintf("constraint oneOrNoneOf{%v} broken, more than one matching policy found", policies),
			)
		}
		return nil
	}
}

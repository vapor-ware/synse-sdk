// Synse SDK
// Copyright (c) 2019 Vapor IO
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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

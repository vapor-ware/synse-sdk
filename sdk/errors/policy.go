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

package errors

import (
	"fmt"
)

// PolicyViolationError is used to designate an error arising from the resolution
// or enforcement of policies.
type PolicyViolationError struct {
	// policy is the name of the policy which caused the violation error.
	policy string

	// msg is the error message.
	msg string
}

// NewPolicyViolationError returns a new instance of a PolicyViolationError.
func NewPolicyViolationError(policy, msg string) *PolicyViolationError {
	return &PolicyViolationError{
		policy: policy,
		msg:    msg,
	}
}

// Error returns the error string and fulfils the error interface.
func (e *PolicyViolationError) Error() string {
	return fmt.Sprintf("policy violation (%s): %s", e.policy, e.msg)
}

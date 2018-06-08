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

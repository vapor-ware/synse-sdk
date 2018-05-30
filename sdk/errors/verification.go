package errors

import (
	"fmt"
)

// VerificationConflict is an error that is used when the config verification
// process detects that a config is invalid because there are some components
// in conflict with one another.
type VerificationConflict struct {
	// configType is the type of config that is being verified.
	configType string

	// msg is the error message.
	msg string
}

// NewVerificationConflictError creates a new instance of the VerificationConflict error.
func NewVerificationConflictError(configType, msg string) *VerificationConflict {
	return &VerificationConflict{
		configType: configType,
		msg:        msg,
	}
}

// Error returns the error string and fulfils the error interface.
func (e *VerificationConflict) Error() string {
	return fmt.Sprintf("conflict detected when verifying %s config: %s", e.configType, e.msg)
}

// VerificationInvalid is an error that is used when the config verification
// process detects that a config is invalid because a piece of data being referenced
// is expected but not found.
type VerificationInvalid struct {
	// configType is the type of config that is being verified.
	configType string

	// msg is the error message.
	msg string
}

// NewVerificationInvalidError creates a new instance of the VerificationInvalid error.
func NewVerificationInvalidError(configType, msg string) *VerificationInvalid {
	return &VerificationInvalid{
		configType: configType,
		msg:        msg,
	}
}

// Error returns the error string and fulfils the error interface.
func (e *VerificationInvalid) Error() string {
	return fmt.Sprintf("%s config invalid: %s", e.configType, e.msg)
}

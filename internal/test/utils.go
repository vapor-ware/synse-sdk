package test

import (
	"testing"
)

// CheckErr is a helper to check the error output for functions
// called during testing.
func CheckErr(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

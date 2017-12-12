package sdk

import (
	"strings"
	"testing"
)

func TestInvalidArgumentErr(t *testing.T) {
	errString := "test error"
	err := invalidArgumentErr(errString)

	if !strings.Contains(err.Error(), "InvalidArgument") {
		t.Error("invalidArgumentErr() -> unexpected error string")
	}
	if !strings.Contains(err.Error(), errString) {
		t.Error("invalidArgumentErr() -> unexpected error string")
	}
}

func TestNotFoundErr(t *testing.T) {
	errString := "test error"
	err := notFoundErr(errString)

	if !strings.Contains(err.Error(), "NotFound") {
		t.Error("notFoundErr() -> unexpected error string")
	}
	if !strings.Contains(err.Error(), errString) {
		t.Error("notFoundErr() -> unexpected error string")
	}
}

package test

import (
	"os"
	"testing"
)

// SetEnv is a wrapper around os.Setenv that handles error handling with the
// testing.T isntance.
func SetEnv(t *testing.T, key, value string) {
	err := os.Setenv(key, value)
	if err != nil {
		t.Fatal(err)
	}
}

// RemoveEnv is a wrapper around os.Unsetenv that handles error handling with the
// testing.T instance so this can be deferred easily.
func RemoveEnv(t *testing.T, key string) {
	err := os.Unsetenv(key)
	if err != nil {
		t.Fatal(err)
	}
}

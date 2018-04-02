package sdk

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewServer tests that a Server is returned when the constructor
// is called.
func TestNewServer(t *testing.T) {
	s, err := NewServer(&Plugin{})
	if err != nil {
		t.Errorf("NewServer should not return an error: %v", err)
	}
	if reflect.TypeOf(*s) != reflect.TypeOf(Server{}) {
		t.Error("NewServer did not return an instance of Server")
	}
}

// TestNewServerNilPlugin tests the NewServer function with a nil
// plugin parameter.
func TestNewServerNilPlugin(t *testing.T) {
	_, err := NewServer(nil)
	assert.Error(t, err)
}

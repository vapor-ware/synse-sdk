package sdk

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewServer tests that a Server is returned when the constructor
// is called.
func TestNewServer(t *testing.T) {
	s, err := newServer(&Plugin{})
	if err != nil {
		t.Errorf("newServer should not return an error: %v", err)
	}
	if reflect.TypeOf(*s) != reflect.TypeOf(server{}) {
		t.Error("newServer did not return an instance of Server")
	}
}

// TestNewServerNilPlugin tests the newServer function with a nil
// plugin parameter.
func TestNewServerNilPlugin(t *testing.T) {
	_, err := newServer(nil)
	assert.Error(t, err)
}

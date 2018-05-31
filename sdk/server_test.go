package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewServer tests that a Server is returned when the constructor
// is called.
func TestNewServer(t *testing.T) {
	s := NewServer("foo", "bar")
	assert.IsType(t, &Server{}, s)
	assert.Equal(t, "foo", s.network)
	assert.Equal(t, "bar", s.address)
}

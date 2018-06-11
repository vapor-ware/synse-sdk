package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfigsNotFound(t *testing.T) {
	err := NewConfigsNotFoundError([]string{"foo", "bar"})

	assert.IsType(t, &ConfigsNotFound{}, err)
	assert.Equal(t, 2, len(err.searchPaths))
	assert.Equal(t, "foo", err.searchPaths[0])
	assert.Equal(t, "bar", err.searchPaths[1])
}

func TestConfigsNotFound_Error(t *testing.T) {
	err := NewConfigsNotFoundError([]string{"foo", "bar"})
	out := err.Error()

	assert.Equal(t, "no configuration file(s) found in: [foo bar]", out)
}

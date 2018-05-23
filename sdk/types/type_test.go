package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestReadingType_Type tests getting the type of the reading
// from the namespaced ReadingType name.
func TestReadingType_Type(t *testing.T) {
	var testTable = []struct {
		name     string
		expected string
	}{
		{
			name:     "foo",
			expected: "foo",
		},
		{
			name:     "foo.bar",
			expected: "bar",
		},
		{
			name:     "test.device.sample.temperature",
			expected: "temperature",
		},
	}

	for _, tc := range testTable {
		readingType := ReadingType{Name: tc.name}
		assert.Equal(t, tc.expected, readingType.Type())
	}
}

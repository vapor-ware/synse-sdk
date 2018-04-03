package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	sys     interface{}
}

// NOTE: this is not the correct way to check the isDir bit, but since
// we will only set it this way for tests, this is fine here.
func (f testFileInfo) IsDir() bool        { return f.mode == os.ModeDir }
func (f testFileInfo) ModTime() time.Time { return f.modTime }
func (f testFileInfo) Mode() os.FileMode  { return f.mode }
func (f testFileInfo) Name() string       { return f.name }
func (f testFileInfo) Size() int64        { return f.size }
func (f testFileInfo) Sys() interface{}   { return f.sys }

// TestToStringMapI tests converting a map[interface{}]interface{} to
// a string map.
func TestToStringMapI(t *testing.T) {
	i := map[interface{}]interface{}{
		"test": "value",
	}

	_, err := toStringMapI(i)
	assert.NoError(t, err)
}

// TestToStringMapI2 tests converting a map[string]interface{} to
// a string map.
func TestToStringMapI2(t *testing.T) {
	i := map[string]interface{}{
		"test": "value",
	}

	_, err := toStringMapI(i)
	assert.NoError(t, err)
}

// TestToStringMapI3 tests converting a string to a string map -
// this should not be possible so we expect this to fail.
func TestToStringMapI3(t *testing.T) {
	i := "test"

	_, err := toStringMapI(i)
	assert.Error(t, err)
}

// TestToSliceStringMapI tests converting a []map[interface{}]interface{} to
// a slice string map.
func TestToSliceStringMapI(t *testing.T) {
	i := []interface{}{}
	i = append(i, map[interface{}]interface{}{"test": "value"})

	_, err := toSliceStringMapI(i)
	assert.NoError(t, err)
}

// TestToSliceStringMapI2 tests converting a []map[string]interface{} to
// a slice string map.
func TestToSliceStringMapI2(t *testing.T) {
	i := []map[string]interface{}{
		{"test": "value"},
	}

	_, err := toSliceStringMapI(i)
	assert.NoError(t, err)
}

// TestToSliceStringMapI3 tests converting a string to a slice string map -
// this should not be possible so we expect this to fail.
func TestToSliceStringMapI3(t *testing.T) {
	i := "test"

	_, err := toSliceStringMapI(i)
	assert.Error(t, err)
}

// TestIsValidConfig tests validating a config file when validation should fail because
// the file is a directory.
func TestIsValidConfig(t *testing.T) {
	fi := testFileInfo{
		name: "test",
		mode: os.ModeDir,
	}

	isValid := isValidConfig(fi)
	assert.False(t, isValid)
}

// TestIsValidConfig2 tests validating a config file when validation should fail because
// the given file does not have a supported file extension.
func TestIsValidConfig2(t *testing.T) {
	fi := testFileInfo{
		name: "test.json",
	}

	isValid := isValidConfig(fi)
	assert.False(t, isValid)
}

// TestIsValidConfig3 tests validating a config file that should be valid.
func TestIsValidConfig3(t *testing.T) {
	fi := testFileInfo{
		name: "test.yml",
	}

	isValid := isValidConfig(fi)
	assert.True(t, isValid)
}

// TestIsValidConfig4 tests validating a config file that should be valid.
func TestIsValidConfig4(t *testing.T) {
	fi := testFileInfo{
		name: "test.yaml",
	}

	isValid := isValidConfig(fi)
	assert.True(t, isValid)
}

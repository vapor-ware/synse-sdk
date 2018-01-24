package config

import (
	"os"
	"testing"
	"time"
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

func TestToStringMapI(t *testing.T) {
	i := map[interface{}]interface{}{
		"test": "value",
	}

	_, err := toStringMapI(i)
	if err != nil {
		t.Error(err)
	}
}

func TestToStringMapI2(t *testing.T) {
	i := map[string]interface{}{
		"test": "value",
	}

	_, err := toStringMapI(i)
	if err != nil {
		t.Error(err)
	}
}

func TestToStringMapI3(t *testing.T) {
	i := "test"

	_, err := toStringMapI(i)
	if err == nil {
		t.Error("expected error, but got nil instead")
	}
}

func TestToSliceStringMapI(t *testing.T) {
	i := []interface{}{}
	i = append(i, map[interface{}]interface{}{"test": "value"})

	_, err := toSliceStringMapI(i)
	if err != nil {
		t.Error(err)
	}
}

func TestToSliceStringMapI2(t *testing.T) {
	i := []map[string]interface{}{
		{"test": "value"},
	}

	_, err := toSliceStringMapI(i)
	if err != nil {
		t.Error(err)
	}
}

func TestToSliceStringMapI3(t *testing.T) {
	i := "test"

	_, err := toSliceStringMapI(i)
	if err == nil {
		t.Errorf("expected error, but got nil instead")
	}
}

func TestIsValidConfig(t *testing.T) {
	fi := testFileInfo{
		name: "test",
		mode: os.ModeDir,
	}

	isValid := isValidConfig(fi)
	if isValid {
		t.Error("expected validation failure: file info is a dir")
	}
}

func TestIsValidConfig2(t *testing.T) {
	fi := testFileInfo{
		name: "test.json",
	}

	isValid := isValidConfig(fi)
	if isValid {
		t.Error("expected validation failure: file info is not in supported file exts")
	}
}

func TestIsValidConfig3(t *testing.T) {
	fi := testFileInfo{
		name: "test.yml",
	}

	isValid := isValidConfig(fi)
	if !isValid {
		t.Error("expected config to be valid, but was not")
	}
}

func TestIsValidConfig4(t *testing.T) {
	fi := testFileInfo{
		name: "test.yaml",
	}

	isValid := isValidConfig(fi)
	if !isValid {
		t.Error("expected config to be valid, but was not")
	}
}

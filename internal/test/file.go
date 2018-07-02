package test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// FileInfo is a struct that fulfils the FileInfo interface that
// can be used for testing.
type FileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	sys     interface{}
}

// NewFileInfo creates a new instance of the test FileInfo for tests to use.
func NewFileInfo(name string, mode os.FileMode) *FileInfo {
	return &FileInfo{
		name: name,
		mode: mode,
	}
}

// NOTE: this is not the correct way to check the isDir bit, but since
// we will only set it this way for tests, this is fine here.

func (f FileInfo) IsDir() bool        { return f.mode == os.ModeDir } // nolint
func (f FileInfo) ModTime() time.Time { return f.modTime }            // nolint
func (f FileInfo) Mode() os.FileMode  { return f.mode }               // nolint
func (f FileInfo) Name() string       { return f.name }               // nolint
func (f FileInfo) Size() int64        { return f.size }               // nolint
func (f FileInfo) Sys() interface{}   { return f.sys }                // nolint

// TempDir holds the current directory being used as the test directory.
// This is generated via ioutil.TempDir.
var TempDir = ""

// SetupTestDir creates a test directory if one does not already exist.
func SetupTestDir(t *testing.T) {
	info, err := os.Stat(TempDir)
	if err != nil {
		if os.IsNotExist(err) {
			dir, e := ioutil.TempDir("", "test")
			if e != nil {
				t.Fatal(e)
			}
			TempDir = dir
		} else {
			t.Fatal(err)
		}
	} else {
		if !info.IsDir() {
			t.Error("testDir is set, but is not a directory")
		}
	}
}

// ClearTestDir removes the test directory, if it exists.
func ClearTestDir(t *testing.T) {
	if TempDir != "" {
		err := os.RemoveAll(TempDir)
		if err != nil {
			t.Fatal(err)
		}
	}
}

// WriteTempFile is a test helper that will write the specified file to
// a test directory. This is essentially a wrapper around ioutil.WriteFile
// that ensures the test directory is in place.
func WriteTempFile(t *testing.T, filename, data string, perm os.FileMode) string {
	if TempDir == "" {
		SetupTestDir(t)
	}

	fullPath := filepath.Join(TempDir, filename)
	err := ioutil.WriteFile(filepath.Join(TempDir, filename), []byte(data), perm)
	if err != nil {
		t.Fatal(err)
	}
	return fullPath
}

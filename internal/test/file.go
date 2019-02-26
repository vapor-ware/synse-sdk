// Synse SDK
// Copyright (c) 2019 Vapor IO
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package test

import (
	"os"
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

//// WriteTempFile is a test helper that will write the specified file to
//// a test directory. This is essentially a wrapper around ioutil.WriteFile
//// that ensures the test directory is in place.
//func WriteTempFile(t *testing.T, filename, data string, perm os.FileMode) string {
//	fullPath := filepath.Join(TempDir, filename)
//	err := ioutil.WriteFile(filepath.Join(TempDir, filename), []byte(data), perm)
//	if err != nil {
//		t.Fatal(err)
//	}
//	return fullPath
//}

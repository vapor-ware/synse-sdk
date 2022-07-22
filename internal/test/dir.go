// Synse SDK
// Copyright (c) 2017-2022 Vapor IO
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
	"io/ioutil"
	"os"
	"testing"
)

// TempDir is a test utility which creates a temporary directory for testing.
// It returns the directory path as well as a function to clean up the directory
// after the test.
func TempDir(t *testing.T) (string, func()) {
	dir, err := ioutil.TempDir("", "synsesdktest")
	if err != nil {
		t.Fatal(err)
	}

	return dir, func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}
}

// Synse SDK
// Copyright (c) 2017-2020 Vapor IO
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
	"testing"
)

// SetEnv is a wrapper around os.Setenv that handles error handling with the
// testing.T isntance.
func SetEnv(t *testing.T, key, value string) {
	err := os.Setenv(key, value)
	if err != nil {
		t.Fatal(err)
	}
}

// RemoveEnv is a wrapper around os.Unsetenv that handles error handling with the
// testing.T instance so this can be deferred easily.
func RemoveEnv(t *testing.T, key string) {
	err := os.Unsetenv(key)
	if err != nil {
		t.Fatal(err)
	}
}

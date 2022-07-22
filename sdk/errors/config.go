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

package errors

import "fmt"

// ConfigsNotFound is an error used when the search for a config file
// results in that file not being found.
type ConfigsNotFound struct {
	// searchPaths is the list of locations where the file was searched for.
	searchPaths []string
}

// NewConfigsNotFoundError returns a new instance of a ConfigsNotFound error.
func NewConfigsNotFoundError(searchPaths []string) *ConfigsNotFound {
	return &ConfigsNotFound{
		searchPaths: searchPaths,
	}
}

// Error returns the error string and fulfils the error interface.
func (e *ConfigsNotFound) Error() string {
	return fmt.Sprintf("no configuration file(s) found in: %s", e.searchPaths)
}

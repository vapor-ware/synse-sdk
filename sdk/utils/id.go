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

package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"strings"
)

// MakeIDString makes a compound string out of the given rack, board, and
// device identifier strings. This string should be a globally unique identifier
// for a given device.
func MakeIDString(rack, board, device string) string {
	return strings.Join([]string{rack, board, device}, "-")
}

// NewUID creates a new unique identifier for a Device. This id should be
// deterministic because it is a hash of various Device configuration components.
// A device's config should be unique, so the hash should be unique.
//
// These device IDs are not guaranteed to be globally unique, but they should
// be unique to the board they reside on.
func NewUID(components ...string) string {
	h := md5.New() // #nosec
	/* #nosec */
	for _, component := range components {
		io.WriteString(h, component) // nolint: errcheck
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

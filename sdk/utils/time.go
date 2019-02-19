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
	"time"

	log "github.com/Sirupsen/logrus"
)

// GetCurrentTime return the current time (time.Now()), with location set to UTC,
// as a string formatted with the RFC3339 layout. This should be the format
// of all timestamps returned by the SDK.
//
// The SDK uses this function to generate all of its timestamps. It is highly
// recommended that plugins use this as well for timestamp generation.
func GetCurrentTime() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// ParseRFC3339 parses a timestamp string in RFC3339 format into a Time struct.
// If it is given an empty string, it will return the zero-value for a Time
// instance. You can check if it is a zero time with the Time's `IsZero` method.
func ParseRFC3339(timestamp string) (t time.Time, err error) {
	if timestamp == "" {
		return
	}
	t, err = time.Parse(time.RFC3339, timestamp)
	if err != nil {
		log.WithField(
			"timestamp", timestamp,
		).Error("[sdk] failed to parse timestamp from RFC3339 format")
	}
	return
}

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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestParseRFC3339_Ok tests successfully parsing an RFC3339 timestamp
// into a Time struct.
func TestParseRFC3339_Ok(t *testing.T) {
	// get the EST location
	est, err := time.LoadLocation("EST")
	assert.NoError(t, err)

	var tests = []struct {
		timestamp string
		expected  time.Time
	}{
		{
			// no timestamp defaults to zero value for time
			timestamp: "",
			expected:  time.Time{},
		},
		{
			// rfc3339 utc
			timestamp: "2018-10-16T18:22:50Z",
			expected:  time.Date(2018, 10, 16, 18, 22, 50, 0, time.UTC),
		},
		{
			// rfc3339nano utc
			timestamp: "2018-10-16T18:22:50.573971054Z",
			expected:  time.Date(2018, 10, 16, 18, 22, 50, 573971054, time.UTC),
		},
		{
			// rcf3339 est
			timestamp: "2018-10-16T13:25:00-05:00",
			expected:  time.Date(2018, 10, 16, 13, 25, 0, 0, est),
		},
		{
			// rfc3339nano est
			timestamp: "2018-10-16T13:25:00.410241272-05:00",
			expected:  time.Date(2018, 10, 16, 13, 25, 0, 410241272, est),
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			actual, err := ParseRFC3339(tt.timestamp)
			assert.NoError(t, err)
			assert.True(t, tt.expected.Equal(actual))
		})
	}
}

// TestParseRFC3339_Error tests unsuccessfully parsing a timestamp into
// a Time struct
func TestParseRFC3339_Error(t *testing.T) {
	var tests = []string{
		"foobar",
		"...",
		"16 Oct 18 18:22 UTC",
		"16 Oct 18 13:25 EST",
		"Tue Oct 16 18:22:50 2018",
		"Tue Oct 16 13:25:00 2018",
		"6:22PM",
		"1:25PM",
		"Tue, 16 Oct 2018 18:22:50 +0000",
		"Tue, 16 Oct 2018 13:25:00 -0500",
		"Oct 16 18:22:50.573997",
		"Oct 16 13:25:00.410271",
		"Tue Oct 16 18:22:50 UTC 2018",
		"Tue Oct 16 13:25:00 EST 2018",
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			actual, err := ParseRFC3339(tt)
			assert.Error(t, err)
			assert.Empty(t, actual)
		})
	}
}

// TestGetCurrentTime tests getting the current time.
func TestGetCurrentTime(t *testing.T) {
	// TODO: figure out how to test the response...
	out := GetCurrentTime()
	assert.NotEmpty(t, out)
}

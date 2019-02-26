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

package health

import (
	"testing"

	"github.com/stretchr/testify/assert"
	synse "github.com/vapor-ware/synse-server-grpc/go"
)

func TestStatus_Encode_statusOk(t *testing.T) {
	s := &Status{
		Name:      "foo",
		Ok:        true,
		Message:   "",
		Timestamp: "now",
		Type:      PeriodicCheck,
	}

	encoded := s.Encode()
	assert.Equal(t, s.Name, encoded.Name)
	assert.Equal(t, s.Message, encoded.Message)
	assert.Equal(t, s.Timestamp, encoded.Timestamp)
	assert.Equal(t, string(s.Type), encoded.Type)
	assert.Equal(t, synse.HealthStatus_OK, encoded.Status)
}

func TestStatus_Encode_statusFailing(t *testing.T) {
	s := &Status{
		Name:      "foo",
		Ok:        false,
		Message:   "error message",
		Timestamp: "now",
		Type:      PeriodicCheck,
	}

	encoded := s.Encode()
	assert.Equal(t, s.Name, encoded.Name)
	assert.Equal(t, s.Message, encoded.Message)
	assert.Equal(t, s.Timestamp, encoded.Timestamp)
	assert.Equal(t, string(s.Type), encoded.Type)
	assert.Equal(t, synse.HealthStatus_FAILING, encoded.Status)
}

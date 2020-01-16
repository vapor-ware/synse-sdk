// Synse SDK
// Copyright (c) 2019-2020 Vapor IO
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
	synse "github.com/vapor-ware/synse-server-grpc/go"
)

// Status represents the status of a health Check at a given time.
type Status struct {
	Name      string
	Ok        bool
	Message   string
	Timestamp string
	Type      CheckType
}

// Encode converts the health Status to its corresponding gRPC message.
func (status *Status) Encode() *synse.V3HealthCheck {
	health := synse.HealthStatus_OK
	if !status.Ok {
		health = synse.HealthStatus_FAILING
	}

	return &synse.V3HealthCheck{
		Name:      status.Name,
		Message:   status.Message,
		Timestamp: status.Timestamp,
		Type:      string(status.Type),
		Status:    health,
	}
}

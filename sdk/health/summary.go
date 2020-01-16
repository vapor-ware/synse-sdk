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

// Summary provides a summary of overall plugin health.
type Summary struct {
	Timestamp string
	Ok        bool
	Checks    []*Status
}

// Encode converts the health Summary to its corresponding gRPC message.
func (summary *Summary) Encode() *synse.V3Health {
	var checks = make([]*synse.V3HealthCheck, len(summary.Checks))
	for i, check := range summary.Checks {
		checks[i] = check.Encode()
	}

	status := synse.HealthStatus_OK
	if !summary.Ok {
		status = synse.HealthStatus_FAILING
	}

	return &synse.V3Health{
		Timestamp: summary.Timestamp,
		Status:    status,
		Checks:    checks,
	}
}

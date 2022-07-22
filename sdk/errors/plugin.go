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

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnsupportedCommandError is an error that can be used to designate that
// a given device does not support an operation, e.g. write. Write is required
// by the PluginHandler interface, but if a device (e.g. a temperature sensor)
// doesn't support writing, this error can be returned.
type UnsupportedCommandError struct{}

func (e *UnsupportedCommandError) Error() string {
	return "Command not supported for given device."
}

// InvalidArgumentErr creates a gRPC InvalidArgument error with the given description.
func InvalidArgumentErr(format string, a ...interface{}) error {
	return status.Errorf(codes.InvalidArgument, format, a...)
}

// NotFoundErr creates a gRPC NotFound error with the given description.
func NotFoundErr(format string, a ...interface{}) error {
	return status.Errorf(codes.NotFound, format, a...)
}

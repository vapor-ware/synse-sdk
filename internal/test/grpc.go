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

package test

import (
	"fmt"

	synse "github.com/vapor-ware/synse-server-grpc/go"
	"google.golang.org/grpc"
)

//
// DEVICES
//

// MockDevicesStream mocks the stream for the Devices request, with no error.
type MockDevicesStream struct {
	grpc.ServerStream
	Results map[string]*synse.V3Device
}

// NewMockDevicesStream creates a new mock devices stream.
func NewMockDevicesStream() *MockDevicesStream {
	return &MockDevicesStream{
		Results: map[string]*synse.V3Device{},
	}
}

// Send fulfils the stream interface for the mock grpc stream.
func (mock *MockDevicesStream) Send(device *synse.V3Device) error {
	mock.Results[device.GetId()] = device
	return nil
}

// MockDevicesStreamErr mocks the stream for the Devices request, with error.
type MockDevicesStreamErr struct {
	grpc.ServerStream
}

// Send fulfils the stream interface for the mock grpc stream.
func (mock *MockDevicesStreamErr) Send(device *synse.V3Device) error {
	return fmt.Errorf("grpc error")
}

//
// READ
//

// MockReadStream mocks the stream for the Read request, with no error.
type MockReadStream struct {
	grpc.ServerStream
	Results []*synse.V3Reading
}

// NewMockReadStream creates a new mock read stream.
func NewMockReadStream() *MockReadStream {
	return &MockReadStream{
		Results: []*synse.V3Reading{},
	}
}

// Send fulfils the stream interface for the mock grpc stream.
func (mock *MockReadStream) Send(reading *synse.V3Reading) error {
	mock.Results = append(mock.Results, reading)
	return nil
}

// MockReadStreamErr mocks the stream for the Read request, with error.
type MockReadStreamErr struct {
	grpc.ServerStream
}

// Send fulfils the stream interface for the mock grpc stream.
func (mock *MockReadStreamErr) Send(reading *synse.V3Reading) error {
	return fmt.Errorf("grpc error")
}

//
// READ CACHED
//

// MockReadCachedStream mocks the stream for the ReadCached request, with no error.
type MockReadCachedStream struct {
	grpc.ServerStream
	Results []*synse.V3Reading
}

// NewMockReadCachedStream creates a new mock read cache stream.
func NewMockReadCachedStream() *MockReadCachedStream {
	return &MockReadCachedStream{
		Results: []*synse.V3Reading{},
	}
}

// Send fulfils the stream interface for the mock grpc stream.
func (mock *MockReadCachedStream) Send(reading *synse.V3Reading) error {
	mock.Results = append(mock.Results, reading)
	return nil
}

// MockReadCachedStreamErr mocks the stream for a ReadCached request, with error.
type MockReadCachedStreamErr struct {
	grpc.ServerStream
}

// Send fulfils the stream interface for the mock grpc stream.
func (mock *MockReadCachedStreamErr) Send(reading *synse.V3Reading) error {
	return fmt.Errorf("grpc error")
}

//
// WRITE ASYNC
//

// MockWriteAsyncStream mocks the stream for the AsyncWrite request, with no error.
type MockWriteAsyncStream struct {
	grpc.ServerStream
	Results map[string]*synse.V3WriteTransaction
}

// NewMockWriteAsyncStream creates a new mock async write stream.
func NewMockWriteAsyncStream() *MockWriteAsyncStream {
	return &MockWriteAsyncStream{
		Results: map[string]*synse.V3WriteTransaction{},
	}
}

// Send fulfils the stream interface for the mock grpc stream.
func (mock *MockWriteAsyncStream) Send(write *synse.V3WriteTransaction) error {
	mock.Results[write.Id] = write
	return nil
}

// MockWriteAsyncStreamErr mocks the stream for the async write request, with error.
type MockWriteAsyncStreamErr struct {
	grpc.ServerStream
}

// Send fulfils the stream interface for the mock grpc stream.
func (mock *MockWriteAsyncStreamErr) Send(write *synse.V3WriteTransaction) error {
	return fmt.Errorf("grpc error")
}

//
// WRITE SYNC
//

// MockWriteSyncStream mocks the stream for the SyncWrite request, with no error.
type MockWriteSyncStream struct {
	grpc.ServerStream
	Results map[string]*synse.V3TransactionStatus
}

// NewMockWriteSyncStream creates a new mock async write stream.
func NewMockWriteSyncStream() *MockWriteSyncStream {
	return &MockWriteSyncStream{
		Results: map[string]*synse.V3TransactionStatus{},
	}
}

// Send fulfils the stream interface for the mock grpc stream.
func (mock *MockWriteSyncStream) Send(write *synse.V3TransactionStatus) error {
	mock.Results[write.Id] = write
	return nil
}

// MockWriteSyncStreamErr mocks the stream for the sync write request, with error.
type MockWriteSyncStreamErr struct {
	grpc.ServerStream
}

// Send fulfils the stream interface for the mock grpc stream.
func (mock *MockWriteSyncStreamErr) Send(write *synse.V3TransactionStatus) error {
	return fmt.Errorf("grpc error")
}

//
// TRANSACTION
//

// MockTransactionStream mocks the stream for the Transaction request, with no error.
type MockTransactionStream struct {
	grpc.ServerStream
	Results map[string]*synse.V3TransactionStatus
}

// NewMockTransactionStream creates a new mock transaction stream.
func NewMockTransactionStream() *MockTransactionStream {
	return &MockTransactionStream{
		Results: map[string]*synse.V3TransactionStatus{},
	}
}

// Send fulfils the stream interface for the mock grpc stream.
func (mock *MockTransactionStream) Send(write *synse.V3TransactionStatus) error {
	mock.Results[write.Id] = write
	return nil
}

// MockTransactionStreamErr mocks the stream for the Transaction request, with error.
type MockTransactionStreamErr struct {
	grpc.ServerStream
}

// Send fulfils the stream interface for the mock grpc stream.
func (mock *MockTransactionStreamErr) Send(write *synse.V3TransactionStatus) error {
	return fmt.Errorf("grpc error")
}

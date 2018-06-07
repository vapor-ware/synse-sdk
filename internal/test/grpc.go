package test

import (
	"fmt"

	"github.com/vapor-ware/synse-server-grpc/go"
	"google.golang.org/grpc"
)

// MockCapabilitiesStream mocks the stream for the Capabilities request, with no error.
type MockCapabilitiesStream struct {
	grpc.ServerStream
	Results map[string]*synse.DeviceCapability
}

// NewMockCapabilitiesStream creates a new mock capabilities stream.
func NewMockCapabilitiesStream() *MockCapabilitiesStream {
	return &MockCapabilitiesStream{
		Results: map[string]*synse.DeviceCapability{},
	}
}

// Send fulfils the stream interface for the mock grpc stream.
func (mock *MockCapabilitiesStream) Send(capability *synse.DeviceCapability) error {
	mock.Results[capability.Kind] = capability
	return nil
}

// MockCapabilitiesStreamErr mocks the stream for the Capabilities request, with error.
type MockCapabilitiesStreamErr struct {
	grpc.ServerStream
}

// Send fulfils the stream interface for the mock grpc stream.
func (mock *MockCapabilitiesStreamErr) Send(capability *synse.DeviceCapability) error {
	return fmt.Errorf("grpc error")
}

// MockDevicesStream mocks the stream for the Devices request, with no error.
type MockDevicesStream struct {
	grpc.ServerStream
	Results map[string]*synse.Device
}

// NewMockDevicesStream creates a new mock devices stream.
func NewMockDevicesStream() *MockDevicesStream {
	return &MockDevicesStream{
		Results: map[string]*synse.Device{},
	}
}

// Send fulfils the stream interface for the mock grpc stream.
func (mock *MockDevicesStream) Send(device *synse.Device) error {
	mock.Results[device.GetUid()] = device
	return nil
}

// MockDevicesStreamErr mocks the stream for the Devices request, with error.
type MockDevicesStreamErr struct {
	grpc.ServerStream
}

// Send fulfils the stream interface for the mock grpc stream.
func (mock *MockDevicesStreamErr) Send(device *synse.Device) error {
	return fmt.Errorf("grpc error")
}

// MockReadStream mocks the stream for the Read request, with no error.
type MockReadStream struct {
	grpc.ServerStream
	Results []*synse.Reading
}

// NewMockReadStream creates a new mock read stream.
func NewMockReadStream() *MockReadStream {
	return &MockReadStream{
		Results: []*synse.Reading{},
	}
}

// Send fulfils the stream interface for the mock grpc stream.
func (mock *MockReadStream) Send(reading *synse.Reading) error {
	mock.Results = append(mock.Results, reading)
	return nil
}

// MockReadStreamErr mocks the stream for the Read request, with error.
type MockReadStreamErr struct {
	grpc.ServerStream
}

// Send fulfils the stream interface for the mock grpc stream.
func (mock *MockReadStreamErr) Send(reading *synse.Reading) error {
	return fmt.Errorf("grpc error")
}

// MockTransactionStream mocks the stream for the Transaction request, with no error.
type MockTransactionStream struct {
	grpc.ServerStream
	Results map[string]*synse.WriteResponse
}

// NewMockTransactionStream creates a new mock transaction stream.
func NewMockTransactionStream() *MockTransactionStream {
	return &MockTransactionStream{
		Results: map[string]*synse.WriteResponse{},
	}
}

// Send fulfils the stream interface for the mock grpc stream.
func (mock *MockTransactionStream) Send(write *synse.WriteResponse) error {
	mock.Results[write.Id] = write
	return nil
}

// MockTransactionStreamErr mocks the stream for the Transaction request, with error.
type MockTransactionStreamErr struct {
	grpc.ServerStream
}

// Send fulfils the stream interface for the mock grpc stream.
func (mock *MockTransactionStreamErr) Send(write *synse.WriteResponse) error {
	return fmt.Errorf("grpc error")
}

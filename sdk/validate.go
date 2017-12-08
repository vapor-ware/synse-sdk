package sdk

import (
	"github.com/vapor-ware/synse-server-grpc/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// validateReadRequest checks to make sure that a ReadRequest has all of the
// fields populated that we need in order to process it as a valid request.
func validateReadRequest(request *synse.ReadRequest) error {
	device := request.GetDevice()
	if device == "" {
		return status.Errorf(codes.InvalidArgument, "No device UID supplied to Read.")
	}
	board := request.GetBoard()
	if board == "" {
		return status.Errorf(codes.InvalidArgument, "No board supplied to Read.")
	}
	rack := request.GetRack()
	if rack == "" {
		return status.Errorf(codes.InvalidArgument, "No rack supplied to Read.")
	}
	return nil
}

// validateWriteRequest checks to make sure that a ReadRequest has all of the
// fields populated that we need in order to process it as a valid request.
func validateWriteRequest(request *synse.WriteRequest) error {
	device := request.GetDevice()
	if device == "" {
		return status.Errorf(codes.InvalidArgument, "No device UID supplied to Write.")
	}
	board := request.GetBoard()
	if board == "" {
		return status.Errorf(codes.InvalidArgument, "No board supplied to Write.")
	}
	rack := request.GetRack()
	if rack == "" {
		return status.Errorf(codes.InvalidArgument, "No rack supplied to Write.")
	}
	return nil
}

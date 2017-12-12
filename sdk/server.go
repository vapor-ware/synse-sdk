package sdk

import (
	"fmt"
	"net"

	"github.com/vapor-ware/synse-server-grpc/go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server is the the Plugin's server component. It acts as the InternalApiServer
// for the gRPC server.
type Server struct {
	plugin *Plugin
}

// NewServer
func NewServer(plugin *Plugin) *Server {
	return &Server{
		plugin: plugin,
	}
}

// serve
func (s *Server) serve() error {
	// set up the gRPC server
	var address string
	var err error
	if Config.Socket.Network == "unix" {
		// if we are configuring for unix socket communication, we first need
		// to create the socket.
		address, err = setupSocket(Config.Socket.Address)
		if err != nil {
			return err
		}
	} else {
		// otherwise, we will just use the address specified in the configuration
		address = Config.Socket.Address
	}

	Logger.Infof("[grpc] listening on network %v with address %v", Config.Socket.Network, address)
	lis, err := net.Listen(Config.Socket.Network, address)
	if err != nil {
		Logger.Fatalf("Failed to listen: %v", err)
		return err
	}

	// create the GRPC server and register our plugin server to it
	svr := grpc.NewServer()
	Logger.Debugf("[grpc] creating new grpc server")
	synse.RegisterInternalApiServer(svr, s)
	Logger.Debugf("[grpc] registering handlers")

	// start gRPC the server
	Logger.Infof("[grpc] serving")
	if err := svr.Serve(lis); err != nil {
		Logger.Fatalf("Failed to serve: %v", err)
		return err
	}
	return nil
}

func (s *Server) Read(req *synse.ReadRequest, stream synse.InternalApi_ReadServer) error {
	responses, err := s.plugin.dm.Read(req)
	if err != nil {
		return err
	}
	for _, response := range responses {
		if err := stream.Send(response); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) Write(ctx context.Context, req *synse.WriteRequest) (*synse.Transactions, error) {
	transactions, err := s.plugin.dm.Write(req)
	if err != nil {
		return nil, err
	}
	return &synse.Transactions{
		Transactions: transactions,
	}, nil
}

func (s *Server) Metainfo(req *synse.MetainfoRequest, stream synse.InternalApi_MetainfoServer) error {
	for _, device := range deviceMap {
		if err := stream.Send(device.toMetainfoResponse()); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) TransactionCheck(ctx context.Context, in *synse.TransactionId) (*synse.WriteResponse, error) {
	transaction := GetTransaction(in.Id)
	if transaction == nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Transaction %v not found.", in.Id))
	}
	return transaction.encode(), nil
}

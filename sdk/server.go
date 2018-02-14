package sdk

import (
	"net"

	"github.com/vapor-ware/synse-sdk/sdk/logger"
	"github.com/vapor-ware/synse-server-grpc/go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Server is the the Plugin's server component. It acts as the InternalApiServer
// for the gRPC server.
type Server struct {
	plugin *Plugin
}

// NewServer creates a new instance of a Server.
func NewServer(plugin *Plugin) (*Server, error) {
	if plugin == nil {
		return nil, invalidArgumentErr("plugin parameter must not be nil")
	}
	server := &Server{
		plugin: plugin,
	}
	return server, nil
}

// setup gets the network and address string which are used as parameters
// to net.Listen(). Any additional setup happens here, e.g. if using the "unix"
// network type, this will create the necessary unix socket.
func (s *Server) setup() (string, string, error) {
	var network = s.plugin.Config.Network.Type
	var address string
	var err error

	if network == "unix" {
		address, err = setupSocket(s.plugin.Config.Network.Address)
		if err != nil {
			return "", "", err
		}
	} else {
		// otherwise, we will just use the address specified in the configuration
		address = s.plugin.Config.Network.Address
	}

	return network, address, nil
}

// serve configures and runs the gRPC server.
func (s *Server) serve() error {

	network, address, err := s.setup()
	if err != nil {
		return err
	}

	logger.Infof("listening on network %v with address %v", network, address)
	lis, err := net.Listen(network, address)
	if err != nil {
		logger.Fatalf("Failed to listen: %v", err)
		return err
	}

	// create the GRPC server and register our plugin server to it
	svr := grpc.NewServer()
	synse.RegisterInternalApiServer(svr, s)

	// start gRPC the server
	logger.Infof("serving")
	if err := svr.Serve(lis); err != nil {
		logger.Fatalf("Failed to serve: %v", err)
		return err
	}
	return nil
}

// Read is the handler for gRPC Read requests.
func (s *Server) Read(req *synse.ReadRequest, stream synse.InternalApi_ReadServer) error {
	responses, err := s.plugin.dataManager.Read(req)
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

// Write is the handler for gRPC Write requests.
func (s *Server) Write(ctx context.Context, req *synse.WriteRequest) (*synse.Transactions, error) {
	transactions, err := s.plugin.dataManager.Write(req)
	if err != nil {
		return nil, err
	}
	return &synse.Transactions{
		Transactions: transactions,
	}, nil
}

// Metainfo is the handler for gRPC Metainfo requests.
func (s *Server) Metainfo(req *synse.MetainfoRequest, stream synse.InternalApi_MetainfoServer) error {
	for _, device := range deviceMap {
		if err := stream.Send(device.encode()); err != nil {
			return err
		}
	}
	return nil
}

// TransactionCheck is the handler for gRPC TransactionCheck requests.
func (s *Server) TransactionCheck(ctx context.Context, in *synse.TransactionId) (*synse.WriteResponse, error) {
	transaction := GetTransaction(in.Id)
	if transaction == nil {
		return nil, notFoundErr("transaction not found: %v", in.Id)
	}
	return transaction.encode(), nil
}

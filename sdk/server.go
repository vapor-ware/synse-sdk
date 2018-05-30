package sdk

import (
	"net"

	"github.com/vapor-ware/synse-server-grpc/go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/vapor-ware/synse-sdk/sdk/logger"
)

type Server struct {
	network string
	address string
}

func NewServer(network, address string) *Server {
	return &Server{
		network: network,
		address: address,
	}
}

func (server *Server) Serve() (err error) {
	var addr string

	switch server.network {
	case "unix":
		addr, err = setupSocket(server.address)
		if err != nil {
			return
		}
	default:
		addr = server.address
	}

	lis, err := net.Listen(server.network, addr)
	if err != nil {
		return
	}

	svr := grpc.NewServer()
	synse.RegisterPluginServer(svr, server)

	if err = svr.Serve(lis); err != nil {
		return
	}
	return nil
}

// Test is the handler for the Synse GRPC Plugin service's `Test` RPC method.
func (server *Server) Test(ctx context.Context, request *synse.Empty) (*synse.Status, error) {
	logger.Debug("gRPC server: test")
	return &synse.Status{Ok: true}, nil
}

// Version is the handler for the Synse GRPC Plugin service's `Version` RPC method.
func (server *Server) Version(ctx context.Context, request *synse.Empty) (*synse.VersionInfo, error) {
	logger.Debug("gRPC server: version")
	ver := GetVersion()
	return &synse.VersionInfo{
		PluginVersion: ver.PluginVersion,
		SdkVersion:    ver.SDKVersion,
		BuildDate:     ver.BuildDate,
		GitCommit:     ver.GitCommit,
		GitTag:        ver.GitTag,
		Arch:          ver.Arch,
		Os:            ver.OS,
	}, nil
}

// Health is the handler for the Synse GRPC Plugin service's `Health` RPC method.
func (server *Server) Health(ctx context.Context, request *synse.Empty) (*synse.PluginHealth, error) {
	logger.Debug("gRPC server: health")
	// TODO
	return nil, nil
}

// Capabilities is the handler for the Synse GRPC Plugin service's `Capabilities` RPC method.
func (server *Server) Capabilities(request *synse.Empty, stream synse.Plugin_CapabilitiesServer) error {
	logger.Debug("gRPC server: capabilities")
	// TODO
	return nil
}

// Devices is the handler for the Synse GRPC Plugin service's `Devices` RPC method.
func (server *Server) Devices(request *synse.DeviceFilter, stream synse.Plugin_DevicesServer) error {
	logger.Debug("gRPC server: devices")
	// TODO
	return nil
}

// Metainfo is the handler for the Synse GRPC Plugin service's `Metainfo` RPC method.
func (server *Server) Metainfo(request *synse.DeviceFilter, stream synse.Plugin_MetainfoServer) error {
	logger.Debug("gRPC server: metainfo")
	// TODO
	return nil
}

// Read is the handler for the Synse GRPC Plugin service's `Read` RPC method.
func (server *Server) Read(request *synse.DeviceFilter, stream synse.Plugin_ReadServer) error {
	logger.Debug("gRPC server: read")
	// TODO
	responses, err := DataManager.Read(request)
	if err != nil {
		logger.Errorf("%v - gRPC read error: %v", request, err)
		return err
	}
	for _, response := range responses {
		logger.Debugf("%v gRPC read response: %v", request, response)
		if err := stream.Send(response); err != nil {
			logger.Errorf("%v - gRPC read error when sending response(s): %v", request, err)
			return err
		}
	}
	return nil
}

// Write is the handler for the Synse GRPC Plugin service's `Write` RPC method.
func (server *Server) Write(ctx context.Context, request *synse.WriteInfo) (*synse.Transactions, error) {
	logger.Debug("gRPC server: write")
	// TODO
	transactions, err := DataManager.Write(request)
	if err != nil {
		logger.Errorf("%v - gRPC write error: %v", request, err)
		return nil, err
	}
	resp := &synse.Transactions{
		Transactions: transactions,
	}
	logger.Debugf("%v gRPC write response: %v", request, resp)
	return resp, nil
}

// Transaction is the handler for the Synse GRPC Plugin service's `Transaction` RPC method.
func (server *Server) Transaction(request *synse.TransactionFilter, stream synse.Plugin_TransactionServer) error {
	logger.Debug("gRPC server: transaction")
	// TODO
	return nil
}

//
//
//// server is the the Plugin's server component. It acts as the InternalApiServer
//// for the gRPC server.
//type server struct {
//	plugin *Plugin
//}
//
//// newServer creates a new instance of a server.
//func newServer(plugin *Plugin) (*server, error) {
//	if plugin == nil {
//		return nil, invalidArgumentErr("plugin parameter must not be nil")
//	}
//	server := &server{
//		plugin: plugin,
//	}
//	return server, nil
//}
//
//// setup gets the network and address string which are used as parameters
//// to net.Listen(). Any additional setup happens here, e.g. if using the "unix"
//// network type, this will create the necessary unix socket.
//func (s *server) setup() (string, string, error) {
//	var network = s.plugin.Config.Network.Type
//	var address string
//	var err error
//
//	if network == "unix" {
//		address, err = setupSocket(s.plugin.Config.Network.Address)
//		if err != nil {
//			return "", "", err
//		}
//	} else {
//		// otherwise, we will just use the address specified in the configuration
//		address = s.plugin.Config.Network.Address
//	}
//
//	return network, address, nil
//}
//
//// serve configures and runs the gRPC server.
//func (s *server) serve() error {
//	logger.Info("Starting gRPC server")
//
//	network, address, err := s.setup()
//	if err != nil {
//		logger.Errorf("gRPC server setup error: %v", err)
//		return err
//	}
//
//	logger.Infof("grpc server listening on %v: %v", network, address)
//	lis, err := net.Listen(network, address)
//	if err != nil {
//		logger.Errorf("Failed to listen on %v %v: %v", network, address, err)
//		return err
//	}
//
//	// create the GRPC server and register our plugin server to it
//	svr := grpc.NewServer()
//	synse.RegisterInternalApiServer(svr, s)
//
//	// start gRPC the server
//	logger.Info("serving...")
//	if err := svr.Serve(lis); err != nil {
//		logger.Fatalf("gRPC server error while serving: %v", err)
//		return err
//	}
//	return nil
//}
//
//// Read is the handler for gRPC Read requests.
//func (s *server) Read(req *synse.ReadRequest, stream synse.InternalApi_ReadServer) error {
//	logger.Debugf("gRPC read: %v", req)
//	responses, err := s.plugin.dataManager.Read(req)
//	if err != nil {
//		logger.Errorf("%v - gRPC read error: %v", req, err)
//		return err
//	}
//	for _, response := range responses {
//		logger.Debugf("%v gRPC read response: %v", req, response)
//		if err := stream.Send(response); err != nil {
//			logger.Errorf("%v - gRPC read error when sending response(s): %v", req, err)
//			return err
//		}
//	}
//	return nil
//}
//
//// Write is the handler for gRPC Write requests.
//func (s *server) Write(ctx context.Context, req *synse.WriteRequest) (*synse.Transactions, error) {
//	transactions, err := s.plugin.dataManager.Write(req)
//	logger.Debugf("gRPC write: %v", req)
//	if err != nil {
//		logger.Errorf("%v - gRPC write error: %v", req, err)
//		return nil, err
//	}
//	resp := &synse.Transactions{
//		Transactions: transactions,
//	}
//	logger.Debugf("%v gRPC write response: %v", req, resp)
//	return resp, nil
//}
//
//// Metainfo is the handler for gRPC Metainfo requests.
//func (s *server) Metainfo(req *synse.MetainfoRequest, stream synse.InternalApi_MetainfoServer) error {
//	logger.Debugf("gRPC metainfo: %v", req)
//	// Preserve device order in responses so that synse gets them in configuration order.
//	for i := 0; i < len(deviceMapOrder); i++ {
//		id := deviceMapOrder[i]
//		device := deviceMap[id]
//		if err := stream.Send(device.encode()); err != nil {
//			logger.Errorf("%v - gRPC metainfo error when sending response(s): %v", req, err)
//			return err
//		}
//	}
//	return nil
//}
//
//// TransactionCheck is the handler for gRPC TransactionCheck requests.
//func (s *server) TransactionCheck(ctx context.Context, in *synse.TransactionId) (*synse.WriteResponse, error) {
//	logger.Debugf("gRPC transaction check: %v", in)
//	transaction, err := getTransaction(in.Id)
//	if err != nil {
//		logger.Errorf("%v - gRPC transaction check error: %v", in, err)
//		return nil, err
//	}
//	if transaction == nil {
//		return nil, notFoundErr("transaction not found: %v", in.Id)
//	}
//	return transaction.encode(), nil
//}

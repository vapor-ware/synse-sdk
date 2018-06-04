package sdk

import (
	"net"

	"github.com/vapor-ware/synse-server-grpc/go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/logger"
)

// Server implements the Synse Plugin gRPC server. It is used by the
// plugin to communicate via gRPC over tcp or unix socket to Synse Server.
type Server struct {
	network string
	address string
}

// NewServer creates a new instance of a Server. This should be used
// by the plugin to create its server instance.
func NewServer(network, address string) *Server {
	return &Server{
		network: network,
		address: address,
	}
}

// Serve sets up the gRPC server and runs it.
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
	return &synse.VersionInfo{
		PluginVersion: Version.PluginVersion,
		SdkVersion:    Version.SDKVersion,
		BuildDate:     Version.BuildDate,
		GitCommit:     Version.GitCommit,
		GitTag:        Version.GitTag,
		Arch:          Version.Arch,
		Os:            Version.OS,
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
	capabilitiesMap := map[string]*synse.DeviceCapability{}

	for _, device := range deviceMap {
		_, hasKind := capabilitiesMap[device.Kind]
		if !hasKind {
			var outputs []string
			for _, o := range device.Outputs {
				outputs = append(outputs, o.Name)
			}
			capabilitiesMap[device.Kind] = &synse.DeviceCapability{
				Kind:    device.Kind,
				Outputs: outputs,
			}
		}
	}

	for _, capability := range capabilitiesMap {
		if err := stream.Send(capability); err != nil {
			return err
		}
	}
	return nil
}

// Devices is the handler for the Synse GRPC Plugin service's `Devices` RPC method.
func (server *Server) Devices(request *synse.DeviceFilter, stream synse.Plugin_DevicesServer) error {
	logger.Debug("gRPC server: devices")
	for _, device := range deviceMap {
		if err := stream.Send(device.encode()); err != nil {
			return err
		}
	}
	return nil
}

// Metainfo is the handler for the Synse GRPC Plugin service's `Metainfo` RPC method.
func (server *Server) Metainfo(ctx context.Context, request *synse.Empty) (*synse.Metadata, error) {
	logger.Debug("gRPC server: metainfo")
	return &synse.Metadata{
		Name:        metainfo.Name,
		Maintainer:  metainfo.Maintainer,
		Description: metainfo.Description,
		Vcs:         metainfo.VCS,
		Version: &synse.VersionInfo{
			PluginVersion: Version.PluginVersion,
			SdkVersion:    Version.SDKVersion,
			BuildDate:     Version.BuildDate,
			GitCommit:     Version.GitCommit,
			GitTag:        Version.GitTag,
			Arch:          Version.Arch,
			Os:            Version.OS,
		},
	}, nil
}

// Read is the handler for the Synse GRPC Plugin service's `Read` RPC method.
func (server *Server) Read(request *synse.DeviceFilter, stream synse.Plugin_ReadServer) error {
	logger.Debug("gRPC server: read")
	responses, err := DataManager.Read(request)
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

// Write is the handler for the Synse GRPC Plugin service's `Write` RPC method.
func (server *Server) Write(ctx context.Context, request *synse.WriteInfo) (*synse.Transactions, error) {
	logger.Debug("gRPC server: write")
	transactions, err := DataManager.Write(request)
	if err != nil {
		return nil, err
	}
	return &synse.Transactions{
		Transactions: transactions,
	}, nil
}

// Transaction is the handler for the Synse GRPC Plugin service's `Transaction` RPC method.
func (server *Server) Transaction(request *synse.TransactionFilter, stream synse.Plugin_TransactionServer) error {
	logger.Debug("gRPC server: transaction")

	// If there is no ID with the incoming request, return all cached transactions.
	if request.Id == "" {
		for _, item := range transactionCache.Items() {
			t, ok := item.Object.(*transaction)
			if ok {
				if err := stream.Send(t.encode()); err != nil {
					return err
				}
			}
		}
	}

	// Otherwise, return the transaction with the specified ID.
	transaction := getTransaction(request.Id)
	if transaction == nil {
		return errors.NotFoundErr("transaction not found: %v", request.Id)
	}
	return stream.Send(transaction.encode())
}

package sdk

import (
	"fmt"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/vapor-ware/synse-server-grpc/go"
)

// PluginServer is the server that is used to run the plugin's read-write loop,
// to track device metainfo, and to server gRPC requests.
type PluginServer struct {
	name           string
	readingManager ReadingManager
	writingManager WritingManager
	pluginDevices  map[string]*Device
	rwLoop         RWLoop
}


// configureDevices is a convenience function to parse all of the plugin
// configuration files, generate Device instances for each of the configured
// devices, and then populate the pluginDevices map which is used to store
// and access these device models. Additionally, if the plugin is set to
// auto-enumerate its devices, this kicks that off.
func (ps *PluginServer) configureDevices(deviceHandler DeviceHandler) error {

	var instanceCfg []*DeviceConfig

	// get any instance configurations from plugin-defined enumeration function
	for _, enumCfg := range Config.AutoEnumerate {
		deviceEnum, err := deviceHandler.EnumerateDevices(enumCfg)
		if err != nil {
			Logger.Errorf("Error enumerating devices with %+v: %v", enumCfg, err)
		} else {
			instanceCfg = append(instanceCfg, deviceEnum...)
		}
	}

	// get any instance configurations from YAML
	deviceCfg, err := parseDeviceConfig(configDir)
	if err != nil {
		return err
	}
	instanceCfg = append(instanceCfg, deviceCfg...)

	// get the prototype configurations from YAML
	protoCfg, err := parsePrototypeConfig(configDir)
	if err != nil {
		return err
	}

	// make the composite device records
	devices := makeDevices(instanceCfg, protoCfg, deviceHandler)

	ps.pluginDevices = make(map[string]*Device)
	for _, device := range devices {
		ps.pluginDevices[device.IDString()] = device
	}
	ps.rwLoop.devices = ps.pluginDevices
	return nil
}


// Read is the gRPC handler for read requests.
func (ps *PluginServer) Read(in *synse.ReadRequest, stream synse.InternalApi_ReadServer) error {
	Logger.Debug("[grpc] READ")

	responses, err := ps.readingManager.read(in)
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

// Write is the gRPC handler for write requests.
func (ps *PluginServer) Write(ctx context.Context, in *synse.WriteRequest) (*synse.Transactions, error) {
	Logger.Debug("[grpc] WRITE")

	transactions, err := ps.writingManager.write(in); if err != nil {
		return nil, err
	}

	return &synse.Transactions{
		Transactions: transactions,
	}, nil
}


// Metainfo is the gRPC handler for metainfo requests.
func (ps *PluginServer) Metainfo(in *synse.MetainfoRequest, stream synse.InternalApi_MetainfoServer) error {
	Logger.Debug("[grpc] METAINFO")

	for _, device := range ps.pluginDevices {
		if err := stream.Send(device.toMetainfoResponse()); err != nil {
			return err
		}
	}
	return nil
}


// TransactionCheck is the gRPC handler for transaction check requests.
func (ps *PluginServer) TransactionCheck(ctx context.Context, in *synse.TransactionId) (*synse.WriteResponse, error) {
	Logger.Debug("[grpc] TRANSACTION CHECK")

	transaction := GetTransaction(in.Id)
	if transaction == nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Transaction %v not found.", in.Id))
	}
	return transaction.toGRPC(), nil
}


// Run starts the PluginServer instance. It starts the background reads,
// the read-write loop, and the gRPC server. The gRPC server is configured
// to communicate over a UNIX socket that is created in a well-known location
// and has the same name as the plugin. Synse server will discover and
// communicate with the plugin using the UNIX socket.
func (ps *PluginServer) Run() error {

	Logger.Infof("[plugin server] Running server with SDK version %v", Version)

	// Start the reading manager
	ps.readingManager.start()

	// start the RW loop
	ps.rwLoop.run()

	// create the socket used to communicate with the gRPC server
	socket, err := setupSocket(ps.name)
	if err != nil {
		return err
	}

	Logger.Infof("[grpc] listening on socket %v", socket)
	lis, err := net.Listen("unix", socket)
	if err != nil {
		Logger.Fatalf("Failed to listen: %v", err)
		return err
	}

	// create the GRPC server and register our plugin server to it
	svr := grpc.NewServer()
	Logger.Debugf("[grpc] creating new grpc server")
	synse.RegisterInternalApiServer(svr, ps)
	Logger.Debugf("[grpc] registering handlers")

	// start gRPC the server
	Logger.Infof("[grpc] serving")
	if err := svr.Serve(lis); err != nil {
		Logger.Fatalf("Failed to serve: %v", err)
		return err
	}
	return nil
}

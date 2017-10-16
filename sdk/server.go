package sdk

import (
	//synse "github.com/vapor-ware/synse-server-grpc/go"
	synse "./synse"

	"fmt"
	"os"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// The PluginServer fulfills the interface needed by the GRPC server. It maps the
// user-defined PluginHandler to the interface for the GRPC server. While they are
// already similar, the PluginServer provides a layer of abstraction.
//
//
//   TODO: update example here
//	 a plugin should get an instance of a PluginServer via:
//		server := sdk.NewPlugin(name, handler)
//		server.Run()
//
type PluginServer struct {
	name string

	readingManager ReadingManager
	writingManager WritingManager

	pluginDevices  map[string]Device
	rwLoop         RWLoop
}

// private method to configure devices. this will always be called on
// plugin server creation via 'NewPlugin'. It will look in the default
// directory for configurations.
func (ps *PluginServer) configureDevices(deviceHandler DeviceHandler) {
	devices := DevicesFromConfig(CONFIG_DIR, deviceHandler)

	ps.pluginDevices = make(map[string]Device)
	for _, device := range devices {
		ps.pluginDevices[device.Uid()] = device
	}
	ps.rwLoop.devices = ps.pluginDevices
}


//
func (ps *PluginServer) getReadings(uid string) []Reading {

	// TODO - need some checking in here for if the UID doesn't exist.
	return ps.readingManager.GetReading(uid)
}



// GRPC READ HANDLER
func (ps *PluginServer) Read(in *synse.ReadRequest, stream synse.InternalApi_ReadServer) error {
	logger.Debug("[grpc] READ")

	uid := in.GetUid()
	if uid == "" {
		logger.Debug("No UID supplied.")
	}

	logger.Debugf("uid: %v\n", uid)

	readings := ps.getReadings(uid)

	for _, r := range readings {
		resp := &synse.ReadResponse{
			Timestamp: r.Timestamp,
			Value: r.Value,
		}
		if err := stream.Send(resp); err != nil {
			return err
		}
	}
	return nil
}

// GRPC WRITE HANDLER
func (ps *PluginServer) Write(ctx context.Context, in *synse.WriteRequest) (*synse.TransactionId, error) {
	logger.Debug("[grpc] WRITE")

	transaction := NewTransactionId()
	UpdateTransactionStatus(transaction.id, PENDING)

	ps.writingManager.channel <- WriteResource{transaction, in.Uid, in.Data}

	return &synse.TransactionId{
		Id: transaction.id,
	}, nil
}


// GRPC METAINFO HANDLER
func (ps *PluginServer) Metainfo(in *synse.MetainfoRequest, stream synse.InternalApi_MetainfoServer) error {
	logger.Debug("[grpc] METAINFO")

	for _, device := range ps.pluginDevices {
		if err := stream.Send(device.ToMetainfoResponse()); err != nil {
			return err
		}
	}
	return nil
}


// GRPC TRANSACTION CHECK HANDLER
func (ps *PluginServer) TransactionCheck(ctx context.Context, in *synse.TransactionId) (*synse.WriteResponse, error) {
	logger.Debug("[grpc] TRANSACTION CHECK")

	transaction := GetTransaction(in.Id)

	// FIXME - need to update GRPC so write response has a 'created' time and 'updated' time
	return &synse.WriteResponse{
		Timestamp: transaction.created,
		Status: transaction.status,
		State: transaction.state,
	}, nil
}



// Run the PluginServer.
// This will first start the read-write loop and then will configure
// and serve the GRPC server and listen for incoming requests. It will be
// configured to listen on a UNIX socket which has the same name as the
// plugin. This socket will be used by Synse to discover and communicate
// with the plugin.
func (ps *PluginServer) Run() error {

	logger.Info("[plugin server] running")

	// Start the read and write managers
	ps.readingManager.Start()
	ps.writingManager.Start()

	// start the RW loop
	ps.rwLoop.Run()

	// start the GRPC server
	socket := fmt.Sprintf("/synse/procs/%s.sock", ps.name)

	var _, err = os.Stat(socket)
	if err == nil {
		os.Remove(socket)
	}

	logger.Infof("[grpc] listening on socket %v", socket)
	lis, err := net.Listen("unix", socket)
	if err != nil {
		logger.Fatalf("Failed to listen: %v", err)
		return err
	}

	// create the GRPC server and register our plugin server to it
	svr := grpc.NewServer()
	logger.Debugf("[grpc] creating new grpc server")
	synse.RegisterInternalApiServer(svr, ps)
	logger.Debugf("[grpc] registering handlers")

	// start the server
	logger.Infof("[grpc] serving")
	if err := svr.Serve(lis); err != nil {
		logger.Fatalf("Failed to serve: %v", err)
		return err
	}

	return nil
}

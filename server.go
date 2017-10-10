package sdk

import (
	"./pb"

	"fmt"
	"os"
	"net"
	"log"

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

	pluginDevices []Device
	rwLoop RWLoop

}

// private method to configure devices. this will always be called on
// plugin server creation via 'NewPlugin'. It will look in the default
// directory for configurations.
func (ps *PluginServer) configureDevices(deviceHandler DeviceHandler) {
	ps.pluginDevices = DevicesFromConfig(CONFIG_DIR, deviceHandler)
	ps.rwLoop.devices = ps.pluginDevices
}


//
func (ps *PluginServer) getReadings(uid string) []Reading {

	// TODO - need some checking in here for if the UID doesn't exist.
	return ps.readingManager.GetReading(uid)
}



// GRPC READ HANDLER
func (ps *PluginServer) Read(in *pb.ReadRequest, stream pb.InternalApi_ReadServer) error {
	fmt.Printf("[grpc] READ\n")

	uid := in.GetUid()
	if uid == "" {
		fmt.Printf("ERROR: No UID supplied.\n")
	}

	fmt.Printf("uid: %v\n", uid)

	readings := ps.getReadings(uid)

	for _, r := range readings {
		resp := &pb.ReadResponse{
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
func (ps *PluginServer) Write(ctx context.Context, in *pb.WriteRequest) (*pb.TransactionId, error) {
	fmt.Printf("[grpc] WRITE\n")

	// TODO -- implement
	return nil, nil
}


// GRPC METAINFO HANDLER
func (ps *PluginServer) Metainfo(in *pb.MetainfoRequest, stream pb.InternalApi_MetainfoServer) error {
	fmt.Printf("[grpc] METAINFO\n")

	for _, device := range ps.pluginDevices {
		if err := stream.Send(device.ToMetainfoResponse()); err != nil {
			return err
		}
	}
	return nil
}


// GRPC TRANSACTION CHECK HANDLER
func (ps *PluginServer) TransactionCheck(ctx context.Context, in *pb.TransactionId) (*pb.WriteResponse, error) {
	fmt.Printf("[grpc] TRANSACTION CHECK\n")

	// TODO -- implement.
	return nil, nil
}



// Run the PluginServer.
// This will first start the read-write loop and then will configure
// and serve the GRPC server and listen for incoming requests. It will be
// configured to listen on a UNIX socket which has the same name as the
// plugin. This socket will be used by Synse to discover and communicate
// with the plugin.
func (ps *PluginServer) Run() error {

	fmt.Printf("[plugin server] running\n")

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

	fmt.Printf("[grpc] listening on socket %v\n", socket)
	lis, err := net.Listen("unix", socket)
	if err != nil {
		log.Fatalf("Failed to listen: %v\n", err)
		return err
	}

	// create the GRPC server and register our plugin server to it
	svr := grpc.NewServer()
	log.Printf("[grpc] creating new grpc server\n")
	pb.RegisterInternalApiServer(svr, ps)
	log.Printf("[grpc] registering handlers\n")

	// start the server
	log.Printf("[grpc] serving\n")
	if err := svr.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v\n", err)
		return err
	}

	return nil
}

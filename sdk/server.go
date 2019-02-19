package sdk

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/vapor-ware/synse-sdk/sdk/utils"

	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/health"

	log "github.com/Sirupsen/logrus"
	"github.com/vapor-ware/synse-server-grpc/go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	networkTypeTCP  = "tcp"
	networkTypeUnix = "unix"
)

var (
	// The directory where Unix sockets are placed for unix-based
	// gRPC communication. This is a var instead of const so that
	// it can be modified for testing.
	socketDir = "/tmp/synse"
)

// server implements the Synse Plugin gRPC server. It is used by the
// plugin to communicate via gRPC over tcp or unix socket to Synse server.
type server struct {
	conf *config.NetworkSettings
	grpc *grpc.Server

	initialized bool

	meta *PluginMetadata

	// Plugin components
	deviceManager *deviceManager
	stateManager  *StateManager
	scheduler     *Scheduler
}

// newServer creates a new instance of a server. This is used by the Plugin
// constructor to create a Plugin's server instance.
func newServer(conf *config.NetworkSettings, dm *deviceManager, sm *StateManager, sched *Scheduler, meta *PluginMetadata) *server {
	return &server{
		conf:          conf,
		deviceManager: dm,
		stateManager:  sm,
		scheduler:     sched,
		meta:          meta,
	}
}

func (server *server) init() error {
	if server.conf == nil {
		// fixme
		return fmt.Errorf("no config")
	}

	log.Debug("[server] setting up server")

	// Depending on the communication protocol, there may be some setup work.
	switch t := server.conf.Type; t {
	case networkTypeUnix:
		// If the path containing the sockets does not exist, create it.
		_, err := os.Stat(socketDir)
		if err != nil {
			if os.IsNotExist(err) {
				if err = os.MkdirAll(socketDir, os.ModePerm); err != nil {
					return err
				}
			} else {
				return err
			}
		}
		// If the socket path does exist, try removing the socket if it is
		// there (left over from a previous run).
		if err = os.Remove(server.address()); !os.IsNotExist(err) {
			return err
		}
		break

	case networkTypeTCP:
		// No setup required.
		break

	default:
		return fmt.Errorf("unsupported network type: %s", t)
	}

	// Get any options for the gRPC server.
	var opts []grpc.ServerOption
	if err := addTLSOptions(&opts, server.conf.TLS); err != nil {
		return err
	}

	// Create the gRPC server instance, passing in any server options.
	server.grpc = grpc.NewServer(opts...)

	server.initialized = true
	return nil
}

// start runs the gRPC server.
func (server *server) start() error {
	if !server.initialized {
		return fmt.Errorf("server is not initialized, can not run")
	}
	if server.grpc == nil {
		return fmt.Errorf("gRPC server not initialized, can not run")
	}

	// Create the listener over the configured protocol and address.
	listener, err := net.Listen(server.conf.Type, server.address())
	if err != nil {
		return err
	}

	// Register the server as an implementation of the gRPC server.
	synse.RegisterV3PluginServer(server.grpc, server)

	log.WithFields(log.Fields{
		"mode": server.conf.Type,
		"addr": server.conf.Address,
	}).Info("[server] serving...")
	return server.grpc.Serve(listener)
}

// stop stops the gRPC server from serving and immediately terminates all open
// connections and listeners.
func (server *server) stop() {
	log.Info("[server] stopping server")
	if server.grpc != nil {
		server.grpc.Stop()
	}
}

// teardown the server post-run. This is called as a PluginAction on plugin
// termination.
func (server *server) teardown() error {
	log.Debug("[server] tearing down server")

	// Stop the server.
	server.stop()

	// Perform any other cleanup.
	switch t := server.conf.Type; t {
	case networkTypeUnix:
		// Remove the unix socket that was being used.
		if err := os.Remove(server.address()); !os.IsNotExist(err) {
			return err
		}
		return nil

	case networkTypeTCP:
		// No cleanup required.
		return nil

	default:
		return fmt.Errorf("unsupported network type: %s", t)
	}
}

// address gets the address for the configured server. The configured address
// may need additional formatting depending on the networking mode, so this
// should be the preferred means of getting the address.
func (server *server) address() string {
	switch t := server.conf.Type; t {
	case networkTypeUnix:
		address := server.conf.Address
		if !strings.HasPrefix(address, socketDir) {
			address = filepath.Join(socketDir, address)
		}
		return address

	case networkTypeTCP:
		return server.conf.Address

	default:
		return ""
	}
}

// registerActions registers pre-run (setup) and post-run (teardown) actions
// for the server.
func (server *server) registerActions(plugin *Plugin) {
	// Register post-run actions.
	plugin.RegisterPostRunActions(
		&PluginAction{
			Name:   "Cleanup gRPC Server",
			Action: func(plugin *Plugin) error { return server.teardown() },
		},
	)
}

// addTLSOptions updates the options slice with any TLS/SSL options for the gRPC server,
// as configured via the plugin network config.
func addTLSOptions(options *[]grpc.ServerOption, settings *config.TLSNetworkSettings) error {
	// If there is no TLS config, there are no options to add here.
	if settings == nil {
		return nil
	}

	tlsLog := log.WithFields(log.Fields{
		"cert":       settings.Cert,
		"key":        settings.Key,
		"ca":         settings.CACerts,
		"skipVerify": settings.SkipVerify,
	})
	tlsLog.Info("[server] configuring for tls/ssl transport")

	cert, err := tls.LoadX509KeyPair(settings.Cert, settings.Key)
	if err != nil {
		tlsLog.WithField("error", err).Error("[server] failed to load TLS key pair")
		return err
	}

	var certPool *x509.CertPool

	// If custom certificate authority certs are specified, use those, otherwise
	// use the system-wide root certs from the OS.
	if len(settings.CACerts) > 0 {
		tlsLog.Info("[server] loading custom CA certs")
		certPool, err = loadCACerts(settings.CACerts)
		if err != nil {
			tlsLog.WithField("error", err).Error("[server] failed to load custom CA certs")
			return err
		}
	} else {
		tlsLog.Info("[server] loading default CA certs from OS")
		certPool, err = x509.SystemCertPool()
		if err != nil {
			tlsLog.WithField("error", err).Error("[server] failed to load default CA certs from OS")
			return err
		}
	}

	clientAuth := tls.RequireAndVerifyClientCert
	if settings.SkipVerify {
		clientAuth = tls.NoClientCert
	}

	creds := credentials.NewTLS(&tls.Config{
		ClientAuth:               clientAuth,
		ClientCAs:                certPool,
		Certificates:             []tls.Certificate{cert},
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		// https://www.acunetix.com/blog/articles/tls-ssl-cipher-hardening/
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
		},
	})
	*options = append(*options, grpc.Creds(creds))

	return nil
}

// loadCACerts loads the certs from the provided certificate authority/authorities.
func loadCACerts(certs []string) (*x509.CertPool, error) {
	certPool := x509.NewCertPool()
	for _, c := range certs {
		ca, err := ioutil.ReadFile(c) // #nosec
		if err != nil {
			log.WithField("error", err).Error("[server] failed to read CA file")
			return nil, err
		}

		if ok := certPool.AppendCertsFromPEM(ca); !ok {
			log.WithField("error", err).Error("[server] failed to append CA cert from PEM")
			return nil, fmt.Errorf("failed to append CA cert from PEM")
		}
	}
	return certPool, nil
}

// --------------------------------------------------------
//
// gRPC API Routes
//
// --------------------------------------------------------

// Test checks whether the plugin is reachable and ready.
//
// It is the handler for the Synse gRPC V3Plugin service's `Test` RPC method.
func (server *server) Test(ctx context.Context, request *synse.Empty) (*synse.V3TestStatus, error) {
	log.Debug("[grpc] TEST request")

	return &synse.V3TestStatus{Ok: true}, nil
}

// Version gets the version information for the plugin.
//
// It is the handler for the Synse gRPC V3Plugin service's `Version` RPC method.
func (server *server) Version(ctx context.Context, request *synse.Empty) (*synse.V3Version, error) {
	log.Debug("[grpc] VERSION request")

	return version.encode(), nil
}

// Health gets the overall health status of the plugin.
//
// It is the handler for the Synse gRPC V3Plugin service's `Health` RPC method.
func (server *server) Health(ctx context.Context, request *synse.Empty) (*synse.V3Health, error) {
	log.Debug("[grpc] HEALTH request")

	// Get the health statuses from the health catalog.
	statuses := health.GetStatus()

	// Determine whether we are ok or failing. If all checks are okay, the plugin is
	// healthy. If there are any checks that are failing, the plugin is considered
	// unhealthy. Synse plugins do not support the notion of partially degraded.
	healthStatus := synse.HealthStatus_OK

	var healthChecks []*synse.V3HealthCheck
	for _, status := range statuses {
		if !status.Ok {
			healthStatus = synse.HealthStatus_FAILING
		}
		healthChecks = append(healthChecks, status.Encode())
	}

	return &synse.V3Health{
		Timestamp: utils.GetCurrentTime(),
		Status:    healthStatus,
		Checks:    healthChecks,
	}, nil
}

// Devices gets all of the devices which a plugin manages.
//
// It is the handler for the Synse gRPC V3Plugin service's `Devices` RPC method.
func (server *server) Devices(request *synse.V3DeviceSelector, stream synse.V3Plugin_DevicesServer) error {
	log.WithFields(log.Fields{
		"tags": request.Tags,
		"id":   request.Id,
	}).Debug("[grpc] DEVICES request")

	// Encode and stream the devices back to the client.
	for _, device := range server.deviceManager.GetDevices(DeviceSelectorToTags(request)...) {
		if err := stream.Send(device.encode()); err != nil {
			return err
		}
	}
	return nil
}

// Metadata gets the meta-information for a plugin.
//
// It is the handler for the Synse gRPC V3Plugin service's `Metadata` RPC method.
func (server *server) Metadata(ctx context.Context, request *synse.Empty) (*synse.V3Metadata, error) {
	log.Debug("[grpc] METADATA request")

	return server.meta.encode(), nil
}

// Read gets readings for the specified plugin device(s).
//
// It is the handler for the Synse gRPC V3Plugin service's `Read` RPC method.
func (server *server) Read(request *synse.V3ReadRequest, stream synse.V3Plugin_ReadServer) error {
	log.WithFields(log.Fields{
		"tags":   request.Selector.Tags,
		"id":     request.Selector.Id,
		"system": request.SystemOfMeasure,
	}).Debug("[grpc] READ request")

	devices := server.deviceManager.GetDevices(DeviceSelectorToTags(request.Selector)...)
	for _, device := range devices {
		readings := server.stateManager.GetReadingsForDevice(device.id)

		// Encode and stream the readings back to the client.
		for _, reading := range readings {
			if err := stream.Send(reading.encode()); err != nil {
				return err
			}
		}
	}
	return nil
}

// ReadCache gets the cached readings from the plugin. If the plugin is not configured
// to cache its readings, this will return a dump of the entire current readings state.
//
// It is the handler for the Synse gRPC V3Plugin service's `ReadCache` RPC method.
func (server *server) ReadCache(request *synse.V3Bounds, stream synse.V3Plugin_ReadCacheServer) error {
	log.WithFields(log.Fields{
		"start": request.Start,
		"end":   request.End,
	}).Debug("[grpc] READCACHE request")

	// Create a channel that will be used to collect the cached readings.
	readings := make(chan *ReadContext, 128)

	go server.stateManager.GetCachedReadings(request.Start, request.End, readings)

	// Encode and stream the readings back to the client.
	for r := range readings {
		for _, data := range r.Reading {
			if err := stream.Send(data.encode()); err != nil {
				return err
			}
		}
	}
	return nil
}

// WriteAsync writes data to the specified plugin device. A transaction ID is returned
// so the status of the write can be checked asynchronously.
//
// It is the handler for the Synse gRPC V3Plugin service's `WriteAsync` RPC method.
func (server *server) WriteAsync(ctx context.Context, request *synse.V3WritePayload) (*synse.V3WriteTransaction, error) {
	log.WithFields(log.Fields{
		"data": request.Data,
		"id":   request.Selector.Id,
	}).Debug("[grpc] WRITE ASYNC request")

	// TODO (etd): update this once various other transaction updates are completed.

	return &synse.V3WriteTransaction{}, nil

	//log.WithField("request", request).Debug("[grpc] write rpc request")
	//transactions, err := DataManager.Write(request)
	//if err != nil {
	//	return nil, err
	//}
	//return &synse.Transactions{
	//	Transactions: transactions,
	//}, nil
}

// WriteSync writes data to the specified plugin device. The request blocks until the
// write resolves so no asynchronous status checking is needed for the write action.
//
// It is the handler for the Synse gRPC V3Plugin service's `WriteSync` RPC method.
func (server *server) WriteSync(request *synse.V3WritePayload, stream synse.V3Plugin_WriteSyncServer) error {
	log.WithFields(log.Fields{
		"data": request.Data,
		"id":   request.Selector.Id,
	}).Debug("[grpc] WRITE SYNC request")

	// TODO (etd): update this once various other transaction updates are completed.

	return nil
}

// Transaction gets the status of an asynchronous write via a transaction ID that
// associated with that action on write.
//
// It is the handler for the Synse gRPC V3Plugin service's `Transaction` RPC method.
func (server *server) Transaction(request *synse.V3TransactionSelector, stream synse.V3Plugin_TransactionServer) error {
	log.WithFields(log.Fields{
		"id": request.Id,
	}).Debug("[grpc] TRANSACTION request")

	// If there is no ID specified with the incoming request, return all of the cached
	// transaction.
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

	// Otherwise, return only the transaction with the specified ID.
	t := getTransaction(request.Id)
	if t == nil {
		return errors.NotFoundErr("transaction not found: %v", request.Id)
	}
	return stream.Send(t.encode())
}

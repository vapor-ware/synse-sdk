package sdk

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/vapor-ware/synse-sdk/sdk/health"
	"io/ioutil"
	"net"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/vapor-ware/synse-server-grpc/go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// server implements the Synse Plugin gRPC server. It is used by the
// plugin to communicate via gRPC over tcp or unix socket to Synse server.
type server struct {
	network string
	address string
	grpc    *grpc.Server
}

// newServer creates a new instance of a server. This should be used
// by the plugin to create its server instance.
func newServer(network, address string) *server {
	return &server{
		network: network,
		address: address,
	}
}

// setup runs any steps needed to set up the environment for the server to run.
// In particular, this makes sure that the proper directories exist if the server
// is running in "unix" mode.
func (server *server) setup() error {
	// Set the server cleanup function as a post-run action for the plugin.
	ctx.postRunActions = append(ctx.postRunActions, func(plugin *Plugin) error {
		return server.cleanup()
	})
	log.WithField("mode", server.network).Debug("[grpc] setting up server")

	switch server.network {
	case networkTypeUnix:
		if !strings.HasPrefix(server.address, sockPath) {
			server.address = fmt.Sprintf("%s/%s", sockPath, server.address)
		}
		// If the path containing the sockets does not exist, create it.
		_, err := os.Stat(sockPath)
		if err != nil {
			if os.IsNotExist(err) {
				if err = os.MkdirAll(sockPath, os.ModePerm); err != nil {
					return err
				}
			} else {
				return err
			}
		}
		// If the socket path does exist, try removing the socket if it is
		// there and left over from a previous run.
		if err = os.Remove(server.address); !os.IsNotExist(err) {
			return err
		}
		return nil

	case networkTypeTCP:
		// There is nothing for us to do in this case.
		return nil

	default:
		return fmt.Errorf("unsupported network type: %s", server.network)
	}
}

// cleanup cleans up the server. The action it takes will depend on the mode it is
// running in. If running in 'unix' mode, it will remove the socket.
func (server *server) cleanup() error {
	log.Info("[grpc] cleaning up server")
	switch server.network {
	case networkTypeUnix:
		if err := os.Remove(server.address); !os.IsNotExist(err) {
			return err
		}
		return nil

	case networkTypeTCP:
		// There is nothing for us to do in this case.
		return nil

	default:
		return fmt.Errorf("unsupported network type: %s", server.network)
	}
}

// Serve sets up the gRPC server and runs it.
func (server *server) Serve() error {
	e := server.setup()
	if e != nil {
		return e
	}

	// Options for the gRPC server to be passed in to the constructor.
	var opts []grpc.ServerOption
	err := setCredsOption(&opts)
	if err != nil {
		return err
	}

	// Create the listener over the configured network type and address.
	lis, err := net.Listen(server.network, server.address)
	if err != nil {
		return err
	}

	// Create the grpc server instance, passing in any server options.
	svr := grpc.NewServer(opts...)
	synse.RegisterV3PluginServer(svr, server)
	server.grpc = svr

	log.Infof("[grpc] listening on %s:%s", server.network, server.address)
	return svr.Serve(lis)
}

// setCredsOptions will add a credentials option to the server options slice, if the
// plugin is configured to use TLS/SSL with gRPC.
func setCredsOption(options *[]grpc.ServerOption) error {
	// If the plugin is configured to use TLS/SSL for communicating with Synse Server,
	// load in the specified certs, make sure everything is happy, and add a gRPC Creds
	// option to the slice of server options.
	if Config.Plugin.Network.TLS != nil {
		tlsConfig := Config.Plugin.Network.TLS
		log.WithFields(log.Fields{
			"cert": tlsConfig.Cert,
			"key":  tlsConfig.Key,
			"ca":   tlsConfig.CACerts,
		}).Debugf("[server] configuring grpc server for tls/ssl transport")

		cert, err := tls.LoadX509KeyPair(tlsConfig.Cert, tlsConfig.Key)
		if err != nil {
			log.Errorf("[server] failed to load TLS key pair: %v", err)
			return err
		}

		var certPool *x509.CertPool

		// If custom certificate authority certs are specified, use those, otherwise
		// use the system-wide root certs from the OS.
		if len(tlsConfig.CACerts) > 0 {
			log.Debugf("[server] loading custom CA certs: %v", tlsConfig.CACerts)
			certPool, err = loadCACerts(tlsConfig.CACerts)
			if err != nil {
				log.Errorf("[server] failed to load custom CA certs: %v", err)
				return err
			}
		} else {
			log.Debug("[server] loading default CA certs from OS")
			certPool, err = x509.SystemCertPool()
			if err != nil {
				log.Errorf("[server] failed to load default OS CA certs: %v", err)
				return err
			}
		}

		clientAuth := tls.RequireAndVerifyClientCert
		if tlsConfig.SkipVerify {
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
	} else {
		log.Debug("[server] configuring grpc server for insecure transport")
	}
	return nil
}

// loadCACerts loads the certs from the provided certificate authority/authorities.
func loadCACerts(cacerts []string) (*x509.CertPool, error) {
	certPool := x509.NewCertPool()
	for _, c := range cacerts {
		ca, err := ioutil.ReadFile(c) // #nosec
		if err != nil {
			log.Errorf("[server] failed to read CA file: %v", err)
			return nil, err
		}

		if ok := certPool.AppendCertsFromPEM(ca); !ok {
			log.Errorf("[server] failed to append CA cert: %v", c)
			return nil, fmt.Errorf("failed to append ca cert")
		}
	}
	return certPool, nil
}

// Stop stops the GRPC server from serving and immediately terminates all open
// connections and listeners.
func (server *server) Stop() {
	log.Info("[grpc] stopping server")
	if server.grpc != nil {
		server.grpc.Stop()
	}
}

//
// gRPC API Routes
//

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

	return version.Encode(), nil
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
		Timestamp: GetCurrentTime(),
		Status: healthStatus,
		Checks: healthChecks,
	}, nil
}

// Devices gets all of the devices which a plugin manages.
//
// It is the handler for the Synse gRPC V3Plugin service's `Devices` RPC method.
func (server *server) Devices(request *synse.V3DeviceSelector, stream synse.V3Plugin_DevicesServer) error {
	log.WithFields(log.Fields{
		"tags": request.Tags,
		"id": request.Id,
	}).Debug("[grpc] DEVICES request")

	// Encode and stream the devices back to the client.
	for _, device := range ctx.devices {
		// TODO (etd): filter upon the tags/id. first, we need to update how devices are
		//  cached/routed to. for now, returning all devices.
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

	return metainfo.Encode(), nil
}

// Read gets readings for the specified plugin device(s).
//
// It is the handler for the Synse gRPC V3Plugin service's `Read` RPC method.
func (server *server) Read(request *synse.V3ReadRequest, stream synse.V3Plugin_ReadServer) error {
	log.WithFields(log.Fields{
		"tags": request.Selector.Tags,
		"id": request.Selector.Id,
		"system": request.SystemOfMeasure,
	}).Debug("[grpc] READ request")

	// Get device readings from the DataManager.
	responses, err := DataManager.Read(request)
	if err != nil {
		return err
	}

	// Encode and stream the readings back to the client.
	for _, response := range responses {
		if err := stream.Send(response); err != nil {
			return err
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
		"end": request.End,
	}).Debug("[grpc] READCACHE request")

	// Create a channel that will be used to collect the cached readings.
	readings := make(chan *ReadContext, 128)
	go getReadingsFromCache(request.Start, request.End, readings)

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
		"id": request.Selector.Id,
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
		"id": request.Selector.Id,
	}).Debug("[grpc] WRITE SYNC request")

	// TODO (etd): update this once various other transaction updates are completed.

	return nil
}

// Transaction is the handler for the Synse GRPC Plugin service's `Transaction` RPC method.
func (server *server) Transaction(request *synse.V3TransactionSelector, stream synse.V3Plugin_TransactionServer) error {
	return nil

	//log.WithField("request", request).Debug("[grpc] transaction rpc request")
	//
	//// If there is no ID with the incoming request, return all cached transactions.
	//if request.Id == "" {
	//	for _, item := range transactionCache.Items() {
	//		t, ok := item.Object.(*transaction)
	//		if ok {
	//			if err := stream.Send(t.encode()); err != nil {
	//				return err
	//			}
	//		}
	//	}
	//	return nil
	//}
	//
	//// Otherwise, return the transaction with the specified ID.
	//transaction := getTransaction(request.Id)
	//if transaction == nil {
	//	return errors.NotFoundErr("transaction not found: %v", request.Id)
	//}
	//return stream.Send(transaction.encode())
}

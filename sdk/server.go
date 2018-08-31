package sdk

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/health"
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

		var CAs *x509.CertPool

		// If custom certificate authority certs are specified, use those, otherwise
		// use the system-wide root certs from the OS.
		if len(tlsConfig.CACerts) > 0 {
			log.Debugf("[server] loading custom CA certs: %v", tlsConfig.CACerts)
			CAs, err = loadCACerts(tlsConfig.CACerts)
			if err != nil {
				log.Errorf("[server] failed to load custom CA certs: %v", err)
				return err
			}
		} else {
			log.Debug("[server] loading default CA certs from OS")
			CAs, err = x509.SystemCertPool()
			if err != nil {
				log.Errorf("[server] failed to load default OS CA certs: %v", err)
				return err
			}
		}

		creds := credentials.NewTLS(&tls.Config{
			Certificates:             []tls.Certificate{cert},
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
			ClientAuth: tls.RequireAndVerifyClientCert,
			ClientCAs:  CAs,
		})

		opts = append(opts, grpc.Creds(creds))
	} else {
		log.Debug("[server] configuring grpc server for insecure transport")
	}

	// Create the listener over the configured network type and address.
	lis, err := net.Listen(server.network, server.address)
	if err != nil {
		return err
	}

	// Create the grpc server instance, passing in any server options.
	svr := grpc.NewServer(opts...)
	synse.RegisterPluginServer(svr, server)
	server.grpc = svr

	log.Infof("[grpc] listening on %s:%s", server.network, server.address)
	return svr.Serve(lis)
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

// Test is the handler for the Synse GRPC Plugin service's `Test` RPC method.
func (server *server) Test(ctx context.Context, request *synse.Empty) (*synse.Status, error) {
	log.WithField("request", request).Debug("[grpc] test rpc request")
	return &synse.Status{Ok: true}, nil
}

// Version is the handler for the Synse GRPC Plugin service's `Version` RPC method.
func (server *server) Version(ctx context.Context, request *synse.Empty) (*synse.VersionInfo, error) {
	log.WithField("request", request).Debug("[grpc] version rpc request")
	return version.Encode(), nil
}

// Health is the handler for the Synse GRPC Plugin service's `Health` RPC method.
func (server *server) Health(ctx context.Context, request *synse.Empty) (*synse.PluginHealth, error) {
	log.WithField("request", request).Debug("[grpc] health rpc request")
	statuses := health.GetStatus()

	// First, we need to determine the overall health of the plugin.
	// If all statuses are good, we are ok. If some are bad, we are partially
	// degraded. If all are bad, we are failing.
	// TODO: do we want partially degraded, or should we just consider it failing
	total := len(statuses)
	ok := 0
	failing := 0

	var healthChecks []*synse.HealthCheck
	for _, status := range statuses {
		if status.Ok {
			ok++
		} else {
			failing++
		}
		healthChecks = append(healthChecks, status.Encode())
	}

	var pluginStatus synse.PluginHealth_Status
	if total == ok {
		pluginStatus = synse.PluginHealth_OK
	} else if total == failing {
		pluginStatus = synse.PluginHealth_FAILING
	} else {
		pluginStatus = synse.PluginHealth_PARTIALLY_DEGRADED
	}

	return &synse.PluginHealth{
		Timestamp: GetCurrentTime(),
		Status:    pluginStatus,
		Checks:    healthChecks,
	}, nil
}

// Capabilities is the handler for the Synse GRPC Plugin service's `Capabilities` RPC method.
func (server *server) Capabilities(request *synse.Empty, stream synse.Plugin_CapabilitiesServer) error {
	log.WithField("request", request).Debug("[grpc] capabilities rpc request")
	capabilitiesMap := map[string]*synse.DeviceCapability{}

	for _, device := range ctx.devices {
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
func (server *server) Devices(request *synse.DeviceFilter, stream synse.Plugin_DevicesServer) error {
	log.WithField("request", request).Debug("[grpc] devices rpc request")
	var (
		rack   = request.GetRack()
		board  = request.GetBoard()
		device = request.GetDevice()
	)
	if device != "" {
		return fmt.Errorf("devices rpc method does not support filtering on device")
	}
	if rack == "" && board != "" {
		return fmt.Errorf("filter specifies board with no rack - must specifiy rack as well")
	}

	for _, device := range ctx.devices {
		if rack != "" {
			if device.Location.Rack != rack {
				continue
			}
			if board != "" {
				if device.Location.Board != board {
					continue
				}
			}
		}
		if err := stream.Send(device.encode()); err != nil {
			return err
		}
	}
	return nil
}

// Metainfo is the handler for the Synse GRPC Plugin service's `Metainfo` RPC method.
func (server *server) Metainfo(ctx context.Context, request *synse.Empty) (*synse.Metadata, error) {
	log.WithField("request", request).Debug("[grpc] metainfo rpc request")
	return &synse.Metadata{
		Name:        metainfo.Name,
		Maintainer:  metainfo.Maintainer,
		Tag:         metainfo.Tag,
		Description: metainfo.Description,
		Vcs:         metainfo.VCS,
		Version:     version.Encode(),
	}, nil
}

// Read is the handler for the Synse GRPC Plugin service's `Read` RPC method.
func (server *server) Read(request *synse.DeviceFilter, stream synse.Plugin_ReadServer) error {
	log.WithField("request", request).Debug("[grpc] read rpc request")
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
func (server *server) Write(ctx context.Context, request *synse.WriteInfo) (*synse.Transactions, error) {
	log.WithField("request", request).Debug("[grpc] write rpc request")
	transactions, err := DataManager.Write(request)
	if err != nil {
		return nil, err
	}
	return &synse.Transactions{
		Transactions: transactions,
	}, nil
}

// Transaction is the handler for the Synse GRPC Plugin service's `Transaction` RPC method.
func (server *server) Transaction(request *synse.TransactionFilter, stream synse.Plugin_TransactionServer) error {
	log.WithField("request", request).Debug("[grpc] transaction rpc request")

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
		return nil
	}

	// Otherwise, return the transaction with the specified ID.
	transaction := getTransaction(request.Id)
	if transaction == nil {
		return errors.NotFoundErr("transaction not found: %v", request.Id)
	}
	return stream.Send(transaction.encode())
}

// Synse SDK
// Copyright (c) 2019 Vapor IO
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package sdk

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk/config"
	sdkError "github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/health"
	synse "github.com/vapor-ware/synse-server-grpc/go"
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

// Server error definitions.
var (
	ErrServerNeedsConfig    = errors.New("server requires configuration to initialize")
	ErrServerNotInitialized = errors.New("unable to run: server not initialized")

	ErrSelectorRequiresID  = sdkError.InvalidArgumentErr("selector must specify device id")
	ErrNoDeviceForSelector = sdkError.NotFoundErr("no device found for specified selector")
	ErrTransactionNotFound = sdkError.NotFoundErr("transaction not found")
)

// server implements the Synse Plugin gRPC server. It is used by the
// plugin to communicate via gRPC over tcp or unix socket to Synse server.
type server struct {
	conf        *config.NetworkSettings
	grpc        *grpc.Server
	meta        *PluginMetadata
	id          *pluginID
	initialized bool

	// Plugin components
	deviceManager *deviceManager
	stateManager  *stateManager
	scheduler     *scheduler
	healthManager *health.Manager
}

// newServer creates a new instance of a server. This is used by the Plugin
// constructor to create a Plugin's server instance.
func newServer(plugin *Plugin) *server {
	return &server{
		id:            plugin.id,
		conf:          plugin.config.Network,
		meta:          plugin.info,
		scheduler:     plugin.scheduler,
		deviceManager: plugin.device,
		stateManager:  plugin.state,
		healthManager: plugin.health,
	}
}

func (server *server) init() error {
	if server.conf == nil {
		return ErrServerNeedsConfig
	}

	log.Debug("[server] initializing")

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
		err = os.Remove(server.address())
		if err != nil && !os.IsNotExist(err) {
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
	log.Info("[server] starting")

	if !server.initialized || server.grpc == nil {
		return ErrServerNotInitialized
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
	}).Info("[server] serving")
	return server.grpc.Serve(listener)
}

// stop stops the gRPC server from serving and immediately terminates all open
// connections and listeners.
func (server *server) stop() {
	log.Info("[server] stopping")
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
	if settings == nil || settings == (&config.TLSNetworkSettings{}) {
		log.Info("[server] tls/ssl not configured, using insecure transport")
		return nil
	}

	// If there is no key and cert, the other options don't matter,
	// so we have nothing to do.
	if settings.Key == "" && settings.Cert == "" {
		log.Info("[server] tls/ssl not configured, using insecure transport")
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
	log.WithFields(log.Fields{
		"route": "TEST",
	}).Info("[grpc] processing request")

	return &synse.V3TestStatus{Ok: true}, nil
}

// Version gets the version information for the plugin.
//
// It is the handler for the Synse gRPC V3Plugin service's `Version` RPC method.
func (server *server) Version(ctx context.Context, request *synse.Empty) (*synse.V3Version, error) {
	log.WithFields(log.Fields{
		"route": "VERSION",
	}).Info("[grpc] processing request")

	return version.encode(), nil
}

// Health gets the overall health status of the plugin.
//
// It is the handler for the Synse gRPC V3Plugin service's `Health` RPC method.
func (server *server) Health(ctx context.Context, request *synse.Empty) (*synse.V3Health, error) {
	log.WithFields(log.Fields{
		"route": "HEALTH",
	}).Info("[grpc] processing request")

	status := server.healthManager.Status()
	return status.Encode(), nil
}

// Devices gets all of the devices which a plugin manages.
//
// It is the handler for the Synse gRPC V3Plugin service's `Devices` RPC method.
func (server *server) Devices(request *synse.V3DeviceSelector, stream synse.V3Plugin_DevicesServer) error {
	rlog := log.WithFields(log.Fields{
		"tags":  request.Tags,
		"id":    request.Id,
		"route": "DEVICES",
	})
	rlog.Info("[grpc] processing request")

	devices, err := server.deviceManager.GetDevices(request)
	if err != nil {
		return err
	}

	//var devices []*Device
	//
	//// If there is no info specified for the selector, assume all devices in the system namespace.
	//// Otherwise, get the set of devices from the specified selector.
	//// TODO (etd): post v3.0: getting all devices in the system namespace means all devices. if/when
	////   we use the namespaces to limit access to devices, this will need to change, as we do not want
	////   to expose all devices to everyone. We are not doing that currently, so it is not an issue
	////   for the initial v3 release.
	//if request.Id == "" && len(request.Tags) == 0 {
	//	//devices = server.deviceManager.GetDevicesByTagNamespace(TagNamespaceDefault)
	//	devices = server.deviceManager.GetDevicesByTagNamespace(TagNamespaceSystem)
	//} else {
	//	devices = server.deviceManager.GetDevicesForTags(DeviceSelectorToTags(request)...)
	//}
	rlog.WithField("devices", len(devices)).Debug("[grpc] got devices")

	// Encode and stream the devices back to the client.
	for _, device := range devices {
		d := device.encode()

		// Set the plugin id here. This is done prior to sending back rather than
		// keeping the plugin id in the device model due to the scoping of the plugin.
		d.Plugin = server.id.uuid.String()

		// Set the device outputs here. This is determined by the device readings.
		var outputs []*synse.V3DeviceOutput
		for _, o := range server.stateManager.GetOutputsForDevice(d.Id) {
			outputs = append(outputs, o.Encode())
		}
		d.Outputs = outputs

		if err := stream.Send(d); err != nil {
			return err
		}
	}
	return nil
}

// Metadata gets the meta-information for a plugin.
//
// It is the handler for the Synse gRPC V3Plugin service's `Metadata` RPC method.
func (server *server) Metadata(ctx context.Context, request *synse.Empty) (*synse.V3Metadata, error) {
	log.WithFields(log.Fields{
		"route": "METADATA",
	}).Info("[grpc] processing request")
	m := server.meta.encode()
	m.Id = server.id.uuid.String()
	return m, nil
}

// Read gets readings for the specified plugin device(s).
//
// It is the handler for the Synse gRPC V3Plugin service's `Read` RPC method.
func (server *server) Read(request *synse.V3ReadRequest, stream synse.V3Plugin_ReadServer) error {
	rlog := log.WithFields(log.Fields{
		"selector": request.Selector,
		"route":    "READ",
	})
	rlog.Info("[grpc] processing request")

	devices, err := server.deviceManager.GetDevices(request.Selector)
	if err != nil {
		return err
	}

	//var devices []*Device
	//
	//// If there is no info specified for the selector, assume all devices in the system namespace.
	//// Otherwise, get the set of devices from the specified selector.
	//// TODO (etd): post v3.0: getting all devices in the system namespace means all devices. if/when
	////   we use the namespaces to limit access to devices, this will need to change, as we do not want
	////   to expose all devices to everyone. We are not doing that currently, so it is not an issue
	////   for the initial v3 release.
	//if request.Selector == nil || (request.Selector.Id == "" && len(request.Selector.Tags) == 0) {
	//	devices = server.deviceManager.GetDevicesByTagNamespace(TagNamespaceSystem)
	//} else {
	//	devices = server.deviceManager.GetDevicesForTags(DeviceSelectorToTags(request.Selector)...)
	//}

	for _, device := range devices {
		rlog.WithField("device", device.id).Debug("[grpc] getting reading(s) for device")
		readings := server.stateManager.GetReadingsForDevice(device.id)

		// Encode and stream the readings back to the client.
		for _, reading := range readings {
			r := reading.Encode()
			r.Id = device.id
			r.DeviceType = device.Type

			if err := stream.Send(r); err != nil {
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
		"route": "READCACHE",
	}).Info("[grpc] processing request")

	// Create a channel that will be used to collect the cached readings.
	readings := make(chan *ReadContext, 128)

	go server.stateManager.GetCachedReadings(request.Start, request.End, readings)

	// Encode and stream the readings back to the client.
	for r := range readings {
		device := server.deviceManager.GetDevice(r.Device)
		for _, data := range r.Reading {
			reading := data.Encode()
			if device != nil {
				reading.Id = device.id
				reading.DeviceType = device.Type
			}
			if err := stream.Send(reading); err != nil {
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
func (server *server) WriteAsync(request *synse.V3WritePayload, stream synse.V3Plugin_WriteAsyncServer) error {
	log.WithFields(log.Fields{
		"data":  request.Data,
		"id":    request.Selector.Id,
		"route": "WRITE ASYNC",
	}).Info("[grpc] processing request")

	if request.Selector.Id == "" {
		return ErrSelectorRequiresID
	}

	devices, err := server.deviceManager.GetDevices(request.Selector)
	if err != nil {
		return err
	}
	if len(devices) != 1 {
		return ErrNoDeviceForSelector
	}

	transactions, err := server.scheduler.Write(devices[0], request.Data)
	if err != nil {
		return err
	}

	for _, txn := range transactions {
		if err := stream.Send(txn); err != nil {
			return err
		}
	}
	return nil
}

// WriteSync writes data to the specified plugin device. The request blocks until the
// write resolves so no asynchronous status checking is needed for the write action.
//
// It is the handler for the Synse gRPC V3Plugin service's `WriteSync` RPC method.
func (server *server) WriteSync(request *synse.V3WritePayload, stream synse.V3Plugin_WriteSyncServer) error {
	log.WithFields(log.Fields{
		"data":  request.Data,
		"id":    request.Selector.Id,
		"route": "WRITE SYNC",
	}).Info("[grpc] processing request")

	if request.Selector.Id == "" {
		return ErrSelectorRequiresID
	}

	devices, err := server.deviceManager.GetDevices(request.Selector)
	if err != nil {
		return err
	}
	if len(devices) != 1 {
		return ErrNoDeviceForSelector
	}

	transactions, err := server.scheduler.WriteAndWait(devices[0], request.Data)
	if err != nil {
		return err
	}

	for _, txn := range transactions {
		if err := stream.Send(txn); err != nil {
			return err
		}
	}
	return nil
}

// Transaction gets the status of an asynchronous write via a transaction ID that
// associated with that action on write.
//
// It is the handler for the Synse gRPC V3Plugin service's `Transaction` RPC method.
//ctx context.Context, request *synse.Empty) (*synse.V3Metadata, error) {
func (server *server) Transaction(ctx context.Context, request *synse.V3TransactionSelector) (*synse.V3TransactionStatus, error) {
	rlog := log.WithFields(log.Fields{
		"id":    request.Id,
		"route": "TRANSACTION",
	})
	rlog.Info("[grpc] processing request")

	t := server.stateManager.getTransaction(request.Id)
	if t == nil {
		rlog.Error("transaction not found")
		return nil, ErrTransactionNotFound
	}
	return t.encode(), nil
}

// Transactions gets the status of all transactions currently being tracked in the
// plugin's transaction cache.
//
// It is the handler for the Synse gRPC V3Plugin service's `Transactions` RPC method.
func (server *server) Transactions(request *synse.Empty, stream synse.V3Plugin_TransactionsServer) error {
	log.WithFields(log.Fields{
		"route": "TRANSACTIONS",
	}).Info("[grpc] processing request")

	for _, item := range server.stateManager.transactions.Items() {
		t, ok := item.Object.(*transaction)
		if ok {
			if err := stream.Send(t.encode()); err != nil {
				return err
			}
		}
	}
	return nil
}

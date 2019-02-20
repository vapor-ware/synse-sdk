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

package config

import "time"

// Plugin contains the configuration for a Synse Plugin.
type Plugin struct {
	// Version is the major version of the plugin configuration.
	Version int `yaml:"version,omitempty"`

	// Debug is a flag to determine whether the plugin should be run with
	// debug logging or regular logging.
	Debug bool `default:"false" yaml:"debug,omitempty"`

	// ID specifies the options for generating a plugin namespace ID.
	ID *IDSettings `default:"{}" yaml:"id,omitempty"`

	// Settings specifies how the plugin should run.
	Settings *PluginSettings `default:"{}" yaml:"settings,omitempty"`

	// Network specifies the networking configuration for the the plugin.
	Network *NetworkSettings `default:"{}" yaml:"network,omitempty"`

	// DynamicRegistration specifies the settings and data for dynamically
	// registering devices to the plugin.
	DynamicRegistration *DynamicRegistrationSettings `default:"{}" yaml:"dynamicRegistration,omitempty"`

	// Health specifies the health settings for the plugin.
	Health *HealthSettings `default:"{}" yaml:"health,omitempty"`
}

// IDSettings are the settings around the plugin ID namespace.
type IDSettings struct {
	// UseMachineID determines whether the machine ID should be used as a
	// part of the namespace for the plugin ID.
	UseMachineID bool `default:"true" yaml:"useMachineID,omitempty"`

	// UsePluginTag determines whether the plugin metadata tag should be used
	// as a part of the namespace for the plugin ID.
	UsePluginTag bool `default:"true" yaml:"usePluginTag,omitempty"`

	// UseEnv allows environment variables to be used when generating the namespace
	// for the plugin ID.
	UseEnv []string `yaml:"useEnv,omitempty"`

	// UseCustom allows setting custom identifiers to be used in generating the namespace
	// for the plugin ID.
	UseCustom []string `yaml:"useCustom,omitempty"`
}

// PluginSettings are the settings around the runtime behavior of a plugin.
type PluginSettings struct {
	// Mode is the run mode of the read and write loops. This can either
	// be "serial" or "parallel".
	Mode string `default:"parallel" yaml:"mode,omitempty"`

	// Listen contains the settings to configure listener behavior.
	Listen *ListenSettings `default:"{}" yaml:"listen,omitempty"`

	// Read contains the settings to configure read behavior.
	Read *ReadSettings `default:"{}" yaml:"read,omitempty"`

	// Write contains the settings to configure write behavior.
	Write *WriteSettings `default:"{}" yaml:"write,omitempty"`

	// Transaction contains the settings to configure transaction
	// handling behavior.
	Transaction *TransactionSettings `default:"{}" yaml:"transaction,omitempty"`

	// Limiter specifies settings for rate limiting for reads/writes.
	Limiter *LimiterSettings `default:"{}" yaml:"limiter,omitempty"`

	// Cache contains the settings to configure local data caching
	// by the plugin.
	Cache *CacheSettings `default:"{}" yaml:"cache,omitempty"`
}

// ListenSettings are the settings for listener behavior.
type ListenSettings struct {
	// Disable can be used to globally disable listening for the plugin.
	// By default, plugin listening is enabled.
	Disable bool `default:"false" yaml:"disable,omitempty"`
}

// ReadSettings are the settings for read behavior.
type ReadSettings struct {
	// Disable can be used to globally disable reading for the plugin.
	// By default, plugin reading is enabled.
	Disable bool `default:"false" yaml:"disable,omitempty"`

	// Interval specifies the duration that the read loop should
	// sleep between iterations. By default, no interval is specified.
	//
	// An interval may be useful for tuning the performance of a plugin. In
	// particular, it can be useful for serial protocols to introduce a
	// bit of a delay so the serial bus is not constantly hammered.
	Interval time.Duration `default:"0s" yaml:"interval,omitempty"`

	// Delay specifies a plugin-global delay between successive reads.
	// By default, no delay is specified.
	//
	// A delay can be useful for tuning the performance of a plugin. In
	// particular, it can be useful for serial protocols to introduce a
	// bit of a delay so the serial bus is not constantly hammered.
	Delay time.Duration `default:"0s" yaml:"delay,omitempty"`

	// QueueSize defines the size of the read queue. This will be the
	// size of the channel that queues up and passes along readings as
	// they are collected.
	//
	// Generally this does not need to be set, but can be used to tune
	// performance for read-intensive plugins.
	QueueSize int `yaml:"queueSize,omitempty"`
}

// WriteSettings are the settings for write behavior.
type WriteSettings struct {
	// Disable can be used to globally disable writing for the plugin.
	// By default, plugin writing is enabled.
	Disable bool `yaml:"disable,omitempty"`

	// Interval specifies the duration that the write loop should
	// sleep between iterations. By default, no interval is specified.
	//
	// An interval may be useful for tuning the performance of a plugin. In
	// particular, it can be useful for serial protocols to introduce a
	// bit of a delay so the serial bus is not constantly hammered.
	Interval time.Duration `default:"0s" yaml:"interval,omitempty"`

	// Delay specifies a plugin-global delay between successive writes.
	// By default, no delay is specified.
	//
	// A delay can be useful for tuning the performance of a plugin. In
	// particular, it can be useful for serial protocols to introduce a
	// bit of a delay so the serial bus is not constantly hammered.
	Delay time.Duration `default:"0s" yaml:"delay,omitempty"`

	// QueueSize defines the size of the write queue. This will be the
	// size of the channel that queues up and passes along write requests.
	//
	// Generally this does not need to be set, but can be used to tune
	// performance for write-intensive plugins.
	QueueSize int `yaml:"queueSize,omitempty"`

	// BatchSize defines the maximum number of writes to process in a
	// single batch.
	//
	// Generally, this does not need to be set, but can be used to tune
	// performance particularly for slow writing serial plugins.
	BatchSize int `yaml:"batchSize,omitempty"`
}

// TransactionSettings are the settings for transaction operations.
type TransactionSettings struct {
	// TTL is the time-to-live for a transaction in the transaction cache.
	TTL time.Duration `default:"5m" yaml:"ttl,omitempty"`
}

// CacheSettings are the settings for an in-memory windowed cache of plugin readings.
type CacheSettings struct {
	// Enabled determines whether a plugin will use a local in-memory cache
	// to store a small window of readings. It is disabled by default.
	Enabled bool `yaml:"enabled,omitempty"`

	// TTL is the time-to-live for a reading in the readings cache. This will
	// only be used if the cache is enabled. Once a reading exceeds this TTL,
	// it is removed from the cache.
	TTL time.Duration `yaml:"ttl,omitempty"`
}

// NetworkSettings are the settings for a plugin's networking behavior.
type NetworkSettings struct {
	// Type is the protocol type. Currently, this must be one of: "tcp"
	// (TCP/IP) or "unix" (Unix Socket).
	Type string `yaml:"type,omitempty"`

	// Address is the address that the gRPC server will run on. For
	// "tcp", this would be the host/port (e.g. "0.0.0.0:5001"). For
	// "unix", this would be the name of the socket (e.g. plugin.sock).
	Address string `yaml:"address,omitempty"`

	// TLS contains the TLS/SSL settings for the gRPC server. If this
	// is not set, insecure transport will be used.
	TLS *TLSNetworkSettings `yaml:"tls,omitempty"`
}

// TLSNetworkSettings are the settings for TLS/SSL for the gRPC server.
type TLSNetworkSettings struct {
	// Cert is the location of the cert file to use for the gRPC server.
	Cert string `yaml:"cert,omitempty"`

	// Key is the location of the cert file to use for the gRPC server.
	Key string `yaml:"key,omitempty"`

	// CACerts are a list of certificate authority certs to use. If none
	// are specified, the OS system-wide TLS certs are used.
	CACerts []string `yaml:"caCerts,omitempty"`

	// SkipVerify is a flag that, when set, will skip certificate checks.
	SkipVerify bool `yaml:"skipVerify,omitempty"`
}

// DynamicRegistrationSettings are the settings for dynamic device registration.
type DynamicRegistrationSettings struct {
	// Config holds the configuration(s) for dynamic device registration. It holds
	// the plugin/protocol/device-specific data which will be used to register
	// devices at runtime, e.g. a server address and port.
	Config []map[string]interface{} `yaml:"config,omitempty"`
}

// LimiterSettings are the settings for rate limiting on reads and writes.
type LimiterSettings struct {
	// Rate is the limit, or maximum frequency of events. A rate of
	// 0 signifies 'unlimited'.
	Rate int `yaml:"rate,omitempty"`

	// Burst defines the bucket size for the limiter, or maximum number
	// of events that can be fulfilled at once. If this is 0, it will take
	// the same value as the rate.
	Burst int `yaml:"burst,omitempty"`
}

// HealthSettings are the settings for plugin health.
type HealthSettings struct {
	// HealthFile is the fully qualified path to the file that will be used
	// to signal that the plugin is healthy. If not set, this will default to
	// "/etc/synse/plugin/healthy".
	HealthFile string `yaml:"healthFile,omitempty"`

	// Checks are the settings for plugin health checks.
	Checks *HealthCheckSettings `yaml:"checks,omitempty"`
}

// HealthCheckSettings are the settings for plugin health checks.
type HealthCheckSettings struct {
	// DisableDefaults determines whether the default plugin health checks
	// should be disabled.
	DisableDefaults bool `yaml:"disableDefaults,omitempty"`
}

/*
Package config contains the configuration definitions and utilities for the SDK.

The logic for reading and parsing configuration files are found in device.go
for device instance configurations, proto.go for device prototype configurations,
and plugin.go for plugin configurations.

Versioned configuration schemes are found in files following the naming pattern
of:

    v<VERSION NUMBER>-<CONFIG TYPE>.go

For example, a file named "v1.0-plugin.go" would define the version 1.0 scheme
for plugin configurations.
*/
package config

/*
Package sdk enables the development of plugins for the Synse Platform.

The Synse SDK implements the core plugin logic and makes it easy to create
new plugins. These plugins form the backbone of the Synse platform, interfacing
with devices and exposing them to Synse Server so all devices can be managed
and controlled though a simple HTTP interface, regardless of backend protocol.

For more information, see: https://synse.readthedocs.io/en/latest/sdk/intro/
*/
package sdk

import (
	log "github.com/sirupsen/logrus"
)

func init() {
	// Logging defaults: use a formatter that gives us millisecond resolution.
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02T15:04:05.999Z07:00",
	})
}

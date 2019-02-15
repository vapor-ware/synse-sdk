/*
Package sdk enables the development of plugins for the Synse Platform.

The Synse SDK implements the core plugin logic and makes it easy to create
new plugins. These plugins form the backbone of the Synse platform, interfacing
with devices and exposing them to Synse Server so all devices can be managed
and controlled though a simple HTTP interface, regardless of backend protocol.

todo: overview of the SDK arch/flow.

*/
package sdk

import (
	log "github.com/Sirupsen/logrus"
)

func init() {
	// Logging defaults: set the level to info and use a formatter that gives
	// us millisecond resolution.
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02T15:04:05.999Z07:00",
	})
}

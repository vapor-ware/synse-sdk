package sdk

import (
	"flag"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
)

var (
	flagDebug   bool
	flagVersion bool
	flagDryRun  bool
)

// TODO (etd): this could be moved to plugin.go once that is cleaned up and
//  simplified a bit. then, it would become more clear to someone looking at
//  how a plugin is initialized which command line flags are being processed.
//  It also makes it easier for plugin setup to have these flags available right
//  there...

func init() {
	flag.BoolVar(&flagDebug, "debug", false, "run the plugin with debug logging")
	flag.BoolVar(&flagVersion, "version", false, "print plugin version information")
	flag.BoolVar(&flagDryRun, "dry-run", false, "perform a dry run to verify the plugin is functional")
}

// parseFlags parses any command line flags passed to the plugin and executes
// appropriate actions for the flags. Not all flags will result in action here.
//
// All flags are parsed here, but only SDK-supported flags are handled here. If
// a plugin specifies additional flags, they should be resolved in a pre-run action.
func parseFlags() {
	flag.Parse()

	// --help is already provided by the flag package, so we don't have to
	// handle it here.

	// --debug will enable debug logging.
	if flagDebug {
		log.SetLevel(log.DebugLevel)
	}

	// --version will print out version info and then exit.
	if flagVersion {
		fmt.Println(version.Format())
		os.Exit(0)
	}
}

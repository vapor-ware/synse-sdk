package sdk

import (
	"flag"
)

var (
	flagDebug   bool
	flagVersion bool
	flagDryRun  bool
)

func init() {
	flag.BoolVar(&flagDebug, "debug", false, "run the plugin with debug logging")
	flag.BoolVar(&flagVersion, "version", false, "print plugin version information")
	flag.BoolVar(&flagDryRun, "dry-run", false, "perform a dry run to verify the plugin is functional")
}

package main

import (
	"fmt"

	"github.com/vapor-ware/synse-sdk/sdk"
)

// In order for GoReleaser to work properly, it needs something to build and
// archive to publish along with the GitHub release. The SDK does not have any
// build artifacts, as it is a base library for other plugins to use. We still
// want to use GoReleaser to for creating the GitHub release and generating the
// changelog, so this simple program serves the purpose of being the build target
// for GoReleaser so it can properly run.

func main() {
	fmt.Printf("SDK Version: %s\n", sdk.Version)
}

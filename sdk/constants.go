package sdk

var (
	// sockPath is the base path for gRPC sockets.
	// It's under /tmp rather than /var/run so that local tests pass.
	sockPath = "/tmp/synse/procs"
)

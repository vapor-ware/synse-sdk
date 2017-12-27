package sdk

const (
	// the directory which contains the device configurations.
	// FIXME - this is currently relative to the binary.. should be configurable?
	configDir = "config"

	// fixme: we should probably make this a more standard place.. /var/run?
	// also - probably doesn't belong here.
	sockPath = "/synse/procs"
)

### Dynamic Device Registration Plugin

This directory contains an example of a simple plugin where the devices are
dynamically registered. While the read and write behavior here is the same as the
`simple-plugin` example, this showcases how dynamic device registration is setup.

Note that in the configuration here, there are no device configs. Although the device
configuration directory is empty, when we run the plugin and check its devices, we see
that devices are configured and returned. This is because those devices were created 
from dynamic registration.

#### Usage

To build the simple plugin, simply
```bash
make build
```
from within this directory. This will create the `plugin` binary which can be
run with
```bash
./plugin
```

Once running, you should see some output, and then it will appear to hang. In
the background, it is performing reads continuously, however messages are only
logged out when incoming gRPC requests are handled. For this you will need to 
use the gRPC client. See the `client` directory for more.
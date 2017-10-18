### Multi Device Plugin

This directory contains an example of a somewhat simple plugin which uses 
functions defined in C as the plugin's read and write functions. This is 
mainly to demonstrate how to integrate C into plugins, so the actual read
and write functionality here is simple.

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
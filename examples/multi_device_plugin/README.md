### Multi Device Plugin

This directory contains an example of a somewhat simple plugin, though more
complex than the "simple plugin". It dispatches reads and writes to perform
different actions based on the characteristics of the device specified for
read or write. 

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
### Simple Plugin

This directory contains an example of an extremely simple plugin. This plugin
serves mainly as an example of the structure of writing a plugin. The read and
write handlers specified here return random values and do nothing, respectively.
The source code has extensive comments describing all of the components and how 
they come together. 

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
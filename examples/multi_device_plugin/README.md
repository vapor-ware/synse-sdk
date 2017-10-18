### Multi Device Plugin

This directory contains an example of a somewhat simple plugin, though more
complex than the "simple plugin". It dispatches reads and writes to perform
different actions based on the characteristics of the device specified for
read or write.

In this case, there are three different kinds of devices. While they are all
different types (temperature, voltage, airflow), they could all be the same
type with different models. In this example, we are differentiating between
devices only by looking at the model, but it shouldn't be hard to extend this
to use a different device feature or compound features, e.g. model and type. 

Since this example is primarily to look at the plugin setup, the reads are kept
very simple for each device model. 
- The airflow device always returns a reading of 100.
- The temperature device always returns a reading of 10.
- The voltage device always returns a reading of 1.

For all devices, the write command does nothing.
 
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
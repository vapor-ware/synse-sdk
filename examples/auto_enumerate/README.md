### Auto Enumerated Device Plugin

This directory contains an example of a simple plugin where the devices are
auto-enumerated. While the read and write behavior here is the same as the
`simple-plugin` example, this showcases how device auto-enumeration is setup.

Note that in the configuration here, there are no device instance configs 
(e.g. configurations in the `config/device` directory) -- only prototype
configs. Prototype configurations are still required for a plugin - only instance
configurations can be auto-enumerated, and they must match up to the prototype
config in order to be recognized as a valid device.

Although the device instance configuration directory is empty, when we run the
plugin and check its metainfo, we see that devices are configured and returned.
This is because those devices were created from auto-enumeration.

```bash
# pcli is the client, built from the synse-sdk/client directory
./pcli --name auto-enum-plugin metainfo
2017/10/25 10:24:24 timestamp:"2017-10-25 10:24:24.251057347 -0400 EDT m=+26.814486143" uid:"f83e9127c9454b2531b21abf36dd92d6" type:"temperature" model:"temp2010" manufacturer:"vaporio" protocol:"emulator" location:<rack:"rack-1" board:"board-1" > output:<type:"temperature" precision:2 unit:<name:"celsius" symbol:"C" > range:<max:100 > > 
2017/10/25 10:24:24 timestamp:"2017-10-25 10:24:24.251146411 -0400 EDT m=+26.814575207" uid:"5bb474a286f1a8874ec9cbc3f5860329" type:"temperature" model:"temp2010" manufacturer:"vaporio" protocol:"emulator" location:<rack:"rack-1" board:"board-1" > output:<type:"temperature" precision:2 unit:<name:"celsius" symbol:"C" > range:<max:100 > > 
2017/10/25 10:24:24 timestamp:"2017-10-25 10:24:24.251165357 -0400 EDT m=+26.814594153" uid:"510a7b2e1c59f40765b25c5668e7eacb" type:"temperature" model:"temp2010" manufacturer:"vaporio" protocol:"emulator" location:<rack:"rack-1" board:"board-1" > output:<type:"temperature" precision:2 unit:<name:"celsius" symbol:"C" > range:<max:100 > > 
2017/10/25 10:24:24 timestamp:"2017-10-25 10:24:24.25117636 -0400 EDT m=+26.814605156" uid:"9da8d1ce54ccc7e1ec84d3bea0391b2c" type:"temperature" model:"temp2010" manufacturer:"vaporio" protocol:"emulator" location:<rack:"rack-1" board:"board-1" > output:<type:"temperature" precision:2 unit:<name:"celsius" symbol:"C" > range:<max:100 > > 
2017/10/25 10:24:24 timestamp:"2017-10-25 10:24:24.25119032 -0400 EDT m=+26.814619116" uid:"2381565715f4108e290c3de7a4309e8d" type:"temperature" model:"temp2010" manufacturer:"vaporio" protocol:"emulator" location:<rack:"rack-1" board:"board-1" > output:<type:"temperature" precision:2 unit:<name:"celsius" symbol:"C" > range:<max:100 > > 
2017/10/25 10:24:24 timestamp:"2017-10-25 10:24:24.251200492 -0400 EDT m=+26.814629288" uid:"b997ed16db905ddcc4469a4cc945941b" type:"temperature" model:"temp2010" manufacturer:"vaporio" protocol:"emulator" location:<rack:"rack-1" board:"board-1" > output:<type:"temperature" precision:2 unit:<name:"celsius" symbol:"C" > range:<max:100 > > 
```

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
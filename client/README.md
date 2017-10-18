### Client

This directory contains a simple CLI client that can be used to issue basic
gRPC requests to a plugin. This is particularly useful for playing around with
the example plugins found in the `examples` directory.

This is not intended for production use and should only be used as a tool for
experimentation, development, or ad hoc testing.

To build the CLI tool, simply just
```bash
make build
```
from within this directory. That will create the `pcli` (plugin cli) binary that
can be run with
```bash
./pcli
```

#### Examples
`pcli` contains four basic commands: `read`, `write`, `transaction`, and `metainfo`.
These correspond to the four gRPC API commands that are supported by Synse Server and
the Plugin SDK.

```
$ ./pcli --help
Simple CLI client for Synse Server gRPC testing.

Usage:
  pcli [command]

Available Commands:
  help        Help about any command
  metainfo    Issue a gRPC Metainfo request
  read        Issue a gRPC Read request
  transaction Issue a gRPC Transaction Check request
  write       Issue a gRPC Write request

Flags:
  -h, --help          help for pcli
  -n, --name string   Name of the plugin (e.g. socket name)

Use "pcli [command] --help" for more information about a command.
```

> ***Note***: 
> *the `--name` flag is required, as it tells us which gRPC server to talk to. This
> could be made easier in the future by extending this CLI, but since it is mainly 
> intended for simple experimentation/demonstration, it was kept simple.*
> 
> *additionally, while the `write` command supports sending both an action string
> and raw bytes for writing, the CLI only supports the action string for simplicity.
> future contributions can be made to improve CLI functionality.*

In the following examples, the CLI was running against the "simple-plugin", which
is defined in the `examples/simple_plugin` directory. 

##### Metainfo
`metainfo` takes no arguments and returns a collection of all the configured
devices for the plugin, along with all the known information about them. This
command should typically be run first, as it gives the device's uid which is
used in subsequent commands.

```
$ ./pcli --name simple-plugin metainfo
2017/10/18 08:58:56 timestamp:"2017-10-18 08:58:56.279858737 -0400 EDT" uid:"847d2d1d7d4e4b9776f6f73bebb8825d" type:"emulated-led" model:"emul8-led" manufacturer:"vaporio" protocol:"emulator" info:"Chamber LED 1" comment:"first emulated led device" location:<rack:"unknown" board:"unknown" > output:<type:"led_state" unit:<> range:<> > output:<type:"led_color" unit:<> range:<> > 
2017/10/18 08:58:56 timestamp:"2017-10-18 08:58:56.280472777 -0400 EDT" uid:"21f96641b1d525d9d81966f5ca7e6213" type:"emulated-led" model:"emul8-led" manufacturer:"vaporio" protocol:"emulator" info:"Chamber LED 2" comment:"second emulated led device" location:<rack:"unknown" board:"unknown" > output:<type:"led_state" unit:<> range:<> > output:<type:"led_color" unit:<> range:<> > 
2017/10/18 08:58:56 timestamp:"2017-10-18 08:58:56.280489013 -0400 EDT" uid:"7ba58b52c2f0e4aad5d7e45cafa63d0f" type:"emulated-temperature" model:"emul8-temp" manufacturer:"vaporio" protocol:"emulator" info:"CEC temp 1" comment:"first emulated temperature device" location:<rack:"unknown" board:"unknown" > output:<type:"temperature" precision:2 unit:<name:"celsius" symbol:"C" > range:<max:100 > > 
2017/10/18 08:58:56 timestamp:"2017-10-18 08:58:56.280501771 -0400 EDT" uid:"82efb65e28b422819be8c59fa21e6e25" type:"emulated-temperature" model:"emul8-temp" manufacturer:"vaporio" protocol:"emulator" info:"CEC temp 2" comment:"second emulated temperature device" location:<rack:"unknown" board:"unknown" > output:<type:"temperature" precision:2 unit:<name:"celsius" symbol:"C" > range:<max:100 > > 
2017/10/18 08:58:56 timestamp:"2017-10-18 08:58:56.280512753 -0400 EDT" uid:"14f3bfcd9fb7e12c1f8756b425ccaad6" type:"emulated-temperature" model:"emul8-temp" manufacturer:"vaporio" protocol:"emulator" info:"CEC temp 3" comment:"third emulated temperature device" location:<rack:"unknown" board:"unknown" > output:<type:"temperature" precision:2 unit:<name:"celsius" symbol:"C" > range:<max:100 > > 
```

##### Read
`read` takes a device uid and returns the latest reading value(s) for that
device. Multiple reading values are returned if the device supports multiple
reading values (e.g. a humidity device may return both humidity and temperature
readings).

```
$ ./pcli --name simple-plugin read 847d2d1d7d4e4b9776f6f73bebb8825d
2017/10/18 08:59:14 timestamp:"2017-10-18 08:59:14.523144898 -0400 EDT" type:"emulated-led" value:"1900820199766248294" 
```

##### Write
`write` takes a device uid and a write action. see note, above, regarding 
raw byte values for the write command. The transaction id and some 
transaction context is returned.

```
$ ./pcli --name simple-plugin write 847d2d1d7d4e4b9776f6f73bebb8825d on
b7jl0b2un4a154rn9u4g  - action:"on"
```


##### Transaction Check
`transaction` takes the transaction id as an argument and returns the status
of that transaction. The transaction id is returned from the `write` command.

```
$ ./pcli --name simple-plugin transaction b7jl0b2un4a154rn9u4g
created:"2017-10-18 08:59:24.700894929 -0400 EDT" updated:"2017-10-18 08:59:24.701427269 -0400 EDT" status:DONE
```

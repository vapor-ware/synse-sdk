### Listener Plugin

This directory contains an example of a simple plugin that defines a listener
function in its device handler. Generally, a device that uses a listener is one
that generates push-based data for the plugin to collect. The listener will
listen for this data and update the plugin state accordingly.

In this case, there is only one kind of device, a "pusher". It will push random
data. In order to collect pushed data, we need something to actually push that
data. A simple program is defined in the "pusher" directory which can be run
alongside this plugin to provide the data. See the next section on how to build
and run the plugin and the pusher data source.
 
#### Usage

To build the pusher data source program, simply
```bash
make pusher
```
from within this directory. The plugin binary can be built with
```bash
make build
```

Both binaries will be output to the 'listener' directory and should
be named `device` and `plugin`, respectively. You can run both simultaneously
in separate shell instances (order doesn't matter):

**Shell 1**
```console
$ ./device 
2018/10/12 11:03:15 Sending data on: :8553
2018/10/12 11:03:15 << 2596996162
2018/10/12 11:03:18 << 4039455774
2018/10/12 11:03:21 << 2854263694
2018/10/12 11:03:24 << 1879968118
2018/10/12 11:03:27 << 1823804162
2018/10/12 11:03:30 << 2949882636
2018/10/12 11:03:33 << 281908850
```

**Shell 2**
```console
./plugin 
DEBU[0000] [sdk] adding 1 devices from config           
DEBU[0000] [sdk] executing 0 pre-run action(s)          
DEBU[0000] [sdk] executing 0 device setup action(s)     
INFO[0000] Plugin Info:                                 
INFO[0000]   Tag:         vaporio/listener-plugin       
INFO[0000]   Name:        listener plugin               
INFO[0000]   Maintainer:  vaporio                       
INFO[0000]   Description: An example plugin with listener device 
INFO[0000]   VCS:                                       
INFO[0000] Version Info:                                
INFO[0000]   Plugin Version: 1.0                        
INFO[0000]   SDK Version:    1.1.0                      
INFO[0000]   Git Commit:     95a2def                    
INFO[0000]   Git Tag:        1.1.0                      
INFO[0000]   Build Date:     2018-10-12T15:01:46        
INFO[0000]   Go Version:     go1.10.2                   
INFO[0000]   OS/Arch:        darwin/amd64               
INFO[0000] Registered Devices:                          
INFO[0000]   rack-1-board-1-f9def8b577bf354577e7c0c907fc5b86 (pusher) 
INFO[0000] --------------------------------             
DEBU[0000] [sdk] starting plugin run                    
DEBU[0000] [sdk] registering default health checks      
DEBU[0000] [health] new periodic health check            interval=30s name="read buffer health"
DEBU[0000] [health] new periodic health check            interval=30s name="write buffer health"
DEBU[0000] [data manager] setting up data manager state 
INFO[0000] [data manager] setting up listeners           handler=pusher
INFO[0000] [data manager] starting read goroutine (reads enabled)  mode=serial
INFO[0000] [data manager] starting write goroutine (writes enabled)  mode=serial
INFO[0000] [data manager] running                       
DEBU[0000] [grpc] setting up server                      mode=unix
DEBU[0000] [server] configuring grpc server for insecure transport 
INFO[0000] [grpc] listening on unix:/tmp/synse/procs/example-plugin.sock 
INFO[0000] [data manager] running listener               device=f9def8b577bf354577e7c0c907fc5b86 handler=pusher
[listener] got data: 2854263694
[listener] got data: 1879968118
[listener] got data: 1823804162
[listener] got data: 2949882636
[listener] got data: 281908850
```


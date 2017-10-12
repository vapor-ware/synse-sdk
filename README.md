# Synse Plugin SDK for Go
An SDK in the Go programming language for creating plugins for Vapor IO's 
[Synse Server][synse-server].

This SDK handles most of the common functionality needed for Synse Server plugins,
such as configuration parsing, background reading, asynchronous writing, generation
and tracking of transaction ids, handling metainfo, etc.

In most cases, a plugin author will only need to write the plugin-specific read and 
write functionality.


## Installing
todo - instructions on install

## Reference

- link to examples
- link to docs

## Overview of SDK Components

- config.go
- interface.go
- managers.go
- models.go
- rw.go
- sdk.go
- server.go
- utils.go


## License
This SDK is licensed under the ____. See LICENSE.txt and NOTICE.txt for more information.



## TODO
 - proper organization
 - figure out how writes will work
    - starting from the grpc command
    - how they get added to the rwloop queue
    - how a write transaction is generated/tracked
    - etc
 - figure out how transaction checks will work
    - this is largely tied to the work above for write
 - proper logging
 - plugin configuration (not prototype/device config) - e.g. debug mode, buffer size, etc?
 - packaging
 - documentation
 - testing
 - figure out error cases / error handling
 - figure out how to properly get the generated protobuf stuff in here (+ also in the python
   synse repo) in a clean and easy way that doesn't require lots of copy + paste, etc. perhaps
   git submodules? may not be too bad this time around since there isn't tons of nesting and
   we can pin it to a release version? or just in its own repo and it can be imported that way?
 - figure out: is there a way to give devices a clearer human-readable name?
 
 
[synse-server]: https://github.com/vapor-ware/synse-server
[![CircleCI](https://circleci.com/gh/vapor-ware/synse-sdk.svg?style=svg&circle-token=a35e96598e3df84da3dc58a4f0f9dcc8632bfbd3)](https://circleci.com/gh/vapor-ware/synse-sdk)

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


 
[synse-server]: https://github.com/vapor-ware/synse-server

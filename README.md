# Synse Plugin SDK for Go
An SDK in the Go programming language for creating plugins for Vapor IO's 
[Synse Server][synse-server].

This SDK handles most of the common functionality needed for Synse Server plugins,
such as configuration parsing, background reading, asynchronous writing, generation
and tracking of transaction ids, handling metainfo, etc.

In most cases, a plugin author will only need to write the plugin-specific read and 
write functionality as well as plugin-specific configuration parsing.


## Installing
```
go get -u github.com/vapor-ware/synse-sdk/sdk
```


## Reference

#### Examples

Before getting started with the SDK, check out the [examples][examples] directory.
It contains various example plugins written using the SDK ranging in complexity.
These examples should give you a good idea on how to start writing your own plugin(s).

#### Documentation

TODO: link to documentation

## Development

#### Testing
Tests for the Synse Plugin SDK are run in CI, the status of which is shown by the 
badge at the top of this README. Tests can be run locally with:

```
go test -v ./sdk
```

for convenience, this can also be done via make.

```
make test
```

#### Linting
Linting is performed as a step in CI. A failure to lint should cause a CI build failure.
The SDK source can also be linted locally using `golint`. To get `golint`, 
```
go get -u github.com/golang/lint/golint
```

then, the source can be linted with
```
golint sdk/...
```

For convenience, this can all be done via make, where linting will also include go
source files in the `client` and `examples` directory.
```
make lint
```


## License
This SDK is licensed under the ____. See LICENSE.txt and NOTICE.txt for more information.


 
[synse-server]: https://github.com/vapor-ware/synse-server
[examples]: https://github.com/vapor-ware/synse-sdk/tree/master/examples
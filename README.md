[![CircleCI](https://circleci.com/gh/vapor-ware/synse-sdk.svg?style=svg&circle-token=a35e96598e3df84da3dc58a4f0f9dcc8632bfbd3)](https://circleci.com/gh/vapor-ware/synse-sdk)

# Synse Plugin SDK for Go
An SDK in the Go programming language for creating plugins for Vapor IO's
[Synse Server][synse-server].

This SDK handles most of the common functionality needed for Synse Server plugins,
such as configuration parsing, background reading, asynchronous writing, generation
and tracking of transaction ids, handling metainfo, etc.

In most cases, a plugin author will only need to write the plugin-specific read and
write functionality as well as plugin-specific configuration parsing.


## Setting up the golang environment on a mac using homebrew.
```
brew install golang
```

### Setup .bashrc
GOPATH is the path to your workspace. It is required to be set and has no default.
```go get``` will install to the first entry in the GOPATH list.

GOROOT is the path to where the Go standard library is located on your local filesystem.

GOBIN is the path to where your Go binaries are installed from running go install.

Sample .bashrc:
```
export GOPATH=$HOME/go:$HOME
export GOROOT=/usr/local/opt/go/libexec
export GOBIN=$HOME/go/bin
export PATH=$PATH:$GOPATH/bin
export PATH=$PATH:$GOROOT/bin
export PATH=$PATH:$GOBIN
```

## Installing
```
go get -u github.com/vapor-ware/synse-sdk/sdk
```


## Reference

### Examples

Before getting started with the SDK, check out the [examples][examples] directory.
It contains various example plugins written using the SDK ranging in complexity.
These examples should give you a good idea on how to start writing your own plugin(s).

### Documentation

TODO: link to documentation

## Development

### Setting up the project
```
make setup
```

### Dependencies
[`go dep`](https://github.com/golang/dep) is used for dependency management. After cloning the repo, you can install `dep` with:

```shell
go get -u github.com/golang/dep/cmd/dep
```

To download the dependencies (or update them in the future) you can run:

```shell
dep ensure -v --vendor-only
```

### Testing
Tests for the Synse Plugin SDK are run in CI, the status of which is shown by the
badge at the top of this README. Tests can be run locally with:

```
go test -v ./sdk
```

for convenience, this can also be done via make.

```
make test
```

### Linting
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

## Ensuring local changes pass in ci.
```
make ci
```

## License
This SDK is licensed under the ____. See LICENSE.txt and NOTICE.txt for more information.



[synse-server]: https://github.com/vapor-ware/synse-server
[examples]: https://github.com/vapor-ware/synse-sdk/tree/master/examples

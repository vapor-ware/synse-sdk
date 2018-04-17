<p align="center"><a href="https://www.vapor.io/"><img src="docs/assets/logo.png" width="360"></a></p>
<p align="center">
    <a href="https://circleci.com/gh/vapor-ware/synse-sdk"><img src="https://circleci.com/gh/vapor-ware/synse-sdk.svg?style=shield&circle-token=a35e96598e3df84da3dc58a4f0f9dcc8632bfbd3"></a>
    <a href="https://codecov.io/gh/vapor-ware/synse-sdk"><img src="https://codecov.io/gh/vapor-ware/synse-sdk/branch/master/graph/badge.svg?token=K9qxpN6AE2" /></a>

<h1 align="center">Synse Plugin SDK</h1>
</p>

<p align="center">A Golang SDK for creating plugins for Synse Server</p>


[Synse Server][synse-server] provides an HTTP API for monitoring and controlling physical
and virtual devices; Synse Plugins provide the backend support for all the devices Synse
Server exposes. This repo contains the official Synse Plugin SDK (written in [Go][go-install])
that can be used to create plugin backends for Synse Server.

The SDK handles most of the common functionality needed for plugins, such as configuration
parsing, background read/write, transaction generation and tracking, meta-info caching, and more.
This means the plugin author should only need to worry about the plugin-specific device support.
See the [SDK Documentation][sdk-docs] for more info.


## The Synse Ecosystem
The Synse SDK is one component of the greater Synse Ecosystem.

- [**vapor-ware/synse-server**][synse-server]: An HTTP server providing a uniform API to interact
  with physical and virtual devices via plugin backends. This can be thought of as a 'front end'
  for Synse Plugins.

- [**vapor-ware/synse-server-grpc**][synse-grpc]: The internal gRPC API that connects Synse
  Server and the Synse Plugins.

- [**vapor-ware/synse-emulator-plugin**][synse-emulator]: A simple plugin with no hardware
  dependencies that can serve as a plugin backend for Synse Server for development,
  testing, and just getting familiar with how Synse Server works or how plugins can be
  written.

- [**vapor-ware/synse-cli**][synse-cli]: A CLI that allows you to easily interact with
  Synse Server (via HTTP) and Plugins (via gRPC) directly from the command line.

- [**vapor-ware/synse-graphql**][synse-graphql]: A GraphQL wrapper around Synse Server's
  HTTP API that provides a powerful query language enabling simple aggregations and
  operations over multiple devices.


## Getting Started
It is strongly recommended that you use a [release][releases] version of the SDK if you are
vendoring dependencies, e.g. with [dep][dep]. The SDK can be installed with:

```
go get -u github.com/vapor-ware/synse-sdk/sdk
```

From there, it is easy to start building your own plugin. The [SDK Documentation][sdk-docs]
provides some useful information on writing plugins. You can also check out the [examples](examples)
directory which contains various example plugins using this SDK. The examples, in conjunction
with the documentation, should get you well on your way to start writing your own plugin(s).


### Developing
If you wish to develop the SDK, see the Developing section in the [SDK Documentation][sdk-docs].


## Sharing Plugins
Have you written a plugin and want to share it with the Synse community? Let us know!
There currently is not a tool or site to search for plugins, so we will maintain a list
here. You can also add the [`synse-plugin`][synse-plugin-tag] tag to your plugin's GitHub repo.

## Feedback
Feedback for the Synse Plugin SDK, or any component of the Synse ecosystem, is greatly appreciated!
If you experience any issues, find the documentation unclear, have requests for features,
or just have questions about it, we'd love to know. Feel free to open an issue for any
feedback you may have.

## Contributing
We welcome contributions to the project. The project maintainers actively manage the issues
and pull requests. If you choose to contribute, we ask that you either comment on an existing
issue or open a new one. This project follows the typical [GitHub Workflow][gh-workflow].

The Synse Plugin SDK is released under the [GPL-2.0](LICENSE) license.


[go-install]: https://golang.org/doc/install
[releases]: https://github.com/vapor-ware/synse-sdk/releases
[dep]: https://github.com/golang/dep
[sdk-docs]: https://vapor-ware.github.io/synse-sdk/
[synse-server]: https://github.com/vapor-ware/synse-server
[synse-cli]: https://github.com/vapor-ware/synse-cli
[synse-emulator]: https://github.com/vapor-ware/synse-emulator-plugin
[synse-graphql]: https://github.com/vapor-ware/synse-graphql
[synse-grpc]: https://github.com/vapor-ware/synse-server-grpc
[gh-workflow]: https://guides.github.com/introduction/flow/
[synse-plugin-tag]: https://github.com/topics/synse-plugin
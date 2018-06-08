/*
Package sdk provides an API for creating plugins for the Synse Platform.

Synse plugins interface with Synse Server via an internal gRPC API which allows
for HTTP access to data and control of devices for any protocol. The SDK serves
as a base for these plugins. It provides the gRPC server needed for Synse Server
to communicate with the plugin, and provides background device reading and
asynchronous device writing capabilities.

While read/write behavior is defined on a per-plugin basis, the SDK provides
built-in support for managing device meta information and for the generation
and management of transaction state for each incoming write request.
*/
package sdk

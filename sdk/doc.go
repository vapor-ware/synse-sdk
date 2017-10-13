/*
Package sdk contains the implementation for the Synse plugin SDK. It
provides built-in handling for the 'Metainfo' and 'Transaction Check'
commands, which the Synse internal gRPC API specify. In addition to
other conveniences, it makes defining new plugins easy by only requiring
the plugin to define the 'Read' and 'Write' behavior.
*/
package sdk

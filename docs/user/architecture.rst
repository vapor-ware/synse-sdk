.. _architecture:

Architecture
============
This page describes the SDK architecture at a high level and provides a summary of
its different components and inner workings.

Overview
--------
The SDK was built to make it easier to develop new plugins. It abstracts away a lot of
the internal state handling and the communication layer from the plugin author, so all
you have to focus on is is implementing plugin logic -- not Synse integration.

At a high level, there are two levels of communication in the SDK. Communication with
Synse Server, and communication with the devices it manages.

Plugin Interaction with Synse Server
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. image:: ../_static/synse-server-simple-arch.svg

When an HTTP API request comes in to Synse Server, e.g. a *read* request, that request
will have some routing information associated with it (``<rack>/<board>/<device>``).
This routing information is used by Synse Server to lookup the device and figure out
which plugin owns it.

Once Synse Server knows where the request is going, it sends over all relevant info
to the plugin via the `Synse gRPC API <https://github.com/vapor-ware/synse-server-grpc>`_.
The capabilities of this API are summarized below in the :ref:`grpcApi` section. The
plugin receives the gRPC request and processes it appropriately, returning the corresponding
response back to Synse Server.

A plugin can be configured to use either TCP or Unix socket for the gRPC transport protocol.


Plugin Interaction with Devices
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. image: ../_static/plugin-arch.svg

When a plugin is run, it will start its "data manager". The data manager will execute reads
and writes for devices continuously (on a configurable interval). The read and write behavior
is defined by the plugin itself, for each device. The diagram above shows the data flow for
reads and writes, starting with an incoming gRPC request from Synse Server.

Reads are executed in a goroutine and the reading values are stored in a local read state
cache. When a gRPC read request comes in, it gets the reading out of the cache. This means that plugin
readings are not always current (e.g. if the read interval is 60s, then a reading in the cache
can be 60s old at most), but with the appropriate read interval, this should be fine. It also
means that device reads can happen asynchronously from API reads.

The same holds true for writes. When a gRPC write request comes in, that write transaction is
put on the write queue, and at some configurable interval, the plugin will execute those writes.

Other incoming gRPC requests, like *transaction* or *device info*, are not handled by the data
manager, since they deal with static information. The handling for these other requests are all
built in to the SDK.


.. _grpcApi:

gRPC API
--------
The `Synse gRPC API <https://github.com/vapor-ware/synse-server-grpc>`_ lets plugins communicate
with Synse Server, and vice versa. Below is a summary of the API methods

:Test:
    Checks that the plugin is reachable.

:Version:
    Gets the version information of the plugin.

:Health:
    Gets the health status of the plugin. A plugin's health status is determined
    by optional health checks.

:Metainfo:
    Get the metadata associated with the plugin. This includes things like the
    plugin name, maintainer, and a brief description of the plugin.

:Capabilities:
    Get the collection of plugin capabilities. This enumerates the different device
    kinds that a plugin supports, and the reading output types supported by each of
    those device kinds.

:Devices:
    Get the information for all devices registered with the plugin.

:Read:
    Read data from a specified device.

:Write:
    Write data to a specified device.

:Transaction:
    Check the status of a write transaction.


The Data Manager
----------------
The Data Manager is a core component of a plugin. While the user should never
have to directly interact with the Data Manager, it is still good to know about.

The data manager is in charge of the read goroutine, the write goroutine, and
the data that gets passed to and from them. It holds the "read cache" and the
"write queue" and manages locking around data access, when necessary.

The data manager supports two run modes:

:serial:
    In serial mode, all readings happen serially, all writing happens serially,
    and the read loop and write loop do not run at the same time.

:parallel:
    In parallel mode, readings happen in parallel, writing happens in parallel,
    and the read loop and write loop can run at the same time.


Reading and writing happens in separate loops, and more specifically, in separate
goroutines altogether. This is done to allow different intervals around reading and
writing (e.g. you may want your plugin to update quickly -- write every 1s, but you
may not need to update readings as quickly -- read every 30s).


Devices
-------
Within the SDK, a `Device <https://godoc.org/github.com/vapor-ware/synse-sdk/sdk#Device>`_
represents the physical or virtual thing that the plugin is interfacing with.

The Device model holds the metadata, config information, and a reference to
its DeviceHandler, which defines how it will be read from/written to.


Readings
--------
A `Reading <https://godoc.org/github.com/vapor-ware/synse-sdk/sdk#Reading>`_
describes a single data point read from a device. It consists of the
reading type, the reading value, and the time at which the reading was
taken.

When generating new readings within a Device's read handler, the timestamp should
follow the RFC3339Nano format, which is the standard time format for plugins and
Synse Server. Built-in helpers, such as ``NewReading`` or ``Output.MakeReading``,
will provide a properly formatted timestamp.

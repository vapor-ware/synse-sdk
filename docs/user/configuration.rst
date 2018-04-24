.. _configuration:

Plugin Configuration
====================
This page describes the different kinds of configuration a plugin has, and gives
examples for each. There are three basic kinds of configuration:

- *Plugin Configuration*: Configuration for how the plugin should behave.
- *Device Prototype*: Meta information for a supported device type.
- *Device Instance*: Instance information for a supported device type.

Device prototype information is relatively static and should not change much. It
is considered safe to package it with the plugin, e.g. in a Docker image. The
plugin configuration and device instance configuration, however, should be defined
on a per-instance basis.


Plugin Configuration
--------------------
The plugin configuration is a YAML file the defines some plugin metainfo
and describes how the plugin should operate.

Default Location
~~~~~~~~~~~~~~~~
The default locations for the plugin configuration (in order of evaluation) are:

.. code-block:: none

    /etc/synse/plugin
    $HOME/.synse/plugin
    $PWD

Where ``$PWD`` (or ``.``) is the directory in which the plugin binary is being run from.


Configuration Options
~~~~~~~~~~~~~~~~~~~~~

:version:
    The version of the configuration scheme.

    .. code-block:: yaml

        version: 1.0


:name:
    The name of the plugin.

    .. code-block:: yaml

        name: example


:debug:
    Enables debug logging.

    .. code-block:: yaml

        debug: true


:network:
    Network settings for the gRPC server.

    :type:
        The type of networking the gRPC server should use. This should
        be one of "tcp" and "unix".

        .. code-block:: yaml

            type: tcp


    :address:
        The network address. For unix socket-based networking, this should
        be the name of the socket. This is typically ``<plugin-name>.sock``,
        e.g. ``example.sock``. For tcp, this can be host/port.

        .. code-block:: yaml

            address: ":5001"


:settings:
    Settings for how the plugin should run, particularly the read/write behavior.

    :mode:
        The run mode. This can be one of "serial" and "parallel". In serial mode,
        locking is done to ensure reads and writes are not done simultaneously. In
        parallel mode, no locking is done so reads and writes can occur simultaneously.

        .. code-block:: yaml

            mode: serial


    :read:
        Settings for device reads.

        :enabled:
            Blanket enable/disable of reading for the plugin. *(default: true)*

            .. code-block:: yaml

                enabled: false


        :interval:
            Perform device reads every *interval*. That is to say for an interval of
            ``1s``, the plugin would read from all devices every second. *(default: 1s)*

            .. code-block:: yaml

                interval: 750ms


        :buffer:
            The size of the read buffer. This is the size of the channel that passes
            readings from the read goroutine to the readings cache. *(default: 100)*.

            .. code-block:: yaml

                buffer: 150


    :write:
        Settings for device writes.

        :enabled:
            Blanket enable/disable of writing for the plugin. *(default: true)*

            .. code-block:: yaml

                enabled: false


        :interval:
            Perform device writes every *interval*. That is to say for an interval of
            ``1s``, the plugin would write *max* writes from the write queue every second.
            *(default: 1s)*

            .. code-block:: yaml

                interval: 750ms

        :buffer:
            The size of the write buffer. This is the size of the channel that passes
            writings from the gRPC write handler to the write goroutine. *(default: 100)*.

            .. code-block:: yaml

                buffer: 150

        :max:
            The max number of write transactions to process in a single pass of
            the write loop. This generally only matters when in *serial* mode.
            *(default: 100)*

            .. code-block:: yaml

                max: 150


    :transaction:
        Settings for write transactions.

        :ttl:
            The time to live for a transaction in the transaction cache,
            after which it will be removed. *(default: 5m)*

            .. code-block:: yaml

                ttl: 10m


:auto_enumerate:
    The auto-enumeration context for a plugin. This is dependent on the plugin
    and the device enumeration handler, but in general it can be anything.
    For more, see :ref:`deviceEnumerationHandler`.

:context:
    Configurable context for the plugin. This is generally not used, but is
    made available as a general map in order to pass values in/around the plugin
    if needed.


:limiter:
    Configurations for a rate limiter against reads and writes. Some backends may
    limit interactions, e.g. some HTTP APIs. This configuration allows a limiter
    to be set up to ensure that a limit is not exceeded.

    :rate:
        The limit, or maximum frequency of events.

        A rate of ``0`` signifies an unlimited rate.

        .. code-block:: yaml

            rate: 500


    :burst:
        The bucket size for the limiter, or maximum number of events that can
        be fulfilled at once.

        If this is ``0``, it will be the same number as the *rate*.

        .. code-block:: yaml

            burst: 30




Example
~~~~~~~
Below is a complete, if contrived, example of a plugin configuration.

.. code-block:: yaml

    version: 1.0
    name: example
    debug: true
    network:
      type: unix
      address: example.sock
    settings:
      mode: parallel
      read:
        interval: 1s
      write:
        interval: 2s


Device Prototype Configuration
------------------------------
Prototype configurations define the static meta-info for a given device type. Additionally,
they define the expected output scheme for those devices.

Default Location
~~~~~~~~~~~~~~~~
The default location for device instance configurations is

.. code-block:: none

    /etc/synse/plugin/config/proto


Configuration Options
~~~~~~~~~~~~~~~~~~~~~

:version:
    The version of the configuration scheme.

    .. code-block:: yaml

        version: 1.0


:prototypes:
    A list of prototype objects.

    :<proto>.type:
        The type of the device. This should match up with the ``type`` of the
        corresponding instance configuration(s).

        .. code-block:: yaml

            type: temperature


    :<proto>.model:
        The model of the device. This should match up with the ``model`` of the
        corresponding instance configuration(s).

        .. code-block:: yaml

            model: example-temp


    :<proto>.manufacturer:
        The manufacturer of the device.

        .. code-block:: yaml

            manufacturer: Vapor IO


    :<proto>.protocol:
        The protocol that the device uses to communicate. This is often the same
        as the kind of plugin, e.g. "ipmi", "rs485".

        .. code-block:: yaml

            protocol: i2c


    :<proto>.output:
        See the output configuration details, below.


The output configuration is a list of reading types. This is separated from
the ``<proto>.output`` above only to give it more room on the page.


:output:
    A list of the supported reading outputs for the device.

    :type:
        The type of the reading. This will be the ``type`` field
        of an `sdk.Reading <https://godoc.org/github.com/vapor-ware/synse-sdk/sdk#Reading>`_.

    :data_type:
        The type of the data. This is the type that the data
        will be cast to in Synse Server, e.g. "int", "float",
        "string".

    :unit:
        The specification for the reading's unit.

        :name:
            The name of the unit, e.g. "millimeters per second"

        :symbol:
            The symbol of the unit, e.g. "mm/s"

    :precision:
        *(Optional)* The decimal precision of the readings. e.g. a precision
        of 3 would round a reading to 3 decimal places.

    :range:
        *(Optional)* The range of permissible values for the reading.

        :min:
            The minimum permissible reading value.

        :max:
            The maximum permissible reading value.



Example
~~~~~~~
Below is a complete, if contrived, example of a device prototype configuration.

.. code-block:: yaml

    version: 1.0
    prototypes:
      - type: temperature
        model: example-temp
        manufacturer: Vapor IO
        protocol: example
        output:
          - type: temperature
            unit:
              name: degrees celsius
              symbol: C
            precision: 2
            range:
              min: 0
              max: 100


Device Instance Configuration
-----------------------------
Device instance configurations define the instance-specific configurations for a device.
This is often, but not exclusively, the information needed to connect to a device, e.g.
an IP address or port. Because device instance configurations should be unique to an instance
of a device, parts of these configurations are also used to generate the composite id hash
for the device.

Default Location
~~~~~~~~~~~~~~~~
The default location for device instance configurations is

.. code-block:: none

    /etc/synse/plugin/config/device


Configuration Options
~~~~~~~~~~~~~~~~~~~~~


:version:
    The version of the configuration scheme.

    .. code-block:: yaml

        version: 1.0


:locations:
    A mapping of location alias to location object. Device instances specify their location
    by referencing the location alias key.

    .. code-block:: yaml

        locations:
          r1b1:
            rack: rack1
            board: board1


    :<location>.rack:
        The name of the rack for the <location> object. This can be either a string, in which
        case it is the rack name, or it can be a mapping. The mapping only supports a single
        key ``from_env``, where the value should be the environment variable to get the name from,
        e.g. ``from_env: HOSTNAME``

        .. code-block:: yaml

            rack:
              from_env: HOSTNAME


    :<location>.board:
        The name of the board for the <location> object. This should be a string.



:devices:
    A list of the device instances, where each item in the list is referenced as ``item``, below.

    .. code-block:: yaml

        devices:
          - type: temperature
            model: example-temp
            instances:
              - channel: "0014"
                location: r1b1
                info: Temperature Device 1


    :<item>.type:
        The type of the device. This should match up with the type specified in the
        corresponding prototype config.

        .. code-block:: yaml

            type: temperature


    :<item>.model:
        The model of the device. This should match up with the model specified in the
        corresponding prototype config.

        .. code-block:: yaml

            model: example-temp


    :<item>.instances:
        A list of instances for the given device type/model. The items in the list are
        objects with no restrictions on the fields/values, except that ``info`` and
        ``location`` are reserved. Each item in the instances list should have a
        ``location`` specified (the value being a valid location alias, defined in
        the ``locations`` object, above). The ``info`` field is not required, but is used
        as a human readable tag for the device which is exposed in the device metainfo.
        All other fields are up to the plugin to define and handle and are typically
        configurations for connecting to or otherwise communicating with the device.

        .. code-block:: yaml

            instances:
              - device_address: "/dev/ttyUSB3"
                base_address: 15
                slave_address: 2
                baud_rate: 19200
                parity: E
                location: r1b1
                info: Example Device 1



Example
~~~~~~~
Below is a complete, if contrived, example of a device instance configuration.

.. code-block:: yaml

    version: 1.0
    locations:
      r1vec:
        rack: rack-1
        board: vec
    devices:
      - type: temperature
        model: example-temp
        instances:
          - id: "1"
            location: r1vec
            info: Example Temperature Sensor 1
          - id: "2"
            location: r1vec
            info: Example Temperature Sensor 2
          - id: "3"
            location: r1vec
            info: Example Temperature Sensor 3


Environment Overrides
---------------------
It may not be convenient to place the configuration files into their default locations,
e.g. when testing locally or mounting into a container. Environment overrides exist that
allow you to tell the plugin where to look for its configuration.

- **PLUGIN_CONFIG** : Specifies the *directory* which contains the plugin configuration
  file, ``config.yml``.
- **PLUGIN_DEVICE_CONFIG** : Specifies the *directory* which contains ``proto`` and ``config``
  subdirectories that hold the configuration YAMLs for the prototype and instance configurations,
  respectively.
- **PLUGIN_PROTO_PATH** : Specifies the *directory* which contains the prototype configuration
  YAMLs.
- **PLUGIN_DEVICE_PATH** : Specifies the *directory* which contains the device instance
  configuration YAMLs.

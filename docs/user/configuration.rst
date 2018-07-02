.. _configuration:

Plugin Configuration
====================
This page describes the different kinds of configuration a plugin has, and gives
examples for each. There are three basic kinds of configuration:

- *Plugin Configuration*: Configuration for how the plugin should behave.
- *Device Configuration*: Configuration for the device instances that the plugin
  should interface with and manage.
- *Output Type Configuration*: Configuration for the supported reading outputs
  for the supported devices.


Plugin Configuration
--------------------
Plugins are configured from a YAML file that defines how the plugin should operate.
Most plugin configurations have sane default values, so it may not even be necessary
to specify your own plugin configuration.

The plugin config file must be named ``config.{yml|yaml}``.


Config Policies
~~~~~~~~~~~~~~~
The following config policies relate to plugin configuration.

- PluginConfigFileOptional *(default)*
- PluginConfigFileRequired
- PluginConfigFileProhibited


Config Locations
~~~~~~~~~~~~~~~~
The default locations for the plugin configuration (in order of evaluation) are:

.. code-block:: none

    $PWD
    $HOME/.synse/plugin
    /etc/synse/plugin

Where ``$PWD`` (or ``.``) is the directory in which the plugin binary is being run from.

A non-default location can be used by setting the ``PLUGIN_CONFIG`` environment variable
to either the directory containing the config file, or to the config file itself.

.. code-block:: none

    PLUGIN_CONFIG=/tmp/plugin/config.yml


Configuration Options
~~~~~~~~~~~~~~~~~~~~~

:version:
    The version of the configuration scheme.

    .. code-block:: yaml

        version: 1.0


:debug:
    Enables debug logging.

    .. code-block:: yaml

        debug: true


:network:
    Network settings for the gRPC server. If this is not specified, it will default
    to a *type* of tcp with an *address* of localhost:5001.

    :type:
        The type of networking protocol the gRPC server should use. This should
        be one of "tcp" or "unix".

        .. code-block:: yaml

            type: tcp


    :address:
        The network address. For unix socket-based networking, this should
        be the path to the socket. For tcp, this can be ip/host[:port].

        .. code-block:: yaml

            address: ":5001"


:settings:
    Settings for how the plugin should run, particularly the read/write behavior.

    :mode:
        The run mode. This can be one of "serial" or "parallel". In serial mode,
        locking is done to ensure reads and writes are not done simultaneously. In
        parallel mode, no locking is done so reads and writes can occur simultaneously.
        *(default: serial)*

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


:dynamicRegistration:
    Settings and configurations for the dynamic registration of devices by a plugin.

    :config:
        The configurations to use for dynamic registration. This should be a list of
        maps, where the key is a string, and the value can be anything. The data in
        each map will be passed to the plugin's configured dynamic registration handler
        function(s).


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


:health:
    Configuration for plugin health checks.

    :useDefaults:
        A flag that determines whether the plugin should use the built-in default
        health checks or not. *(default: true)*


:context:
    Configurable context for the plugin. This is generally not used, but is
    made available as a general map in order to pass values in/around the plugin
    if needed.


Example
~~~~~~~
Below is an example of a plugin configuration.

.. code-block:: yaml

    version: 1.0
    debug: true
    network:
      type: tcp
      address: ":5001"
    settings:
      mode: parallel
      read:
        interval: 1s
      write:
        interval: 2s



Device Configuration
--------------------
Device configurations define the devices that a plugin will interface with and expose to
Synse Server.

All device configs are unified into a single config when the plugin reads them in and
validates them. Device configurations can be specified in a single file, or across multiple
files. The file name does not matter, but it must have a .yml or .yaml extension.


Config Policies
~~~~~~~~~~~~~~~
The following config policies relate to device configuration.

For file configuration:

- DeviceConfigFileOptional
- DeviceConfigFileRequired *(default)*
- DeviceConfigFileProhibited

For dynamic configuration:

- DeviceConfigDynamicOptional *(default)*
- DeviceConfigDynamicRequired
- DeviceConfigDynamicProhibited


Config Locations
~~~~~~~~~~~~~~~~
The default locations for the device configuration(s) (in order of evaluation) are:

.. code-block:: none

    ./config/device
    /etc/synse/plugin/config/device

A non-default location can be used by setting the ``PLUGIN_DEVICE_CONFIG`` environment variable
to either the directory containing the config file, or to the config file itself.

.. code-block:: none

    PLUGIN_DEVICE_CONFIG=/tmp/device/config.yml


Configuration Options
~~~~~~~~~~~~~~~~~~~~~


:version:
    The version of the configuration scheme.

    .. code-block:: yaml

        version: 1.0


:locations:
    A list of location definitions. Device instances specify their location
    by referencing the locations defined here.

    .. code-block:: yaml

        locations:
          - name: r1b1
            rack:
                fromEnv: RACK
            board:
                name: board1


    :<location>.name:
        The name given to the location. This is how the location is identified and
        referenced. There cannot be different locations with the same name.


    :<location>.rack:
        The specification for the rack location. This is a map that contains one of
        the following:

        :name:
            The name of the rack.

        :fromEnv:
            The name of the environment variable holding the name of the rack.


    :<location>.board:
        The specification for the board location. This is a map that contains one of
        the following:

        :name:
            The name of the board.

        :fromEnv:
            The name of the environment variable holding the name of the board.



:devices:
    A list of device kinds, where each item in the list is referenced as ``kind``, below.

    .. code-block:: yaml

        devices:
          - name: temperature
            metadata:
                model: example-temp
            instances:
              - channel: "0014"
                location: r1b1
                info: Temperature Device 1


    :<item>.name:
        The name of the device kind. This name should be unique to the device kind
        for the plugin. This can be arbitrarily namespaced, but the last element of
        the namespace should be the type of device, e.g. "temperature".

        .. code-block:: yaml

            name: foo.bar.temperature


    :<item>.metadata:
        Metadata associated with the device kind. This is a mapping of string to string.
        There is no limit to the amount of metadata stored here. This metadata should be
        for the device kind level, so it could contain information like a product ID,
        model number, manufacturer, etc. This is optional and just used to help identify
        the devices.

        .. code-block:: yaml

            metadata:
                model: example-temp
                manufacturer: vaporio


    :<item>.handlerName:
        Specifies the name of the DeviceHandler to match to this device kind. By default,
        a device kind will match to a handler using its `Name` field. If this field is set,
        it will override that behavior and match to a handler with the name specified here.
        This field is optional.

        .. code-block:: yaml

            handlerName: foo.example


    :<item>.outputs:
        A list of the reading output types provided by device instances for this device kind.
        A device instance can specify its own outputs, but if all instances for a kind will
        support the same outputs, it is cumbersome to re-specify them for every device, so
        they can be specified here and will be inherited by the device instances. See the output
        type config options, below.


    :<item>.instances:
        A list of device instances configured for this device kind. The instance configurations
        define the devices that the plugin will ultimately read from and write to. See the device
        instance config options, below.


**Output Type Config Options**

:type:
    The name of the output type that describes the output format for a device reading output.

    .. code-block:: yaml

        type: foo.temperature


:info:
    Any info that can be used to provide a short human-understandable label, description, or
    summary of the reading output. This is optional.

    .. code-block:: yaml

        info: On-board temperature reading value


:data:
    A map where the key is a string and the value is anything. This data contains any protocol/output
    specific configuration associated with the device output. Most device outputs will not need their
    own configuration data specified here, in which case this can be left empty. It is the responsibility
    of the plugin to handle these values correctly.

    .. code-block:: yaml

        data:
            channel: 3
            port: /dev/ttyUSB0


**Device Instance Config Options**

:info:
    A short human-understandable label, description, or summary of the device instance. While
    this is not required, it is recommended to used, as it makes identifying devices much easier.

    .. code-block:: yaml

        info: top right temperature sensor


:location:
    The location of the device. This should be a string that references the ``name`` of a
    location that was specified in the ``locations`` block of the config. This field is required.

    .. code-block:: yaml

        location: r1b1


:data:
    Any protocol/device specific configuration for this device instance. This will often be
    data used to communicate with the device. It is the responsibility of the plugin to handle
    these values correctly.

    .. code-block:: yaml

        data:
            channel: 5
            port: /dev/ttyUSB0
            id: 14


:outputs:
    A list of the output types for the readings that this device supports. A device instance will need
    to have at least one output type, but can have more. It can inherit output types from its
    device kind. For more, see the section on device outputs, above.

    .. code-block:: yaml

        outputs:
            - type: foo.temperature
            - type: foo.humidity


:disableOutputInheritance:
    A flag that, when set, will prevent this instance from inheriting output types from its
    parent device kind. This is false by default (so it will inherit by default).

    .. code-block:: yaml

        disableOutputInheritance: true


:handlerName:
    The name of a device handler to match to this device instance. By default, a device instance
    will match with a device handler using the Name field of its device kind. This field can be
    set to override that behavior and match to a handler with the name specified here.
    This field is optional.

    .. code-block:: yaml

        handlerName: foo.bar.something


Example
~~~~~~~
Below is an example of a device configuration.

.. code-block:: yaml

    version: 1.0
    locations:
      - name: r1vec
        rack:
          name: rack-1
        board:
          name: vec
    devices:
      - name: temperature
        metadata:
          model: example-temp
          manufacturer: vaporio
        outputs:
          - type: temperature
        instances:
          - info: Example Temperature Sensor 1
            location: r1vec
            data:
              id: 1
          - info: Example Temperature Sensor 2
            location: r1vec
            data:
              id: 2
          - info: Example Temperature Sensor 3
            location: r1vec
            data:
              id: 3



Output Type Configuration
-------------------------
Output type configurations define output types which describe how a device reading
should be formatted and adds context info around the reading output. Output type
configurations can be specified directly in the code, so they do not need to be set
via config file. Since these should not change frequently, it is recommended to
define them in-code, but that may not work well for all plugins, so the option to
define them via config exists.


Config Policies
~~~~~~~~~~~~~~~
The following config policies relate to output type configuration.

- TypeConfigFileOptional *(default)*
- TypeConfigFileRequired
- TypeConfigFileProhibited



Config Locations
~~~~~~~~~~~~~~~~
The default locations for the output type configuration(s) (in order of evaluation) are:

.. code-block:: none

    ./config/type
    /etc/synse/plugin/config/type

A non-default location can be used by setting the ``PLUGIN_TYPE_CONFIG`` environment variable
to either the directory containing the config file, or to the config file itself.

.. code-block:: none

    PLUGIN_DEVICE_CONFIG=/tmp/type/config.yml


Configuration Options
~~~~~~~~~~~~~~~~~~~~~

:version:
    The version of the configuration scheme.

    .. code-block:: yaml

        version: 1.0


:name:
    The name of the output type. Output type names should be unique for a plugin.
    The name can be arbitrarily namespaced.

    .. code-block:: yaml

        name: foo.temperature


:precision:
    The decimal precision that the reading should be rounded to. This is only
    applied to readings that provide float values. This specifies the number of
    decimal places to round to.

    .. code-block:: yaml

        precision: 3


:unit:
    The unit of reading.

    .. code-block:: yaml

        unit:
          name: millimeters per second
          symbol: mm/s


    :name:
        The full name of the unit.

    :symbol:
        The symbolic representation of the unit.


:scalingFactor:
    A factor that the reading value can be multiplied by to get the final
    output value. This is optional and will be 1 if not specified (e.g. the
    reading value will not change). This value should resolve to a numeric.
    Negatives and fractional values are supported. This can be the value itself,
    e.g. "0.01", or a mathematical representation of the value, e.g. "1e-2".

    .. code-block:: yaml

        scalingFactor: -.4E10

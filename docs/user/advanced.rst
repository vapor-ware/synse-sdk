.. _advancedUsage:

Advanced Usage
==============
This page describes some of the more advanced features of the SDK for plugin development.


- Dynamic Registration
√ Pre Run Actions
√ Post Run Actions
√ Device Setup Actions
- Health Checks
- Configuration Policies
√ Plugin Options
√ C backend
√ Command line args



Command Line Arguments
----------------------
The SDK has some built-in command line arguments for plugins. These can be seen by running
the plugin with the ``--help`` flag.

.. code-block:: none

    $ ./plugin --help
    Usage of ./plugin:
      -debug
            run the plugin with debug logging
      -dry-run
            perform a dry run to verify the plugin is functional
      -version
            print plugin version information


A plugin can add its own command line args if it needs to as well. This can be done simply
by defining the flags that the plugin needs, e.g.

.. code-block:: go

    import (
        "flag"
    )

    var customFlag bool

    func init() {
        flag.BoolVar(&customFlag, "custom", false, "some custom functionality")
    }

This flag will be parsed on plugin ``Run()``, so it can only be used after the plugin
has been run.


Pre Run Actions
---------------
Pre Run Actions are actions that the plugin will perform before it starts to
run the gRPC server and start the data manager's read/write goroutines. These actions
can be used for plugin-wide setup, should a plugin require it. For example, this could
be used to perform some kind of authentication, verifying that some backend exists and is
reachable, or to do additional config validation, etc.

Pre Run Actions should fulfil the ``pluginAction`` type and should be registered with the
plugin before it is run. An (abridged) example:

.. code-block:: go

    // preRunAction defines a function that will run before the
    // plugin starts its main run logic.
    func preRunAction(p *sdk.Plugin) error {
        return backend.VerifyRunning()  // do some action
    }

    func main() {
        plugin  := sdk.NewPlugin()

        plugin.RegisterPreRunActions(
            preRunAction,
        )
    }


For more, see the `Device Actions Example Plugin <https://github.com/vapor-ware/synse-sdk/tree/master/examples/device_actions>`_.

Post Run Actions
----------------
Post Run Actions are actions that the plugin will perform after it is shut down gracefully.
A graceful shutdown of a plugin is done by passing the SIGTERM or SIGINT signal to the plugin.
These actions can be used for plugin-wide shutdown/cleanup, such as cleaning up state, terminating
connections, etc.

Post Run Actions should fulfil the ``pluginAction`` type and should be registered with the
plugin before it is run. An (abridged) example:

.. code-block:: go

    // postRunAction defines a function that will run after the plugin
    // has gracefully terminated.
    func postRunAction(p *sdk.Plugin) error {
        return db.closeConnection() // do some action
    }

    func main() {
        plugin := sdk.NewPlugin()

        plugin.RegisterPostRunActions(
            postRunAction,
        )
    }


For more, see the `Device Actions Example Plugin <https://github.com/vapor-ware/synse-sdk/tree/master/examples/device_actions>`_.

Device Setup Actions
--------------------
Some devices might need a setup action performed before the plugin starts to read
or write to them. As an example, this could be performing some type of authentication,
or setting some bit in a register. The action itself is plugin (and protocol) specific
and does not matter to the SDK.

Device Setup Actions should fulfil the ``deviceAction`` type and should be registered with
the plugin before it is run.

When a device setup action is registered, it should be registered with a filter. This filter
is used to identify which devices the action should apply to. An (abridged) example:

.. code-block:: go

    // deviceSetupAction defines a function we will use as a
    // device setup action.
    func deviceSetupAction(p *sdk.Plugin, d *sdk.Device) error {
        return utils.Validate(d) // do some action
    }

    func main() {
        // Create a new Plugin
        plugin := sdk.NewPlugin()

        // Register the action with all devices that have
        // the type "airflow".
        plugin.RegisterDeviceSetupActions(
            "type=airflow",
            deviceSetupAction,
        )
    }


For more, see the `Device Actions Example Plugin <https://github.com/vapor-ware/synse-sdk/tree/master/examples/device_actions>`_.

Plugin Options
--------------
As other sections here describe in more detail, there may be cases where a plugin would want
to override some default plugin functionality. As an example, the SDK provides a default device
identifier function. What this function does is take the config for a particular device and creates
a hash out of that config info in order to create a deterministic ID for the device.

The premise of the ID determinism is that a device config will generally define how to address that
device (e.g. for a serial device, it could be the serial bus, channel, etc). If the config changes,
we are talking to something different, so we assume that a change in config equates to a change in
device identity.

Obviously, this is not always the case, which is where having a custom identifier function becomes
useful. If we wanted to only take a subset of the device config, we could define a simple device
identifier override function, but in order to register it with the plugin, we'd need to use a Plugin
Option.

Plugin Options are passed to the plugin when it is initialized via ``sdk.NewPlugin``.

.. code-block:: go

    // ProtocolIdentifier gets the unique identifiers out of the plugin-specific
    // configuration to be used in UID generation.
    func ProtocolIdentifier(data map[string]interface{}) string {
    	return fmt.Sprint(data["id"])
    }

    func main() {
        plugin := sdk.NewPlugin(
            sdk.CustomDeviceIdentifier(ProtocolIdentifier),
        )
    }

An example of this can be found in the
`Device Actions Example Plugin <https://github.com/vapor-ware/synse-sdk/tree/master/examples/device_actions>`_.


C Backend
---------
Plugins can be written with C backends. In general, this means that the read/write
handlers or some related logic is written in C. This feature is not specific to the
SDK, but is a feature of Go itself.

For more information on this, see the `CGo Documentation <https://golang.org/cmd/cgo/>`_
and the `C Plugin <https://github.com/vapor-ware/synse-sdk/tree/master/examples/c_plugin>`_ example.

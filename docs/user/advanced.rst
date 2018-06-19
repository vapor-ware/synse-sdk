.. _advancedUsage:

Advanced Usage
==============
This page describes some of the more advanced features of the SDK for plugin development.


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

Dynamic Registration
--------------------
Dynamic Registration is when devices are configured not from config YAML files, but
dynamically at runtime. There are two kinds of dynamic registration functions:

- one that creates DeviceConfig(s) (e.g. it creates the configuration for a device)
- one that creates Device(s) (e.g. it creates the device directly)

By default, a plugin will not do any dynamic device registration. In enable dynamic registration
for a plugin, the dynamic registration function will have to be defined, and then it will
have to be passed to the plugin constructor via a PluginOption.

Dynamic registration can be useful when you do not know what devices may exist at any given
time. A good example of this is IPMI. While you should know the BMC IP address, you may not
know all the devices on all your BMCs. Even if you do, it would be cumbersome to have to manually
enumerate these in a config file.

With device enumeration, you can just create a function that will query the BMC for its
devices and then use that response to generate the devices (or the device configs) at runtime.

An extremely simple example of this can be found in the
`Dynamic Registration Example Plugin <https://github.com/vapor-ware/synse-sdk/tree/master/examples/dynamic_registration>`_.

Configuration Policies
----------------------
The SDK exposes different configuration policies that a plugin can set to modify its
behavior. By default, the policies dictate that

- plugin config is optional (e.g. a plugin can use defaults)
- device config(s) are required (e.g. YAML files must be specified for device configs)
- dynamic device config(s) are optional
- output type config file(s) are optional

For many plugins, the default policies will be good enough. Some plugins may require some
explicit configuration, so to enforce it, they can set the appropriate policy. As an example,
there could be a hypothetical plugin that will only allow the pre-defined output types, will
not allow device configs from file, requires devices to be registered from dynamic registration.
The config policies allow that behavior to be enforced, and cause the plugin to terminate if
any of the policies are violated.

Below is a table that lists all of the current config policies. There can only be one (or none)
policy chosen from each column below at any given time, e.g. you cannot have ``PluginConfigFileOptional``
and ``PluginConfigFileRequired`` specified at the same time.

==========================   ==========================   =============================   =========================
Plugin (File)                Device Config (File)         Device Config (Dynamic)         Output Type Config (File)
==========================   ==========================   =============================   =========================
PluginConfigFileOptional     DeviceConfigFileOptional     DeviceConfigDynamicOptional     TypeConfigFileOptional
PluginConfigFileRequired     DeviceConfigFileRequired     DeviceConfigDynamicRequired     TypeConfigFileRequired
PluginConfigFileProhibited   DeviceConfigFileProhibited   DeviceConfigDynamicProhibited   TypeConfigFileProhibited
==========================   ==========================   =============================   =========================

Setting config policies for the plugin is simple:

.. code-block:: go

    import (
        "github.com/vapor-ware/synse-sdk/sdk"
        "github.com/vapor-ware/synse-sdk/sdk/policies"
    )

    func main() {
        plugin := sdk.NewPlugin()

        policies.Add(policies.DeviceConfigFileOptional)
        policies.Add(policies.TypeConfigFileOptional)
    }


An example of this can be found in the
`Dynamic Registration Example Plugin <https://github.com/vapor-ware/synse-sdk/tree/master/examples/dynamic_registration>`_.

Health Checks
-------------
The SDK supports plugin health checks. The health of the plugin derived from these checks is
surfaced via the Synse gRPC API, and can be seen via the Synse Server HTTP API.

A health check is just a function that returns an error. When run, if the function returns
``nil``, the check passed. If an error is returned, the check has failed. Health checks can
be registered and run in different ways, but the SDK only natively supports *periodic* checks
currently.

Writing and registering a health check is simple. As an example, we could define a health check
that will periodically hit a URL to see if it is reachable:

.. code-block:: go

    import (
        "github.com/vapor-ware/synse-sdk/sdk"
        "github.com/vapor-ware/synse-sdk/sdk/health"
    )

    func checkURL() error {
        resp, err := http.Get(someURL)
        if err != nil {
            return err
        }
        if !resp.Ok {
            return fmt.Errorf("Got non-200 response from URL")
        }
        return nil
    }

    func main() {
        plugin := sdk.NewPlugin()

        health.RegisterPeriodicCheck("example health check", 30*time.Second, checkURL)
    }



C Backend
---------
Plugins can be written with C backends. In general, this means that the read/write
handlers or some related logic is written in C. This feature is not specific to the
SDK, but is a feature of Go itself.

For more information on this, see the `CGo Documentation <https://golang.org/cmd/cgo/>`_
and the `C Plugin <https://github.com/vapor-ware/synse-sdk/tree/master/examples/c_plugin>`_ example.

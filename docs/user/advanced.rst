.. _advancedUsage:

Advanced Usage
==============
This page describes some of the more advanced features of the SDK for plugin development.


.. _deviceEnumerationHandler:

Device Enumeration Handler
--------------------------
The `Device Enumeration <https://godoc.org/github.com/vapor-ware/synse-sdk/sdk#DeviceEnumerator>`_ Handler,

.. code-block:: go

    type DeviceEnumerator func(map[string]interface{}) ([]*config.DeviceConfig, error)


is a handler that allows a plugin to register device instances programmatically, not through
pre-defined YAML. A good use case for this is IPMI, where the plugin will know which BMCs to
reach out to, but not which devices are on the BMCs. Instead of manually going through and
constructing the configuration for each server, this can be done through a device enumeration
handler that connects with the BMC, get all devices it has, and then initialize plugin Device
instances for those found devices.

The ``map[string]interface{}`` that is the input parameter to the ``DeviceEnumerator`` function
type is the map defined in the plugin configuration under the ``auto_enumerate`` key. Any values
can be specified there, under any nesting, but it is up to the plugin writer to parse them correctly.


For more, see the `Auto Enumerate Example Plugin <https://github.com/vapor-ware/synse-sdk/tree/master/examples/auto_enumerate>`_.


Pre Run Actions
---------------
Pre Run Actions are actions that the plugin should perform before it starts to
run the gRPC server and start the read/write goroutines. These are actions that
should be used for plugin-wide setup actions, should a plugin require it. This could
be performing some kind of authentication, verifying that some backend exists and is
reachable, etc.

Pre Run Actions should be defined as part of plugin initialization and should
be registered with the plugin before it is run.

A pre run action should fulfil the ``pluginAction`` type

.. code-block:: go

    type pluginAction func(p *Plugin) error

The ``pluginAction`` should then be registered with the plugin via the
`plugin.RegisterPreRunActions <https://godoc.org/github.com/vapor-ware/synse-sdk/sdk#Plugin.RegisterPreRunActions>`_
function.

An (abridged) example:

.. code-block:: go

    // preRunAction defines a function we will run before the
    // plugin starts its main run logic.
    func preRunAction(p *sdk.Plugin) error {
        return backend.VerifyRunning()  // do some action
    }


    func main() {
        // Create a new Plugin
        plugin, err := sdk.NewPlugin(handlers, nil)
        if err != nil {
            log.Fatal(err)
        }

        // Register the action with the plugin.
        plugin.RegisterPreRunActions(
            preRunAction,
        )
    }


For more, see the `Pre Run Actions Example Plugin <https://github.com/vapor-ware/synse-sdk/tree/master/examples/pre_run_actions>`_.


Device Setup Actions
--------------------
Some devices might need a setup action performed before the plugin starts to read
or write to them. As an example, this could be performing some type of authentication,
or setting some bit in a register. The action itself is plugin (and protocol) specific
and does not matter to the SDK.

Device Setup Actions should be defined as part of plugin initialization and should
be registered with the plugin before it is run.

A device setup action should fulfil the ``deviceAction`` type

.. code-block:: go

    type deviceAction func(p *Plugin, d *Device) error


The ``deviceAction`` should then be registered with the plugin via the
`plugin.RegisterDeviceSetupActions <https://godoc.org/github.com/vapor-ware/synse-sdk/sdk#Plugin.RegisterDeviceSetupActions>`_
function.

An (abridged) example:

.. code-block:: go

    // deviceSetupAction defines a function we will use as a
    // device setup action.
    func deviceSetupAction(p *sdk.Plugin, d *sdk.Device) error {
        return utils.Validate(d) // do some action
    }


    func main() {
        // Create a new Plugin
        plugin, err := sdk.NewPlugin(handlers, nil)
        if err != nil {
            log.Fatal(err)
        }

        // Register the action with all devices that have
        // the type "airflow".
        plugin.RegisterDeviceSetupActions(
            "type=airflow",
            deviceSetupAction,
        )
    }


For more, see the `Pre Run Actions Example Plugin <https://github.com/vapor-ware/synse-sdk/tree/master/examples/pre_run_actions>`_.

C Backend
---------
Plugins can be written with C backends. In general, this means that the read/write
handlers or some related logic is written in C. This feature is not specific to the
SDK, but is a feature of Go itself.

For more information on this, see the `CGo Documentation <https://golang.org/cmd/cgo/>`_
and the `C Plugin <https://github.com/vapor-ware/synse-sdk/tree/master/examples/c_plugin>`_ example.

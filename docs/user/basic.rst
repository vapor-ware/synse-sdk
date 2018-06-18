.. _basics:

Basic Usage
===========
This page describes some of basic features of the SDK and provides an example
of a simple plugin. See the :ref:`advancedUsage` page for an overview of some of
the more advanced features of the plugin SDK.


Creating a Plugin
-----------------
Creating a new plugin is as simple as:

.. code-block:: go

    import (
        "log"
        "github.com/vapor-ware/synse-sdk/sdk"
    )

    func main() {
        plugin := sdk.NewPlugin()

        if err := plugin.Run(); err != nil {
            log.Fatal(err)
        }
    }

This creates a new Plugin instance, but doesn't do much more than that. It is always
advised to use ``sdk.NewPlugin`` to create your plugin instance. The plugin should
always be run via ``plugin.Run()``.

Setting Plugin Metadata
-----------------------
At a minimum, a plugin requires a name. Ideally, a plugin should include more than a
name. The current set of plugin metadata includes:

- name
- maintainer
- description
- vcs link
- tag

The Plugin tag is automatically generated from the ``name`` and ``maintainer`` info,
following the template ``{maintainer}/{name}``, where both the maintainer and name fields
are lower-cased, dashes (``-``) are converted to underscored (``_``), and spaces converted
to dashes (``-``).

The plugin metadata should be set via the ``SetPluginMeta`` function, e.g.

.. code-block:: go

    const (
        pluginName       = "example"
        pluginMaintainer = "vaporio"
        pluginDesc       = "example plugin description"
        pluginVcs        = "github.com/foo/bar"
    )

    func main() {
        sdk.SetPluginMeta(
            pluginName,
            pluginMaintainer,
            pluginDesc,
            pluginVcs,
        )
    }


Registering Output Types
------------------------
All plugins will need to define *output types*. An output type is a definition of a
device reading output, providing metadata and requirements around the reading. For
example, a plugin might have a temperature sensor. The plugin will implement the logic
for how to read from a temperature sensor, but that will ultimately resolve to some
value. To make sense of that value, we want to associate to an output type to give it
context.

.. code-block:: go

    var Temperature = sdk.OutputType{
        Name: "temperature",
        Precision: 3,
        Unit: sdk.Unit{
            Name: "celsius",
            Symbol: "C",
        },
    }

With this context, we know that the reading value corresponds to a temperature reading
with unit "celsius", and it will be rounded to a precision of 3 decimal places. The name
of the output type identifies the type, so it should be unique. It is convention to namespace
output types. This allows for multiple similar types to be specified, e.g.

.. code-block:: go

    var Temperature1 = sdk.OutputType{
        Name: "modelX.temperature",
        Precision: 3,
        Unit: sdk.Unit{
            Name: "celsius",
            Symbol: "C",
        },
    }

    var Temperature2 = sdk.OutputType{
        Name: "modelY.temperature",
        Precision: 2,
        Unit: sdk.Unit{
            Name: "Kelvin",
            Symbol: "K",
        },
    }

The namespacing is arbitrary, so it is up to the plugin author to decide what makes the
most sense. With OutputTypes defined, they can be registered with the plugin simply:

.. code-block:: go

    func main() {
        plugin := sdk.NewPlugin()

        err := plugin.RegisterOutputTypes(
            &Temperature1,
            &Temperature2,
        )
    }

OutputTypes can also be defined via config file, in which case, they will not need to
be explicitly registered with the plugin, as seen in the example above. They will be
registered when the configs are read in, during the pre-run setup.


Registering Device Handlers
---------------------------
All plugins need to define *device handlers*. A device handler defines how a particular
device will be read from/written to, and if it is even capable of reads/writes. There are
currently three types of functionality that a device handler can define:

- Read
- Write
- Bulk Read

*Read* defines how an individual device should be read. *Write* defines how an individual
device should bw written to. *Bulk Read* defines read functionality for all devices that
use that handler. That is to say, while a *read* happens one at a time, a *bulk read* will
read all devices at once. While bulk reads have a more limited use case, they can simplify
some device readings, for example, if a give device/protocol requires all registers to be
read through to get a single reading (as can be the case for I2C), it can be easier to just
do that bulk read once instead of re-doing it for every device on that bus.

.. note:: If both a "read" function and "bulk read" function are specified for a single
   device handler, the bulk read will be ignored and the SDK will only use the read function.
   If bulk read function is desired, make sure that no individual read function is specified.

If no function is specified for any of these, the SDK takes that to mean that the handler
does not support that functionality. That is to say, a device handler with only a read
function defined implies that those devices cannot be written to.

Defining a handler is as simple as giving it a name, and the appropriate functions:

.. code-block:: go

    var TemperatureHandler = sdk.DeviceHandler{
        Name: "temperature",
        Read: func(device *sdk.Device) ([]*sdk.Reading, error) {
            ...
        },
    }

See the `GoDoc <https://godoc.org/github.com/vapor-ware/synse-sdk/sdk>`_ for more details on
how handlers should be defined.

Like DeviceOutputs, a DeviceHandler name identifies that handler, so it should be unique.
If necessary, handler names should be namespaced, but the namespacing is arbitrary and left
to the plugin to decide. With DeviceHandlers defined, they can be registered with the plugin
simply:

.. code-block:: go

    func main() {
        plugin := sdk.NewPlugin()

        plugin.RegisterDeviceHandlers(
            &TemperatureHandler,
        )
    }


Creating New Readings
---------------------
When creating a new reading in a handler's read function, it is highly recommended
to use the built-in constructors, as they automatically fill in some fields. In particular,
it is important to note that the Synse platform has standardized on RFC3339 timestamp
formatting, which the built-in constructors do for you.

One of the easiest ways to create a new reading is with the following pattern. Below,
we have some ``value``, which is whatever reading we got. The input to ``GetOutput`` is
the name of the output type. If the output type does not exist for the device, this will cause
the plugin to panic (in this particular pattern), which is typically desirable, since it is
indicative of a mis-configuration in the device configs.

.. code-block:: go

    var someHandler = sdk.DeviceHandler{
        Name: "example.reader",
        Read: func(device *sdk.Device) ([]*sdk.Reading, error) {
            // plugin-specific read logic
            ...

            return []*sdk.Reading{
                device.GetOutput("example.temperature").MakeReading(value),
            }, nil
        },
    }


A Complete Example
------------------
A complete example of a simple plugin that exercises all of these pieces can be found in the
SDK repo's `examples/simple_plugin <https://github.com/vapor-ware/synse-sdk/tree/master/examples/simple_plugin>`_
directory.

For a slightly more complex example, see the `Emulator Plugin <https://github.com/vapor-ware/synse-emulator-plugin>`_.

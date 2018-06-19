.. _tutorial:

Tutorial
========
This page will go through a step by step tutorial on how to create a plugin. The plugin
we will create in this tutorial will be simple and provide readings for a single device. For
examples of more complex plugins, see the ``examples`` directory in the source repo, or see
the `Emulator Plugin <https://github.com/vapor-ware/synse-emulator-plugin>`_.

The plugin we will build here will provide a single "memory" device which will give readings for
total memory, free memory, and the used percent. To get this memory info we will use
`<https://github.com/shirou/gopsutil>`_.


0. Figure out desired behavior
------------------------------
Before we dive in and start writing code, its always a good idea to lay out what we want
that code to do. In this case, we'll outline what we want the plugin to do, and what data
we want it to provide. Since this is a simple, somewhat contrived plugin, these are all
pretty basic.


Goals
~~~~~
- Provide readings for memory usage
- Do not support writing (doesn't make sense in this case)
- Have the readings be updated every 5 seconds

Devices
~~~~~~~
- One kind of device will be supported: a "memory usage" device. It will provide readings for:
    - Total Memory (in bytes)
    - Free Memory (in bytes)
    - Used Percent (percent)

With this outline of what we want in mind, we can start framing up the plugin.


1. Create the plugin skeleton
-----------------------------
If you have read the documentation on plugin configuration, you will know that there are
three types of configurations that a plugin uses: plugin config, device config, and
output type config. What each does is explained in the configuration documentation.

We will not need to define the output type config, since we will have our output types
built directly into the plugin. That means we only need to specify the device config
and the plugin config.

We will include those with our plugin, as well as a file to define the plugin.

.. code-block:: none

    ▼ tutorial-plugin
        ▼ config
            ▼ device
                mem.yml
        config.yml
        plugin.go


.. note:: There are different ways a plugin can be structured. This example does
   not aim to define the "correct" way. Since it is a simple plugin, it just has
   a simple structure.


First, we will focus on writing the configuration for the plugin and the supported
devices. Note that the plugin configuration does not need to be written first. For
this tutorial we are writing if first, though, to help build an understanding of
how devices are defined and how the plugin will ultimately use them.


2. Write the configurations
----------------------------------
First we'll start with the plugin configuration, then we will look at the device
configuration.

Plugin Configuration
~~~~~~~~~~~~~~~~~~~~
The plugin configuration defines how the plugin itself will operate. Since
this is a simple, somewhat contrived plugin with only a single readable device,
the configuration will not be too complicated. See the plugin configuration
documentation for more info on how to configure plugins.

First, we will want to decide what protocol we want the plugin to use. In this
case, we will use unix socket, but it should be trivial to use TCP instead, should
you decide to.

As per the Goals we laid out in section 0, we want the readings to
be updated every 5 seconds. That means we will need to set the read interval
to ``5s``. All together, this would look like:

.. code-block:: yaml
    :caption: config.yml
    :name: config.yml

    version: 1.0
    debug: false
    network:
      type: unix
      address: memory.sock
    settings:
      read:
        interval: 5s


In the above, ``version`` refers to the version of the configuration file scheme,
not the version of the plugin itself. We've also set ``debug: false`` to disable
debug logging. If you wish to see debug logs, just set this to ``true``.


Device Configuration
~~~~~~~~~~~~~~~~~~~~
Next, we will define the device configuration for our memory device.

In this simple case, we can say that our device is a "memory" type device. Although
optional, we will also specify some metadata with it, namely a model (that we will
make up for the sake of the tutorial). The name of the device kind needs to be unique,
but since this is the only device we have here, we don't need to worry about it.

Another component to the instance configurations is defining the device location. If you
are familiar with Synse Server, you will know that we currently reference devices via a
rack/board/device hierarchy, e.g. ``read/rack-1/board-1/device-1``. These are effectively
just labels to namespace devices, so they can be whatever you want them to be. For this
tutorial, we'll say that the rack is ``local`` and the board is ``host``. This should result
in the Synse Server URI ``read/local/host/<device-id>``.

.. note::

    Synse Server 2.0 uses the ``<rack>/<board>/<device>`` notation for identifying
    all devices. This notation is largely historical from the initial design of
    Synse Server, which did not aim to be as generalized as it is now. In future
    versions (e.g. 3.0), early planning and discussion has the strict rack-board-device
    requirements phased out in favor of more generalized labeling. This should not
    be any concern now, but something to look for in the future.


Additionally, we will need to specify the output types of the device readings. We have not
defined those in code yet, but we know from section 0 that we want a single device that outputs:

- Total Memory (in bytes)
- Free Memory (in bytes)
- Used Percent (percent)

So we can call those outputs ``memory.total``, ``memory.free``, and ``percent_used``,
respectively. Later, we will define the output types corresponding to those names.

The final piece to our configuration is specifying the config for the memory device
instance. Here we will only want one device instance (we're only getting memory from one place,
so we only need a single device to do it). As we will see in the next section, we
will need a way to reliably identify this device. For protocols like HTTP, RS-485, and
others, we can do this by using the addressing configuration as part of the ID composite
(if device X can only be reached via unique address A, then address A can help to identify
device X). Since we do not need any protocol-specific configurations for our memory
device, we will just add in an ``id`` field that will provide a reliable unique identifier
for that device (since we only have one device, it may seem weird, but if we were to have
two memory devices, we'd need a way to differentiate).

.. code-block:: yaml
    :caption: config/device/mem.yml
    :name: config/device/mem.yml

    version: 1.0
    locations:
      - name: local
        rack:
          name: local
        board:
          name: host
    devices:
      - name: memory
        metadata:
          model: tutorial-mem
        outputs:
          - type: memory.total
          - type: memory.free
          - type: percent_used
        devices:
          - info: Virtual Memory Usage
            location: local
            data:
              id: 1


In the above config, the ``version`` is the version of the configuration scheme.

3. Define the output types
--------------------------
As mentioned in the previous section, we still need to define the output types that
we used in the device configuration. While we could define these in their own config
files, its easier to just define them right in the code.

We know that both free memory and total memory should describe the number of bytes
and percent used should be a percentage. Knowing this and what we are calling these
output types is all we need

.. code-block:: go

    var (
        memoryTotal = sdk.OutputType{
            Name: "memory.total",
            Unit: sdk.Unit{
                Name: "bytes",
                Symbol: "B",
            },
        }

        memoryFree = sdk.OutputType{
            Name: "memory.free",
            Unit: sdk.Unit{
                Name: "bytes",
                Symbol: "B",
            },
        }

        percentUsed = sdk.OutputType{
            Name: "percent_used",
            Unit: sdk.Unit{
                Name: "percent",
                Symbol: "%",
            },
        }
    )



4. Write handlers for the device(s)
-----------------------------------
If you've read through some of the documentation on plugin basics, you should know that
in order to handle the configured devices, handlers for those devices need to be defined.

We only want our memory device to support reading, so we only need to define a read function
for our device handler. To read the memory info, we will use `<https://github.com/shirou/gopsutil>`_
which can be gotten via

.. code-block:: console

    $ go get github.com/shirou/gopsutil/mem


Using that package, we will define the read functionality for the ``memory`` device. Note that because
this tutorial is simple, we are putting everything in one file, but this is not required and is
discouraged for plugins that do anything beyond serve as an example. See the SDK repo's ``examples``
directory or the emulator plugin for examples of how to structure plugins.



Device Handler
~~~~~~~~~~~~~~
Next we'll define the read-write handler for our device. We won't do any writing for the device, so
its more of a read handler in this case. To read the memory info, we can use
`<https://github.com/shirou/gopsutil>`_ which can be gotten via

.. code-block:: console

    $ go get github.com/shirou/gopsutil/mem


We can use that package to define our read functionality for the ``memory`` device. Note that because
this tutorial is simple, we are putting everything in one file, but this is not required and is
discouraged for plugins that do anything beyond serve as an example. See the SDK repo's ``examples``
directory or the emulator plugin for examples of how to structure plugins.

.. code-block:: go

    var memoryHandler = sdk.DeviceHandler{
    	Name: "memory",
        Read: func(device *sdk.Device) ([]*sdk.Reading, error) {
    		v, err := mem.VirtualMemory()
    		if err != nil {
    			return nil, err
    		}
    		return []*sdk.Reading{
    			device.GetOutput("memory.total").MakeReading(v.Total),
    			device.GetOutput("memory.free").MakeReading(v.Free),
    			device.GetOutput("percent_used").MakeReading(v.UsedPercent),
    		}, nil
    	},
    }

Now we have our configuration defined and our handler defined. Next, we put together
the plugin, configure it, and register the handlers.


5. Create and configure the plugin
----------------------------------
The creation, configuration, registration, and running of a plugin can all be done
within the ``main()`` function. In short, the things that need to happen are:

- register plugin metadata
- create the ``Plugin``
- register the output types
- register all handlers
- run the plugin

If that sounds simple -- that's because it should be!

All plugins have some metadata associated with them. At a minimum, all plugins
require a name, but should also have a maintainer and short description and can
have a VCS link as well. We will call the plugin "tutorial plugin" and will have
"vaporio" be the maintainer.

.. code-block:: go

    func main() {
    	// Set plugin metadata
    	sdk.SetPluginMeta(
    		"tutorial plugin",
    		"vaporio",
    		"a simple plugin that reads virtual memory - used as a tutorial",
    		"",
    	)

    	// Create the plugin
    	plugin := sdk.NewPlugin()

    	// Register output types
    	err := plugin.RegisterOutputTypes(
    		&memoryTotal,
    		&memoryFree,
    		&percentUsed,
    	)
    	if err != nil {
    		log.Fatal(err)
    	}

    	// Register the device handler
    	plugin.RegisterDeviceHandlers(
    		&memoryHandler,
    	)

        // Run the plugin.
        if err := plugin.Run(); err != nil {
            log.Fatal(err)
        }
    }


.. note:: There are more things that can be done during plugin setup, from registering
   pre-run/post-run actions, to modifying various behaviors, to adding health checks. For
   more on this, see the :ref:`advancedUsage` section.


6. Plugin Summary
-----------------
To summarize, we should now have a file structure that looks like:

.. code-block:: none

    ▼ tutorial-plugin
        ▼ config
            ▼ device
                mem.yml
        config.yml
        plugin.go


With the configuration files:

.. code-block:: yaml
    :caption: config.yml

    version: 1.0
    debug: false
    network:
      type: unix
      address: memory.sock
    settings:
      read:
        interval: 5s


.. code-block:: yaml
    :caption: config/device/mem.yml

    version: 1.0
    locations:
      - name: local
        rack:
          name: local
        board:
          name: host
    devices:
      - name: memory
        metadata:
          model: tutorial-mem
        outputs:
          - type: memory.total
          - type: memory.free
          - type: percent_used
        devices:
          - info: Virtual Memory Usage
            location: local
            data:
              id: 1


And the plugin source code:

.. code-block:: go
    :caption: plugin.go

    package main

    import (
        "log"

        "github.com/shirou/gopsutil/mem"

        "github.com/vapor-ware/synse-sdk/sdk"
    )

    var (
    	memoryTotal = sdk.OutputType{
    		Name: "memory.total",
    		Unit: sdk.Unit{
    			Name: "bytes",
    			Symbol: "B",
    		},
    	}

    	memoryFree = sdk.OutputType{
    		Name: "memory.free",
    		Unit: sdk.Unit{
    			Name: "bytes",
    			Symbol: "B",
    		},
    	}

    	percentUsed = sdk.OutputType{
    		Name: "percent_used",
    		Unit: sdk.Unit{
    			Name: "percent",
    			Symbol: "%",
    		},
    	}
    )

    var memoryHandler = sdk.DeviceHandler{
    	Name: "memory",
        Read: func(device *sdk.Device) ([]*sdk.Reading, error) {
    		v, err := mem.VirtualMemory()
    		if err != nil {
    			return nil, err
    		}
    		return []*sdk.Reading{
    			device.GetOutput("memory.total").MakeReading(v.Total),
    			device.GetOutput("memory.free").MakeReading(v.Free),
    			device.GetOutput("percent_used").MakeReading(v.UsedPercent),
    		}, nil
    	},
    }

    func main() {
    	// Set plugin metadata
    	sdk.SetPluginMeta(
    		"tutorial plugin",
    		"vaporio",
    		"a simple plugin that reads virtual memory - used as a tutorial",
    		"",
    	)

    	// Create the plugin
    	plugin := sdk.NewPlugin()

    	// Register output types
    	err := plugin.RegisterOutputTypes(
    		&memoryTotal,
    		&memoryFree,
    		&percentUsed,
    	)
    	if err != nil {
    		log.Fatal(err)
    	}

    	// Register the device handler
    	plugin.RegisterDeviceHandlers(
    		&memoryHandler,
    	)

        // Run the plugin.
        if err := plugin.Run(); err != nil {
            log.Fatal(err)
        }
    }



7. Build and run the plugin
---------------------------
Next we will build and run the plugin locally, without Synse Server in front of it. In order
to interface with the plugin, we'll use the `Synse CLI <https://github.com/vapor-ware/synse-cli>`_.

From within the ``tutorial-plugin`` directory,

.. code-block:: console

    $ go build -o plugin


Congratulations, the plugin is now built! Now we can run it

.. code-block:: console

    $ ./plugin

You should see a single registered ``memory`` device and no errors. To interact
with the plugin, we can use the CLI.

.. warning:: The CLI may not be fully updated for SDK 1.0 yet, so not all of the
   CLI commands below may work. These docs will be updated once the CLI is updated.

Getting the plugin device info

.. code-block:: console

    $ synse plugin -u /tmp/synse/procs/memory.sock meta
    ID                                 TYPE      MODEL          PROTOCOL   RACK      BOARD
    65f660ac428556804060c13349e500de   memory    tutorial-mem   os         local     host


Getting a reading from the device

.. code-block:: console

    $ synse plugin -u /tmp/synse/procs/memory.sock read local host 65f660ac428556804060c13349e500de
    TYPE           VALUE               TIMESTAMP
    total          8589934592          Thu Apr 19 11:19:36 EDT 2018
    free           324714496           Thu Apr 19 11:19:36 EDT 2018
    percent_used   73.24576377868652   Thu Apr 19 11:19:36 EDT 2018


The device doesn't support writes, so writing should fail

.. code-block:: console

    $ synse plugin -u /tmp/synse/procs/memory.sock write local host 65f660ac428556804060c13349e500de total 123
    rpc error: code = Unknown desc = writing not enabled for device local-host-65f660ac428556804060c13349e500de (no write handler)


Now, you've configured, created, and run a plugin. The only thing left to do is
connect it with Synse Server and access the data it provides via Synse Server's
HTTP API.


8. Using with Synse Server
--------------------------
In this section, we'll go over how to deploy a plugin with Synse Server. While there are a few
ways of doing it, the recommended way is to run the plugin as a container and link it to the
Synse Server container. This means the plugin will be getting memory info from the container, not
the host machine, but this section just serves as an example of how to do it.

The first thing we will need to do is containerize the plugin. For this, we can write a Dockerfile.
For our Dockerfile, we'll assume that the binary was built locally, but examples exist in other repos
of how to use docker build stages to containerize the build process as well.

It is also important to note that all configs can be included in the Dockerfile with the plugin,
but it is best practice to not do this. The prototype configs can be included, since they should
not change based on the deployment, but the instance and plugin configs may change, so they should
be provided at runtime.

First, we'll make sure we have our plugin build locally. We will use the alpine linux base image,
so we want to build it for linux. If you are running on linux, this can be done simply with

.. code-block:: console

    $ go build -o plugin

If running on a non linux/amd64 architecture, e.g. Darwin, you will need to cross-compile

.. code-block:: console

    $ GOOS=linux GOARCH=amd64 go build -o plugin

Now, we can write our Dockerfile. While the configs can be built-in, we will not
do so here, since it is good practice to provide the configs at runtime for that
particular deployment.

.. code-block:: dockerfile
    :caption: Dockerfile

    FROM alpine

    COPY plugin plugin

    CMD ["./plugin"]


We can build the image as ``vaporio/tutorial-plugin``

.. code-block:: console

    $ docker build -t vaporio/tutorial-plugin .


Before we run the image, we'll want to update the plugin configuration that we will use.
Instead of using unix sockets for networking, we'll use TCP over port 5001. Change
``config.yml`` to:

.. code-block:: yaml

    version: 1.0
    name: memory
    debug: false
    network:
      type: tcp
      address: ":5001"
    settings:
      read:
        interval: 5s


Running via Docker
~~~~~~~~~~~~~~~~~~

Now we can run the plugin, supplying the plugin and instance configurations. We will also need
to specify environment variables so the plugin knows where to look for these configurations.

.. code-block:: console

    $ docker run -d \
        -p 5001:5001 \
        --name=tutorial-plugin \
        -v $PWD/config/device:/etc/synse/plugin/config/device \
        -v $PWD/config.yml:/tmp/config.yml \
        -e PLUGIN_CONFIG=/tmp \
        vaporio/tutorial-plugin


The plugin should now be running and waiting. You can check ``docker logs tutorial-plugin``
to view the logs and make sure everything is running correctly.

To connect it to Synse Server, you'll need the Synse Server image. The easiest way is to
just pull it from DockerHub:

.. code-block:: console

    $ docker pull vaporio/synse-server

We'll also need to create a network to link them across.

.. code-block:: console

    $ docker network create synse
    $ docker network connect synse tutorial-plugin


We'll now run Synse Server and connect it to the network. Here, we register the tutorial
plugin with Synse Server by using its environment configuration.

.. code-block:: console

    $ docker run -d \
        --name=synse-server \
        --network=synse \
        -p 5000:5000 \
        -e SYNSE_PLUGIN_TCP=tutorial-plugin:5001 \
        vaporio/synse-server


Now, you should be ready to use Synse Server to interact with the plugin. See the
:ref:`interactingViaSynseServer` section, below.


Running via Docker Compose
~~~~~~~~~~~~~~~~~~~~~~~~~~
All of the above can be done somewhat simpler via docker compose, using a compose file

.. code-block:: yaml
    :caption: tutorial.yml

    version: "3"
    services:
      synse-server:
        container_name: synse-server
        image: vaporio/synse-server
        ports:
          - 5000:5000
        environment:
          SYNSE_PLUGIN_TCP: tutorial-plugin:5001
        links:
          - tutorial-plugin

      tutorial-plugin:
        container_name: tutorial-plugin
        image: vaporio/tutorial-plugin
        ports:
          - 5001:5001
        volumes:
          - ./config/device:/etc/synse/plugin/config/device
          - ./config.yml:/tmp/config.yml
        environment:
          PLUGIN_CONFIG: /tmp


Then, just bring up the compose file

.. code-block:: console

    $ docker-compose -f tutorial.yml up -d


You should now be ready to use Synse Server to interact with the plugin. See the next
section for how to do so.


.. _interactingViaSynseServer:

Interacting via Synse Server
~~~~~~~~~~~~~~~~~~~~~~~~~~~~
With Synse Server now running locally, we can interact with its HTTP API using ``curl``.

- Check that the server is up and ready

.. code-block:: console

    $curl localhost:5000/synse/test
    {
      "status":"ok",
      "timestamp":"2018-04-19T16:56:16.085286Z"
    }


- Get ``scan`` information (e.g., see which devices are available). We should expect
  to see the single memory device managed by the plugin.

.. code-block:: console

    $ curl localhost:5000/synse/2.1/scan
    {
      "racks":[
        {
          "id":"local",
          "boards":[
            {
              "id":"host",
              "devices":[
                {
                  "id":"baeb1223219e634446c4af115be089e7",
                  "info":"Virtual Memory Usage",
                  "type":"memory"
                }
              ]
            }
          ]
        }
      ]
    }


- We can ``read`` from that device, and we should expect to get back the total, free, and
  percent_used readings from the memory device.

.. code-block:: console

    $ curl localhost:5000/synse/2.1/read/local/host/baeb1223219e634446c4af115be089e7
    {
      "kind":"memory",
      "data":{
        "total":{
          "value":2096058368,
          "timestamp":"2018-06-19T13:28:31.0881264Z",
          "unit":{
            "symbol":"B",
            "name":"bytes"
          },
          "type":"total",
          "info":""
        },
        "free":{
          "value":211611648,
          "timestamp":"2018-06-19T13:28:31.0881454Z",
          "unit":{
            "symbol":"B",
            "name":"bytes"
          },
          "type":"free",
          "info":""
        },
        "percent_used":{
          "value":69.7154570841,
          "timestamp":"2018-06-19T13:28:31.0881577Z",
          "unit":{
            "symbol":"%",
            "name":"percent"
          },
          "type":"percent_used",
          "info":""
        }
      }
    }


Now, you have successfully created, configured, and ran a Synse Plugin both on its own
and as part of a deployment with Synse Server. Explore the
`Synse Server API <https://vapor-ware.github.io/synse-server/>`_ to see what
else you can do with it.
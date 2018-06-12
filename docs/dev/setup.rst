.. _setup:

Developer Setup
===============
This section goes into detail on how to get set up to develop the SDK as well
as various development workflow steps that we use here at Vapor IO.

Getting Started
---------------
When first getting started with developing the SDK, you will first need to have `Go <https://golang.org/doc/install>`_
(version 1.9+) installed. To check which version you have, e.g.,

.. code-block:: console

    $ go version
    go version go1.9.1 darwin/amd64


Then, you will need to get the SDK source either by checking out the repo via git,

.. code-block:: console

    $ git clone https://github.com/vapor-ware/synse-sdk.git
    $ cd synse-sdk

Or via ``go get``

.. code-block:: console

    $ go get -u github.com/vapor-ware/synse-sdk/sdk
    $ cd $GOPATH/src/github.com/vapor-ware/synse-sdk

Now, you should be ready to start developing on the SDK.


Workflow
--------
To aid in the developer workflow, Makefile targets are provided for common development
tasks. To see what targets are provided, see the project ``Makefile``, or run ``make help``
out of the project repo root.

.. code-block:: console

    $ make help
    build           Build the SDK locally
    check-examples  Check that the examples run without failing.
    ci              Run CI checks locally (build, test, lint)
    clean           Remove temporary files
    cover           Run tests and open the coverage report
    dep             Ensure and prune dependencies
    dep-update      Ensure, update, and prune dependencies
    docs            Build the docs locally
    examples        Build the examples
    fmt             Run goimports on all go files
    github-tag      Create and push a tag with the current version
    godoc           Run godoc to get a local version of docs on port 8080
    help            Print usage information
    lint            Lint project source files
    setup           Install the build and development dependencies
    test            Run all tests
    version         Print the version of the SDK


In general when developing, tests should be run (e.g. ``make test``) and the could should
be formatted (``make fmt``) and linted (``make lint``). This ensures that the code works
and is consistent and readable. Tests should also be added or updated as appropriate
(see the :ref:`testing` section).


CI
--
All commits and pull requests to the Synse Plugin SDK trigger a build in `Circle CI <https://circleci.com/gh/vapor-ware/synse-sdk>`_.
The CI configuration can be found in the repo's ``.circleci/config.yml`` file. In summary,
a build triggered by a commit will:

- Install dependencies
- Run linting
- Check formatting
- Run tests with coverage reporting (and upload results to CodeCov)
- Build the example plugins in the ``examples`` directory

When a tag is pushed to the repo, CI checks that the tag version matches the SDK version
specified in the repo, then generates a changelog and drafts a new release for that version.

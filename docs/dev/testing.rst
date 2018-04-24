.. _testing:

Testing
=======
The Synse Plugin SDK strives to follow the `Golang testing <https://golang.org/pkg/testing/>`_
best practices. Tests for each file are found in the same directory following the pattern
``FILENAME_test.go``, so given a file named ``plugin.go``, the test file would be ``plugin_test.go``.


Writing Tests
-------------
There are many `articles <https://blog.alexellis.io/golang-writing-unit-tests/>`_ and tutorials
out there on how to write unit tests for Golang. In general, this repository tries to follow them
as best as possible and also tries to be consistent with how tests are written. This makes
them easier to read and maintain. When writing new tests, use the existing ones as a guide.

Whenever additions or changes are made to the code base, there should be tests that cover
them. Many unit tests already exists, so some changes may not require tests to be added.
To help ensure that the SDK is well-tested, we upload coverage reports to
`CodeCov <https://codecov.io/gh/vapor-ware/synse-sdk>`_. While good code coverage does not
ensure bug-free code, it can still be a useful indicator.


Running Tests
-------------
Tests can be run with ``go test``, e.g.

.. code-block:: console

    $ go test ./sdk/...

For convenience, there is a make target to do this

.. code-block:: console

    $ make test

While the above make target will report coverage at a high level, it can be useful to
see a detailed coverage report that shows which lines were hit and which were missed.
For that, you can use the make target

.. code-block:: console

    make cover

This will run tests and collect and join coverage reports for all packages/sub-packages
and output them as an HTML page.
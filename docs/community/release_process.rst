.. _release_process:

Release Process
===============

The following guidelines describe the criteria for new releases. The Synse
Plugin SDK is versioned with the format ``major.minor.micro``.


Major Version
-------------

A major release will include breaking changes. When a new major release
is cut, it will be versioned as ``X.0.0``. For example, if the previous
release version was ``1.4.2``, the next version would be ``2.0.0``.

Breaking changes are changes which break backwards compatibility with previous
versions. Typically, this would mean changes to the API. Major releases may also
include bug fixes.


Minor Version
-------------

A minor release will not include breaking changes to the API, but may
otherwise include additions, updates, or bug fixes. If the previous release
version was ``1.4.2``, the next minor release would be ``1.5.0``.

Minor version releases are backwards compatible with releases of the same major
version number.


Micro Version
-------------

A micro release will not include any breaking changes and will typically only
include minor changes or bug fixes that were missed with the previous minor
version release. If the previous release version was ``1.4.2``, the next micro
release would be ``1.4.3``.

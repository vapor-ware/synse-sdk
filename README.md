### Synse Backgound Process SDK


This is a simple SDK to aid in writing plugins for Synse. It handles a lot
of common functionality and requires that the plugin only write handlers for
plugin-specific actions, such as read and write. 

See the `example` directory for examples on how to use the SDK.


TODO:
 - proper organization
 - figure out how writes will work
    - starting from the grpc command
    - how they get added to the rwloop queue
    - how a write transaction is generated/tracked
    - etc
 - figure out how transaction checks will work
    - this is largely tied to the work above for write
 - proper logging
 - plugin configuration (not prototype/device config) - e.g. debug mode, buffer size, etc?
 - packaging
 - documentation
 - testing
 - figure out error cases / error handling
 - figure out how to properly get the generated protobuf stuff in here (+ also in the python
   synse repo) in a clean and easy way that doesn't require lots of copy + paste, etc. perhaps
   git submodules? may not be too bad this time around since there isn't tons of nesting and
   we can pin it to a release version? or just in its own repo and it can be imported that way?
 - figure out: is there a way to give devices a clearer human-readable name?
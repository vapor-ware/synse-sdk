#!/usr/bin/env bash

#
# Run the example plugins included in the repository.
#
# This script is used by CI as a means to easily test that all the example
# plugins (dry)run as expected. This requires the plugins to be built first.
#
# This should be run from the project root directory.
#

for d in examples/*/ ; do \
    echo "\n\033[32m$d\033[0m" ; \
    cd $d ; \
    if [ ! -f "plugin" ]; then echo "\033[31mplugin binary not found\033[0m"; fi; \
    if ! ./plugin --dry-run; then exit 1; fi; \
    cd ../.. ; \
done

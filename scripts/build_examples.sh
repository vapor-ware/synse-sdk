#!/usr/bin/env bash

#
# Build the examples included in the repository.
#
# This script is used by CI as a means to easily test that all the example
# plugins can be successfully built.
#
# This should be run from the project root directory.
#

for d in examples/*/ ; do \
    echo "\n\033[32m$d\033[0m" ; \
    cd $d ; \
    go build -v -o plugin ; \
    cd ../.. ; \
done

#!/usr/bin/env bash

#
# coverage.sh
#
# create a unified coverage report for all packages
#

set -e
echo "mode: atomic" > coverage.txt

for d in $(go list ./... | grep -v vendor | grep -v examples); do
    go test -race -coverprofile=profile.out -covermode=atomic ${d}
    if [ -f profile.out ]; then
        cat profile.out | awk '{if(NR>1)print}' >> coverage.txt
        rm profile.out
    fi
done
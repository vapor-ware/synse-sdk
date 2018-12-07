#!/usr/bin/env bash

sdk_version="${SDK_VERSION}"
ci_tag="${TAG_NAME}"

if [ ! "${sdk_version}" ] && [ ! "${ci_tag}" ]; then
    echo "No version or tag specified."
    exit 1
fi

if [ "${sdk_version}" != "${ci_tag}" ]; then
    echo "Versions do not match: sdk@${sdk_version} tag@${ci_tag}"
    exit 1
fi

echo "Versions match: sdk@${sdk_version} tag@${ci_tag}"
#!/usr/bin/env bash

sdk_version="${SDK_VERSION}"
circle_tag="${CIRCLE_TAG}"

if [ ! "${sdk_version}" ] && [ ! "${circle_tag}" ]; then
    echo "No version or tag specified."
    exit 1
fi

if [ "${sdk_version}" != "${circle_tag}" ]; then
    echo "Versions do not match: sdk@${sdk_version} tag@${circle_tag}"
    exit 1
fi

echo "Versions match: sdk@${sdk_version} tag@${circle_tag}"
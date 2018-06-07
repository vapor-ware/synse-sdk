#
# Dynamic Device Plugin Example
#

PLUGIN_VERSION := 1.0

GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2> /dev/null || true)
GIT_TAG    ?= $(shell git describe --tags 2> /dev/null || true)
BUILD_DATE := $(shell date -u +%Y-%m-%dT%T 2> /dev/null)
GO_VERSION := $(shell go version | awk '{ print $$3 }')

PKG_CTX := github.com/vapor-ware/synse-sdk/sdk
LDFLAGS := -w \
	-X ${PKG_CTX}.BuildDate=${BUILD_DATE} \
	-X ${PKG_CTX}.GitCommit=${GIT_COMMIT} \
	-X ${PKG_CTX}.GitTag=${GIT_TAG} \
	-X ${PKG_CTX}.GoVersion=${GO_VERSION} \
	-X ${PKG_CTX}.PluginVersion=${PLUGIN_VERSION}


build:
	@go build -ldflags "${LDFLAGS}" -o plugin


.PHONY: build
.DEFAULT_GOAL := build
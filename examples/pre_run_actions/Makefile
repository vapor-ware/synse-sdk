PLUGIN_VERSION=1.0

PKG_CTX=main
BUILD_DATE := $(shell date -u +%Y-%m-%dT%T 2> /dev/null)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2> /dev/null || true)
GIT_TAG := $(shell git describe --tags 2> /dev/null || true)
GO_VERSION := $(shell go version | awk '{ print $$3 }')

LDFLAGS="-X ${PKG_CTX}.BuildDate=${BUILD_DATE} \
		 -X ${PKG_CTX}.GitCommit=${GIT_COMMIT} \
		 -X ${PKG_CTX}.GitTag=${GIT_TAG} \
		 -X ${PKG_CTX}.GoVersion=${GO_VERSION} \
         -X ${PKG_CTX}.VersionString=${PLUGIN_VERSION}"

build:
	@go build -ldflags ${LDFLAGS} -o plugin


.PHONY: build
.DEFAULT_GOAL := build
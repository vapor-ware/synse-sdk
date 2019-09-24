#
# Synse Plugin SDK
#

SDK_VERSION := $(shell cat sdk/version.go | grep 'const Version' | awk '{print $$4}')

HAS_LINT := $(shell which golangci-lint)

.PHONY: build
build:  ## Build the SDK locally
	go build -v ./sdk || exit

.PHONY: ci
ci:  ## Run CI checks locally (build, test, lint)
	@$(MAKE) build test lint

.PHONY: clean
clean:  ## Remove temporary files
	go clean -v || exit

.PHONY: cover
cover: test  ## Run tests and open the coverage report
	go tool cover -html=coverage.out

.PHONY: docs
docs:  ## Build the docs locally
	(cd docs ; make html)

.PHONY: check-examples
check-examples:  ## Check that the examples run without failing.
	@for d in examples/*/ ; do \
		echo "\n\033[32m$$d\033[0m" ; \
		cd $$d ; \
		if [ ! -f "plugin" ]; then echo "\033[31mplugin binary not found\033[0m"; fi; \
		if ! ./plugin --dry-run; then exit 1; fi; \
		cd ../.. ; \
	done


.PHONY: examples
examples:  ## Build the examples
	@for d in examples/*/ ; do \
		echo "\n\033[32m$$d\033[0m" ; \
		cd $$d ; \
		go build -v -o plugin || exit ; \
		cd ../.. ; \
	done

.PHONY: fmt
fmt:  ## Run goimports on all go files
	find . -name '*.go' -not -wholename './vendor/*' | while read -r file; do goimports -w "$$file" || exit; done

.PHONY: github-tag
github-tag:  ## Create and push a tag with the current version
	git tag -a ${SDK_VERSION} -m "Synse SDK version ${SDK_VERSION}"
	git push -u origin ${SDK_VERSION}

.PHONY: godoc
godoc:  ## Run godoc to get a local version of docs on port 8080
	open http://localhost:8080/pkg/github.com/vapor-ware/synse-sdk/sdk/
	godoc -http ":8080"

.PHONY: lint
lint:  ## Lint project source files
ifndef HAS_LINT
	$(shell curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b $$(go env GOPATH)/bin v1.18.0)
endif
	golangci-lint run --deadline=5m

.PHONY: setup
setup:  ## Install the build and development dependencies
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/alecthomas/gometalinter
	go get -u golang.org/x/tools/cmd/cover
	gometalinter --install
	@$(MAKE) dep

.PHONY: test
test:  ## Run all tests
	@ # Note: this requires go1.10+ in order to do multi-package coverage reports
	go test -race -coverprofile=coverage.out -covermode=atomic ./...  || exit

.PHONY: version
version: ## Print the version of the SDK
	@echo "$(SDK_VERSION)"

.PHONY: help
help:  ## Print usage information
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

.DEFAULT_GOAL := help


#
# CI Targets
#

.PHONY: ci-check-version
ci-check-version:
	SDK_VERSION=$(SDK_VERSION) ./bin/ci/check_version.sh

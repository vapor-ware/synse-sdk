#
# Synse SDK
#

SDK_VERSION := $(shell cat sdk/version.go | grep 'const Version' | awk '{print $$4}')


.PHONY: clean
clean:  ## Remove temporary files
	go clean -v
	rm -f coverage.out

.PHONY: cover
cover: test  ## Run tests and open the coverage report
	go tool cover -html=coverage.out

.PHONY: check-examples
check-examples:  ## Check that the example plugins run without failing
	@./scripts/run_examples.sh

.PHONY: examples
examples:  ## Build the example plugins
	@./scripts/build_examples.sh

.PHONY: fmt
fmt:  ## Run goimports on all go files
	find . -name '*.go' -not -wholename './vendor/*' | while read -r file; do goimports -w "$$file"; done

.PHONY: github-tag
github-tag:  ## Create and push a tag with the current version
	git tag -a ${SDK_VERSION} -m "Synse SDK version ${SDK_VERSION}"
	git push -u origin ${SDK_VERSION}

.PHONY: godoc
godoc:  ## Server godocs locally on port 8080
	open http://localhost:8080/pkg/github.com/vapor-ware/synse-sdk/sdk/
	godoc -http ":8080"

.PHONY: lint
lint:  ## Lint project source files
	staticcheck ./sdk/... ./examples/...
	go vet ./sdk/... ./examples/...

.PHONY: test
test:  ## Run all tests
	@ # Note: this requires go1.10+ in order to do multi-package coverage reports
	go test -race -coverprofile=coverage.out -covermode=atomic ./sdk/...

.PHONY: version
version:  ## Print the version of the SDK
	@echo "${SDK_VERSION}"

.PHONY: help
help:  ## Print usage information
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

.DEFAULT_GOAL := help


# Targets for CI

.PHONY: unit-test integration-test

unit-test: test

integration-test: examples check-examples

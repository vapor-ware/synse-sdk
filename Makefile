
SDK_VERSION := $(shell cat sdk/version.go | grep 'const Version' | awk '{print $$4}')


.PHONY: build
build:  ## Build the SDK locally
	go build -v ./sdk

.PHONY: ci
ci:  ## Run CI checks locally (build, test, lint)
	@$(MAKE) build test lint

.PHONY: clean
clean:  ## Remove temporary files
	go clean -v

.PHONY: cover
cover:  ## Run tests and open the coverage report
	go test -covermode=atomic -coverprofile=coverage.txt -v -race -timeout=30s ./sdk
	go tool cover -html=coverage.txt
	rm coverage.txt

.PHONY: dep
dep:  ## Ensure and prune dependencies
	dep ensure -v
	dep prune

.PHONY: examples
examples:  ## Build the examples
	@for d in examples/*/ ; do \
		echo "\n\033[32m$$d\033[0m" ; \
		cd $$d ; \
		go build -v -o plugin ; \
		cd ../.. ; \
	done

.PHONY: fmt
fmt:  ## Run goimports on all go files
	find . -name '*.go' -not -wholename './vendor/*' | while read -r file; do goimports -w "$$file"; done

.PHONY: lint
lint:  ## Lint project source files
	gometalinter ./... --vendor --tests --deadline=5m \
		--disable=gas --disable=errcheck

.PHONY: setup
setup:  ## Install the build and development dependencies
	go get -u github.com/alecthomas/gometalinter
	go get -u golang.org/x/tools/cmd/cover
	go get -u github.com/golang/dep/cmd/dep
	gometalinter --install --update
	@$(MAKE) dep

.PHONY: test
test:  ## Run all tests
	go test -cover -v ./sdk

.PHONY: version
version: ## Print the version of the SDK
	@echo "$(SDK_VERSION)"

.PHONY: help
help:  ## Print usage information
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

.DEFAULT_GOAL := help

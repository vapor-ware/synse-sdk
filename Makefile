
HAS_LINT := $(shell command -v golint)


lint:  ## Lint the project source files
ifndef HAS_LINT
	@go get -u github.com/golang/lint/golint
endif
	@golint -set_exit_status sdk/... client/... examples/... && echo "ok"


test:  ## Run the SDK tests
	go test -v ./sdk


help:  ## Print usage information
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort


.PHONY: help lint test
.DEFAULT_GOAL := help

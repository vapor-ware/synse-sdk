
HAS_LINT := $(shell command -v golint)


lint:
ifndef HAS_LINT
	@go get -u github.com/golang/lint/golint
endif
	@golint -set_exit_status sdk/... client/... examples/...


.PHONY: lint
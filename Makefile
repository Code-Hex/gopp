PROJECT = github.com/Code-Hex/gopp

# https://docs.codecov.io/docs/flags
CODECOV_CLI_FLAG=
ifneq ($(CI),)
	CODECOV_CLI_FLAG+=-Z
	ifneq ($(CODECOV_PULL_NUM),)
	CODECOV_CLI_FLAG+=-P ${CODECOV_PULL_NUM}
	endif
endif

.PHONY: vet
vet:
	@go vet ./...

.PHONY: lint
lint:
	golint ./...

.PHONY: test
test:
	@echo "+ $@"
	@go test -count=1 -timeout 60s -v -race ./...

.PHONY: coverage
coverage:
	@echo "+ $@"
	@go test -count=1 -timeout 60s -v -race -covermode=atomic -coverpkg=./... -coverprofile=coverage.txt ./...

.PHONY: codecov
codecov: SHELL=/usr/bin/env bash
codecov:
	@echo "+ $@"
	bash <(curl -s https://codecov.io/bash) ${CODECOV_CLI_FLAG}

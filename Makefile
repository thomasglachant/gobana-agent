APPNAME  = spooter
PACKAGE  = github.com/thomasglachant/spooter
DATE    ?= $(shell date +%FT%T%z)
VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || \
			cat $(CURDIR)/.version 2> /dev/null || echo v0)
BIN      = $(CURDIR)/bin
GOPATH   = $(CURDIR)/.gopath~
BASE     = $(CURDIR)
PKGS     = $(or $(PKG),$(shell cd $(BASE) && env GOPATH=$(GOPATH) $(GO) list ./... | grep -v "^$(PACKAGE)/vendor/"))

GO      = GO111MODULE=on go
GOFMT   = $(shell go env GOPATH)/bin/gofumpt
GOLINT  = $(shell go env GOPATH)/bin/golangci-lint
TIMEOUT = 15
V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell printf "\033[34;1m▶\033[0m")

.DEFAULT_GOAL := help

.PHONY: build
build: $(BASE) ; $(info $(M) building executable…) @ ## Build program binary (without checking lint and format)
	$Q cd $(BASE) && $(GO) build \
		-tags release \
		-ldflags "-X main.Version=$(VERSION) -X main.BuildDate=$(DATE)" \
		-o $(BIN)/$(APPNAME) *.go

$(GOLINT): | $(BASE) ; $(info $(M) building lint…)
	$Q GOPATH=$(shell go env GOPATH) go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: lint
lint: $(BASE) $(GOLINT) ; $(info $(M) running lint…) @ ## Run lint
	$Q cd $(BASE) && $(GOLINT) run --color auto --fix

$(GOFMT): | $(BASE) ; $(info $(M) building fmt…)
	$Q GOPATH=$(shell go env GOPATH) go install mvdan.cc/gofumpt@latest

.PHONY: fmt
fmt: $(BASE) $(GOFMT) ; $(info $(M) running gofmt…) @ ## Run gofmt on all source files
	@ret=0 && for d in $$($(GO) list -f '{{.Dir}}' ./... | grep -v /vendor/); do \
		$(GOFMT) -l -w $$d/*.go || ret=$$? ; \
	 done ; exit $$ret

.PHONY: start-agent
start-agent: build $(BASE) ; $(info $(M) launch app...) @ ## Launch application
	@$(BIN)/$(APPNAME) $(RUN_ARGS) -agent -config=$(config)

.PHONY: clean
clean: ; $(info $(M) cleaning…)	@ ## Cleanup everything
	@rm -rf $(GOPATH)
	@find bin ! -name '.gitkeep' -type f -exec rm -f {} +
	@rm -rf test/tests.* test/coverage.*

.PHONY: test
test: fmt lint build test-unit ## Run tests

.PHONY: test-unit
test-unit:  ## Run tests
	$(info $(M) run tests…)
	$(GO) test -v ./...

.PHONY: help
help:
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: version
version:
	@echo $(VERSION)

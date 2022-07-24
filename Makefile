APPNAME_AGENT  = spooter-agent
APPNAME_SERVER  = spooter-server
PACKAGE  = github.com/thomasglachant/spooter
DATE    ?= $(shell date +%FT%T%z)
VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || \
			cat $(CURDIR)/.version 2> /dev/null || echo v0)
BIN      = $(CURDIR)/bin
GOPATH   = $(CURDIR)/.gopath~
BASE     = $(CURDIR)

LIST_PKGS= ./agent/... ./core/... ./server/...
PKGS     = agent core server

GO      = GO111MODULE=on go
GOFMT   = $(shell go env GOPATH)/bin/gofumpt
GOLINT  = $(shell go env GOPATH)/bin/golangci-lint
TIMEOUT = 15
V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell printf "\033[34;1m▶\033[0m")
Y = $(shell printf "\033[33;1m▶\033[0m")

.DEFAULT_GOAL := help

.PHONY: build
build: build-agent build-server # Build all

.PHONY: build-agent
build-agent: $(BASE) ; $(info $(M) building agent executable…) @ ## Build program binary (without checking lint and format)
	$Q cd $(BASE) && $(GO) build \
		-tags release \
		-ldflags "-X main.Version=$(VERSION) -X main.BuildDate=$(DATE)" \
		-o $(BIN)/$(APPNAME_AGENT) $(PACKAGE)/agent

.PHONY: start-agent
start-agent: build-agent $(BASE) ; $(info $(M) launch agent...) @ ## Launch application
	@$(BIN)/$(APPNAME_AGENT) $(RUN_ARGS) -config=$(config)

.PHONY: build-server
build-server: $(BASE) ; $(info $(M) building server executable…) @ ## Build program binary (without checking lint and format)
	$Q cd $(BASE) && $(GO) build \
		-tags release \
		-ldflags "-X main.Version=$(VERSION) -X main.BuildDate=$(DATE)" \
		-o $(BIN)/$(APPNAME_SERVER) $(PACKAGE)/server

.PHONY: start-server
start-server: build-server $(BASE) ; $(info $(M) launch server...) @ ## Launch application
	@$(BIN)/$(APPNAME_SERVER) $(RUN_ARGS) -config=$(config)

$(GOLINT): | $(BASE) ; $(info $(M) building lint…)
	$Q GOPATH=$(shell go env GOPATH) go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: lint
lint: lint-core lint-agent lint-server ## Run lint

.PHONY: lint-core
lint-core: $(BASE) $(GOLINT) ; $(info $(M) lint module core…) @ ## Run lint on core
	$Q cd $(BASE)/core && $(GOLINT) run --color auto --fix

.PHONY: lint-agent
lint-agent: $(BASE) $(GOLINT) ; $(info $(M) lint module agent…) @ ## Run lint on agent
	$Q cd $(BASE)/agent && $(GOLINT) run --color auto --fix

.PHONY: lint-server
lint-server: $(BASE) $(GOLINT) ; $(info $(M) lint module server…) @ ## Run lint on server
	$Q cd $(BASE)/server && $(GOLINT) run --color auto --fix

$(GOFMT): | $(BASE) ; $(info $(M) building fmt…)
	$Q GOPATH=$(shell go env GOPATH) go install mvdan.cc/gofumpt@latest

.PHONY: fmt
fmt: $(BASE) $(GOFMT) ; $(info $(M) running gofmt…) @ ## Run gofmt on all source files
	@ret=0 && for d in $$($(GO) list -f '{{.Dir}}' $(LIST_PKGS) | grep -v /vendor/); do \
		$(GOFMT) -l -w $$d/*.go || ret=$$? ; \
	 done ; exit $$ret

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
	$(GO) test -v $(LIST_PKGS)

.PHONY: help
help:
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

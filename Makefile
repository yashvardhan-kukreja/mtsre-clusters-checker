# These shell flags are REQUIRED for an early exit in case any program called by make errors!
.SHELLFLAGS=-euo pipefail -c
SHELL := /bin/bash

.PHONY: all fmt clean check build tidy docker-build docker-push goimports golangci-lint

REPO := quay.io/mtsre/mtsre-clusters-checker
TAG := $(shell git rev-parse --short HEAD)

# Set the GOBIN environment variable so that dependencies will be installed
# always in the same place, regardless of the value of GOPATH
CACHE := $(PWD)/.cache
export GOBIN := $(CACHE)/bin
export PATH := $(GOBIN):$(PATH)

all: build

clean: ## Clean this directory
	@rm -fr $(CACHE) $(GOBIN) bin/* dist/ || true

build: tidy ## Build binaries
	@CGO_ENABLED=0 go build -a -o bin/clusters-checker main.go

run: build
	@bin/mtsre-clusters-checker

tidy:
	@go mod tidy
	@go mod verify

fmt:
	@go fmt ./...

check: golangci-lint goimports

docker-build:
	@docker build -t $(REPO):$(TAG) .

docker-push:
	@if [ -z "$(DOCKER_CONF)" ]; then echo "Please set DOCKER_CONF. Exiting." && exit 1; fi
	@docker tag $(REPO):$(TAG) $(REPO):latest
	@docker --config=$(DOCKER_CONF) push $(REPO):$(TAG)
	@docker --config=$(DOCKER_CONF) push $(REPO):latest

GOIMPORTS := $(GOBIN)/goimports
goimports:
	@$(call go-get-tool,$(GOIMPORTS),golang.org/x/tools/cmd/goimports)
	@$(GOIMPORTS) -w -l $(shell find . -type f -name "*.go" -not -path "./vendor/*")

GOLANGCI_LINT := $(GOBIN)/golangci-lint
golangci-lint:
	@$(call go-get-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint@v1.43.0)
	@echo "Running golangci-lint..."
	@$(GOLANGCI_LINT) run --timeout=10m -E unused,gosimple,staticcheck --skip-dirs-use-default --verbose


# go-get-tool will 'go get' any package $2 and install it to $1.
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
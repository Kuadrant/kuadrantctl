SHELL := /bin/bash

MKFILE_PATH := $(abspath $(firstword $(MAKEFILE_LIST)))
PROJECT_PATH := $(patsubst %/,%,$(dir $(MKFILE_PATH)))
GO ?= go

all: help

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY : help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

include ./make/tools/*.mk

##@ Development

.PHONY : test
test: clean-cov fmt vet $(GINKGO) ## Run unit tests
	mkdir -p $(PROJECT_PATH)/coverage
	# Shuffle both the order in which specs within a suite run, and the order in which different suites run
	# You can always rerun a given ordering later by passing the --seed flag a matching seed.
	$(GINKGO) \
		--randomize-all \
		--randomize-suites \
		--coverpkg ./pkg/...,./cmd/... \
		--output-dir $(PROJECT_PATH)/coverage \
		--coverprofile cover.out \
		./pkg/... ./cmd/...


.PHONY : install
install: fmt vet ## Build and install kuadrantctl binary ($GOBIN or GOPATH/bin)
	@set -e; \
	GIT_SHA=$$(git rev-parse --short=7 HEAD 2>/dev/null) || { \
		GIT_HASH=$${GITHUB_SHA:-NO_SHA}; \
	}; \
	if [ -z "$$GIT_HASH" ]; then \
		GIT_DIRTY=$$(git diff --stat); \
		if [ -n "$$GIT_DIRTY" ]; then \
			GIT_HASH=$${GIT_SHA}-dirty; \
		else \
			GIT_HASH=$${GIT_SHA}; \
		fi; \
	fi; \
	LDFLAGS="-X 'github.com/kuadrant/kuadrantctl/version.GitHash=$$GIT_HASH'"; \
	GOBIN=$(PROJECT_PATH)/bin $(GO) install -ldflags "$$LDFLAGS";


.PHONY: prepare-local-cluster
prepare-local-cluster: $(KIND) ## Deploy locally kuadrant operator from the current code
	$(MAKE) kind-delete-cluster
	$(MAKE) kind-create-cluster

.PHONY: env-setup
env-setup:
	$(MAKE) gateway-api-install

.PHONY: local-setup
local-setup: ## Sets up Kind cluster with GatewayAPI manifests
	$(MAKE) prepare-local-cluster
	$(MAKE) env-setup

.PHONY: local-cleanup
local-cleanup: ## Delete local cluster
	$(MAKE) kind-delete-cluster

.PHONY : fmt
fmt: ## Run go fmt ./...
	$(GO) fmt ./...

.PHONY : vet
vet: ## Run go vet ./...
	$(GO) vet ./...

.PHONY: clean-cov
clean-cov: ## Remove coverage reports
	rm -rf $(PROJECT_PATH)/coverage

.PHONY: run-lint
run-lint: $(GOLANGCI-LINT) ## Run linter tool (golangci-lint)
	$(GOLANGCI-LINT) run --timeout 2m

# Include last to avoid changing MAKEFILE_LIST used above
include ./make/*.mk

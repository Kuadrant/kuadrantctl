SHELL := /bin/bash

MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_PATH := $(patsubst %/,%,$(dir $(MKFILE_PATH)))
GO ?= go

all: help

.PHONY : help
help: Makefile
	@sed -n 's/^##//p' $<

# Ginkgo tool
GINKGO = $(PROJECT_PATH)/bin/ginkgo
$(GINKGO):
	# In order to make sure the version of the ginkgo cli installed
	# is the same as the version of go.mod,
	# instead of calling go-install-tool,
	# running go install from the current module will pick version from current go.mod file.
	GOBIN=$(PROJECT_PATH)/bin go install github.com/onsi/ginkgo/v2/ginkgo

.PHONY: ginkgo
ginkgo: $(GINKGO) ## Download ginkgo locally if necessary.

KIND = $(PROJECT_PATH)/bin/kind
KIND_VERSION = v0.22.0
$(KIND):
	$(call go-install-tool,$(KIND),sigs.k8s.io/kind@$(KIND_VERSION))

.PHONY: kind
kind: $(KIND) ## Download kind locally if necessary.

KUSTOMIZE = $(PROJECT_PATH)/bin/kustomize
$(KUSTOMIZE):
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v4@v4.5.5)

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.

## test: Run unit tests
.PHONY : test
test: clean-cov fmt vet $(GINKGO)
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

## install: Build and install kuadrantctl binary ($GOBIN or GOPATH/bin)
.PHONY : install
install: fmt vet
	GOBIN=$(PROJECT_PATH)/bin $(GO) install

.PHONY: prepare-local-cluster
prepare-local-cluster: $(KIND) ## Deploy locally kuadrant operator from the current code
	$(MAKE) kind-delete-cluster
	$(MAKE) kind-create-cluster

.PHONY: env-setup
env-setup:
	$(MAKE) olm-install
	$(MAKE) gateway-api-install
	$(MAKE) istio-install
	$(MAKE) deploy-gateway

## local-setup: Sets up Kind cluster with GatewayAPI manifests and istio GW, nothing Kuadrant. Build and install kuadrantctl binary
.PHONY: local-setup
local-setup:
	$(MAKE) prepare-local-cluster
	$(MAKE) env-setup

## local-cleanup: Delete local cluster
.PHONY: local-cleanup
local-cleanup: ## Delete local cluster
	$(MAKE) kind-delete-cluster

.PHONY : fmt
fmt:
	$(GO) fmt ./...

.PHONY : vet
vet:
	$(GO) vet ./...

.PHONY: clean-cov
clean-cov: ## Remove coverage reports
	rm -rf $(PROJECT_PATH)/coverage

# Include last to avoid changing MAKEFILE_LIST used above
include ./make/*.mk

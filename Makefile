SHELL := /bin/bash

MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_PATH := $(patsubst %/,%,$(dir $(MKFILE_PATH)))
GO ?= go
KUADRANT_NAMESPACE=kuadrant-system

all: help

.PHONY : help
help: Makefile
	@sed -n 's/^##//p' $<

# Ginkgo tool
GINKGO = $(PROJECT_PATH)/bin/ginkgo
$(GINKGO):
	$(call go-install-tool,$(GINKGO),github.com/onsi/ginkgo/ginkgo@v1.16.4)

KIND = $(PROJECT_PATH)/bin/kind
KIND_VERSION = v0.20.0
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
test: fmt vet $(GINKGO)
	# huffle both the order in which specs within a suite run, and the order in which different suites run
	# You can always rerun a given ordering later by passing the --seed flag a matching seed.
	$(GINKGO) --randomizeAllSpecs --randomizeSuites -v -progress --trace --cover ./...

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
	$(MAKE) cert-manager-install
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

# Include last to avoid changing MAKEFILE_LIST used above
include ./make/*.mk

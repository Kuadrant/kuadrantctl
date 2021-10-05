SHELL := /bin/bash

MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_PATH := $(patsubst %/,%,$(dir $(MKFILE_PATH)))
GO ?= go
KUADRANT_NAMESPACE=kuadrant-system

include utils.mk

all: help

.PHONY : help
help: Makefile
	@sed -n 's/^##//p' $<

# Kind tool
KIND = $(PROJECT_PATH)/bin/kind
KIND_CLUSTER_NAME = kuadrant-local
$(KIND):
	$(call go-get-tool,$(KIND),sigs.k8s.io/kind@v0.11.1)

.PHONY : kind
kind: $(KIND)

# istioctl tool
ISTIOCTL = $(PROJECT_PATH)/bin/istioctl
ISTIOCTLVERSION = 1.9.4
istioctl:
ifeq (,$(wildcard $(ISTIOCTL)))
	@{ \
	set -e ;\
	mkdir -p $(dir $(ISTIOCTL)) ;\
	curl -sSL https://raw.githubusercontent.com/istio/istio/master/release/downloadIstioCtl.sh | ISTIO_VERSION=$(ISTIOCTLVERSION) HOME=$(PROJECT_PATH)/bin/ sh - > /dev/null 2>&1;\
	mv $(PROJECT_PATH)/bin/.istioctl/bin/istioctl $(ISTIOCTL) ;\
	rm -r $(PROJECT_PATH)/bin/.istioctl ;\
	chmod +x $(ISTIOCTL) ;\
	}
endif

# Ginkgo tool
GINKGO = $(PROJECT_PATH)/bin/ginkgo
$(GINKGO):
	$(call go-get-tool,$(GINKGO),github.com/onsi/ginkgo/ginkgo@v1.16.4)

## test: Run unit tests
.PHONY : test
test: fmt vet $(GINKGO)
	# huffle both the order in which specs within a suite run, and the order in which different suites run
	# You can always rerun a given ordering later by passing the --seed flag a matching seed.
	$(GINKGO) --randomizeAllSpecs --randomizeSuites -v -progress --trace --cover ./...

## install: Build and install kuadrantctl binary ($GOBIN or GOPATH/bin)
.PHONY : install
install: fmt vet
	$(GO) install

.PHONY : fmt
fmt:
	$(GO) fmt ./...

.PHONY : vet
vet:
	$(GO) vet ./...

# Generates istio manifests with patches.
.PHONY: generate-istio-manifests
generate-istio-manifests: istioctl
	$(ISTIOCTL) manifest generate --set profile=minimal --set values.gateways.istio-ingressgateway.autoscaleEnabled=false --set values.pilot.autoscaleEnabled=false --set values.global.istioNamespace=kuadrant-system -f istiomanifests/patches/istio-externalProvider.yaml -o istiomanifests/autogenerated

.PHONY: istio-manifest-update-test
istio-manifest-update-test: generate-istio-manifests
	git diff --exit-code ./istiomanifests/autogenerated
	[ -z "$$(git ls-files --other --exclude-standard --directory --no-empty-directory ./istiomanifests/autogenerated)" ]

# Generates kuadrant manifests.
KUADRANTVERSION=v0.1.1
KUADRANT_CONTROLLER_IMAGE=quay.io/3scale/kuadrant-controller:$(KUADRANTVERSION)
.PHONY: generate-kuadrant-manifests
generate-kuadrant-manifests:
	$(eval TMP := $(shell mktemp -d))
	cd $(TMP); git clone --depth 1 --branch $(KUADRANTVERSION) https://github.com/kuadrant/kuadrant-controller.git
	cd $(TMP)/kuadrant-controller; make kustomize
	cd $(TMP)/kuadrant-controller/config/manager; $(TMP)/kuadrant-controller/bin/kustomize edit set image controller=${KUADRANT_CONTROLLER_IMAGE}
	cd $(TMP)/kuadrant-controller/config/default; $(TMP)/kuadrant-controller/bin/kustomize edit set namespace $(KUADRANT_NAMESPACE)
	cd $(TMP)/kuadrant-controller; bin/kustomize build config/default -o $(PROJECT_PATH)/kuadrantmanifests/autogenerated/kuadrant.yaml
	-rm -rf $(TMP)

.PHONY: kuadrant-manifest-update-test
kuadrant-manifest-update-test: generate-kuadrant-manifests
	git diff --exit-code ./kuadrantmanifests/autogenerated
	[ -z "$$(git ls-files --other --exclude-standard --directory --no-empty-directory ./kuadrantmanifests/autogenerated)" ]

# Generates limitador manifests.
LIMITADOR_OPERATOR_VERSION=v0.2.0
LIMITADOR_OPERATOR_IMAGE=quay.io/3scale/limitador-operator:$(LIMITADOR_OPERATOR_VERSION)
.PHONY: generate-limitador-operator-manifests
generate-limitador-operator-manifests:
	$(eval TMP := $(shell mktemp -d))
	cd $(TMP); git clone --depth 1 --branch $(LIMITADOR_OPERATOR_VERSION) https://github.com/kuadrant/limitador-operator.git
	cd $(TMP)/limitador-operator; make kustomize
	cd $(TMP)/limitador-operator/config/manager; $(TMP)/limitador-operator/bin/kustomize edit set image controller=$(LIMITADOR_OPERATOR_IMAGE)
	cd $(TMP)/limitador-operator/config/default; $(TMP)/limitador-operator/bin/kustomize edit set namespace $(KUADRANT_NAMESPACE)
	cd $(TMP)/limitador-operator; bin/kustomize build config/default -o $(PROJECT_PATH)/limitadormanifests/autogenerated/limitador-operator.yaml
	-rm -rf $(TMP)

.PHONY: limitador-operator-manifest-update-test
limitador-operator-manifest-update-test: generate-limitador-operator-manifests
	git diff --exit-code ./limitadormanifests/autogenerated
	[ -z "$$(git ls-files --other --exclude-standard --directory --no-empty-directory ./limitadormanifests/autogenerated)" ]

# Generates authorino manifests.
AUTHORINO_VERSION=v0.4.0
AUTHORINO_IMAGE=quay.io/3scale/authorino:$(AUTHORINO_VERSION)
AUTHORINO_DEPLOYMENT=cluster-wide-notls
AUTHORINO_MANIFEST_FILE=$(PROJECT_PATH)/authorinomanifests/autogenerated/authorino.yaml
.PHONY: generate-authorino-manifests
generate-authorino-manifests:
	$(eval TMP := $(shell mktemp -d))
	cd $(TMP); git clone --depth 1 --branch $(AUTHORINO_VERSION) https://github.com/kuadrant/authorino.git
	cd $(TMP)/authorino; GOBIN=$(TMP)/authorino/bin make kustomize;
	cd $(TMP)/authorino/deploy/base; $(TMP)/authorino/bin/kustomize edit set image authorino=${AUTHORINO_IMAGE}
	cd $(TMP)/authorino/deploy/overlays/$(AUTHORINO_DEPLOYMENT); $(TMP)/authorino/bin/kustomize edit set namespace $(KUADRANT_NAMESPACE)
	cd $(TMP)/authorino; $(TMP)/authorino/bin/kustomize build install > $(AUTHORINO_MANIFEST_FILE)
	echo "---" >> $(AUTHORINO_MANIFEST_FILE)
	cd $(TMP)/authorino; $(TMP)/authorino/bin/kustomize build deploy/overlays/$(AUTHORINO_DEPLOYMENT) >> $(AUTHORINO_MANIFEST_FILE)
	-rm -rf $(TMP)

.PHONY: authorino-manifest-update-test
authorino-manifest-update-test: generate-authorino-manifests
	git diff --exit-code ./authorinomanifests/autogenerated
	[ -z "$$(git ls-files --other --exclude-standard --directory --no-empty-directory ./authorinomanifests/autogenerated)" ]

.PHONY : cluster-cleanup
cluster-cleanup: $(KIND)
	$(KIND) delete cluster --name $(KIND_CLUSTER_NAME)

.PHONY : cluster-setup
cluster-setup: $(KIND) cluster-cleanup
	$(KIND) create cluster --name $(KIND_CLUSTER_NAME) --config utils/kind/cluster.yaml

GOLANGCI-LINT=$(PROJECT_PATH)/bin/golangci-lint
$(GOLANGCI-LINT):
	mkdir -p $(PROJECT_PATH)/bin
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(PROJECT_PATH)/bin v1.41.1

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI-LINT)

.PHONY: run-lint
run-lint: $(GOLANGCI-LINT)
	$(GOLANGCI-LINT) run --timeout 2m


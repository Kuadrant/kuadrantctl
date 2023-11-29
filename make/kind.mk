
##@ Kind

## Targets to help install and use kind for development https://kind.sigs.k8s.io

KIND_CLUSTER_NAME ?= kuadrantctl-local

.PHONY: kind-create-cluster
kind-create-cluster: kind ## Create the "kuadrantctl-local" kind cluster.
	$(KIND) create cluster --name $(KIND_CLUSTER_NAME) --config utils/kind-cluster.yaml

.PHONY: kind-delete-cluster
kind-delete-cluster: kind ## Delete the "kuadrantctl-local" kind cluster.
	- $(KIND) delete cluster --name $(KIND_CLUSTER_NAME)

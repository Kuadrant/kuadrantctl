##@ Gateway API resources

.PHONY: gateway-api-install
gateway-api-install: kustomize ## Install Gateway API CRDs
	$(KUSTOMIZE) build config/gateway-api | kubectl apply -f -

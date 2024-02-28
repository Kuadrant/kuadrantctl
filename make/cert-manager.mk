##@ Install cert-manager, a tool to help manage the TLS certificates.

.PHONY: cert-manager-install
cert-manager-install:
	kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.3/cert-manager.yaml

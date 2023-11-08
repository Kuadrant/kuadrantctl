##@ Install Operator Lifecycle Manager (OLM), a tool to help manage the Operators running on your cluster.

.PHONY: olm-install
olm-install:
	curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/v0.26.0/install.sh | bash -s v0.26.0

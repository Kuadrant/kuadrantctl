
##@ Istio

## Targets to help install and configure istio

# istioctl tool
ISTIOCTL=$(shell pwd)/bin/istioctl
ISTIOVERSION = 1.19.3
$(ISTIOCTL):
	mkdir -p $(shell pwd)/bin
	$(eval TMP := $(shell mktemp -d))
	cd $(TMP); curl -sSL https://istio.io/downloadIstio | ISTIO_VERSION=$(ISTIOVERSION) sh -
	cp $(TMP)/istio-$(ISTIOVERSION)/bin/istioctl ${ISTIOCTL}
	-rm -rf $(TMP)

.PHONY: istioctl
istioctl: $(ISTIOCTL) ## Download istioctl locally if necessary.

.PHONY: istio-install
istio-install: istioctl ## Install istio.
	$(ISTIOCTL) install -f utils/istio-operator.yaml -y

.PHONY: istio-uninstall
istio-uninstall: istioctl ## Uninstall istio.
	$(ISTIOCTL) uninstall -y --purge

KIND = $(PROJECT_PATH)/bin/kind
KIND_VERSION = v0.22.0
$(KIND):
	$(call go-install-tool,$(KIND),sigs.k8s.io/kind@$(KIND_VERSION))

.PHONY: kind
kind: $(KIND) ## Download kind locally if necessary.

GINKGO = $(PROJECT_PATH)/bin/ginkgo
$(GINKGO):
	# In order to make sure the version of the ginkgo cli installed
	# is the same as the version of go.mod,
	# instead of calling go-install-tool,
	# running go install from the current module will pick version from current go.mod file.
	GOBIN=$(PROJECT_PATH)/bin go install github.com/onsi/ginkgo/v2/ginkgo

.PHONY: ginkgo
ginkgo: $(GINKGO) ## Download ginkgo locally if necessary.

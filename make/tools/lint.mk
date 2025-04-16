
GOLANGCI-LINT=$(PROJECT_PATH)/bin/golangci-lint
$(GOLANGCI-LINT):
	mkdir -p $(PROJECT_PATH)/bin
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(PROJECT_PATH)/bin v1.64.8

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI-LINT) ## Download golangci-lint 

SHELL := /bin/bash

MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_PATH := $(patsubst %/,%,$(dir $(MKFILE_PATH)))

GO ?= go

all: help

.PHONY : help
help: Makefile
	@sed -n 's/^##//p' $<


## test: Run unit tests
.PHONY : test
test: fmt vet
	$(GO) test  -v ./...

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

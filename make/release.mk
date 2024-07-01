##@ Release tasks

prepare-release: 
ifndef VERSION
	$(error VERSION is undefined)
endif
	sed -i 's/Version = ".*"/Version = "$(VERSION)"/' version/version.go

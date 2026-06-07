APP        := mimixbox
PREPARE_UT := test/ut/prepareUnitTest.sh
INSTALLER  := scripts/installer.sh
MK_JAIL    := scripts/mkJailForDebianFamily.sh
RELEASE    := scripts/release.sh

build:  ## Build the mimixbox binary
	go build "-ldflags=-s -w" -trimpath -o $(APP) cmd/mimixbox/main.go
	$(MAKE) licenses

clean: ## Clean project
	-rm mimixbox
	-rm -rf vendor
	-rm cover.*
	-rm -rf /tmp/mimixbox/ut/*
	-rm -rf release
	-rm -rf licenses

docker: ## Run container for testing mimixbox
	docker image build -t mimixbox/test:latest .
	docker container run --rm -it mimixbox/test:latest

install: ## Install mimixbox (with symbolic links) on your system
	$(INSTALLER)

full-install: ## Install mimixbox and create symbolic links for every applet
	-$(INSTALLER)
	mimixbox --full-install /usr/local/bin

remove: ## Remove mimixbox-symbolic link
	mimixbox --remove /usr/local/bin

test: pre_ut  ## Run unit tests with coverage (writes cover.out / cover.html)
	-@go test -cover ./... -coverpkg=./... -coverprofile=cover.out
	-@go tool cover -html=cover.out -o cover.html
	-@rm -rf /tmp/mimixbox/ut/*

test-e2e: ## Run the shellspec end-to-end tests against the built binary
	cd test/it && shellspec --shell /bin/bash

lint: ## Run golangci-lint
	golangci-lint run ./...

jail:  ## Make jail environment for testing chroot/ischroot
	$(MK_JAIL)

release: ## Make release files.
	$(RELEASE)

licenses: ## Get licenses for dependent libraries
	-@go-licenses save ./cmd/mimixbox --force --save_path "licenses/"

pre_ut:
	@echo "Make files for test at test directory."
	-@rm -rf /tmp/mimixbox/ut/*
	@$(PREPARE_UT)

# Backwards-compatible aliases for the previous target names.
ut: test  ## Alias for "make test"
it: test-e2e  ## Alias for "make test-e2e"

.DEFAULT_GOAL := help
.PHONY: build clean docker install full-install remove test test-e2e lint jail release licenses pre_ut ut it help

help:
	@grep -E '^[0-9a-zA-Z_-]+[[:blank:]]*:.*?## .*$$' $(MAKEFILE_LIST) | sort \
	| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[1;32m%-15s\033[0m %s\n", $$1, $$2}'

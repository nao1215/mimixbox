APP        := mimixbox
PREPARE_UT := test/ut/prepareUnitTest.sh
INSTALLER  := scripts/installer.sh
MK_JAIL    := scripts/mkJailForDebianFamily.sh
MK_MAN     := scripts/mkManpages.sh
RELEASE    := scripts/release.sh

build: deps ## Build mimixbox and make man-pages
	go build "-ldflags=-s -w" -trimpath -o $(APP) cmd/mimixbox/main.go
	$(MAKE) doc
	$(MAKE) licenses

clean: ## Clean project
	-rm mimixbox
	-rm -rf vendor
	-rm cover.*
	-rm -rf /tmp/mimixbox/ut/*
	-rm -rf release
	-rm -rf licenses
	-find . -name "*.1.gz" | xargs rm -f

doc: ## Make man-pages
	$(MK_MAN)

docker: ## Run container for testing mimixbox 
	docker image build -t mimixbox/test:latest .
	docker container run --rm -it mimixbox/test:latest

install: ## Install mimixbox (with symbolic link) and man-pages on your system
	$(INSTALLER)

full-install: ## Full Install mimixbox (with symbolic link) and man-pages on your system
	$(INSTALLER)
	mimixbox --full-install /usr/local/bin

remove: ## Remove mimixbox-symbolic link
	mimixbox --remove /usr/local/bin

it: ## Execute integration test
	cd test/it && shellspec --shell /bin/bash

jail:  ## Make jail environment for testing chroot/ischroot
	$(MK_JAIL)

release: ## Make release files.
	$(RELEASE)

licenses: ## Get licenses for dependent libraries
	-@go-licenses save ./cmd/mimixbox --force --save_path "licenses/"

deps: ## Dependency resolution for build
	go mod vendor

pre_ut:
	@echo "Clean test directory."
	-@rm -rf /tmp/mimixbox/ut/*
	@echo "Make files for test at test directory."
	@$(PREPARE_UT)

ut: pre_ut  ## Execute unit test
	-@go test -cover ./... -v -coverpkg=./... -coverprofile=cover.out
	-@go tool cover -html=cover.out -o cover.html
	@echo "--------------------------------------------------------------------"
	-@rm -rf /tmp/mimixbox/ut/*
	@echo "The tool saved the coverage information in an HTML. See cover.html"


.DEFAULT_GOAL := help
.PHONY: build clean doc docker install jail release deps it ut pre_ut

help:  
	@grep -E '^[0-9a-zA-Z_-]+[[:blank:]]*:.*?## .*$$' $(MAKEFILE_LIST) | sort \
	| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[1;32m%-15s\033[0m %s\n", $$1, $$2}'
build: deps ## Build mimixbox and make man-pages
	go build "-ldflags=-s -w" -trimpath -o mimixbox cmd/mimixbox/main.go
	$(MAKE) doc
	$(MAKE) licenses

clean: ## Clean project
	-rm mimixbox
	-rm -rf vendor
	-rm cover.*
	-rm -rf release
	-rm -rf licenses
	-find . -name "*.1.gz" | xargs rm -f

doc: ## Make man-pages
	./scripts/mkManpages.sh

docker: ## Run container for testing mimixbox 
	$(MAKE) build CGO_ENABLED=0 
	docker image build -t mimixbox/test:latest .
	docker container run --rm -it mimixbox/test:latest

install: ## Install mimixbox and man-pages on your system
	./scripts/installer.sh

jail:  ## Make jail environment for testing chroot/ischroot
	./scripts/mkJailForDebianFamily.sh

release: ## Make release files.
	./scripts/release.sh

licenses: ## Get licenses for dependent libraries
	-@go-licenses save ./cmd/mimixbox --force --save_path "licenses/"

deps: ## Dependency resolution for build
	go mod vendor

.DEFAULT_GOAL := help
.PHONY: build clean doc docker install jail release deps

help:  
	@grep -E '^[0-9a-zA-Z_-]+[[:blank:]]*:.*?## .*$$' $(MAKEFILE_LIST) | sort \
	| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[1;32m%-15s\033[0m %s\n", $$1, $$2}'
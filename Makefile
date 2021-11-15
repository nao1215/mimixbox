build: deps ## Build mimixbox and make man-pages
	go build -o mimixbox cmd/mimixbox/main.go
	$(MAKE) doc

clean: ## Clean project
	-rm mimixbox
	-rm -rf vendor
	-rm cover.*
	-rm -rf release
	-find . -name "*.1.gz" | xargs rm -f

doc: ## Make man-pages
	./scripts/mkManpages.sh

install: ## Install mimixbox and man-pages on your system
	./scripts/installer.sh

jail:  ## Make jail environment for testing chroot/ischroot
	./scripts/mkJailForDebianFamily.sh

release: ## Make release files.
	./scripts/release.sh

deps: ## Dependency resolution for build
	dep ensure
	go mod vendor

.DEFAULT_GOAL := help
.PHONY: build clean doc install jail release deps
 
help:  
	@grep -E '^[0-9a-zA-Z_-]+[[:blank:]]*:.*?## .*$$' $(MAKEFILE_LIST) | sort \
	| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[1;32m%-15s\033[0m %s\n", $$1, $$2}'
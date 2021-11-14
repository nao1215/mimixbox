build: deps
	go build -o mimixbox cmd/mimixbox/main.go
	$(MAKE) doc

clean:
	-rm mimixbox
	-rm -rf vendor
	-rm cover.*
	-rm -rf release
	-find . -name "*.1.gz" | xargs rm -f

doc:
	./scripts/mkManpages.sh

install:
	./scripts/installer.sh

release:
	./scripts/release.sh

deps:
	dep ensure
	go mod vendor
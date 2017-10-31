SHELL := /bin/bash

REV := $(shell git rev-parse HEAD)
CHANGES := $(shell test -n "$$(git status --porcelain)" && echo '+CHANGES' || true)

TARGET := packer-provisioner-itamae-local
VERSION ?= $(shell cat VERSION)

VENDOR ?= vendor
PACKAGES ?= $(shell go list ./... | grep -vE $(VENDOR))
FILES ?= $(shell find . -type f -name '*.go' | grep -vE $(VENDOR))

OS ?= darwin freebsd linux openbsd
ARCH ?= 386 amd64
LDFLAGS := -X github.com/kwilczynski/$(TARGET)/itamaelocal.Revision=$(REV)$(CHANGES)

GPG_SIGNING_KEY ?=

.SUFFIXES:

.PHONY: \
	help \
	default \
	clean \
	clean-artifacts \
	clean-releases \
	clean-vendor \
	tools \
	deps \
	test \
	coverage \
	vet \
	errors \
	assignments \
	static \
	lint \
	imports \
	fmt \
	env \
	build \
	build-all \
	doc \
	release \
	package-release \
	sign-release \
	check \
	vendor \
	version


all: imports fmt lint vet errors assignments static build

help:
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@echo '    help               Show this help screen.'
	@echo '    clean              Remove binaries, artifacts and releases.'
	@echo '    clean-artifacts    Remove build artifacts only.'
	@echo '    clean-releases     Remove releases only.'
	@echo '    clean-vendor       Remove content of the vendor directory.'
	@echo '    tools              Install tools needed by the project.'
	@echo '    deps               Update and save project build time dependencies.'
	@echo '    test               Run unit tests.'
	@echo '    coverage           Report code tests coverage.'
	@echo '    vet                Run go vet.'
	@echo '    errors             Run errcheck.'
	@echo '    assignments        Run ineffassign.'
	@echo '    static             Run staticcheck.'
	@echo '    lint               Run golint.'
	@echo '    imports            Run goimports.'
	@echo '    fmt                Run gofmt.'
	@echo '    env                Display Go environment.'
	@echo '    build              Build project for current platform.'
	@echo '    build-all          Build project for all supported platforms.'
	@echo '    doc                Start Go documentation server on port 8080.'
	@echo '    release            Package and sing project for release.'
	@echo '    package-release    Package release and compress artifacts.'
	@echo '    sign-release       Sign release and generate checksums.'
	@echo '    check              Verify compiled binary.'
	@echo '    vendor             Download and install build time dependencies.'
	@echo '    version            Display Go version.'
	@echo ''
	@echo 'Targets run by default are: imports, fmt, lint, vet, errors, assignments, static and build.'
	@echo ''

print-%:
	@echo $* = $($*)

clean: clean-artifacts clean-releases
	go clean -i ./...
	rm -f \
		$(CURDIR)/coverage.* \
		$(CURDIR)/$(TARGET)_*

clean-artifacts:
	rm -Rf artifacts/*

clean-releases:
	rm -Rf releases/*

clean-vendor:
	find $(VENDOR) -type d -print0 2>/dev/null | xargs -0 rm -Rf

clean-all: clean clean-artifacts clean-vendor

tools:
	for name in zip shasum gpg; do \
		which $$name &>/dev/null || (echo "Please install $$name to continue."; exit 1); \
	done
	which dep &>/dev/null; if (( $$? > 0)); then \
		go get github.com/golang/dep/cmd/dep; \
	fi
	go get github.com/axw/gocov/gocov
	go get github.com/golang/lint/golint
	go get github.com/gordonklaus/ineffassign
	go get github.com/kisielk/errcheck
	go get github.com/matm/gocov-html
	go get github.com/mitchellh/gox
	go get golang.org/x/tools/cmd/goimports
	go get honnef.co/go/tools/cmd/staticcheck

deps:
	dep ensure -update
	dep prune

test:
	go test -v $(PACKAGES)

coverage:
	gocov test $(PACKAGES) > $(CURDIR)/coverage.out 2>/dev/null
	gocov report $(CURDIR)/coverage.out
	if [[ -z "$$CI" ]]; then \
		gocov-html $(CURDIR)/coverage.out > $(CURDIR)/coverage.html; \
	  	if which open &>/dev/null; then \
	   		open $(CURDIR)/coverage.html; \
	  	fi; \
	fi

vet:
	$(eval QUIET := $(shell test "$(MAKECMDGOALS)" == "vet" || echo 1))
	@go vet -v $(PACKAGES) $(shell test -z "$(QUIET)" || echo '&>/dev/null'); \
	if (( $$? > 0 )); then \
		if [[ -n "$(QUIET)" ]]; then \
			echo "go vet found number of issues. Run 'make vet' to check directly."; \
		else \
			exit $$?; \
		fi; \
	fi

errors:
	$(eval QUIET := $(shell test "$(MAKECMDGOALS)" == "errors" || echo 1))
	@errcheck -ignoretests -blank $(PACKAGES) $(shell test -z "$(QUIET)" || echo '&>/dev/null'); \
	if (( $$? > 0 )); then \
		if [[ -n "$(QUIET)" ]]; then \
			echo "errcheck found number of issues. Run 'make errors' to check directly."; \
		else \
			exit $$?; \
		fi; \
	fi

assignments:
	$(eval QUIET := $(shell test "$(MAKECMDGOALS)" == "assignments" || echo 1))
	@ineffassign . $(shell test -z "$(QUIET)" || echo '&>/dev/null'); \
	if (( $$? > 0 )); then \
		if [[ -n "$(QUIET)" ]]; then \
			echo "ineffassign found number of issues. Run 'make assignments' to check directly."; \
		else \
			exit $$?; \
		fi; \
	fi

static:
	$(eval QUIET := $(shell test "$(MAKECMDGOALS)" == "static" || echo 1))
	@staticcheck $(PACKAGES) $(shell test -z "$(QUIET)" || echo '&>/dev/null'); \
	if (( $$? > 0 )); then \
		if [[ -n "$(QUIET)" ]]; then \
			echo "staticcheck found number of issues. Run 'make static' to check directly."; \
		else \
			exit $$?; \
		fi; \
	fi
lint:
	$(eval QUIET := $(shell test "$(MAKECMDGOALS)" == "lint" || echo 1))
	@golint $(PACKAGES) $(shell test -z "$(QUIET)" || echo '&>/dev/null'); \
	if (( $$? > 0 )); then \
		if [[ -n "$(QUIET)" ]]; then \
			echo "golint found number of issues. Run 'make lint' to check directly."; \
		else \
			exit $$?; \
		fi; \
	fi

imports:
	$(eval QUIET := $(shell test "$(MAKECMDGOALS)" == "imports" || echo 1))
	@goimports -l $(FILES) $(shell test -z "$(QUIET)" || echo '&>/dev/null'); \
	if (( $$? > 0 )); then \
		if [[ -n "$(QUIET)" ]]; then \
			echo "goimports found number of issues. Run 'make imports' to check directly."; \
		else \
			exit $$?; \
		fi; \
	fi

fmt:
	$(eval QUIET := $(shell test "$(MAKECMDGOALS)" == "fmt" || echo 1))
	@gofmt -l $(FILES) $(shell test -z "$(QUIET)" || echo '&>/dev/null'); \
	if (( $$? > 0 )); then \
		if [[ -n "$(QUIET)" ]]; then \
			echo "gofmt found number of issues. Run 'make fmt' to check directly."; \
		else \
			exit $$?; \
		fi; \
	fi

env:
	@go env

build:
	go build \
		-ldflags "$(LDFLAGS)" \
		-o "$(TARGET)" .

build-all: vendor
	mkdir -p $(CURDIR)/artifacts/$(VERSION)
	gox \
		-os "$(OS)" -arch "$(ARCH)" \
		-ldflags "$(LDFLAGS)" \
		-output "$(CURDIR)/artifacts/$(VERSION)/{{.OS}}_{{.Arch}}/$(TARGET)" .
	cp -f $(CURDIR)/artifacts/$(VERSION)/$$(go env GOOS)_$$(go env GOARCH)/$(TARGET) .

doc:
	godoc -http=:8080 -index

release: build-all package-release sign-release

package-release:
	@test -x $(CURDIR)/artifacts/$(VERSION) || (echo 'Please make a release first.'; exit 1)
	mkdir -p $(CURDIR)/releases/$(VERSION)
	for release in $$(find $(CURDIR)/artifacts/$(VERSION) -mindepth 1 -maxdepth 1 -type d 2>/dev/null); do \
		platform=$$(basename $$release); \
		pushd $$release &>/dev/null; \
			zip $(CURDIR)/releases/$(VERSION)/$(TARGET)_$${platform}.zip $(TARGET) &>/dev/null; \
		popd &>/dev/null; \
	done

sign-release:
	@test -x $(CURDIR)/artifacts/$(VERSION) || (echo 'Please make a release first.'; exit 1)
	pushd $(CURDIR)/releases/$(VERSION) &>/dev/null; \
	shasum -a 256 -b $(TARGET)_* > SHA256SUMS; \
	if test -n "$(GPG_SIGNING_KEY)"; then \
		gpg \
			--default-key $(GPG_SIGNING_KEY) \
			-a -o SHA256SUMS.sign -b SHA256SUMS; \
	fi; \
	popd &>/dev/null

check:
	@if $(CURDIR)/$(TARGET) --version | grep -qF '$(VERSION)'; then \
		echo "$(CURDIR)/$(TARGET): OK"; \
	else \
		exit 1; \
	fi

vendor:
	dep ensure -vendor-only

version:
	@go version

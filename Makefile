#
# Makefile
#
# Copyright 2016 Krzysztof Wilczynski
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

SHELL := /bin/bash

REV := $(shell git rev-parse HEAD)
CHANGES := $(shell test -n "$$(git status --porcelain)" && echo '+CHANGES' || true)

TARGET := packer-provisioner-itamae
VERSION := $(shell cat VERSION)

OS := darwin freebsd linux openbsd
ARCH := 386 amd64
LDFLAGS := -X github.com/kwilczynski/$(TARGET)/itamae.Revision=$(REV)$(CHANGES)

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
	vet \
	lint \
	imports \
	fmt \
	env \
	compile \
	build \
	doc \
	release \
	sign-release \
	check \
	vendor \
	version

all: imports fmt lint vet build

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
	@echo '    deps               Download and install build time dependencies.'
	@echo '    test               Run unit tests.'
	@echo '    vet                Run go vet.'
	@echo '    lint               Run golint.'
	@echo '    imports            Run goimports.'
	@echo '    fmt                Run go fmt.'
	@echo '    env                Display Go environment.'
	@echo '    compile            Compile binary for current system and architecture.'
	@echo '    build              Test and compile project for supported platforms.'
	@echo '    doc                Start Go documentation server on port 8080.'
	@echo '    release            Prepare project for release.'
	@echo '    sign-release       Sign release and generate checksums.'
	@echo '    check              Verify compiled binary.'
	@echo '    vendor             Update and save project build time dependencies.'
	@echo '    version            Display Go version.'
	@echo ''
	@echo 'Targets run by default are: imports, fmt, lint, vet and build.'
	@echo ''

print-%:
	@echo $* = $($*)

clean: clean-artifacts clean-releases
	go clean -x -i ./...
	rm -vf packer-provisioner-itamae_*

clean-artifacts:
	rm -vRf artifacts/*

clean-releases:
	rm -vRf releases/*

clean-vendor:
	find $(CURDIR)/vendor -type d -print0 2>/dev/null | xargs -0 rm -vRf

clean-all: clean clean-artifacts clean-vendor

tools:
	go get golang.org/x/tools/cmd/vet
	go get golang.org/x/tools/cmd/goimports
	go get github.com/golang/lint/golint
	go get github.com/tools/godep
	go get github.com/mitchellh/gox

deps:
	godep restore

test: deps
	go test -v ./...

vet:
	go vet -v ./...

lint:
	golint ./...

imports:
	goimports -l -w .

fmt:
	go fmt ./...

env:
	@go env

compile: compile-binary check

compile-binary: env deps
	go build -v \
	   -ldflags "$(LDFLAGS)" \
	   -o "$(TARGET)" .

build: env test
	mkdir -v -p $(CURDIR)/artifacts/$(VERSION)
	gox -verbose \
	    -os "$(OS)" -arch "$(ARCH)" \
	    -ldflags "$(LDFLAGS)" \
	    -output "$(CURDIR)/artifacts/$(VERSION)/{{.OS}}_{{.Arch}}/$(TARGET)" .
	cp -v -f \
	   "$(CURDIR)/artifacts/$(VERSION)/$$(go env GOOS)_$$(go env GOARCH)/$(TARGET)" .

doc:
	godoc -http=:8080 -index

release:
	@test -x $(CURDIR)/artifacts/$(VERSION) || exit 1
	mkdir -v -p $(CURDIR)/releases/$(VERSION)
	for release in $$(find $(CURDIR)/artifacts/$(VERSION) -mindepth 1 -maxdepth 1 -type d); do \
	  platform=$$(basename $$release); \
	  pushd $$release &>/dev/null; \
	  zip -v $(CURDIR)/releases/$(VERSION)/$(TARGET)_$${platform}.zip $(TARGET); \
	  popd &>/dev/null; \
	done

sign-release:
	@test -x $(CURDIR)/releases/$(VERSION) || exit 1
	pushd $(CURDIR)/releases/$(VERSION) &>/dev/null; \
	shasum -a 256 -b $(TARGET)_* | tee SHA256SUMS; \
	popd &>/dev/null

check:
	@test -x $(CURDIR)/$(TARGET) || exit 1
	if $(CURDIR)/$(TARGET) --version | grep -qF '$(VERSION)'; then \
	  echo "$(CURDIR)/$(TARGET): OK"; \
	else \
	  exit 1; \
	fi

vendor: deps
	godep save

version:
	@go version

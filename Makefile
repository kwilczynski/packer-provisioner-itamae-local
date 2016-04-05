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

OS := $(OS:darwin freebsd linux openbsd)
ARCH := $(ARCH: 386 amd64)
LDFLAGS := -X github.com/kwilczynski/$(TARGET)/itamae.Revision=$(REV)$(CHANGES)

.PHONY: \
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

default: all

all: imports fmt lint vet build

clean: clean-artifacts clean-releases
	go clean -x -i ./...
	rm -vf packer-provisioner-itamae_*

clean-artifacts:
	rm -vRf artifacts/*

clean-releases:
	rm -vRf releases/*

clean-vendor:
	find $(CURDIR)/vendor -type d -print0 | xargs -0 rm -vRf || true

clean-all: clean clean-packages clean-vendor

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
	go env

compile: compile-binary check

compile-binary: env deps
	go build -v \
	   -ldflags "$(LDFLAGS)" \
	   -o "$(TARGET)" .

build: env test
	test -x $(CURDIR)/artifacts || mkdir -v -p $(CURDIR)/artifacts/$(VERSION)
	gox -verbose \
	    -os "$(OS)" -arch "$(ARCH)" \
	    -ldflags "$(LDFLAGS)" \
	    -output "$(CURDIR)/artifacts/$(VERSION)/{{.OS}}_{{.Arch}}/$(TARGET)" .
	cp -v -f \
	   "$(CURDIR)/artifacts/$(VERSION)/$$(go env GOOS)_$$(go env GOARCH)/$(TARGET)" .

doc:
	godoc -http=:8080 -index

release:
	test -x $(CURDIR)/releases || mkdir -v -p $(CURDIR)/releases/$(VERSION)
	for release in $$(find $(CURDIR)/artifacts/$(VERSION) -mindepth 1 -maxdepth 1 -type d); do \
	  platform=$$(basename $$release); \
	  pushd $$release >/dev/null 2>&1; \
	  zip -v $(CURDIR)/releases/$(VERSION)/$(TARGET)_$(VERSION)_$${platform}.zip $(TARGET); \
	  popd >/dev/null 2>&1; \
	done

sign-release:
	pushd $(CURDIR)/releases/$(VERSION) >/dev/null 2>&1; \
	shasum -a 256 -b $(TARGET)_$(VERSION)_* > $(TARGET)_$(VERSION)_SHA256SUMS; \
	popd >/dev/null 2>&1

check:
	test -x $(CURDIR)/$(TARGET) || exit 1
	if $(CURDIR)/$(TARGET) --version | grep -qF '$(VERSION)'; then \
	  echo "$(CURDIR)/$(TARGET): OK"; \
	else \
	  exit 1; \
	fi

vendor: deps
	godep save

version: env
	go version

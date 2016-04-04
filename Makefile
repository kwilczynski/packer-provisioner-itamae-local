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

#GOPATH := $(CURDIR)/vendor:${GOPATH}

#export GOPATH
#export GO15VENDOREXPERIMENT=1

REV := $(shell git rev-parse HEAD)
CHANGES := $(shell test -n "$$(git status --porcelain)" && echo '+CHANGES' || true)

PKG := packer-provisioner-itamae
VERSION := $(shell cat VERSION)

OS := linux darwin
ARCH := 386 amd64
LDFLAGS := -X github.com/kwilczynski/$(PKG)/itamae.Revision=$(REV)$(CHANGES)

.PHONY: default clean clean-vendor deps tools vet test lint fmt release vendor

default: all

all: fmt lint vet build

clean: clean-packages
	go clean -x -i ./...
	rm -vf packer-provisioner-itamae_*

clean-packages:
	rm -vdRf packages/*

clean-vendor:
	find $(CURDIR)/vendor -type d -print0 | xargs -0 rm -vdRf || true

clean-all: clean clean-packages clean-vendor

deps: tools
	godep restore

tools:
	go get golang.org/x/tools/cmd/vet
	go get github.com/golang/lint/golint
	go get github.com/tools/godep
	go get github.com/mitchellh/gox

test: deps
	go test -v ./...

vet:
	go vet -v ./...

lint:
	golint ./...

fmt:
	go fmt ./...

env:
	go env

compile: env deps
	go build -v \
	   -ldflags "$(LDFLAGS)" \
	   -o "$(PKG)" .

build: env test
	@test -x $(CURDIR)/packages || mkdir -v $(CURDIR)/packages
	gox -verbose \
	    -os "$(OS)" -arch "$(ARCH)" \
	    -ldflags "$(LDFLAGS)" \
	    -output "$(CURDIR)/packages/{{.OS}}_{{.Arch}}/$(PKG)" .
	cp -v -f \
	   $(CURDIR)/packages/$$(go env GOOS)_$$(go env GOARCH)/$(PKG) .

doc:
	godoc -http=:8080 -index

release:
	for release in $$(find $(CURDIR)/packages -mindepth 1 -maxdepth 1 -type d); do \
	  platform=$$(basename $$release); \
	  pushd $$release >/dev/null 2>&1; \
	  zip -v $(CURDIR)/$(PKG)_$(VERSION)_$${platform}.zip $(PKG); \
	  popd >/dev/null 2>&1; \
	done

sign-release:
	shasum -a 256 -b $(PKG)_$(VERSION)_* > ./$(PKG)_${VERSION}_SHA256SUMS

vendor:
	godep restore
	godep save

version: env
	go version

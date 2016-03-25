include golang.mk
.DEFAULT_GOAL := test # override default goal set in library makefile

.PHONY: all build test vendor $(PKGS)
SHELL := /bin/bash
PKG = github.com/Clever/http-science
PKGS := $(shell go list ./... | grep -v /vendor)
EXECUTABLE := $(shell basename $(PKG))
$(eval $(call golang-version-check,1.5))

BUILDS := \
	build/linux-amd64 \
	build/darwin-amd64

all: test build

build/darwin-amd64:
	GOARCH=amd64 GOOS=darwin go build -o "$@/$(EXECUTABLE)" $(PKG)
build/linux-amd64:
	GOARCH=amd64 GOOS=linux go build -o "$@/$(EXECUTABLE)" $(PKG)

build: $(BUILDS)

test: $(PKGS)
$(PKGS): golang-test-all-strict-deps
	$(call golang-test-all-strict,$@)

vendor: golang-godep-vendor-deps
	$(call golang-godep-vendor,$(PKGS))

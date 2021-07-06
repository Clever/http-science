include sfncli.mk
include golang.mk
.DEFAULT_GOAL := test # override default goal set in library makefile

.PHONY: all build test vendor $(PKGS) install_deps
SHELL := /bin/bash
PKG = github.com/Clever/http-science
PKGS := $(shell go list ./... | grep -v /vendor)
EXECUTABLE := $(shell basename $(PKG))
SFNCLI_VERSION := latest
$(eval $(call golang-version-check,1.13))

BUILDS := \
	build/linux-amd64

all: test build

build/linux-amd64:
	GOARCH=amd64 GOOS=linux go build -o "$@/$(EXECUTABLE)" $(PKG)

build: clean $(BUILDS) bin/sfncli

clean:
	-rm -rf build

test: $(PKGS)
$(PKGS): golang-test-all-strict-deps
	$(call golang-test-all-strict,$@)




run: build
	gearcmd --name http-science --cmd build/linux-amd64/http-science --parseargs=false --pass-sigterm


install_deps:
	go mod vendor

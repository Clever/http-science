SHELL := /bin/bash
PKG = github.com/Clever/http-science
PKGS := $(shell go list ./... | grep -v /vendor)
EXECUTABLE := http-science
.PHONY: build all vendor $(PKGS)

all: test build

GOLINT := $(GOPATH)/bin/golint
$(GOLINT):
	go get github.com/golang/lint/golint

GODEP := $(GOPATH)/bin/godep
$(GODEP):
	go get -u github.com/tools/godep

test: $(PKGS)

$(PKGS): $(GOLINT)
	go install $@
	gofmt -w=true $(GOPATH)/src/$@/*.go
	$(GOLINT) $(GOPATH)/src/$@/*.go
	@echo "TESTING $@..."
	go test -v $@

BUILDS := \
	build/linux-amd64 \
	build/darwin-amd64

build/darwin-amd64:
	GOARCH=amd64 GOOS=darwin go build -o "$@/$(EXECUTABLE)" $(PKG)
build/linux-amd64:
	GOARCH=amd64 GOOS=linux go build -o "$@/$(EXECUTABLE)" $(PKG)

build: $(BUILDS)

vendor: $(GODEP)
	$(GODEP) save $(PKGS)
	find vendor/ -path '*/vendor' -type d | xargs -IX rm -r X # remove any nested vendor directories

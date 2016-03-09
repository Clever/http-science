SHELL := /bin/bash
PKG = github.com/Clever/http-science
GOLINT := $(GOPATH)/bin/golint

test: $(PKG)

all: build test

$(GOLINT):
	go get github.com/golang/lint/golint

$(PKG): $(GOLINT) $(GODEP)
	go get $@
	go install $@
	gofmt -w=true $(GOPATH)/src/$@/*.go
	$(GOLINT) $(GOPATH)/src/$@/*.go
ifeq ($(COVERAGE),1)
	go test -cover -coverprofile=$(GOPATH)/src/$@/c.out $@ -test.v
	go tool cover -html=$(GOPATH)/src/$@/c.out
else
	@echo "TESTING $@..."
	go test -v $@
endif

EXECUTABLE := http-science
BUILDS := \
	build/linux-amd64 \
	build/darwin-amd64

$(GOPATH)/bin/gox:
	go get github.com/mitchellh/gox
build/darwin-amd64: $(GOPATH)/bin/gox
	sudo PATH=$$PATH:`go env GOROOT`/bin $(GOPATH)/bin/gox -build-toolchain -os darwin -arch amd64
	GOARCH=amd64 GOOS=darwin go build -o "$@/$(EXECUTABLE)" $(PKG)
build/linux-amd64: $(GOPATH)/bin/gox
	sudo PATH=$$PATH:`go env GOROOT`/bin $(GOPATH)/bin/gox -build-toolchain -os linux -arch amd64
	GOARCH=amd64 GOOS=linux go build -o "$@/$(EXECUTABLE)" $(PKG)

build: $(BUILDS)

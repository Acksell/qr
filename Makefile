BINARY := qr
GOBIN  := $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN := $(shell go env GOPATH)/bin
endif

.PHONY: build install clean

build:
	go build -o $(BINARY) ./

install:
	go build -o $(GOBIN)/$(BINARY) ./

clean:
	rm -f $(BINARY)

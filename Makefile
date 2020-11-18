VERSION := $(shell git describe --tags --always)
BUILD := go build -v -ldflags "-s -w -X main.Version=$(VERSION)"

BINARY = check_netdev_linux

.PHONY : all clean build test

all: build test

test:
	go test -v ./...

clean:
	rm -rf build/

build:
	mkdir -p build
	GOOS=linux GOARCH=amd64 $(BUILD) -o build/$(BINARY).amd64 .
	GOOS=linux GOARCH=arm64 $(BUILD) -o build/$(BINARY).arm64 .
	GOOS=linux GOARCH=arm $(BUILD) -o build/$(BINARY).arm .

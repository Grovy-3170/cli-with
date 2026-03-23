# Makefile for cli-with

# Go build variables
BINARY_NAME=with
CMD_PATH=./cmd/with
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

.PHONY: all build test test-race lint vet fmt clean install uninstall deps

# Default target
all: build test

# Build the with binary
build:
	go build -buildvcs=false $(LDFLAGS) -o $(BINARY_NAME) $(CMD_PATH)

# Run tests
test:
	go test -v ./...

# Run tests with race detector
test-race:
	go test -v -race ./...

# Run golangci-lint
lint:
	golangci-lint run ./...

# Run go vet
vet:
	go vet ./...

# Run gofmt
fmt:
	gofmt -w .

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)

# Install binary to $GOPATH/bin
install:
	go install -buildvcs=false $(LDFLAGS) $(CMD_PATH)

# Uninstall binary from GOPATH/bin
uninstall:
	rm -f $(shell go env GOPATH)/bin/$(BINARY_NAME)
	rm -f $(BINARY_NAME)

# Download dependencies
deps:
	go mod download

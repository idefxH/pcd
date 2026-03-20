# Makefile for pcdp-wizard

BINARY_NAME=pcdp-wizard
GO_FILES=$(wildcard *.go)

.PHONY: build test install clean

build: $(BINARY_NAME)

$(BINARY_NAME): $(GO_FILES) go.mod
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINARY_NAME) .

test:
	go test -v ./...

install: build
	install -m 755 $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)
	go clean

# Development targets
fmt:
	go fmt ./...

vet:
	go vet ./...

mod-tidy:
	go mod tidy

# Build for multiple architectures
build-linux-amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINARY_NAME)-linux-amd64 .

build-linux-arm64:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINARY_NAME)-linux-arm64 .

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINARY_NAME)-darwin-amd64 .

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINARY_NAME)-darwin-arm64 .

build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64
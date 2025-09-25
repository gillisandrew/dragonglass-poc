# Dragonglass CLI Build System
BINARY_NAME=dragonglass-build
CLI_BINARY_NAME=dragonglass
VERSION?=dev
COMMIT_HASH=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d %H:%M:%S UTC')

# Go build flags
LDFLAGS=-ldflags="-s -w -X 'main.Version=$(VERSION)' -X 'main.Commit=$(COMMIT_HASH)' -X 'main.BuildTime=$(BUILD_TIME)'"

# Default target
.PHONY: all
all: build build-cli

# Build the build tool binary
.PHONY: build
build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/dragonglass-build

# Build the main CLI binary
.PHONY: build-cli
build-cli:
	go build $(LDFLAGS) -o bin/$(CLI_BINARY_NAME) ./cmd/dragonglass

# Build for multiple platforms
.PHONY: build-all
build-all: build-darwin build-linux

.PHONY: build-darwin
build-darwin:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/dragonglass-build
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/dragonglass-build
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(CLI_BINARY_NAME)-darwin-amd64 ./cmd/dragonglass
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(CLI_BINARY_NAME)-darwin-arm64 ./cmd/dragonglass

.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/dragonglass-build
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64 ./cmd/dragonglass-build
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(CLI_BINARY_NAME)-linux-amd64 ./cmd/dragonglass
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(CLI_BINARY_NAME)-linux-arm64 ./cmd/dragonglass

# Run tests
.PHONY: test
test:
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install to GOPATH/bin
.PHONY: install
install:
	go install $(LDFLAGS) ./cmd/dragonglass-build
	go install $(LDFLAGS) ./cmd/dragonglass

# Format code
.PHONY: fmt
fmt:
	go fmt ./...

# Lint code
.PHONY: lint
lint:
	golangci-lint run

# Download dependencies
.PHONY: deps
deps:
	go mod download
	go mod tidy

# Development build (with debug symbols)
.PHONY: dev
dev:
	go build -o bin/$(BINARY_NAME) ./cmd/dragonglass-build
	go build -o bin/$(CLI_BINARY_NAME) ./cmd/dragonglass

# Run with dagger (for development/testing)
.PHONY: run-local
run-local: build
	dagger run ./bin/$(BINARY_NAME) . --directory example-plugin

.PHONY: run-remote
run-remote: build
	dagger run ./bin/$(BINARY_NAME) https://github.com/gillisandrew/dragonglass-poc.git --ref main --directory example-plugin

# Run the main CLI
.PHONY: run-cli
run-cli: build-cli
	./bin/$(CLI_BINARY_NAME)

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build        - Build the build tool binary"
	@echo "  build-cli    - Build the main CLI binary"
	@echo "  build-all    - Build both binaries for all supported platforms"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  clean        - Clean build artifacts"
	@echo "  install      - Install both binaries to GOPATH/bin"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  dev          - Development build of both binaries"
	@echo "  run-local    - Build and run build tool with local example"
	@echo "  run-remote   - Build and run build tool with remote example"
	@echo "  run-cli      - Build and run the main CLI"
	@echo "  help         - Show this help"
	@echo ""
	@echo "Environment variables:"
	@echo "  VERSION      - Build version (default: dev)"
	@echo ""
	@echo "Examples:"
	@echo "  make build VERSION=v1.0.0"
	@echo "  make build-cli"
	@echo "  make run-cli"
	@echo "  make build-all"
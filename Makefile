APP_NAME := infra-cli
BUILD_DIR := bin
MODULE := github.com/jasonjacinth/infra-cli

# Build metadata — injected into the binary at link time via ldflags.
# VERSION can be overridden at invocation: make build VERSION=1.2.3
VERSION   ?= 0.2.0
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# ldflags injects version, commit, and build time into the cmd package variables.
LDFLAGS := -ldflags "\
  -X github.com/jasonjacinth/infra-cli/cmd.Version=$(VERSION) \
  -X github.com/jasonjacinth/infra-cli/cmd.GitCommit=$(GIT_COMMIT) \
  -X github.com/jasonjacinth/infra-cli/cmd.BuildTime=$(BUILD_TIME)"

.PHONY: build build-all clean run test vet help

## build: Compile the binary for the current platform into ./bin/
build:
	@echo "Building $(APP_NAME) v$(VERSION) ($(GIT_COMMIT))..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) .
	@echo "Binary built at $(BUILD_DIR)/$(APP_NAME)"

## build-all: Cross-compile binaries for macOS (arm64/amd64) and Linux (amd64)
build-all:
	@echo "Cross-compiling $(APP_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)

	@echo "  -> darwin/arm64 (macOS Apple Silicon)"
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 .

	@echo "  -> darwin/amd64 (macOS Intel)"
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 .

	@echo "  -> linux/amd64"
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 .

	@echo ""
	@echo "Binaries:"
	@ls -lh $(BUILD_DIR)/$(APP_NAME)-*

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

## run: Build and run with optional ARGS (e.g., make run ARGS="status --app my-service")
run: build
	./$(BUILD_DIR)/$(APP_NAME) $(ARGS)

## test: Run all tests
test:
	go test ./... -v

## vet: Run go vet for static analysis
vet:
	go vet ./...

## help: Show this help message
help:
	@echo "Available targets:"
	@grep -E '^## ' Makefile | sed 's/## /  /'

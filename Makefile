APP_NAME := infra-cli
BUILD_DIR := bin
MODULE := github.com/jasonjacinth/infra-cli

# Build metadata
VERSION ?= 0.1.0
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

.PHONY: build build-all clean run test help

## build: Compile the binary for the current platform into ./bin/
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) .
	@echo "Binary built at $(BUILD_DIR)/$(APP_NAME)"

## build-all: Cross-compile binaries for macOS (arm64/amd64) and Linux (amd64)
build-all:
	@echo "Cross-compiling $(APP_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)

	@echo "  → darwin/arm64 (macOS Apple Silicon)"
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 .

	@echo "  → darwin/amd64 (macOS Intel)"
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 .

	@echo "  → linux/amd64"
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 .

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
	go test ./...

## help: Show this help message
help:
	@echo "Available targets:"
	@grep -E '^## ' Makefile | sed 's/## /  /'

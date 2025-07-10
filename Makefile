# Bitbucket Runner Makefile

# Binary name
BINARY_NAME=bitbucket-runner

# Build directory
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Version information
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
BUILD_USER := $(shell whoami)

# Linker flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.BuildUser=$(BUILD_USER)"

.PHONY: all build clean test test-race test-integration install help

all: clean build test ## Build and test

help: ## Show this help message
	@echo 'Usage:'
	@echo '  make <target>'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) printf "  %-20s %s\n", $$1, $$2 \
	}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-race: ## Run tests with race detection
	@echo "Running tests with race detection..."
	$(GOTEST) -race -v ./...

test-integration: ## Run integration tests (requires Docker)
	@echo "Running integration tests..."
	$(GOTEST) -tags=integration -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

install: build ## Install binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t bitbucket-runner:$(VERSION) .
	docker build -t bitbucket-runner:latest .

docker-run: ## Run in Docker
	docker run --rm -v $(PWD):/workspace bitbucket-runner:latest

release: clean build-all test ## Build release artifacts
	@echo "Creating release..."
	cd $(BUILD_DIR) && \
	for binary in *; do \
		tar -czf $$binary.tar.gz $$binary; \
	done

dev: ## Run in development mode
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) . && ./$(BUILD_DIR)/$(BINARY_NAME)

watch: ## Watch and rebuild on changes
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air

.DEFAULT_GOAL := help
# Stroidex CLI Makefile
# Build and test automation for Stroidex CLI

# Go parameters
GOCMD=go
GOBUILD=build
GOTEST=test
GOFMT=fmt
GOVET=vet

# Build parameters
BINARY_NAME=stroidex
BUILD_DIR=build
DIST_DIR=dist

# Version and build info
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOCMD) $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Install binary
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/ || cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Install complete"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) ./internal/cli/... -v

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) ./internal/cli/... -cover -v

# Run specific test
.PHONY: test-unit
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) ./internal/cli/... -v -run TestProgressBar

# Run benchmarks
.PHONY: benchmark
benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) ./internal/cli/... -bench=. -benchmem

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DIST_DIR)
	@echo "Clean complete"

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

# Vet code
.PHONY: vet
vet:
	@echo "Vetting code..."
	$(GOVET) ./...

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed, skipping lint"; \
	fi

# Run all checks
.PHONY: check
check: fmt vet test

# Development build with race detector
.PHONY: dev
dev:
	@echo "Building for development..."
	$(GOCMD) $(GOBUILD) -race $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

# Build for multiple platforms
.PHONY: build-all
build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(DIST_DIR)

	# Linux AMD64
	@echo "Building for Linux AMD64..."
	GOOS=linux GOARCH=amd64 $(GOCMD) $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 .

	# Darwin AMD64
	@echo "Building for Darwin AMD64..."
	GOOS=darwin GOARCH=amd64 $(GOCMD) $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 .

	# Windows AMD64
	@echo "Building for Windows AMD64..."
	GOOS=windows GOARCH=amd64 $(GOCMD) $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe .

# Create release
.PHONY: release
release: clean test build-all
	@echo "Creating release..."
	@cd $(DIST_DIR) && \
		for file in $(BINARY_NAME)-*; do \
			if [[ $$file == *.exe ]]; then \
				zip -r $(BINARY_NAME)-$$file-$(VERSION).zip $$file; \
			else \
				tar -czf $(BINARY_NAME)-$$file-$(VERSION).tar.gz $$file; \
			fi; \
		done
	@echo "Release created in $(DIST_DIR)"

# Run the application
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME) --help

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	$(GOCMD) mod download
	$(GOCMD) mod tidy

# Update dependencies
.PHONY: update
update:
	@echo "Updating dependencies..."
	$(GOCMD) get -u ./...
	$(GOCMD) mod tidy

# Generate documentation
.PHONY: docs
docs:
	@echo "Generating documentation..."
	@mkdir -p docs/generated
	$(GOCMD) doc -all -html -doc docs/generated

# Show help
.PHONY: help
help:
	@echo "Stroidex CLI Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build       - Build the binary"
	@echo "  install     - Install the binary"
	@echo "  test        - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  benchmark   - Run benchmarks"
	@echo "  clean       - Clean build artifacts"
	@echo "  fmt         - Format code"
	@echo "  vet         - Vet code"
	@echo "  lint        - Lint code"
	@echo "  check       - Run fmt, vet and test"
	@echo "  dev         - Build for development (with race detector)"
	@echo "  build-all   - Build for all platforms"
	@echo "  release     - Create release packages"
	@echo "  run         - Build and run the application"
	@echo "  deps        - Install dependencies"
	@echo "  update      - Update dependencies"
	@echo "  docs        - Generate documentation"
	@echo "  help        - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make test"
	@echo "  make test-coverage"
	@echo "  make build-all"
# WallFetch Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet
GOLINT=golangci-lint

# Binary info
BINARY_NAME=wallfetch
BINARY_PATH=./cmd/wallfetch
BUILD_DIR=build

# Version info
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Install paths
PREFIX?=/usr/local
BINDIR=$(PREFIX)/bin

# Build flags
LDFLAGS=-ldflags "-s -w -X github.com/AccursedGalaxy/wallfetch/internal/cli.Version=$(VERSION)"
GCFLAGS=
TEST_FLAGS=-v -race -cover

.PHONY: all build build-dev clean test test-cover test-verbose fmt vet lint install uninstall deps update-deps completions install-automation uninstall-automation automation-status dev-setup check help

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(BINARY_PATH)

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(BINARY_PATH)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(BINARY_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(BINARY_PATH)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(BINARY_PATH)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# Run tests
test:
	$(GOTEST) $(TEST_FLAGS) ./...

# Run tests with coverage
test-cover:
	$(GOTEST) $(TEST_FLAGS) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests verbosely
test-verbose:
	$(GOTEST) $(TEST_FLAGS) -cover ./...

# Format code
fmt:
	$(GOFMT) ./...

# Vet code
vet:
	$(GOVET) ./...

# Lint code (requires golangci-lint)
lint:
	@if command -v $(GOLINT) >/dev/null 2>&1; then \
		$(GOLINT) run; \
	else \
		echo "$(GOLINT) not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Check code quality (fmt, vet, lint, test)
check: fmt vet lint test

# Install dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Update dependencies
update-deps:
	$(GOMOD) tidy
	$(GOGET) -u ./...

# Install the binary (requires sudo for system-wide install)
install: build
	@echo "Installing $(BINARY_NAME) to $(BINDIR)..."
	@mkdir -p $(BINDIR)
	@if [ -w "$(BINDIR)" ]; then \
		install -m 755 $(BUILD_DIR)/$(BINARY_NAME) $(BINDIR)/$(BINARY_NAME); \
	else \
		sudo install -m 755 $(BUILD_DIR)/$(BINARY_NAME) $(BINDIR)/$(BINARY_NAME); \
	fi
	@echo "Installation complete! Run 'wallfetch --help' to get started."

# Install locally (user-only, no sudo required)
install-local: build
	@echo "Installing $(BINARY_NAME) locally to ~/bin..."
	@mkdir -p ~/bin
	install -m 755 $(BUILD_DIR)/$(BINARY_NAME) ~/bin/$(BINARY_NAME)
	@echo "Installation complete! Make sure ~/bin is in your PATH."
	@echo "Run 'export PATH=$$PATH:~/bin' or add it to your shell profile."

# Uninstall the binary
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@if [ -w "$(BINDIR)" ]; then \
		rm -f $(BINDIR)/$(BINARY_NAME); \
	else \
		sudo rm -f $(BINDIR)/$(BINARY_NAME); \
	fi
	@echo "Uninstall complete!"

# Development commands
dev-build:
	@echo "Building for development..."
	$(GOBUILD) -o $(BINARY_NAME)-dev $(BINARY_PATH)

# Build with race detector for development
build-race:
	@echo "Building with race detector..."
	$(GOBUILD) -race $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-race $(BINARY_PATH)

# Run the application
run:
	@echo "Building and running $(BINARY_NAME)..."
	$(GOBUILD) -o $(BINARY_NAME)-dev $(BINARY_PATH) && ./$(BINARY_NAME)-dev

# Development setup (install common tools)
dev-setup:
	@echo "Setting up development environment..."
	$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOCMD) install golang.org/x/tools/cmd/goimports@latest
	@echo "Development tools installed!"

# Clean all artifacts including dev binaries
clean-all: clean
	rm -f $(BINARY_NAME) $(BINARY_NAME)-dev coverage.out coverage.html

# Generate shell completions
completions: build
	@echo "Generating shell completions..."
	@chmod +x scripts/generate-completions.sh
	@./scripts/generate-completions.sh $(BUILD_DIR)/$(BINARY_NAME) completions

# Install weekly automation (Linux only)
install-automation:
	@echo "Installing weekly wallpaper automation..."
	@chmod +x scripts/install-weekly-automation.sh
	@./scripts/install-weekly-automation.sh install

# Uninstall weekly automation
uninstall-automation:
	@echo "Uninstalling weekly wallpaper automation..."
	@chmod +x scripts/install-weekly-automation.sh
	@./scripts/install-weekly-automation.sh uninstall

# Show automation status
automation-status:
	@echo "Checking automation status..."
	@chmod +x scripts/install-weekly-automation.sh
	@./scripts/install-weekly-automation.sh status

# Show help
help:
	@echo "WallFetch - Professional Wallpaper Management for Linux"
	@echo ""
	@echo "Available targets:"
	@echo ""
	@echo "BUILD TARGETS:"
	@echo "  build              - Build the binary"
	@echo "  build-all          - Build for multiple platforms"
	@echo "  build-race         - Build with race detector"
	@echo "  dev-build          - Quick build for development"
	@echo ""
	@echo "TEST & QUALITY:"
	@echo "  test               - Run tests"
	@echo "  test-cover         - Run tests with coverage report"
	@echo "  test-verbose       - Run tests verbosely"
	@echo "  fmt                - Format Go code"
	@echo "  vet                - Vet Go code"
	@echo "  lint               - Lint code with golangci-lint"
	@echo "  check              - Run fmt, vet, lint, and test"
	@echo ""
	@echo "DEPENDENCIES:"
	@echo "  deps               - Install dependencies"
	@echo "  update-deps        - Update dependencies to latest versions"
	@echo ""
	@echo "INSTALLATION:"
	@echo "  install            - Install binary system-wide (requires sudo)"
	@echo "  install-local      - Install binary locally (~/bin, no sudo needed)"
	@echo "  uninstall          - Remove binary from system"
	@echo ""
	@echo "DEVELOPMENT:"
	@echo "  run                - Build and run the application"
	@echo "  dev-setup          - Install development tools"
	@echo "  completions        - Generate shell completions"
	@echo ""
	@echo "AUTOMATION (Linux only):"
	@echo "  install-automation - Install weekly wallpaper automation"
	@echo "  uninstall-automation - Remove weekly wallpaper automation"
	@echo "  automation-status  - Show automation status"
	@echo ""
	@echo "CLEANUP:"
	@echo "  clean              - Clean build artifacts"
	@echo "  clean-all          - Clean all artifacts including dev files"
	@echo ""
	@echo "OTHER:"
	@echo "  help               - Show this help message" 
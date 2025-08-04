# WallFetch Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary info
BINARY_NAME=wallfetch
BINARY_PATH=./cmd/wallfetch
BUILD_DIR=build

# Install paths
PREFIX?=/usr/local
BINDIR=$(PREFIX)/bin

# Build flags
LDFLAGS=-ldflags "-s -w"

.PHONY: all build clean test install uninstall deps update-deps completions install-automation uninstall-automation automation-status

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
	$(GOTEST) -v ./...

# Install dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Update dependencies
update-deps:
	$(GOMOD) tidy
	$(GOGET) -u ./...

# Install the binary
install: build
	@echo "Installing $(BINARY_NAME) to $(BINDIR)..."
	@mkdir -p $(BINDIR)
	cp $(BUILD_DIR)/$(BINARY_NAME) $(BINDIR)/$(BINARY_NAME)
	@echo "Installation complete!"

# Uninstall the binary
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	rm -f $(BINDIR)/$(BINARY_NAME)
	@echo "Uninstall complete!"

# Development commands
dev-build:
	$(GOBUILD) -o $(BINARY_NAME) $(BINARY_PATH)

# Run the application
run:
	$(GOBUILD) -o $(BINARY_NAME) $(BINARY_PATH) && ./$(BINARY_NAME)

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
	@echo "Available targets:"
	@echo "  build              - Build the binary"
	@echo "  build-all          - Build for multiple platforms"
	@echo "  clean              - Clean build artifacts"
	@echo "  test               - Run tests"
	@echo "  deps               - Install dependencies"
	@echo "  update-deps        - Update dependencies"
	@echo "  install            - Install binary to $(BINDIR)"
	@echo "  uninstall          - Remove binary from $(BINDIR)"
	@echo "  dev-build          - Quick build for development"
	@echo "  run                - Build and run the application"
	@echo "  completions        - Generate shell completions"
	@echo "  install-automation - Install weekly wallpaper automation (Linux)"
	@echo "  uninstall-automation - Remove weekly wallpaper automation"
	@echo "  automation-status  - Show automation status"
	@echo "  help               - Show this help message" 
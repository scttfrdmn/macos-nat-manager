# macOS NAT Manager Makefile

# Variables
BINARY_NAME=nat-manager
PACKAGE=github.com/scttfrdmn/macos-nat-manager
VERSION=$(shell git describe --tags --always --dirty)
COMMIT=$(shell git rev-parse HEAD)
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_TIME)"

# Go build flags
GOFLAGS=-v

.PHONY: help build clean test install uninstall deps check fmt lint release homebrew

# Default target
all: build

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: deps ## Build the binary
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	go build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME) cmd/nat-manager/main.go
	@echo "Build complete: ./$(BINARY_NAME)"

build-release: deps ## Build optimized release binary
	@echo "Building release $(BINARY_NAME) $(VERSION)..."
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_NAME) cmd/nat-manager/main.go
	strip $(BINARY_NAME)
	@echo "Release build complete: ./$(BINARY_NAME)"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f dist/*
	rm -rf build/
	go clean

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

check: deps ## Run all checks (lint, vet, fmt)
	@echo "Running checks..."
	@make fmt
	@make lint
	@make vet

fmt: ## Format Go code
	@echo "Formatting code..."
	go fmt ./...

lint: ## Run golint
	@echo "Running golint..."
	@if command -v golint >/dev/null 2>&1; then \
		golint ./...; \
	else \
		echo "golint not installed, skipping..."; \
	fi

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

install: build ## Install binary to system
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo cp $(BINARY_NAME) /usr/local/bin/
	sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "Installation complete. Run with: sudo $(BINARY_NAME)"

uninstall: ## Remove binary from system
	@echo "Removing $(BINARY_NAME) from system..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Uninstallation complete."

install-deps: ## Install development dependencies
	@echo "Installing development dependencies..."
	@if ! command -v golint >/dev/null 2>&1; then \
		echo "Installing golint..."; \
		go install golang.org/x/lint/golint@latest; \
	fi
	@if ! command -v dnsmasq >/dev/null 2>&1; then \
		echo "Installing dnsmasq..."; \
		brew install dnsmasq; \
	fi

# Release targets
release: clean build-release ## Create a release
	@echo "Creating release $(VERSION)..."
	mkdir -p dist
	cp $(BINARY_NAME) dist/
	cp LICENSE dist/
	cp README.md dist/
	cp CHANGELOG.md dist/
	cd dist && tar -czf $(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz *
	@echo "Release created: dist/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz"

homebrew: ## Generate Homebrew formula
	@echo "Generating Homebrew formula..."
	@./scripts/generate-homebrew.sh $(VERSION)

# Development targets
run: build ## Build and run with TUI
	sudo ./$(BINARY_NAME)

run-cli: build ## Build and run with CLI help
	sudo ./$(BINARY_NAME) --help

debug: ## Build with debug info
	@echo "Building debug version..."
	go build -gcflags="-N -l" $(LDFLAGS) -o $(BINARY_NAME)-debug cmd/nat-manager/main.go
	@echo "Debug build complete: ./$(BINARY_NAME)-debug"

# Docker targets (for CI/testing)
docker-build: ## Build in Docker (for CI)
	docker build -t $(BINARY_NAME):$(VERSION) .

docker-test: ## Run tests in Docker
	docker run --rm $(BINARY_NAME):$(VERSION) make test

# Git hooks
install-hooks: ## Install git hooks
	@echo "Installing git hooks..."
	cp scripts/pre-commit .git/hooks/
	chmod +x .git/hooks/pre-commit

# Project setup
setup: deps install-deps install-hooks ## Set up development environment
	@echo "Development environment setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "1. Run 'make build' to build the project"
	@echo "2. Run 'make test' to run tests"
	@echo "3. Run 'sudo make run' to test the application"

# Show version info
version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build time: $(BUILD_TIME)"

# Quick development cycle
dev: clean build test ## Quick development cycle: clean, build, test
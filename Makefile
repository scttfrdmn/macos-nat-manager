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

test: test-unit test-security ## Run all safe tests (no root required)
	@echo "✅ All safe tests completed"

test-unit: ## Run unit tests only
	@echo "🧪 Running unit tests..."
	go test -v ./internal/config ./internal/nat ./internal/tui

test-integration: ## Run integration tests (requires root)
	@echo "🔧 Running integration tests (requires root)..."
	@if [ "$$(id -u)" != "0" ]; then \
		echo "❌ Integration tests require root privileges"; \
		echo "   Option 1: sudo make test-integration"; \
		echo "   Option 2: make test-integration-askpass (uses ASKPASS)"; \
		exit 1; \
	fi
	@go test -v ./test/integration/...

test-integration-askpass: ensure-askpass ## Run integration tests using external ASKPASS
	@echo "🔧 Running integration tests with external ASKPASS..."
	@if [ -z "$${SUDO_ASKPASS:-}" ]; then \
		echo "⚡ Setting up ASKPASS environment..."; \
		export SUDO_ASKPASS="$$(which askpass)"; \
	fi
	@SUDO_ASKPASS="$$(which askpass)" sudo -A go test -v ./test/integration/...

test-security: ## Run security tests  
	@echo "🔒 Running security tests..."
	@go test -v ./test/security/...

test-coverage: ## Run unit tests with coverage
	@echo "📊 Running tests with coverage..."
	go test -coverprofile=coverage.out ./internal/config ./internal/nat ./internal/tui
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out | tail -1
	@echo "📈 Coverage report generated: coverage.html"

test-race: ## Run tests with race detection
	@echo "🏃 Running tests with race detection..."
	go test -race ./internal/...

test-all: test-unit test-security ## Run complete test suite (requires root for integration)
	@echo "🔧 Running integration tests (requires root)..."
	@sudo go test -v ./test/integration/...
	@echo "✅ Complete test suite finished"

test-all-askpass: test-unit test-security ensure-askpass ## Run complete test suite using external ASKPASS
	@echo "🔧 Running integration tests with external ASKPASS..."
	@SUDO_ASKPASS="$$(which askpass)" sudo -A go test -v ./test/integration/...
	@echo "✅ Complete test suite with external ASKPASS finished"

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

security-scan: ## Run security scanners
	@echo "🛡️  Running security scanners..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "⚠️  gosec not installed. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi
	@make test-security

deps-audit: ## Audit dependencies for vulnerabilities
	@echo "🔍 Auditing dependencies..."
	@if command -v nancy >/dev/null 2>&1; then \
		go list -json -m all | nancy sleuth; \
	else \
		echo "⚠️  nancy not installed for vulnerability scanning"; \
		echo "   Install with: go install github.com/sonatypecommunity/nancy@latest"; \
	fi
	@go mod verify

quality-check: ## Run pre-commit quality checks
	@echo "⚡ Running pre-commit quality checks..."
	@./.git/hooks/pre-commit

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

install-security-tools: ## Install security scanning tools
	@echo "🔧 Installing security tools..."
	@go install golang.org/x/lint/golint@latest
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@go install github.com/sonatypecommunity/nancy@latest
	@go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	@go install github.com/client9/misspell/cmd/misspell@latest
	@go install github.com/gordonklaus/ineffassign@latest
	@echo "✅ Security tools installed"

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

# External ASKPASS dependency management
ensure-askpass: ## Ensure external ASKPASS is installed
	@echo "🔧 Checking for external ASKPASS..."
	@if ! command -v askpass >/dev/null 2>&1; then \
		echo "❌ External ASKPASS not found!"; \
		echo ""; \
		echo "Install options:"; \
		echo "  1. Homebrew: brew tap scttfrdmn/macos-askpass && brew install macos-askpass"; \
		echo "  2. Direct:   curl -fsSL https://raw.githubusercontent.com/scttfrdmn/macos-askpass/main/install.sh | bash"; \
		echo "  3. Project:  https://github.com/scttfrdmn/macos-askpass"; \
		echo ""; \
		exit 1; \
	fi
	@echo "✅ External ASKPASS found: $$(which askpass)"
	@echo "   Version: $$(askpass version | head -1)"

install-askpass: ## Install external ASKPASS via Homebrew
	@echo "🍺 Installing external ASKPASS via Homebrew..."
	@if ! command -v brew >/dev/null 2>&1; then \
		echo "❌ Homebrew not found. Install from https://brew.sh first"; \
		exit 1; \
	fi
	@brew tap scttfrdmn/macos-askpass
	@brew install macos-askpass
	@echo "✅ External ASKPASS installed successfully"

test-askpass: ensure-askpass ## Test external ASKPASS functionality
	@echo "🧪 Testing external ASKPASS functionality..."
	@askpass test

# Project setup
setup: deps install-deps install-hooks ## Set up development environment
	@echo "Development environment setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "1. Run 'make build' to build the project"
	@echo "2. Run 'make test' to run tests"
	@echo "3. Run 'sudo make run' to test the application"
	@echo "4. Run 'make install-askpass' for automated testing support"

# Show version info
version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build time: $(BUILD_TIME)"

# Quick development cycle
dev: clean build test ## Quick development cycle: clean, build, test
# Testing Strategy for macOS NAT Manager

This document outlines the comprehensive testing approach for the NAT Manager project, addressing unit tests, integration tests, security tests, and end-to-end testing scenarios.

## 🧪 Testing Pyramid

```
    🔺 E2E/Integration (Manual/CI-Optional)
   🔺🔺 Security & Dependency Tests
  🔺🔺🔺 Functional Unit Tests (CI)
 🔺🔺🔺🔺 Static Analysis & Linting (Pre-commit)
```

## 📋 Test Categories

### 1. Unit Tests (Automated - CI/CD)

**Location**: `internal/*/test.go`  
**Command**: `go test ./internal/...`  
**Coverage**: ~65% focused on critical functionality

**What we test**:
- ✅ Data structure validation
- ✅ Configuration loading/saving
- ✅ Error handling and edge cases
- ✅ Utility functions and formatters
- ✅ Manager lifecycle operations
- ✅ Network interface parsing

**What we DON'T test** (requires root):
- ❌ Actual network configuration changes
- ❌ pfctl rule creation/removal
- ❌ Bridge interface creation
- ❌ DHCP server startup
- ❌ IP forwarding changes

### 2. Integration Tests (Manual - Root Required)

**Location**: `test/integration/`  
**Command**: `sudo go test ./test/integration/...`

**What we test**:
- 🔧 Full NAT lifecycle (start/stop/status)
- 🔧 Real network interface creation/destruction
- 🔧 Bridge interface configuration
- 🔧 pfctl rule verification
- 🔧 DHCP server functionality
- 🔧 Configuration persistence
- 🔧 System cleanup after operations

**When to run**:
- Before releases
- After significant network-related changes
- On development machines with proper permissions
- In dedicated testing VMs/containers

### 3. Security Tests (Automated - CI/CD Safe)

**Location**: `test/security/`  
**Command**: `go test ./test/security/...`

**What we test**:
- 🔒 Source code scanning for hardcoded secrets
- 🔒 Input validation against injection attacks
- 🔒 Configuration file permissions
- 🔒 Dependency vulnerability scanning
- 🔒 Race condition testing
- 🔒 Privilege escalation prevention

## 🚀 Running Tests

### Quick Test Suite (CI/CD Compatible)
```bash
# Run all safe tests (no root required)
make test

# With coverage report
make test-coverage

# Security scanning
make test-security
```

### Full Integration Testing (Root Required)

#### Traditional sudo approach
```bash
# Complete integration test suite
sudo make test-integration

# Single integration test
sudo go test ./test/integration/ -v -run TestNATFullLifecycle

# Run with race detection
sudo go test ./test/integration/ -race
```

#### Using External ASKPASS (Automated/CI-Friendly)
```bash
# Install external ASKPASS (one-time setup)
make install-askpass

# Or install manually:
# brew tap scttfrdmn/macos-askpass && brew install macos-askpass

# Run integration tests with ASKPASS
make test-integration-askpass

# Complete test suite with ASKPASS
make test-all-askpass

# Test ASKPASS functionality
make test-askpass
```

**External ASKPASS Setup:**
```bash
# For local development - interactive setup
askpass setup

# For CI/CD - use environment variables
export CI_SUDO_PASSWORD="your_password"
export SUDO_ASKPASS="$(which askpass)"

# Enable debug mode
export ASKPASS_DEBUG=1

# Force CLI mode (disable GUI dialogs)
export ASKPASS_FORCE_CLI=1
```

**External ASKPASS Project:**
- **Repository**: https://github.com/scttfrdmn/macos-askpass
- **Documentation**: Complete setup guides and examples
- **Features**: GUI dialogs, keychain integration, multi-source auth

### Manual E2E Testing Scenarios

#### Scenario 1: Complete NAT Setup and Teardown
```bash
# 1. Start NAT service
sudo nat-manager start -e en0 -i bridge200 -n 192.168.200

# 2. Verify configuration
sudo nat-manager status
sudo nat-manager interfaces
sudo pfctl -s nat

# 3. Connect a device to the bridge (requires additional setup)
# 4. Verify internet connectivity through NAT
# 5. Monitor connections
sudo nat-manager monitor --follow

# 6. Clean teardown
sudo nat-manager stop
sudo nat-manager status # Should show inactive
```

#### Scenario 2: Configuration Management
```bash
# Test configuration persistence
sudo nat-manager start -e en0 -i bridge201 -n 192.168.201
sudo nat-manager stop

# Verify config is saved
cat ~/.config/nat-manager/config.yaml

# Test loading saved config
sudo nat-manager start # Should use saved config
```

#### Scenario 3: Error Handling and Recovery
```bash
# Test with invalid interface
sudo nat-manager start -e nonexistent0 -i bridge202 -n 192.168.202
# Should fail gracefully

# Test cleanup after failure
sudo nat-manager status # Should show clean state

# Test force cleanup
sudo nat-manager stop --force
```

## 🔍 Security Testing Details

### Dependency Scanning
```bash
# Check for known vulnerabilities
go list -json -m all | nancy sleuth

# Audit Go modules
go mod audit

# Check for outdated dependencies
go list -u -m all
```

### Static Security Analysis
```bash
# Run gosec security scanner
gosec ./...

# Check for common security issues
go vet ./...

# Custom security tests
go test ./test/security/...
```

### Manual Security Review Checklist

- [ ] No hardcoded passwords or secrets in source code
- [ ] All external inputs are validated and sanitized
- [ ] No command injection vulnerabilities
- [ ] Proper error handling (no information disclosure)
- [ ] Configuration files have appropriate permissions
- [ ] No unnecessary system calls or privilege escalation
- [ ] Proper cleanup on failures
- [ ] Race condition protection for concurrent operations

## 🎯 Testing Best Practices

### For Developers
1. **Write tests first** for new functionality
2. **Test error paths** as thoroughly as success paths
3. **Use table-driven tests** for multiple scenarios
4. **Mock external dependencies** in unit tests
5. **Test edge cases** and boundary conditions

### For CI/CD
1. **Run unit tests** on every commit
2. **Run security tests** on every PR
3. **Run integration tests** on release candidates
4. **Fail fast** on security issues
5. **Generate coverage reports** for visibility

### For Manual Testing
1. **Test on clean systems** to verify setup/teardown
2. **Test with different network configurations**
3. **Test error recovery scenarios**
4. **Test with real network traffic**
5. **Verify system state after operations**

## 🐛 Debugging Failed Tests

### Unit Test Failures
```bash
# Run specific test with verbose output
go test ./internal/nat -v -run TestSpecificFunction

# Run with race detection
go test ./internal/nat -race

# Generate detailed coverage
go test ./internal/nat -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Integration Test Failures
```bash
# Check system state
sudo pfctl -s nat
sudo pfctl -s state
ifconfig bridge200

# Check DHCP server
sudo lsof -i :67
ps aux | grep dnsmasq

# Manual cleanup if needed
sudo pfctl -d
sudo killall dnsmasq
sudo ifconfig bridge200 destroy
sudo sysctl -w net.inet.ip.forwarding=0
```

### Security Test Failures
```bash
# Review flagged code
grep -r "suspicious_pattern" internal/

# Check file permissions
find . -name "*.yaml" -exec ls -la {} \;

# Review dependencies
go mod graph | grep vulnerable_package
```

### ASKPASS Test Failures
```bash
# Check ASKPASS script
./scripts/setup-askpass.sh test

# Verify environment variables
echo "SUDO_ASKPASS: $SUDO_ASKPASS"
echo "CI_SUDO_PASSWORD: ${CI_SUDO_PASSWORD:+SET}"

# Test password retrieval
$SUDO_ASKPASS  # Should output password without error

# Check keychain (local development)
security find-generic-password -a "$USER" -s "nat-manager-sudo" -w

# Debug ASKPASS execution
export NAT_ASKPASS_DEBUG=1
sudo -A echo "ASKPASS test"
```

## 🔐 External ASKPASS Integration

### Architecture Overview
The NAT Manager uses the external **macOS ASKPASS** project for automated sudo authentication:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   NAT Manager   │    │ External ASKPASS│    │ Password Source │
│                 ├───►│                 ├───►│                 │
│ make test-*     │    │ askpass binary  │    │ Multi-source    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### External Project Features
- **Multi-source Authentication**: CI vars → Local vars → Keychain → GUI Dialog → Terminal
- **Smart Environment Detection**: Automatically detects GUI vs CLI environments
- **Native GUI Dialogs**: macOS password dialogs for interactive use
- **Security First**: No persistent password storage, comprehensive validation
- **Zero Dependencies**: Pure bash implementation, works everywhere

### Integration Benefits

#### Clean Architecture
- **Separation of Concerns**: NAT Manager focuses on networking, ASKPASS handles authentication
- **Reusable Component**: ASKPASS can be used by other projects
- **External Maintenance**: Security updates and features handled by dedicated project

#### Enhanced Capabilities
- **GUI Support**: Native macOS dialogs for interactive development
- **Better Documentation**: Dedicated project with comprehensive guides
- **Community Driven**: Standalone project encourages contributions

### Setup and Usage

#### Installation
```bash
# Via NAT Manager convenience target
make install-askpass

# Or manually via Homebrew
brew tap scttfrdmn/macos-askpass
brew install macos-askpass

# Or direct installation
curl -fsSL https://raw.githubusercontent.com/scttfrdmn/macos-askpass/main/install.sh | bash
```

#### Configuration
- **Project Repository**: https://github.com/scttfrdmn/macos-askpass
- **Setup Guide**: Complete documentation and examples available
- **Security Analysis**: Comprehensive threat model and best practices

## 📈 Coverage Goals

| Package | Current | Target | Priority |
|---------|---------|---------|----------|
| `nat` | 58.7% | 70% | High |
| `config` | 70.2% | 75% | Medium |
| `cli` | Partial* | 60%* | Medium |
| `tui` | 16.5% | 40% | Low |

*CLI coverage limited by root requirement for real functionality

## 🔄 Continuous Improvement

### Test Automation Roadmap
- [ ] Set up GitHub Actions for automated testing
- [ ] Integrate security scanning in CI pipeline  
- [ ] Add performance benchmarking
- [ ] Create Docker containers for isolated integration testing
- [ ] Set up automated dependency vulnerability scanning

### Test Coverage Improvements
- [ ] Add more error condition testing
- [ ] Increase configuration validation coverage
- [ ] Add performance and load testing
- [ ] Create mock network interfaces for broader CLI testing
- [ ] Add chaos engineering scenarios

---

**Remember**: The goal is not 100% coverage, but 100% confidence that the critical functionality works correctly and securely.
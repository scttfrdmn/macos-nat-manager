# ASKPASS Setup Guide for macOS NAT Manager

This guide explains how to set up automated sudo authentication using ASKPASS for testing the macOS NAT Manager, enabling CI/CD integration and streamlined local development.

## 🎯 Overview

ASKPASS allows automated sudo authentication without interactive password prompts, essential for:
- **CI/CD Pipelines**: GitHub Actions, Jenkins, etc.
- **Automated Testing**: Integration tests requiring root privileges
- **Development Workflow**: Streamlined local testing

## 🚀 Quick Start

### 1. Set Up ASKPASS Environment
```bash
# Clone and navigate to project
git clone https://github.com/scttfrdmn/macos-nat-manager
cd macos-nat-manager

# Set up ASKPASS (interactive setup)
make setup-askpass
```

### 2. Test ASKPASS Functionality
```bash
# Test ASKPASS configuration
make test-askpass

# Run integration tests with ASKPASS
make test-integration-askpass
```

### 3. Run Complete Test Suite
```bash
# Full test suite with ASKPASS
make test-all-askpass
```

## 🔧 Configuration Options

### Local Development Setup

#### Option 1: Keychain Storage (Recommended)
```bash
# Interactive setup with keychain storage
./scripts/setup-askpass.sh setup

# Follow prompts to store password securely
```

#### Option 2: Environment Variables
```bash
# Set password via environment variable
export SUDO_PASSWORD="your_password"
export SUDO_ASKPASS="$(pwd)/scripts/askpass.sh"

# Test configuration
make test-askpass
```

### CI/CD Environment Setup

#### GitHub Actions
```yaml
# .github/workflows/ci.yml
- name: Set up ASKPASS environment
  run: |
    chmod +x scripts/askpass.sh
    export SUDO_ASKPASS="$(pwd)/scripts/askpass.sh"
    echo "SUDO_ASKPASS=$(pwd)/scripts/askpass.sh" >> $GITHUB_ENV

- name: Run integration tests
  env:
    CI_SUDO_PASSWORD: ${{ secrets.MACOS_SUDO_PASSWORD }}
  run: make test-integration-askpass
```

#### Jenkins/Other CI Systems
```bash
# Set environment variables
export CI_SUDO_PASSWORD="${SUDO_PASSWORD_SECRET}"
export SUDO_ASKPASS="$(pwd)/scripts/askpass.sh"

# Run tests
make test-integration-askpass
```

## 🛠️ Available Commands

| Command | Description |
|---------|-------------|
| `make setup-askpass` | Interactive ASKPASS setup |
| `make test-askpass` | Test ASKPASS functionality |
| `make test-integration-askpass` | Run integration tests with ASKPASS |
| `make test-all-askpass` | Complete test suite with ASKPASS |
| `make clean-askpass` | Remove ASKPASS configuration |

## 🔍 Troubleshooting

### Common Issues

#### 1. "sudo: no askpass program specified"
```bash
# Solution: Ensure SUDO_ASKPASS is set
export SUDO_ASKPASS="$(pwd)/scripts/askpass.sh"
```

#### 2. "ASKPASS: Failed to retrieve password"
```bash
# Check password sources
echo "Environment: ${SUDO_PASSWORD:+SET}"
echo "CI Password: ${CI_SUDO_PASSWORD:+SET}"

# Check keychain (local)
security find-generic-password -a "$USER" -s "nat-manager-sudo" -w
```

#### 3. Permission Denied
```bash
# Ensure script is executable
chmod +x scripts/askpass.sh scripts/setup-askpass.sh
```

### Debug Mode

Enable detailed logging:
```bash
export NAT_ASKPASS_DEBUG=1
make test-askpass
```

Sample debug output:
```
ASKPASS: Called by sudo (PID: 12345)
ASKPASS: User: username
ASKPASS: PWD: /path/to/project
✅ ASKPASS script executes without error
```

### Manual Testing

Test ASKPASS script directly:
```bash
# Test password retrieval
./scripts/askpass.sh

# Test with sudo
sudo -A echo "ASKPASS working"
```

## 🔐 Security Considerations

### Password Storage Priority
1. **CI_SUDO_PASSWORD** environment variable (CI/CD)
2. **SUDO_PASSWORD** environment variable (local)
3. **macOS Keychain** (local development)
4. **Interactive prompt** (fallback)

### Best Practices
- ✅ Use repository secrets for CI/CD passwords
- ✅ Store local passwords in macOS Keychain
- ✅ Rotate passwords regularly
- ✅ Limit secret scope to specific repositories
- ❌ Never commit passwords to version control
- ❌ Avoid plain text password files

### Keychain Security
```bash
# View stored password
security find-generic-password -a "$USER" -s "nat-manager-sudo" -w

# Remove stored password
security delete-generic-password -a "$USER" -s "nat-manager-sudo"
```

## 🏗️ Architecture

### Component Overview
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Make Target   │    │  ASKPASS Script │    │ Password Source │
│                 │    │                 │    │                 │
│ test-*-askpass  ├───►│ scripts/        ├───►│ ENV/Keychain/   │
│                 │    │ askpass.sh      │    │ Interactive     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Script Flow
1. **sudo -A** command triggers ASKPASS script
2. Script checks password sources in priority order
3. Password returned to sudo for authentication
4. Test execution continues with root privileges

### File Structure
```
scripts/
├── askpass.sh              # Main ASKPASS implementation
├── setup-askpass.sh        # Interactive setup script
└── askpass-env.sh          # Generated environment file
```

## 🔄 Integration Examples

### Local Development Workflow
```bash
# One-time setup
make setup-askpass

# Daily development
make test-integration-askpass
make build
sudo -A ./nat-manager start -e en0
```

### CI/CD Pipeline Integration
```bash
# In your CI script
export SUDO_ASKPASS="$(pwd)/scripts/askpass.sh"
export CI_SUDO_PASSWORD="${SECRETS_SUDO_PASSWORD}"

# Run tests
make test-all-askpass

# Build and test
make build
sudo -A ./nat-manager --version
```

## 📝 Environment File Example

After running `make setup-askpass`, an environment file is created:
```bash
# scripts/askpass-env.sh
export SUDO_ASKPASS="/path/to/project/scripts/askpass.sh"
# export NAT_ASKPASS_DEBUG=1  # Uncomment for debugging
```

Source this file for consistent environment:
```bash
source scripts/askpass-env.sh
```

## 🆘 Support

If you encounter issues with ASKPASS setup:

1. **Check Prerequisites**: Ensure you have sudo privileges on the system
2. **Verify Scripts**: Run `make setup-askpass` to validate installation
3. **Test Configuration**: Use `make test-askpass` to verify setup
4. **Enable Debugging**: Set `NAT_ASKPASS_DEBUG=1` for detailed logging
5. **Review Logs**: Check terminal output for specific error messages

For additional help, see [TESTING.md](../TESTING.md) or open an issue on GitHub.
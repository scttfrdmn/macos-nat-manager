# Changelog

All notable changes to macOS NAT Manager will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Homebrew tap distribution (`scttfrdmn/macos-nat-manager`)
- Integration with macos-askpass for automated testing
- Comprehensive security testing framework
- Enhanced integration tests requiring root privileges
- Complete dependency management via Homebrew

### Changed
- Refactored ASKPASS implementation to use external macos-askpass project
- Improved testing architecture with separate unit and integration test suites
- Updated documentation with Homebrew installation instructions
- Enhanced GoReleaser configuration for automated releases

### Security
- Added security vulnerability scanning for dependencies
- Implemented input validation and sanitization tests
- Added privilege escalation prevention checks
- Enhanced configuration file security validation

## [1.0.0] - TBD

### Added
- Interactive Terminal User Interface (TUI) with bubbletea
- Full command-line interface with cobra
- True NAT functionality using pfctl (not bridging)
- Network interface management and selection
- YAML-based configuration with validation
- Real-time connection and device monitoring
- Automatic bridge interface creation and cleanup
- DNS forwarding and resolution
- DHCP server integration with dnsmasq
- Comprehensive error handling and validation
- Cross-architecture support (Intel and Apple Silicon)
- Shell completion support (bash, zsh, fish)
- Professional logging and debugging features

### Features
- **True NAT**: Actual address translation, not transparent bridging
- **Privacy**: Complete network isolation - internal devices hidden from upstream
- **802.1x Compatible**: Appears as single device to enterprise networks
- **Monitoring**: Real-time connection tracking and device discovery
- **Configuration**: Persistent YAML configuration with validation
- **Automation**: Full CLI support for scripts and automation
- **Security**: Input validation, privilege management, clean teardown

### Dependencies
- **macOS**: 12.0+ (Monterey or later)
- **Go**: 1.21+ for building from source
- **dnsmasq**: DHCP server functionality
- **macos-askpass**: Automated sudo authentication for testing
- **pfctl**: Built into macOS (packet filter control)

### Installation Methods
- **Homebrew**: `brew install scttfrdmn/macos-nat-manager/nat-manager`
- **Direct Binary**: Download from GitHub releases
- **Source**: `go install github.com/scttfrdmn/macos-nat-manager/cmd/nat-manager@latest`

### Usage Examples
```bash
# Interactive mode
sudo nat-manager

# CLI mode
sudo nat-manager start --external en0 --internal bridge100 --network 192.168.100

# Monitor connections
sudo nat-manager status

# Clean shutdown
sudo nat-manager stop
```

## [0.1.0] - Development

### Added
- Initial project structure
- Core NAT functionality proof of concept
- Basic pfctl integration
- Configuration framework foundation
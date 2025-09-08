# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- Web-based management interface
- Configuration profiles/presets
- Port forwarding rules
- Traffic shaping/QoS
- Multi-interface support
- Logging and analytics
- Shell completion scripts

## [1.0.0] - 2025-09-03

### Added
- Initial release of macOS NAT Manager
- **TUI Interface**: Interactive terminal interface using bubbletea
- **CLI Interface**: Full command-line interface using cobra
- **True NAT Implementation**: Address translation using pfctl (not bridging)
- **Interface Management**: Easy selection and configuration of network interfaces
- **DHCP Server**: Built-in DHCP using dnsmasq
- **Real-time Monitoring**: Live connection and device monitoring
- **Configuration Management**: YAML-based configuration with validation
- **Clean Setup/Teardown**: Proper cleanup of all network changes
- **Homebrew Installation**: Professional Homebrew formula

#### Core Features
- Start/stop NAT service with configurable parameters
- List and select network interfaces
- Configure IP ranges and DHCP settings
- Monitor active connections and connected devices
- Status reporting with detailed system information

#### CLI Commands
- `nat-manager` - Launch TUI interface
- `nat-manager start` - Start NAT service
- `nat-manager stop` - Stop NAT service  
- `nat-manager status` - Show NAT status
- `nat-manager interfaces` - List network interfaces
- `nat-manager monitor` - Monitor connections

#### Technical Implementation
- **NAT Rules**: Uses macOS pfctl for proper address translation
- **DHCP Server**: Integrates with dnsmasq for IP assignment
- **Interface Creation**: Automatic bridge interface creation
- **IP Forwarding**: Controlled kernel IP forwarding
- **State Management**: Persistent configuration and runtime state
- **Error Handling**: Comprehensive error handling and recovery

#### Platform Support
- **macOS**: Full support for macOS 12+ (Monterey and later)
- **Architecture**: Native support for Intel and Apple Silicon

#### Installation Methods  
- **Homebrew**: `brew install scttfrdmn/macos-nat-manager/nat-manager`
- **Source**: Build from source with Go 1.21+
- **Binary**: Download from GitHub releases

### Security
- All network changes are temporary and cleaned up
- No permanent system modifications
- Requires explicit sudo privileges
- pfctl rules isolated to NAT manager
- Configuration stored in user directory

### Documentation
- Comprehensive README with setup instructions
- Built-in help system and command documentation  
- Troubleshooting guide with common issues
- Example configurations and use cases

---

## Version History

### Pre-1.0.0 Development
- [0.9.0] - Beta release with core functionality
- [0.8.0] - Alpha release for testing
- [0.7.0] - Initial TUI implementation  
- [0.6.0] - CLI framework setup
- [0.5.0] - NAT engine development
- [0.4.0] - Interface management
- [0.3.0] - DHCP integration
- [0.2.0] - Configuration system
- [0.1.0] - Project initialization

---

## Release Notes

### 1.0.0 Release Notes

This is the first stable release of macOS NAT Manager, providing a complete solution for true NAT functionality on macOS. Unlike macOS's built-in Internet Sharing which operates as a transparent bridge, this tool provides proper Network Address Translation with complete privacy and network isolation.

**Key Highlights:**
- **True NAT**: Actual address translation, not bridging
- **Privacy**: Internal devices completely hidden from upstream network  
- **Professional**: Enterprise-ready with comprehensive error handling
- **User-Friendly**: Both TUI and CLI interfaces for different use cases
- **Reliable**: Extensive testing and validation on macOS systems

**Upgrade Path:**
- This is the initial release, no upgrade considerations

**Breaking Changes:**
- N/A (initial release)

**Deprecations:**
- N/A (initial release)

**Migration Guide:**
- For users coming from manual pfctl/dnsmasq setups, this tool provides a cleaner, more reliable alternative with automatic cleanup

---

## Contributors

- **Lead Developer**: Your Name (@scttfrdmn)
- **Contributors**: See [GitHub contributors](https://github.com/scttfrdmn/macos-nat-manager/graphs/contributors)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
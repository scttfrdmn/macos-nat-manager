# macOS NAT Manager

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![macOS](https://img.shields.io/badge/macOS-12+-green.svg)](https://www.apple.com/macos/)
[![Release](https://img.shields.io/github/release/scttfrdmn/macos-nat-manager.svg)](https://github.com/scttfrdmn/macos-nat-manager/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/scttfrdmn/macos-nat-manager)](https://goreportcard.com/report/github.com/scttfrdmn/macos-nat-manager)
[![Build Status](https://github.com/scttfrdmn/macos-nat-manager/workflows/CI/badge.svg)](https://github.com/scttfrdmn/macos-nat-manager/actions)
[![CodeQL](https://github.com/scttfrdmn/macos-nat-manager/workflows/CodeQL/badge.svg)](https://github.com/scttfrdmn/macos-nat-manager/actions/workflows/codeql.yml)
[![Quality Gate](https://img.shields.io/badge/Quality%20Gate-A+-brightgreen.svg)](https://github.com/scttfrdmn/macos-nat-manager)
[![Pre-commit](https://img.shields.io/badge/pre--commit-enabled-brightgreen?logo=pre-commit&logoColor=white)](https://github.com/pre-commit/pre-commit)

A comprehensive Network Address Translation (NAT) manager for macOS with both Terminal UI and CLI interfaces. Unlike macOS's built-in Internet Sharing which operates as a bridge, this tool creates **true NAT** with address translation, providing better privacy and network isolation.

## ✨ Features

- 🖥️ **Interactive TUI** - Beautiful terminal interface built with [bubbletea](https://github.com/charmbracelet/bubbletea)
- ⌨️ **Full CLI** - Complete command-line interface using [cobra](https://github.com/spf13/cobra)
- 🔀 **True NAT** - Actual address translation using pfctl, not transparent bridging
- 🌐 **Interface Management** - Easy selection and configuration of network interfaces
- ⚙️ **Flexible Configuration** - YAML-based config with validation and persistence
- 📊 **Real-time Monitoring** - Live connection and device monitoring
- 🔧 **Clean Setup/Teardown** - Proper cleanup with no permanent system changes
- 🛡️ **Network Isolation** - Internal devices completely hidden from upstream network
- 🍺 **Homebrew Ready** - Professional installation with dependency management

## 🆚 Why Not macOS Internet Sharing?

| Feature | macOS Internet Sharing | NAT Manager |
|---------|----------------------|-------------|
| **Operation** | Transparent Bridge | True NAT |
| **MAC Visibility** | ❌ Devices visible to upstream | ✅ Hidden behind single MAC |
| **802.1x Compatibility** | ❌ Easily detected | ✅ Appears as single device |
| **Privacy** | ❌ Limited | ✅ Full network isolation |
| **Monitoring** | ❌ No built-in tools | ✅ Real-time monitoring |
| **Configuration** | ❌ GUI only | ✅ CLI + TUI + Config files |

## 🚀 Quick Start

### Homebrew Installation (Recommended)

```bash
# Add tap and install
brew tap scttfrdmn/tap
brew install nat-manager

# Run with TUI interface
sudo nat-manager

# Or use CLI
sudo nat-manager start --external en0 --internal bridge100 --network 192.168.100
```

### Manual Installation

```bash
# Clone repository
git clone https://github.com/scttfrdmn/macos-nat-manager.git
cd macos-nat-manager

# Build and install
make setup
make build
sudo make install

# Run
sudo nat-manager
```

## 📖 Usage

### TUI Interface

Launch the interactive terminal interface:

```bash
sudo nat-manager
```

Navigate through menus to configure interfaces, start NAT, and monitor connections.

### CLI Interface

#### Start NAT Service

```bash
# Basic usage
sudo nat-manager start --external en0 --internal bridge100

# With custom network
sudo nat-manager start -e en0 -i bridge100 -n 192.168.100 \
  --dhcp-start 192.168.100.100 --dhcp-end 192.168.100.200

# With custom DNS
sudo nat-manager start -e en1 -i bridge101 -n 10.0.1 \
  --dns 1.1.1.1,1.0.0.1
```

#### Monitor and Manage

```bash
# Show status
sudo nat-manager status
sudo nat-manager status --json  # JSON output

# List interfaces
sudo nat-manager interfaces
sudo nat-manager interfaces --all  # Include inactive

# Monitor connections
sudo nat-manager monitor
sudo nat-manager monitor --follow --devices  # Continuous mode

# Stop service
sudo nat-manager stop
sudo nat-manager stop --force  # Force cleanup
```

#### Interface Management

```bash
# List all interfaces
sudo nat-manager interfaces

# Filter by type
sudo nat-manager interfaces --type bridge
sudo nat-manager interfaces --type ethernet
```

## ⚙️ Configuration

### Configuration File

NAT Manager uses YAML configuration stored at `~/.config/nat-manager/config.yaml`:

```yaml
external_interface: en0
internal_interface: bridge100
internal_network: 192.168.100
dhcp_range:
  start: 192.168.100.100
  end: 192.168.100.200
  lease: 12h
dns_servers:
  - 8.8.8.8
  - 8.8.4.4
```

### Environment Variables

- `NAT_MANAGER_CONFIG` - Custom config file path
- `NAT_MANAGER_VERBOSE` - Enable verbose logging

### Command Line Options

```bash
sudo nat-manager --help

Global Flags:
  --config string      config file (default: ~/.nat-manager.yaml)
  --verbose, -v        verbose output
  --config-path string path to store configuration
```

## 🏗️ Architecture

```
Internet → [External Interface] → NAT Engine → [Internal Interface] → Connected Devices
           (en0/en1/etc)         (pfctl)      (bridge100)          (192.168.x.x)
                                     ↕
                              [DHCP Server]
                               (dnsmasq)
```

### Technical Components

- **pfctl Integration** - macOS packet filter for NAT rules
- **dnsmasq** - DHCP and DNS services for internal network
- **Bridge Interfaces** - Virtual interfaces for internal networks
- **IP Forwarding** - Kernel-level packet forwarding
- **Interface Management** - Dynamic interface creation/destruction

## 🛠️ Development

### Prerequisites

- Go 1.21+
- macOS 12+ (Monterey or later)
- Homebrew (for dnsmasq)
- Root privileges

### Build from Source

```bash
# Clone and setup
git clone https://github.com/scttfrdmn/macos-nat-manager.git
cd macos-nat-manager
make setup

# Build
make build

# Run tests
make test

# Install development dependencies
make install-deps

# Quick development cycle
make dev  # clean, build, test
```

### Project Structure

```
macos-nat-manager/
├── cmd/nat-manager/     # Main application entry
├── internal/
│   ├── cli/            # Cobra CLI commands
│   ├── tui/            # Bubbletea TUI interface
│   ├── nat/            # NAT management logic
│   └── config/         # Configuration management
├── homebrew/           # Homebrew formula
├── docs/              # Documentation
└── scripts/           # Build and utility scripts
```

### Available Make Targets

```bash
make help          # Show all available targets
make build         # Build binary
make test          # Run tests
make install       # Install to system
make clean         # Clean build artifacts
make release       # Create release
make homebrew      # Generate Homebrew formula
```

## 🔧 Troubleshooting

### Common Issues

**"This tool requires root privileges"**
```bash
# Solution: Always use sudo
sudo nat-manager
```

**"dnsmasq not found"**
```bash
# Solution: Install dnsmasq
brew install dnsmasq
```

**"Failed to create bridge interface"**
```bash
# Solution: Use different bridge number
sudo nat-manager start -e en0 -i bridge101 -n 192.168.101
```

**No internet access for connected devices**
```bash
# Debug steps
sudo nat-manager status              # Check overall status
sudo pfctl -s nat                   # Check NAT rules
sysctl net.inet.ip.forwarding       # Check IP forwarding
ps aux | grep dnsmasq               # Check DHCP server
```

### Debug Commands

```bash
# Check NAT rules
sudo pfctl -s nat
sudo pfctl -s state

# Check IP forwarding
sysctl net.inet.ip.forwarding

# Check interfaces
ifconfig
sudo nat-manager interfaces --all

# Check DHCP
sudo lsof -i :67  # DHCP server port
```

### Clean Manual Cleanup

If something goes wrong:

```bash
# Stop everything
sudo nat-manager stop --force

# Manual cleanup
sudo pfctl -d                        # Disable pfctl
sudo killall dnsmasq                 # Stop DHCP
sudo ifconfig bridge100 destroy      # Remove bridge
sudo sysctl -w net.inet.ip.forwarding=0  # Disable forwarding
```

## 📊 Monitoring

### Built-in Monitoring

```bash
# Real-time monitoring
sudo nat-manager monitor --follow

# Show connected devices
sudo nat-manager monitor --devices

# JSON output for scripts
sudo nat-manager status --json
```

### Integration with System Tools

```bash
# Network statistics
netstat -rn                    # Routing table
netstat -i                     # Interface statistics
lsof -i                       # Network connections

# System monitoring
sudo fs_usage | grep nat-manager  # File system usage
sudo dtrace -n 'syscall:::entry /execname == "nat-manager"/'  # System calls
```

## 🔒 Security Considerations

- **Root Privileges** - Required for network configuration
- **Temporary Changes** - All modifications are reversed on exit
- **Isolated Rules** - pfctl rules don't interfere with other services
- **Clean State** - No permanent system modifications
- **Process Isolation** - Dedicated processes for each component

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass (`make test`)
6. Commit your changes (`git commit -am 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Create a Pull Request

### Code Quality Standards

This project maintains **A+ grade** equivalent to [Go Report Card](https://goreportcard.com/) standards:

#### Automated Quality Checks

The repository includes a comprehensive **pre-commit hook** that enforces the same quality standards as Go Report Card:

```bash
# Hook runs these checks automatically on every commit:
1. gofmt      - Code formatting (100% compliance)
2. go vet     - Static analysis (no issues)
3. golint     - Style guide compliance (warnings allowed)
4. gocyclo    - Cyclomatic complexity (≤15 per function)
5. misspell   - Spelling errors (zero tolerance)
6. ineffassign- Ineffectual assignments (zero tolerance)
```

#### Quality Requirements

- **Grade A or A+** - Only commits that achieve grade A (≥80%) or A+ (≥90%) are allowed
- **Comprehensive Testing** - All tests must pass before commit
- **Zero Static Issues** - `go vet` must report no problems
- **Complexity Control** - Functions with complexity >15 must be refactored

#### Manual Quality Checks

```bash
# Run quality checks manually
make lint          # Run all linters
make test          # Run test suite
make vet          # Static analysis
make fmt          # Format code

# Check Go Report Card score locally
.git/hooks/pre-commit  # Run the same checks as Git hook
```

#### Development Standards

- Follow Go conventions and `gofmt` formatting
- Maintain cyclomatic complexity ≤15 per function
- Add comprehensive tests for new features
- Update documentation as needed
- Use meaningful, conventional commit messages
- Ensure 100% compatibility with Go Report Card standards

## 📋 Roadmap

- [ ] **v1.1.0** - Port forwarding support
- [ ] **v1.2.0** - Web-based management interface
- [ ] **v1.3.0** - Configuration profiles/presets
- [ ] **v2.0.0** - Multi-interface support
- [ ] **v2.1.0** - Traffic shaping and QoS
- [ ] **v2.2.0** - Advanced logging and analytics

See [CHANGELOG.md](CHANGELOG.md) for detailed release history.

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Charm](https://charm.sh/) for the excellent TUI libraries
- [Cobra](https://cobra.dev/) for the CLI framework
- [Viper](https://github.com/spf13/viper) for configuration management
- The Go community for fantastic tooling and libraries

## 📞 Support

- 📖 **Documentation** - Check our [docs](docs/) directory
- 🐛 **Bug Reports** - [GitHub Issues](https://github.com/scttfrdmn/macos-nat-manager/issues)
- 💬 **Discussions** - [GitHub Discussions](https://github.com/scttfrdmn/macos-nat-manager/discussions)
- 📧 **Email** - your.email@example.com

---

<p align="center">
  <strong>Built with ❤️ for the macOS community</strong><br>
  <sub>Providing true NAT where bridging isn't enough</sub>
</p>
# macOS NAT Manager

A comprehensive Terminal User Interface (TUI) tool for setting up and managing Network Address Translation (NAT) on macOS. Unlike macOS's built-in Internet Sharing which operates as a bridge, this tool creates true NAT with address translation, providing better privacy and network isolation.

## Features

- 🖥️ **Interactive TUI Interface** - Easy-to-use terminal interface
- 🔀 **True NAT Implementation** - Address translation, not bridging
- 🌐 **Interface Management** - Select and configure network interfaces
- ⚙️ **Flexible Configuration** - Customizable IP ranges and DHCP settings
- 📊 **Connection Monitoring** - Real-time view of active connections
- 🔧 **Clean Setup/Teardown** - Proper cleanup on exit
- 🛡️ **Network Isolation** - Internal devices hidden from upstream network

## Why Use This Instead of macOS Internet Sharing?

macOS's built-in Internet Sharing operates as a transparent bridge, which means:
- ❌ Connected devices' MAC addresses are visible to the upstream network
- ❌ Devices can be detected by 802.1x systems
- ❌ Less privacy and security

This NAT Manager provides:
- ✅ True NAT with address translation
- ✅ Connected devices are completely hidden
- ✅ Single MAC address visible to upstream network
- ✅ Better privacy and 802.1x compatibility

## Prerequisites

- macOS (tested on macOS 12+)
- Go 1.21+ (for building)
- Homebrew
- Root/sudo privileges

## Installation

### Option 1: Quick Install

```bash
# Clone or download the source files
# Make sure you have main.go, go.mod, and build.sh

# Run the build script
chmod +x build.sh
./build.sh

# Run the application
sudo ./nat-manager
```

### Option 2: Manual Build

```bash
# Install dependencies
brew install dnsmasq

# Initialize Go module and install dependencies
go mod init macos-nat-manager
go get github.com/charmbracelet/bubbles@v0.18.0
go get github.com/charmbracelet/bubbletea@v0.25.0
go get github.com/charmbracelet/lipgloss@v0.9.1

# Build
go build -o nat-manager main.go

# Run
sudo ./nat-manager
```

### Option 3: System Installation

```bash
# After building, install system-wide
sudo cp nat-manager /usr/local/bin/
sudo chmod +x /usr/local/bin/nat-manager

# Run from anywhere
sudo nat-manager
```

## Usage

### Main Interface

The application presents a menu-driven interface:

```
macOS NAT Manager

🔴 NAT Inactive

1. Configure Interfaces
2. Configure NAT Settings  
3. Start NAT
4. Monitor Connections
5. Stop NAT

Press number to select, 'q' to quit
```

### Step-by-Step Setup

1. **Configure Interfaces** (Option 1)
   - Select your external interface (e.g., `en0` for Ethernet, `en1` for WiFi)
   - Select or create an internal interface (e.g., `bridge100`)
   - Press `e` to set external, `i` to set internal

2. **Configure NAT Settings** (Option 2)
   - Set internal network (default: `192.168.100`)
   - Configure DHCP range (default: `192.168.100.100` - `192.168.100.200`)
   - DNS servers are pre-configured (Google DNS)

3. **Start NAT** (Option 3)
   - Enables IP forwarding
   - Creates bridge interface (if needed)
   - Sets up pfctl NAT rules
   - Starts DHCP server

4. **Monitor Connections** (Option 4)
   - View active connections through the NAT
   - Real-time updates every 2 seconds
   - Shows source, destination, protocol, and state

5. **Stop NAT** (Option 5)
   - Clean shutdown of all services
   - Removes NAT rules
   - Destroys created interfaces
   - Stops DHCP server

### Keyboard Shortcuts

**Main Menu:**
- `1-5`: Select menu options
- `q`: Quit application

**Interface Selection:**
- `e`: Set selected interface as external
- `i`: Set selected interface as internal  
- `r`: Refresh interface list
- `↑/↓`: Navigate interfaces
- `esc`: Back to main menu

**Configuration:**
- `1-3`: Edit configuration items
- `esc`: Back to main menu

**Monitor:**
- `r`: Refresh connection list
- `↑/↓`: Navigate connections
- `esc`: Back to main menu

**Input Fields:**
- `Enter`: Save value
- `Esc`: Cancel changes

## Configuration Details

### Default Settings

- **Internal Network**: `192.168.100.0/24`
- **Gateway**: `192.168.100.1`
- **DHCP Range**: `192.168.100.100` - `192.168.100.200`
- **DHCP Lease**: `12h`
- **DNS Servers**: `8.8.8.8`, `8.8.4.4`

### Network Topology

```
Internet → [External Interface] → NAT Router → [Internal Interface] → Connected Devices
           (en0/en1/etc)         (Your Mac)    (bridge100)          (192.168.100.x)
```

## Technical Details

### What It Does

1. **IP Forwarding**: Enables packet forwarding between interfaces
2. **NAT Rules**: Uses `pfctl` to create NAT translation rules
3. **Bridge Interface**: Creates virtual interface for internal network
4. **DHCP Server**: Runs `dnsmasq` to assign IP addresses
5. **Traffic Translation**: Rewrites packet headers for address translation

### Files Created

- `/tmp/nat_rules.conf` - Temporary pfctl rules file
- DHCP leases managed by dnsmasq

### System Changes

**Temporary (cleaned up on exit):**
- IP forwarding enabled
- pfctl NAT rules loaded
- Bridge interface created
- dnsmasq process running

**None permanent** - All changes are reverted when stopping NAT or exiting

## Troubleshooting

### Common Issues

**"This tool requires root privileges"**
- Solution: Run with `sudo ./nat-manager`

**"dnsmasq not found"**
- Solution: Install with `brew install dnsmasq`

**"Failed to create bridge interface"**
- Check if interface name is available
- Try different bridge number (bridge101, bridge102, etc.)

**"Failed to load pfctl rules"**
- Check if pfctl is enabled on your system
- Ensure no conflicting firewall rules

**No internet access for connected devices**
- Verify external interface has internet connectivity
- Check NAT rules are properly loaded: `sudo pfctl -s nat`
- Ensure IP forwarding is enabled: `sysctl net.inet.ip.forwarding`

### Debug Commands

```bash
# Check NAT rules
sudo pfctl -s nat

# Check IP forwarding
sysctl net.inet.ip.forwarding

# Check DHCP server
ps aux | grep dnsmasq

# Check interfaces
ifconfig

# Check routing table
netstat -rn
```

### Clean Manual Cleanup

If the application exits unexpectedly:

```bash
# Stop NAT
sudo pfctl -d

# Remove bridge interface (adjust number as needed)
sudo ifconfig bridge100 destroy

# Stop DHCP server
sudo killall dnsmasq

# Disable IP forwarding (optional)
sudo sysctl -w net.inet.ip.forwarding=0
```

## Security Considerations

- **Root Privileges**: Required for network configuration
- **Firewall Integration**: Works with pfctl (macOS's built-in firewall)
- **Network Isolation**: Internal devices are NATed, not bridged
- **Clean Shutdown**: Proper cleanup prevents security holes

## Contributing

This tool can be extended with additional features:
- Custom DNS server configuration
- Port forwarding rules
- Traffic shaping/QoS
- Multiple internal networks
- Configuration file support
- Logging capabilities

## License

This is a utility tool for macOS network management. Use responsibly and in accordance with your network policies.

## Disclaimer

This tool modifies network configuration and requires root privileges. Always test in a safe environment first. The authors are not responsible for any network disruptions or security issues that may arise from its use.
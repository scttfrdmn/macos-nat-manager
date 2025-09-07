#!/bin/bash
#
# macOS NAT Manager - Installation Script
# This script installs nat-manager from GitHub releases or builds from source
#

set -e

# Configuration
REPO="scttfrdmn/macos-nat-manager"
BINARY_NAME="nat-manager"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.config/nat-manager"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ️  ${1}${NC}"
}

log_success() {
    echo -e "${GREEN}✅ ${1}${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  ${1}${NC}"
}

log_error() {
    echo -e "${RED}❌ ${1}${NC}"
}

log_step() {
    echo -e "${PURPLE}🔧 ${1}${NC}"
}

# Check if running on macOS
check_platform() {
    if [[ "$OSTYPE" != "darwin"* ]]; then
        log_error "This tool only works on macOS"
        exit 1
    fi
    
    log_success "Platform: macOS ✓"
}

# Check if running with appropriate privileges
check_privileges() {
    if [[ $EUID -eq 0 ]]; then
        log_warning "Running as root. Installation will be system-wide."
        INSTALL_MODE="system"
    else
        log_info "Running as user. Will need sudo for installation."
        INSTALL_MODE="user"
    fi
}

# Check dependencies
check_dependencies() {
    log_step "Checking dependencies..."
    
    # Check for curl
    if ! command -v curl &> /dev/null; then
        log_error "curl is required but not installed"
        exit 1
    fi
    
    # Check for Homebrew (recommended for dnsmasq)
    if ! command -v brew &> /dev/null; then
        log_warning "Homebrew not found. You'll need to install dnsmasq manually."
        log_info "Install Homebrew: https://brew.sh"
    else
        log_success "Homebrew found"
        
        # Check for dnsmasq
        if ! command -v dnsmasq &> /dev/null; then
            log_step "Installing dnsmasq via Homebrew..."
            brew install dnsmasq
            log_success "dnsmasq installed"
        else
            log_success "dnsmasq already installed"
        fi
    fi
}

# Get the latest release version
get_latest_version() {
    log_step "Getting latest release version..."
    
    LATEST_VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | \
        grep '"tag_name":' | \
        sed -E 's/.*"([^"]+)".*/\1/')
    
    if [[ -z "$LATEST_VERSION" ]]; then
        log_error "Failed to get latest version"
        exit 1
    fi
    
    log_success "Latest version: $LATEST_VERSION"
}

# Detect architecture
detect_arch() {
    ARCH=$(uname -m)
    case $ARCH in
        x86_64) ARCH="amd64" ;;
        arm64) ARCH="arm64" ;;
        *) 
            log_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    
    log_success "Architecture: $ARCH"
}

# Download and install binary
install_binary() {
    log_step "Downloading and installing binary..."
    
    # Construct download URL
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_VERSION}/${BINARY_NAME}-${LATEST_VERSION#v}-darwin-${ARCH}.tar.gz"
    
    # Create temporary directory
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    # Download archive
    log_info "Downloading from: $DOWNLOAD_URL"
    if curl -L "$DOWNLOAD_URL" -o "${BINARY_NAME}.tar.gz"; then
        log_success "Download complete"
    else
        log_error "Download failed"
        exit 1
    fi
    
    # Extract archive
    tar -xzf "${BINARY_NAME}.tar.gz"
    
    # Install binary
    if [[ "$INSTALL_MODE" == "system" ]]; then
        cp "$BINARY_NAME" "$INSTALL_DIR/"
        chmod +x "$INSTALL_DIR/$BINARY_NAME"
    else
        sudo cp "$BINARY_NAME" "$INSTALL_DIR/"
        sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
    fi
    
    # Cleanup
    cd - > /dev/null
    rm -rf "$TEMP_DIR"
    
    log_success "Binary installed to $INSTALL_DIR/$BINARY_NAME"
}

# Create configuration directory
setup_config() {
    log_step "Setting up configuration directory..."
    
    if [[ ! -d "$CONFIG_DIR" ]]; then
        mkdir -p "$CONFIG_DIR"
        log_success "Created config directory: $CONFIG_DIR"
    else
        log_info "Config directory already exists: $CONFIG_DIR"
    fi
    
    # Create default config if it doesn't exist
    if [[ ! -f "$CONFIG_DIR/config.yaml" ]]; then
        cat > "$CONFIG_DIR/config.yaml" << EOF
# macOS NAT Manager Configuration
# Edit these values according to your setup

external_interface: ""        # e.g., en0, en1
internal_interface: bridge100
internal_network: 192.168.100
dhcp_range:
  start: 192.168.100.100
  end: 192.168.100.200
  lease: 12h
dns_servers:
  - 8.8.8.8
  - 8.8.4.4
EOF
        log_success "Created default config file"
    else
        log_info "Config file already exists"
    fi
}

# Test installation
test_installation() {
    log_step "Testing installation..."
    
    if command -v "$BINARY_NAME" &> /dev/null; then
        VERSION_OUTPUT=$($BINARY_NAME --version 2>&1 || true)
        log_success "Installation verified: $VERSION_OUTPUT"
    else
        log_error "Installation failed - binary not found in PATH"
        exit 1
    fi
}

# Show usage information
show_usage_info() {
    log_success "Installation completed successfully! 🎉"
    echo ""
    echo -e "${CYAN}📖 Quick Start:${NC}"
    echo "   sudo $BINARY_NAME                    # Launch TUI interface"
    echo "   sudo $BINARY_NAME interfaces         # List network interfaces"
    echo "   sudo $BINARY_NAME start --help       # Get help for CLI"
    echo ""
    echo -e "${CYAN}⚠️  Important:${NC}"
    echo "   • Root privileges required (use sudo)"
    echo "   • Only works on macOS"
    echo "   • Requires dnsmasq for DHCP functionality"
    echo ""
    echo -e "${CYAN}📁 Configuration:${NC}"
    echo "   Config file: $CONFIG_DIR/config.yaml"
    echo "   Edit the config file to set your preferred interfaces and settings"
    echo ""
    echo -e "${CYAN}🔗 Links:${NC}"
    echo "   • Documentation: https://github.com/${REPO}/blob/main/README.md"
    echo "   • Issues: https://github.com/${REPO}/issues"
    echo "   • Releases: https://github.com/${REPO}/releases"
    echo ""
}

# Installation options
install_from_release() {
    check_platform
    check_privileges
    check_dependencies
    get_latest_version
    detect_arch
    install_binary
    setup_config
    test_installation
    show_usage_info
}

install_from_source() {
    log_step "Installing from source..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is required to build from source"
        log_info "Install Go: https://golang.org/dl/"
        exit 1
    fi
    
    # Check Go version
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    log_success "Go version: $GO_VERSION"
    
    # Clone repository
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    log_step "Cloning repository..."
    git clone "https://github.com/${REPO}.git" .
    
    # Build
    log_step "Building from source..."
    make build
    
    # Install
    if [[ "$INSTALL_MODE" == "system" ]]; then
        cp "$BINARY_NAME" "$INSTALL_DIR/"
        chmod +x "$INSTALL_DIR/$BINARY_NAME"
    else
        sudo cp "$BINARY_NAME" "$INSTALL_DIR/"
        sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
    fi
    
    # Cleanup
    cd - > /dev/null
    rm -rf "$TEMP_DIR"
    
    setup_config
    test_installation
    show_usage_info
}

# Main installation logic
main() {
    echo -e "${PURPLE}🚀 macOS NAT Manager Installer${NC}"
    echo ""
    
    # Parse command line arguments
    case "${1:-release}" in
        "source"|"src"|"build")
            install_from_source
            ;;
        "release"|"")
            install_from_release
            ;;
        "help"|"-h"|"--help")
            echo "Usage: $0 [release|source|help]"
            echo ""
            echo "Options:"
            echo "  release (default)  Install from GitHub releases"
            echo "  source             Build and install from source"
            echo "  help               Show this help message"
            echo ""
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
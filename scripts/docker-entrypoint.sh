#!/bin/bash
#
# Docker entrypoint script for macOS NAT Manager
# This script is used when running the application in a Docker container
# Primarily for testing and CI/CD purposes
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}ℹ️  [DOCKER] ${1}${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  [DOCKER] ${1}${NC}"
}

log_error() {
    echo -e "${RED}❌ [DOCKER] ${1}${NC}"
}

# Initialize function
init_container() {
    log_info "Initializing macOS NAT Manager container..."
    
    # Check if running in privileged mode (required for network operations)
    if [[ ! -w /proc/sys/net/ipv4/ip_forward ]]; then
        log_warning "Container not running in privileged mode"
        log_warning "Network operations will be limited"
        log_info "Run with: docker run --privileged ..."
    fi
    
    # Display container information
    log_info "Container User: $(whoami)"
    log_info "Container OS: $(uname -s)"
    log_info "Available Commands:"
    echo "  nat-manager --help"
    echo "  nat-manager version"
    echo "  nat-manager interfaces (limited in container)"
    echo ""
    
    # Show important notes for Docker usage
    cat << 'EOF'
📝 DOCKER USAGE NOTES:

This Docker container is primarily intended for:
• Testing and CI/CD pipelines
• Documentation and examples  
• Development environment setup

❌ LIMITATIONS IN DOCKER:
• Cannot actually manage macOS network interfaces
• pfctl and dnsmasq functionality is limited
• Root privileges and macOS-specific features unavailable

✅ WHAT WORKS:
• Command-line interface testing
• Configuration file validation
• Help and documentation access
• Build and testing workflows

💡 FOR ACTUAL NAT FUNCTIONALITY:
Install natively on macOS using:
  brew tap scttfrdmn/tap
  brew install nat-manager

EOF
}

# Handle special cases
case "$1" in
    --init|init)
        init_container
        exit 0
        ;;
    --test|test)
        log_info "Running container tests..."
        nat-manager --version
        nat-manager --help > /dev/null
        log_info "Container tests passed!"
        exit 0
        ;;
    --shell|shell|bash)
        log_info "Starting interactive shell..."
        exec /bin/bash
        ;;
    --docs|docs)
        log_info "Available documentation:"
        ls -la /usr/local/share/doc/nat-manager/
        echo ""
        echo "To view README:"
        echo "  cat /usr/local/share/doc/nat-manager/README.md"
        echo ""
        echo "To view CHANGELOG:"  
        echo "  cat /usr/local/share/doc/nat-manager/CHANGELOG.md"
        exit 0
        ;;
esac

# If no arguments or help requested, show container-specific help
if [[ $# -eq 0 || "$1" == "--help" || "$1" == "-h" ]]; then
    init_container
    echo "🐳 DOCKER-SPECIFIC COMMANDS:"
    echo "  --init        Initialize and show container info"
    echo "  --test        Run basic container tests"
    echo "  --shell       Start interactive bash shell"
    echo "  --docs        List available documentation"
    echo ""
    echo "📋 PASS-THROUGH TO NAT-MANAGER:"
fi

# Pass all arguments to nat-manager
exec nat-manager "$@"
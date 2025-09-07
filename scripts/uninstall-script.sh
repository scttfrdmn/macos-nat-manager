#!/bin/bash
#
# macOS NAT Manager - Uninstall Script
# Safely removes nat-manager and optionally removes configuration files
#

set -e

# Configuration
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

# Confirmation prompt
confirm() {
    local prompt="$1"
    local default="${2:-N}"
    
    if [[ "$default" == "Y" ]]; then
        prompt="$prompt [Y/n]"
    else
        prompt="$prompt [y/N]"
    fi
    
    read -p "$(echo -e "${CYAN}❓ $prompt${NC} ")" -r
    
    if [[ -z "$REPLY" ]]; then
        REPLY="$default"
    fi
    
    case "$REPLY" in
        [Yy]|[Yy][Ee][Ss]) return 0 ;;
        *) return 1 ;;
    esac
}

# Stop NAT service if running
stop_nat_service() {
    log_step "Checking for running NAT service..."
    
    if command -v "$BINARY_NAME" &> /dev/null; then
        # Check if NAT is running (this will require sudo)
        if sudo "$BINARY_NAME" status &> /dev/null; then
            log_warning "NAT service is currently running"
            if confirm "Stop NAT service before uninstalling?" "Y"; then
                log_step "Stopping NAT service..."
                if sudo "$BINARY_NAME" stop --force; then
                    log_success "NAT service stopped"
                else
                    log_warning "Failed to stop NAT service cleanly"
                    log_info "Manual cleanup may be required"
                fi
            else
                log_warning "Uninstalling while service is running may leave network configuration active"
            fi
        else
            log_info "NAT service is not running"
        fi
    else
        log_info "NAT manager binary not found in PATH"
    fi
}

# Remove binary
remove_binary() {
    log_step "Removing binary..."
    
    if [[ -f "$INSTALL_DIR/$BINARY_NAME" ]]; then
        if [[ $EUID -eq 0 ]]; then
            rm -f "$INSTALL_DIR/$BINARY_NAME"
        else
            sudo rm -f "$INSTALL_DIR/$BINARY_NAME"
        fi
        log_success "Removed binary from $INSTALL_DIR/$BINARY_NAME"
    else
        log_info "Binary not found at $INSTALL_DIR/$BINARY_NAME"
    fi
    
    # Check other possible locations
    local alt_locations=(
        "/usr/bin/$BINARY_NAME"
        "/bin/$BINARY_NAME"
        "$HOME/bin/$BINARY_NAME"
        "$HOME/.local/bin/$BINARY_NAME"
    )
    
    for location in "${alt_locations[@]}"; do
        if [[ -f "$location" ]]; then
            log_info "Found binary at alternative location: $location"
            if confirm "Remove this binary as well?"; then
                if [[ -w "$(dirname "$location")" ]]; then
                    rm -f "$location"
                else
                    sudo rm -f "$location"
                fi
                log_success "Removed binary from $location"
            fi
        fi
    done
}

# Remove configuration
remove_configuration() {
    if [[ -d "$CONFIG_DIR" ]]; then
        log_step "Configuration directory found: $CONFIG_DIR"
        
        # Show what will be removed
        if [[ -f "$CONFIG_DIR/config.yaml" ]]; then
            log_info "Found configuration file: $CONFIG_DIR/config.yaml"
        fi
        if [[ -f "$CONFIG_DIR/state.yaml" ]]; then
            log_info "Found state file: $CONFIG_DIR/state.yaml"
        fi
        
        # Count total files
        local file_count
        file_count=$(find "$CONFIG_DIR" -type f | wc -l)
        log_info "Total files in config directory: $file_count"
        
        if confirm "Remove all configuration files and directory?"; then
            rm -rf "$CONFIG_DIR"
            log_success "Removed configuration directory"
        else
            log_info "Configuration files preserved"
        fi
    else
        log_info "No configuration directory found"
    fi
}

# Remove Homebrew installation
remove_homebrew() {
    if command -v brew &> /dev/null; then
        log_step "Checking Homebrew installation..."
        
        # Check if installed via Homebrew
        if brew list | grep -q "^${BINARY_NAME}\$" 2>/dev/null; then
            log_info "Found Homebrew installation"
            if confirm "Remove via Homebrew?" "Y"; then
                brew uninstall "$BINARY_NAME"
                log_success "Removed Homebrew installation"
                return 0
            fi
        else
            log_info "Not installed via Homebrew (or tap not available)"
        fi
    else
        log_info "Homebrew not found"
    fi
    return 1
}

# Clean up shell completions
remove_completions() {
    log_step "Checking for shell completions..."
    
    local completion_files=(
        "/usr/local/share/bash-completion/completions/$BINARY_NAME"
        "/etc/bash_completion.d/$BINARY_NAME"
        "/usr/local/share/zsh/site-functions/_$BINARY_NAME"
        "/usr/local/share/fish/completions/$BINARY_NAME.fish"
        "$HOME/.bash_completion.d/$BINARY_NAME"
        "$HOME/.zsh/completions/_$BINARY_NAME"
    )
    
    local found_completions=false
    for completion_file in "${completion_files[@]}"; do
        if [[ -f "$completion_file" ]]; then
            log_info "Found completion file: $completion_file"
            found_completions=true
        fi
    done
    
    if $found_completions; then
        if confirm "Remove shell completion files?"; then
            for completion_file in "${completion_files[@]}"; do
                if [[ -f "$completion_file" ]]; then
                    if [[ -w "$(dirname "$completion_file")" ]]; then
                        rm -f "$completion_file"
                    else
                        sudo rm -f "$completion_file"
                    fi
                    log_success "Removed: $completion_file"
                fi
            done
        fi
    else
        log_info "No completion files found"
    fi
}

# Manual cleanup instructions
show_manual_cleanup() {
    log_step "Manual cleanup instructions..."
    
    echo ""
    echo -e "${CYAN}🔧 If NAT service was running, you may need to manually clean up:${NC}"
    echo ""
    echo "1. Disable pfctl (if enabled):"
    echo "   sudo pfctl -d"
    echo ""
    echo "2. Stop any dnsmasq processes:"
    echo "   sudo killall dnsmasq"
    echo ""
    echo "3. Remove bridge interfaces (if created):"
    echo "   sudo ifconfig bridge100 destroy"
    echo "   sudo ifconfig bridge101 destroy"
    echo ""
    echo "4. Disable IP forwarding (if desired):"
    echo "   sudo sysctl -w net.inet.ip.forwarding=0"
    echo ""
    echo -e "${YELLOW}⚠️  Only run these commands if you're sure NAT manager created them${NC}"
    echo ""
}

# Verification
verify_removal() {
    log_step "Verifying removal..."
    
    local issues=false
    
    # Check binary
    if command -v "$BINARY_NAME" &> /dev/null; then
        log_warning "Binary still found in PATH: $(which "$BINARY_NAME")"
        issues=true
    else
        log_success "Binary removed from PATH"
    fi
    
    # Check configuration
    if [[ -d "$CONFIG_DIR" ]]; then
        log_info "Configuration directory still exists (preserved by user choice)"
    else
        log_success "Configuration directory removed"
    fi
    
    if $issues; then
        log_warning "Some components may still be present"
    else
        log_success "Uninstallation appears complete"
    fi
}

# Main uninstall process
main() {
    echo -e "${PURPLE}🗑️  macOS NAT Manager Uninstaller${NC}"
    echo ""
    
    # Parse arguments
    local remove_config=false
    local force=false
    local homebrew_first=true
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --remove-config|--purge)
                remove_config=true
                shift
                ;;
            --force)
                force=true
                shift
                ;;
            --no-homebrew)
                homebrew_first=false
                shift
                ;;
            --help|-h)
                echo "Usage: $0 [OPTIONS]"
                echo ""
                echo "Options:"
                echo "  --remove-config    Also remove configuration files"
                echo "  --purge            Same as --remove-config"
                echo "  --force            Skip confirmation prompts"
                echo "  --no-homebrew      Skip Homebrew uninstall attempt"
                echo "  --help, -h         Show this help message"
                echo ""
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                echo "Use '$0 --help' for usage information"
                exit 1
                ;;
        esac
    done
    
    # Confirmation
    if ! $force; then
        echo -e "${CYAN}This will remove macOS NAT Manager from your system.${NC}"
        echo ""
        if ! confirm "Continue with uninstallation?"; then
            log_info "Uninstallation cancelled"
            exit 0
        fi
        echo ""
    fi
    
    # Stop service first
    stop_nat_service
    
    # Try Homebrew removal first
    local homebrew_removed=false
    if $homebrew_first; then
        if remove_homebrew; then
            homebrew_removed=true
        fi
    fi
    
    # Remove binary (if not removed via Homebrew)
    if ! $homebrew_removed; then
        remove_binary
    fi
    
    # Remove completions
    remove_completions
    
    # Remove configuration
    if $remove_config || ! $homebrew_removed; then
        remove_configuration
    fi
    
    # Verify removal
    verify_removal
    
    # Show manual cleanup if needed
    if ! $force; then
        show_manual_cleanup
    fi
    
    echo ""
    log_success "Uninstallation completed! 👋"
    echo ""
    echo -e "${CYAN}Thank you for using macOS NAT Manager!${NC}"
    echo ""
    echo -e "${BLUE}If you encountered any issues or have feedback:${NC}"
    echo "  https://github.com/scttfrdmn/macos-nat-manager/issues"
    echo ""
}

# Run main function
main "$@"
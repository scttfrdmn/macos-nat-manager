#!/bin/bash
#
# Generate Homebrew Formula for macOS NAT Manager
# This script generates a .rb formula file with the correct version and SHA256
#

set -e

# Configuration
REPO="scttfrdmn/macos-nat-manager"
FORMULA_NAME="nat-manager"
OUTPUT_DIR="homebrew"
OUTPUT_FILE="$OUTPUT_DIR/${FORMULA_NAME}.rb"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}ℹ️  ${1}${NC}"
}

log_success() {
    echo -e "${GREEN}✅ ${1}${NC}"
}

log_error() {
    echo -e "${RED}❌ ${1}${NC}"
}

log_step() {
    echo -e "${BLUE}🔧 ${1}${NC}"
}

# Help message
show_help() {
    cat << EOF
Usage: $0 [VERSION] [OPTIONS]

Generate Homebrew formula for macOS NAT Manager

Arguments:
  VERSION           Version to generate formula for (e.g., v1.0.0)
                   If not provided, uses latest git tag

Options:
  --output FILE     Output file path (default: $OUTPUT_FILE)
  --repo REPO       GitHub repository (default: $REPO)
  --dry-run         Generate formula but don't write to file
  --help, -h        Show this help message

Examples:
  $0 v1.0.0
  $0 --dry-run
  $0 v1.2.0 --output /tmp/formula.rb

EOF
}

# Parse arguments
VERSION=""
DRY_RUN=false
CUSTOM_OUTPUT=""
CUSTOM_REPO=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --help|-h)
            show_help
            exit 0
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --output)
            CUSTOM_OUTPUT="$2"
            shift 2
            ;;
        --repo)
            CUSTOM_REPO="$2"
            shift 2
            ;;
        v*.*.*)
            VERSION="$1"
            shift
            ;;
        *.*.*)
            VERSION="v$1"
            shift
            ;;
        *)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Use custom values if provided
if [[ -n "$CUSTOM_OUTPUT" ]]; then
    OUTPUT_FILE="$CUSTOM_OUTPUT"
fi

if [[ -n "$CUSTOM_REPO" ]]; then
    REPO="$CUSTOM_REPO"
fi

# Get version if not provided
if [[ -z "$VERSION" ]]; then
    log_step "Getting latest version from git tags..."
    VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.1.0")
    log_info "Using version: $VERSION"
fi

# Validate version format
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-.*)?$ ]]; then
    log_error "Invalid version format: $VERSION"
    log_info "Expected format: vX.Y.Z or vX.Y.Z-suffix"
    exit 1
fi

# Remove 'v' prefix for formula
CLEAN_VERSION="${VERSION#v}"

log_step "Generating Homebrew formula for version $VERSION..."

# Calculate SHA256 for the tarball
TARBALL_URL="https://github.com/${REPO}/archive/refs/tags/${VERSION}.tar.gz"
log_step "Calculating SHA256 for $TARBALL_URL..."

# Download and calculate SHA256
TEMP_FILE=$(mktemp)
if curl -sL "$TARBALL_URL" -o "$TEMP_FILE" 2>/dev/null; then
    SHA256=$(shasum -a 256 "$TEMP_FILE" | cut -d' ' -f1)
    rm -f "$TEMP_FILE"
    log_success "SHA256: $SHA256"
else
    log_error "Failed to download tarball from $TARBALL_URL"
    log_info "Make sure the version tag exists and is pushed to GitHub"
    rm -f "$TEMP_FILE"
    exit 1
fi

# Get current date for formula
FORMULA_DATE=$(date -u +"%Y-%m-%d")

# Generate Ruby class name (capitalize and remove hyphens)
CLASS_NAME=$(echo "$FORMULA_NAME" | sed 's/-//g' | sed 's/\b\w/\U&/g')

# Generate the formula
log_step "Generating formula content..."

FORMULA_CONTENT=$(cat << EOF
# Homebrew Formula for macOS NAT Manager
# Generated on $FORMULA_DATE

class $CLASS_NAME < Formula
  desc "macOS NAT Manager - True NAT with address translation"
  homepage "https://github.com/${REPO}"
  url "$TARBALL_URL"
  sha256 "$SHA256"
  license "MIT"
  head "https://github.com/${REPO}.git", branch: "main"

  # macOS only
  depends_on :macos

  # Build dependencies
  depends_on "go" => :build
  
  # Runtime dependencies  
  depends_on "dnsmasq"

  def install
    # Set build variables
    ldflags = %W[
      -s -w
      -X main.version=${CLEAN_VERSION}
      -X main.commit=#{tap.user}
      -X main.date=#{time.strftime("%Y-%m-%dT%H:%M:%SZ")}
    ]

    # Build the binary
    system "go", "build", *std_go_args(ldflags: ldflags), "./cmd/nat-manager"
    
    # Install the binary  
    bin.install "nat-manager"
    
    # Install documentation
    doc.install "README.md"
    doc.install "CHANGELOG.md"
    doc.install "LICENSE"
    
    # Install man page if it exists
    if File.exist?("docs/nat-manager.1")
      man1.install "docs/nat-manager.1"
    end
    
    # Install completion scripts
    if File.exist?("completions/nat-manager.bash")
      bash_completion.install "completions/nat-manager.bash" => "nat-manager"
    end
    if File.exist?("completions/_nat-manager")
      zsh_completion.install "completions/_nat-manager"
    end
    if File.exist?("completions/nat-manager.fish")
      fish_completion.install "completions/nat-manager.fish"
    end
  end

  def post_install
    # Ensure dnsmasq is available
    unless which("dnsmasq")
      ohai "Installing dnsmasq dependency..."
      system "brew", "install", "dnsmasq"
    end

    # Create configuration directory
    config_dir = "#{Dir.home}/.config/nat-manager"
    mkdir_p config_dir unless Dir.exist?(config_dir)

    # Welcome message
    ohai "macOS NAT Manager installed successfully!"
    puts <<~EOS
      
      🎉 Installation complete!
      
      📖 Usage:
        sudo nat-manager                    # Launch TUI interface
        sudo nat-manager start --help       # CLI help
        sudo nat-manager interfaces         # List interfaces
      
      ⚠️  Important:
        • Root privileges required (use sudo)
        • Only works on macOS
        • Requires dnsmasq for DHCP functionality
      
      📚 Documentation:
        • README: #{doc}/README.md
        • Issues: https://github.com/${REPO}/issues
      
      🔧 Configuration:
        Config file: ~/.config/nat-manager/config.yaml
        
      💡 Quick start:
        1. sudo nat-manager interfaces      # List available interfaces  
        2. sudo nat-manager start -e en0 -i bridge100 -n 192.168.100
        3. sudo nat-manager monitor         # Monitor connections
        4. sudo nat-manager stop            # Stop NAT service
        
    EOS
  end

  test do
    # Test that the binary was installed correctly
    assert_match "${CLEAN_VERSION}", shell_output("#{bin}/nat-manager --version")
    
    # Test that help works
    help_output = shell_output("#{bin}/nat-manager --help")
    assert_match "macOS NAT Manager", help_output
    assert_match "True NAT with address translation", help_output
    
    # Test subcommands exist
    assert_match "start", help_output
    assert_match "stop", help_output
    assert_match "status", help_output
    assert_match "interfaces", help_output
    assert_match "monitor", help_output
  end

  def caveats
    <<~EOS
      ⚠️  IMPORTANT NOTES:
      
      🔐 Root Privileges Required:
        This tool modifies network configuration and requires root privileges.
        Always run with sudo: sudo nat-manager
        
      🍺 Dependencies:
        • dnsmasq: Installed automatically as dependency
        • pfctl: Built into macOS (used for NAT rules)
        
      🚫 Limitations:
        • macOS only (uses macOS-specific networking tools)
        • Requires active network interface for external connection
        • Bridge interfaces created automatically if needed
        
      🛡️  Security:
        • Creates temporary pfctl rules (cleaned up on exit)
        • Enables IP forwarding (disabled on stop)
        • No permanent system changes
        
      📖 Documentation:
        • README: #{doc}/README.md
        • GitHub: https://github.com/${REPO}
        
      🐛 Issues:
        Report bugs: https://github.com/${REPO}/issues
    EOS
  end
end
EOF
)

# Output or write the formula
if $DRY_RUN; then
    log_info "Generated formula (dry-run mode):"
    echo "----------------------------------------"
    echo "$FORMULA_CONTENT"
    echo "----------------------------------------"
else
    # Create output directory if it doesn't exist
    mkdir -p "$(dirname "$OUTPUT_FILE")"
    
    # Write the formula
    echo "$FORMULA_CONTENT" > "$OUTPUT_FILE"
    log_success "Formula written to: $OUTPUT_FILE"
    
    # Show file info
    log_info "Formula details:"
    echo "  File size: $(wc -c < "$OUTPUT_FILE") bytes"
    echo "  Lines: $(wc -l < "$OUTPUT_FILE")"
    echo "  Version: $CLEAN_VERSION"
    echo "  SHA256: $SHA256"
fi

log_success "Homebrew formula generation completed! 🍺"

if ! $DRY_RUN; then
    log_info "Next steps:"
    echo "1. Test the formula: brew install --build-from-source $OUTPUT_FILE"
    echo "2. Add to your tap repository"
    echo "3. Update tap with: git add $OUTPUT_FILE && git commit && git push"
fi
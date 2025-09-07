class NatManager < Formula
  desc "macOS NAT Manager - True NAT with address translation"
  homepage "https://github.com/scttfrdmn/macos-nat-manager"
  url "https://github.com/scttfrdmn/macos-nat-manager/archive/refs/tags/v1.0.0.tar.gz"
  sha256 "YOUR_SHA256_HASH_HERE"
  license "MIT"
  head "https://github.com/scttfrdmn/macos-nat-manager.git", branch: "main"

  depends_on "go" => :build
  depends_on "dnsmasq"

  def install
    # Set build variables
    ldflags = %W[
      -s -w
      -X main.version=#{version}
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
    
    # Install man page if it exists
    if File.exist?("docs/nat-manager.1")
      man1.install "docs/nat-manager.1"
    end
    
    # Install completion scripts if they exist
    if File.exist?("completions/nat-manager.bash")
      bash_completion.install "completions/nat-manager.bash"
    end
    if File.exist?("completions/nat-manager.zsh")
      zsh_completion.install "completions/nat-manager.zsh"
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
        • Issues: https://github.com/scttfrdmn/macos-nat-manager/issues
      
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
    assert_match version.to_s, shell_output("#{bin}/nat-manager --version")
    
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
        • GitHub: https://github.com/scttfrdmn/macos-nat-manager
        
      🐛 Issues:
        Report bugs: https://github.com/scttfrdmn/macos-nat-manager/issues
    EOS
  end
end
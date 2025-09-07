# macOS NAT Manager - Complete Project Structure

```
macos-nat-manager/
├── LICENSE                    # MIT License
├── README.md                  # Main documentation
├── CHANGELOG.md               # Keep a changelog format
├── Makefile                   # Build automation
├── go.mod                     # Go module file
├── go.sum                     # Go checksum file
├── .goreleaser.yaml           # Release automation
├── .github/
│   └── workflows/
│       ├── ci.yml            # Continuous Integration
│       └── release.yml       # Release workflow
├── cmd/
│   └── nat-manager/
│       └── main.go           # Main entry point
├── internal/
│   ├── cli/                  # CLI commands (Cobra)
│   │   ├── root.go
│   │   ├── start.go
│   │   ├── stop.go
│   │   ├── status.go
│   │   ├── interfaces.go
│   │   └── monitor.go
│   ├── tui/                  # TUI implementation
│   │   ├── app.go
│   │   ├── models.go
│   │   └── views.go
│   ├── nat/                  # NAT core functionality
│   │   ├── manager.go
│   │   ├── pfctl.go
│   │   ├── dhcp.go
│   │   └── interfaces.go
│   └── config/               # Configuration management
│       └── config.go
├── pkg/                      # Public API (if needed later)
├── scripts/                  # Build and utility scripts
│   ├── install.sh
│   └── uninstall.sh
├── homebrew/                 # Homebrew formula
│   └── nat-manager.rb
└── docs/                     # Additional documentation
    ├── usage.md
    └── troubleshooting.md
```
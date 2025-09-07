#!/bin/bash

# macOS NAT Manager - Build and Install Script

set -e

echo "🔧 Building macOS NAT Manager..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go from https://golang.org/dl/"
    exit 1
fi

# Check if Homebrew is installed (for dnsmasq)
if ! command -v brew &> /dev/null; then
    echo "❌ Homebrew is not installed. Please install from https://brew.sh/"
    exit 1
fi

# Install dnsmasq if not present
if ! command -v dnsmasq &> /dev/null; then
    echo "📦 Installing dnsmasq..."
    brew install dnsmasq
fi

# Initialize Go module if go.mod doesn't exist
if [ ! -f "go.mod" ]; then
    echo "📝 Initializing Go module..."
    go mod init macos-nat-manager
    go get github.com/charmbracelet/bubbles@v0.18.0
    go get github.com/charmbracelet/bubbletea@v0.25.0
    go get github.com/charmbracelet/lipgloss@v0.9.1
else
    echo "📝 Downloading dependencies..."
    go mod tidy
fi

# Build the application
echo "🔨 Compiling..."
go build -o nat-manager main.go

# Make it executable
chmod +x nat-manager

echo "✅ Build complete!"
echo ""
echo "To run the NAT manager:"
echo "  sudo ./nat-manager"
echo ""
echo "Note: Root privileges are required for network configuration."
echo ""
echo "Optional: Move to system PATH:"
echo "  sudo cp nat-manager /usr/local/bin/"
echo "  sudo chmod +x /usr/local/bin/nat-manager"
echo ""
echo "Then run with: sudo nat-manager"
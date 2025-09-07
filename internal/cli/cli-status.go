package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/scttfrdmn/macos-nat-manager/internal/config"
	"github.com/scttfrdmn/macos-nat-manager/internal/nat"
)

var jsonOutput bool

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show NAT service status",
	Long: `Display the current status of the NAT service including:
- Running state
- Interface configuration  
- Network settings
- Active connections
- System resource usage

Example:
  nat-manager status
  nat-manager status --json  # JSON output for scripting`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("⚠️  No configuration found\n")
			cfg = config.Default()
		}

		// Create NAT manager
		manager := nat.NewManager(cfg)

		// Get status
		status, err := manager.GetStatus()
		if err != nil {
			return fmt.Errorf("failed to get NAT status: %w", err)
		}

		if jsonOutput {
			return printStatusJSON(status)
		}

		return printStatusHuman(status)
	},
}

func printStatusHuman(status *nat.Status) error {
	// Overall status
	if status.Running {
		fmt.Printf("🟢 NAT Status: %s\n", "ACTIVE")
	} else {
		fmt.Printf("🔴 NAT Status: %s\n", "INACTIVE")
		return nil
	}

	fmt.Printf("\n📡 Configuration:\n")
	fmt.Printf("   External Interface: %s (%s)\n", status.Config.ExternalInterface, status.ExternalIP)
	fmt.Printf("   Internal Interface: %s (%s.1/24)\n", status.Config.InternalInterface, status.Config.InternalNetwork)
	fmt.Printf("   DHCP Range: %s - %s\n", status.Config.DHCPRange.Start, status.Config.DHCPRange.End)
	fmt.Printf("   DNS Servers: %s\n", strings.Join(status.Config.DNSServers, ", "))

	fmt.Printf("\n🔧 System Status:\n")
	fmt.Printf("   IP Forwarding: %s\n", formatBool(status.IPForwarding))
	fmt.Printf("   pfctl NAT Rules: %s\n", formatBool(status.PFCTLEnabled))
	fmt.Printf("   DHCP Server: %s\n", formatBool(status.DHCPRunning))

	if len(status.ConnectedDevices) > 0 {
		fmt.Printf("\n📱 Connected Devices (%d):\n", len(status.ConnectedDevices))
		for _, device := range status.ConnectedDevices {
			fmt.Printf("   %s - %s (%s)\n", device.IP, device.MAC, device.Hostname)
		}
	}

	if len(status.ActiveConnections) > 0 {
		fmt.Printf("\n🌐 Active Connections (%d):\n", len(status.ActiveConnections))
		for i, conn := range status.ActiveConnections {
			if i >= 10 { // Limit display to prevent spam
				fmt.Printf("   ... and %d more\n", len(status.ActiveConnections)-10)
				break
			}
			fmt.Printf("   %s → %s (%s)\n", conn.Source, conn.Destination, conn.Protocol)
		}
	}

	fmt.Printf("\n📊 Statistics:\n")
	fmt.Printf("   Uptime: %s\n", status.Uptime)
	fmt.Printf("   Bytes In/Out: %s / %s\n", formatBytes(status.BytesIn), formatBytes(status.BytesOut))

	return nil
}

func printStatusJSON(status *nat.Status) error {
	// For JSON output, you'd typically use encoding/json
	// This is a simplified version
	fmt.Printf(`{
  "running": %t,
  "external_interface": "%s",
  "internal_interface": "%s",
  "external_ip": "%s",
  "internal_network": "%s",
  "ip_forwarding": %t,
  "pfctl_enabled": %t,
  "dhcp_running": %t,
  "connected_devices": %d,
  "active_connections": %d,
  "uptime": "%s",
  "bytes_in": %d,
  "bytes_out": %d
}`,
		status.Running,
		status.Config.ExternalInterface,
		status.Config.InternalInterface,
		status.ExternalIP,
		status.Config.InternalNetwork,
		status.IPForwarding,
		status.PFCTLEnabled,
		status.DHCPRunning,
		len(status.ConnectedDevices),
		len(status.ActiveConnections),
		status.Uptime,
		status.BytesIn,
		status.BytesOut,
	)
	return nil
}

func formatBool(b bool) string {
	if b {
		return "✅ Enabled"
	}
	return "❌ Disabled"
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func init() {
	rootCmd.AddCommand(statusCmd)

	statusCmd.Flags().BoolVar(&jsonOutput, "json", false, "output status in JSON format")
}
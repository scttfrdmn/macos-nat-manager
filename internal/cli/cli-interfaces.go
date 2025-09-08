// Package cli provides command line interface commands for the NAT manager
package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/scttfrdmn/macos-nat-manager/internal/nat"
)

var (
	showAll    bool
	filterType string
)

// interfacesCmd represents the interfaces command
var interfacesCmd = &cobra.Command{
	Use:     "interfaces",
	Aliases: []string{"iface", "if"},
	Short:   "List available network interfaces",
	Long: `List all available network interfaces on the system.

This shows interfaces that can be used for NAT configuration,
including their current status, IP addresses, and types.

Example:
  nat-manager interfaces
  nat-manager interfaces --all          # Show all interfaces including loopback
  nat-manager interfaces --type bridge  # Filter by interface type`,
	RunE: func(_ *cobra.Command, _ []string) error {
		// Create a temporary manager to get interfaces
		manager := nat.NewManager(nil)
		interfaces, err := manager.GetNetworkInterfaces()
		if err != nil {
			return fmt.Errorf("failed to list interfaces: %w", err)
		}

		// Filter by type if specified
		if filterType != "" {
			filtered := make([]nat.NetworkInterface, 0)
			for _, iface := range interfaces {
				if strings.EqualFold(iface.Type, filterType) {
					filtered = append(filtered, iface)
				}
			}
			interfaces = filtered
		}

		if len(interfaces) == 0 {
			fmt.Printf("No interfaces found\n")
			return nil
		}

		// Print header
		fmt.Printf("%-12s %-10s %-15s %-8s %s\n", "INTERFACE", "TYPE", "IP ADDRESS", "STATUS", "DESCRIPTION")
		fmt.Printf("%-12s %-10s %-15s %-8s %s\n", 
			strings.Repeat("-", 12),
			strings.Repeat("-", 10),
			strings.Repeat("-", 15),
			strings.Repeat("-", 8),
			strings.Repeat("-", 20))

		// Print interfaces
		for _, iface := range interfaces {
			status := "Down"
			statusIcon := "❌"
			if iface.Status == "Up" {
				status = "Up"
				statusIcon = "✅"
			}

			ip := iface.IP
			if ip == "" {
				ip = "N/A"
			}

			description := getInterfaceDescription(iface)

			fmt.Printf("%-12s %-10s %-15s %s%-7s %s\n", 
				iface.Name, 
				iface.Type, 
				ip, 
				statusIcon, 
				status,
				description)
		}

		fmt.Printf("\nSuitable for:\n")
		fmt.Printf("  External: Interfaces with internet connectivity (en0, en1, etc.)\n")
		fmt.Printf("  Internal: Bridge interfaces for NAT (bridge100, bridge101, etc.)\n")
		fmt.Printf("\nNote: Bridge interfaces will be created automatically if they don't exist\n")

		return nil
	},
}

func getInterfaceDescription(iface nat.NetworkInterface) string {
	switch {
	case strings.HasPrefix(iface.Name, "en"):
		if strings.Contains(iface.Name, "0") {
			return "Ethernet (Primary)"
		}
		return "Ethernet/WiFi"
	case strings.HasPrefix(iface.Name, "bridge"):
		return "Virtual Bridge"
	case strings.HasPrefix(iface.Name, "utun"):
		return "VPN Tunnel"
	case strings.HasPrefix(iface.Name, "awdl"):
		return "AirDrop/AirPlay"
	case strings.HasPrefix(iface.Name, "lo"):
		return "Loopback"
	case strings.HasPrefix(iface.Name, "gif"):
		return "Generic Tunnel"
	case strings.HasPrefix(iface.Name, "stf"):
		return "6to4 Tunnel"
	default:
		return "Network Interface"
	}
}

func init() {
	rootCmd.AddCommand(interfacesCmd)

	interfacesCmd.Flags().BoolVarP(&showAll, "all", "a", false, "show all interfaces including loopback and inactive")
	interfacesCmd.Flags().StringVarP(&filterType, "type", "t", "", "filter by interface type (ethernet, bridge, vpn, etc.)")
}
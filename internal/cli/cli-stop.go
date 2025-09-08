package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/scttfrdmn/macos-nat-manager/internal/config"
	"github.com/scttfrdmn/macos-nat-manager/internal/nat"
)

var force bool

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop NAT service",
	Long: `Stop the NAT service and clean up all configuration.

This will:
- Disable pfctl NAT rules
- Stop DHCP server
- Remove/destroy internal interface
- Disable IP forwarding
- Clean up temporary files

Example:
  nat-manager stop
  nat-manager stop --force  # Force stop even if some cleanup fails`,
	RunE: func(_ *cobra.Command, _ []string) error {
		// Load config
		cfg, err := config.Load()
		if err != nil {
			if !force {
				return fmt.Errorf("failed to load config: %w", err)
			}
			// Use default config for force stop
			cfg = config.Default()
		}

		// Convert config to NAT config
		natConfig := &nat.NATConfig{
			ExternalInterface: cfg.ExternalInterface,
			InternalInterface: cfg.InternalInterface,
			InternalNetwork:   cfg.InternalNetwork,
			DHCPRange: nat.DHCPRange{
				Start: cfg.DHCPRange.Start,
				End:   cfg.DHCPRange.End,
				Lease: cfg.DHCPRange.Lease,
			},
			DNSServers: cfg.DNSServers,
			Active:     cfg.Active,
		}

		// Create NAT manager
		manager := nat.NewManager(natConfig)

		// Check if running
		if !manager.IsActive() && !force {
			return fmt.Errorf("NAT is not running")
		}

		// Stop NAT
		if err := manager.StopNAT(); err != nil {
			if !force {
				return fmt.Errorf("failed to stop NAT: %w", err)
			}
			fmt.Printf("Warning: some cleanup failed: %v\n", err)
		}

		fmt.Printf("✅ NAT stopped successfully\n")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)

	stopCmd.Flags().BoolVarP(&force, "force", "f", false, "force stop even if some operations fail")
}

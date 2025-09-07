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
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config
		cfg, err := config.Load()
		if err != nil {
			if !force {
				return fmt.Errorf("failed to load config: %w", err)
			}
			// Use default config for force stop
			cfg = config.Default()
		}

		// Create NAT manager
		manager := nat.NewManager(cfg)

		// Check if running
		if running, err := manager.IsRunning(); err != nil {
			if !force {
				return fmt.Errorf("failed to check NAT status: %w", err)
			}
			fmt.Printf("Warning: could not check NAT status: %v\n", err)
		} else if !running && !force {
			return fmt.Errorf("NAT is not running")
		}

		// Stop NAT
		if err := manager.Stop(); err != nil {
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
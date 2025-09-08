package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/scttfrdmn/macos-nat-manager/internal/config"
	"github.com/scttfrdmn/macos-nat-manager/internal/nat"
)

var (
	externalInterface string
	internalInterface string
	internalNetwork   string
	dhcpStart         string
	dhcpEnd           string
	dnsServers        []string
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start NAT service",
	Long: `Start the NAT service with the specified configuration.
	
This will:
- Enable IP forwarding
- Create/configure internal interface  
- Set up pfctl NAT rules
- Start DHCP server
- Begin routing traffic between interfaces

Example:
  nat-manager start --external en0 --internal bridge100 --network 192.168.100
  nat-manager start -e en1 -i bridge101 -n 10.0.1 --dhcp-start 10.0.1.100 --dhcp-end 10.0.1.200`,
	RunE: func(_ *cobra.Command, _ []string) error {
		// Load existing config
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Override with command line flags
		if externalInterface != "" {
			cfg.ExternalInterface = externalInterface
		}
		if internalInterface != "" {
			cfg.InternalInterface = internalInterface
		}
		if internalNetwork != "" {
			cfg.InternalNetwork = internalNetwork
		}
		if dhcpStart != "" {
			cfg.DHCPRange.Start = dhcpStart
		}
		if dhcpEnd != "" {
			cfg.DHCPRange.End = dhcpEnd
		}
		if len(dnsServers) > 0 {
			cfg.DNSServers = dnsServers
		}

		// Validate required fields
		if cfg.ExternalInterface == "" {
			return fmt.Errorf("external interface is required (use --external or -e)")
		}
		if cfg.InternalInterface == "" {
			return fmt.Errorf("internal interface is required (use --internal or -i)")
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

		// Check if already running
		if manager.IsActive() {
			return fmt.Errorf("NAT is already running")
		}

		// Start NAT
		if err := manager.StartNAT(); err != nil {
			return fmt.Errorf("failed to start NAT: %w", err)
		}

		// Save config for future use
		if err := cfg.Save(); err != nil {
			fmt.Printf("Warning: failed to save config: %v\n", err)
		}

		fmt.Printf("✅ NAT started successfully\n")
		fmt.Printf("   External: %s\n", cfg.ExternalInterface)
		fmt.Printf("   Internal: %s (%s.1/24)\n", cfg.InternalInterface, cfg.InternalNetwork)
		fmt.Printf("   DHCP Range: %s - %s\n", cfg.DHCPRange.Start, cfg.DHCPRange.End)
		fmt.Printf("   DNS Servers: %s\n", strings.Join(cfg.DNSServers, ", "))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Interface flags
	startCmd.Flags().StringVarP(&externalInterface, "external", "e", "", "external network interface (e.g., en0, en1)")
	startCmd.Flags().StringVarP(&internalInterface, "internal", "i", "", "internal network interface (e.g., bridge100)")

	// Network configuration flags
	startCmd.Flags().StringVarP(&internalNetwork, "network", "n", "", "internal network (e.g., 192.168.100)")
	startCmd.Flags().StringVar(&dhcpStart, "dhcp-start", "", "DHCP range start (e.g., 192.168.100.100)")
	startCmd.Flags().StringVar(&dhcpEnd, "dhcp-end", "", "DHCP range end (e.g., 192.168.100.200)")
	startCmd.Flags().StringSliceVar(&dnsServers, "dns", []string{}, "DNS servers (comma-separated)")

	// Mark required flags with helpful messages
	_ = startCmd.MarkFlagRequired("external")
	_ = startCmd.MarkFlagRequired("internal")
}

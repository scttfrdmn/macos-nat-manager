package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/scttfrdmn/macos-nat-manager/internal/config"
	"github.com/scttfrdmn/macos-nat-manager/internal/nat"
)

var (
	refreshInterval time.Duration
	maxConnections  int
	showDevices     bool
	followMode      bool
)

// monitorCmd represents the monitor command
var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor NAT traffic and connections",
	Long: `Monitor active NAT traffic, connections, and connected devices in real-time.

This displays:
- Active network connections through NAT
- Connected devices and their DHCP leases
- Real-time traffic statistics
- Connection state changes

Example:
  nat-manager monitor
  nat-manager monitor --interval 5s --max 50  # Custom refresh and limit
  nat-manager monitor --devices               # Show connected devices
  nat-manager monitor --follow                # Continuous monitoring mode`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Create NAT manager
		manager := nat.NewManager(cfg)

		// Check if NAT is running
		if running, err := manager.IsRunning(); err != nil {
			return fmt.Errorf("failed to check NAT status: %w", err)
		} else if !running {
			return fmt.Errorf("NAT is not running. Start it first with 'nat-manager start'")
		}

		if followMode {
			return runFollowMode(manager)
		}

		return runSnapshotMode(manager)
	},
}

func runSnapshotMode(manager *nat.Manager) error {
	status, err := manager.GetStatus()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	fmt.Printf("📊 NAT Monitor - %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("External: %s (%s) → Internal: %s (%s.1/24)\n\n",
		status.Config.ExternalInterface,
		status.ExternalIP,
		status.Config.InternalInterface,
		status.Config.InternalNetwork)

	if showDevices && len(status.ConnectedDevices) > 0 {
		fmt.Printf("📱 Connected Devices (%d):\n", len(status.ConnectedDevices))
		fmt.Printf("%-15s %-18s %-15s %s\n", "IP ADDRESS", "MAC ADDRESS", "HOSTNAME", "LEASE TIME")
		fmt.Printf("%s %s %s %s\n", 
			fmt.Sprintf("%-15s", strings.Repeat("-", 15)),
			fmt.Sprintf("%-18s", strings.Repeat("-", 18)),
			fmt.Sprintf("%-15s", strings.Repeat("-", 15)),
			strings.Repeat("-", 15))

		for _, device := range status.ConnectedDevices {
			hostname := device.Hostname
			if hostname == "" {
				hostname = "Unknown"
			}
			fmt.Printf("%-15s %-18s %-15s %s\n", 
				device.IP, device.MAC, hostname, device.LeaseTime)
		}
		fmt.Println()
	}

	if len(status.ActiveConnections) > 0 {
		fmt.Printf("🌐 Active Connections (%d):\n", len(status.ActiveConnections))
		fmt.Printf("%-8s %-25s %-25s %-12s\n", "PROTO", "SOURCE", "DESTINATION", "STATE")
		fmt.Printf("%-8s %-25s %-25s %-12s\n",
			strings.Repeat("-", 8),
			strings.Repeat("-", 25),
			strings.Repeat("-", 25),
			strings.Repeat("-", 12))

		count := 0
		for _, conn := range status.ActiveConnections {
			if count >= maxConnections {
				fmt.Printf("... and %d more connections\n", len(status.ActiveConnections)-maxConnections)
				break
			}
			fmt.Printf("%-8s %-25s %-25s %-12s\n", 
				conn.Protocol, conn.Source, conn.Destination, conn.State)
			count++
		}
	} else {
		fmt.Printf("🌐 No active connections\n")
	}

	fmt.Printf("\n📈 Statistics:\n")
	fmt.Printf("Uptime: %s | Traffic: %s in, %s out\n",
		status.Uptime,
		formatBytes(status.BytesIn),
		formatBytes(status.BytesOut))

	return nil
}

func runFollowMode(manager *nat.Manager) error {
	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Printf("\n\n👋 Monitoring stopped\n")
		cancel()
	}()

	fmt.Printf("🔄 NAT Monitor (Follow Mode) - Press Ctrl+C to stop\n")
	fmt.Printf("Refresh interval: %s | Max connections: %d\n\n", refreshInterval, maxConnections)

	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()

	// Initial display
	if err := displayMonitorData(manager); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			// Clear screen and redisplay
			fmt.Print("\033[2J\033[H") // ANSI clear screen and move cursor to top
			if err := displayMonitorData(manager); err != nil {
				fmt.Printf("Error updating display: %v\n", err)
			}
		}
	}
}

func displayMonitorData(manager *nat.Manager) error {
	status, err := manager.GetStatus()
	if err != nil {
		return err
	}

	fmt.Printf("📊 NAT Monitor - %s (Uptime: %s)\n",
		time.Now().Format("15:04:05"),
		status.Uptime)
	fmt.Printf("External: %s (%s) → Internal: %s (%s.1/24)\n",
		status.Config.ExternalInterface,
		status.ExternalIP,
		status.Config.InternalInterface,
		status.Config.InternalNetwork)
	fmt.Printf("Traffic: %s in, %s out | Devices: %d | Connections: %d\n\n",
		formatBytes(status.BytesIn),
		formatBytes(status.BytesOut),
		len(status.ConnectedDevices),
		len(status.ActiveConnections))

	if showDevices && len(status.ConnectedDevices) > 0 {
		fmt.Printf("📱 Connected Devices:\n")
		for _, device := range status.ConnectedDevices {
			hostname := device.Hostname
			if hostname == "" {
				hostname = "Unknown"
			}
			fmt.Printf("  %s - %s (%s)\n", device.IP, hostname, device.MAC[:8]+"...")
		}
		fmt.Println()
	}

	if len(status.ActiveConnections) > 0 {
		fmt.Printf("🌐 Recent Connections:\n")
		count := 0
		for _, conn := range status.ActiveConnections {
			if count >= maxConnections {
				break
			}
			fmt.Printf("  %s %s → %s (%s)\n",
				conn.Protocol, conn.Source, conn.Destination, conn.State)
			count++
		}
		if len(status.ActiveConnections) > maxConnections {
			fmt.Printf("  ... and %d more\n", len(status.ActiveConnections)-maxConnections)
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(monitorCmd)

	monitorCmd.Flags().DurationVarP(&refreshInterval, "interval", "i", 2*time.Second, "refresh interval for follow mode")
	monitorCmd.Flags().IntVarP(&maxConnections, "max", "m", 20, "maximum connections to display")
	monitorCmd.Flags().BoolVarP(&showDevices, "devices", "d", false, "show connected devices")
	monitorCmd.Flags().BoolVarP(&followMode, "follow", "f", false, "continuous monitoring mode")
}
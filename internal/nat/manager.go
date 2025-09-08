// Package nat provides core NAT functionality for macOS systems
package nat

import (
	"bufio"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strings"
)

// Config represents the configuration for NAT
type Config struct {
	ExternalInterface string
	InternalInterface string
	InternalNetwork   string
	DHCPRange         DHCPRange
	DNSServers        []string
	Active            bool
}

// DHCPRange represents DHCP IP range configuration
type DHCPRange struct {
	Start string
	End   string
	Lease string
}

// NetworkInterface represents a network interface
type NetworkInterface struct {
	Name   string
	Type   string
	Status string
	IP     string
}

// Connection represents a network connection
type Connection struct {
	Source      string
	Destination string
	Protocol    string
	State       string
}

// Manager manages NAT operations
type Manager struct {
	config  *Config
	dhcpPid int
}

// NewManager creates a new NAT manager
func NewManager(config *Config) *Manager {
	return &Manager{
		config: config,
	}
}

// GetNetworkInterfaces returns a list of available network interfaces
func (m *Manager) GetNetworkInterfaces() ([]NetworkInterface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	var result []NetworkInterface
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		var ip string
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ip = ipnet.IP.String()
					break
				}
			}
		}

		status := "down"
		if iface.Flags&net.FlagUp != 0 {
			status = "up"
		}

		result = append(result, NetworkInterface{
			Name:   iface.Name,
			Type:   getInterfaceType(iface.Name),
			Status: status,
			IP:     ip,
		})
	}

	return result, nil
}

// StartNAT starts the NAT service
func (m *Manager) StartNAT() error {
	if m.config == nil {
		return fmt.Errorf("NAT config is nil")
	}

	// Create bridge interface if it doesn't exist
	if strings.HasPrefix(m.config.InternalInterface, "bridge") {
		cmd := exec.Command("ifconfig", m.config.InternalInterface, "create")
		_ = cmd.Run() // Interface might already exist, which is fine

		// Configure bridge interface
		bridgeIP := m.config.InternalNetwork + ".1"
		cmd = exec.Command("ifconfig", m.config.InternalInterface, "inet", bridgeIP, "netmask", "255.255.255.0")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to configure bridge interface: %w", err)
		}
	}

	// Enable IP forwarding
	cmd := exec.Command("sysctl", "-w", "net.inet.ip.forwarding=1")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable IP forwarding: %w", err)
	}

	// Set up NAT rules with pfctl
	natRule := fmt.Sprintf("nat on %s from %s.0/24 to any -> (%s)",
		m.config.ExternalInterface, m.config.InternalNetwork, m.config.ExternalInterface)

	cmd = exec.Command("pfctl", "-e")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable pfctl: %w", err)
	}

	// Write NAT rule to pfctl
	cmd = exec.Command("sh", "-c", fmt.Sprintf("echo '%s' | pfctl -f -", natRule))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set NAT rule: %w", err)
	}

	// Start DHCP server
	if err := m.startDHCPServer(); err != nil {
		return fmt.Errorf("failed to start DHCP server: %w", err)
	}

	m.config.Active = true
	return nil
}

// StopNAT stops the NAT service
func (m *Manager) StopNAT() error {
	if m.config == nil {
		return fmt.Errorf("NAT config is nil")
	}

	// Disable pfctl
	_ = exec.Command("pfctl", "-d").Run()

	// Destroy bridge interface if we created it
	if strings.HasPrefix(m.config.InternalInterface, "bridge") {
		_ = exec.Command("ifconfig", m.config.InternalInterface, "destroy").Run()
	}

	// Stop DHCP server
	_ = exec.Command("killall", "dnsmasq").Run()

	// Disable IP forwarding
	_ = exec.Command("sysctl", "-w", "net.inet.ip.forwarding=0").Run()

	m.config.Active = false
	return nil
}

// GetActiveConnections returns active network connections
func (m *Manager) GetActiveConnections() ([]Connection, error) {
	connections := make([]Connection, 0)

	cmd := exec.Command("netstat", "-n")
	output, err := cmd.Output()
	if err != nil {
		// Return empty slice instead of error to avoid breaking status
		return connections, nil
	}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	re := regexp.MustCompile(`^(tcp|udp)\s+\d+\s+\d+\s+(\S+)\s+(\S+)\s+(\S+)`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if len(matches) == 5 {
			connections = append(connections, Connection{
				Protocol:    strings.ToUpper(matches[1]),
				Source:      matches[2],
				Destination: matches[3],
				State:       matches[4],
			})
		}
	}

	return connections, nil
}

// IsActive returns whether NAT is currently active
func (m *Manager) IsActive() bool {
	if m.config == nil {
		return false
	}
	return m.config.Active
}

// GetConfig returns the current NAT configuration
func (m *Manager) GetConfig() *Config {
	return m.config
}

// Cleanup performs cleanup operations
func (m *Manager) Cleanup() {
	_ = exec.Command("pfctl", "-d").Run()
	_ = exec.Command("killall", "dnsmasq").Run()
	_ = exec.Command("sysctl", "-w", "net.inet.ip.forwarding=0").Run()
}

// startDHCPServer starts the DHCP server using dnsmasq
func (m *Manager) startDHCPServer() error {
	dhcpRange := fmt.Sprintf("%s.%s,%s.%s,%s",
		m.config.InternalNetwork, m.config.DHCPRange.Start,
		m.config.InternalNetwork, m.config.DHCPRange.End,
		m.config.DHCPRange.Lease)

	args := []string{
		"--interface=" + m.config.InternalInterface,
		"--dhcp-range=" + dhcpRange,
		"--no-daemon",
		"--log-queries",
		"--log-dhcp",
	}

	// Add DNS servers
	for _, dns := range m.config.DNSServers {
		args = append(args, "--server="+dns)
	}

	cmd := exec.Command("dnsmasq", args...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start dnsmasq: %w", err)
	}

	m.dhcpPid = cmd.Process.Pid
	return nil
}

// ConnectedDevice represents a connected device
type ConnectedDevice struct {
	IP        string
	MAC       string
	Hostname  string
	LeaseTime string
}

// Status represents NAT status information
type Status struct {
	Active            bool
	Running           bool // Alias for Active for backward compatibility
	ExternalIP        string
	Uptime            string
	ConnectedDevices  []ConnectedDevice
	ActiveConnections []Connection
	BytesIn           uint64
	BytesOut          uint64
	IPForwarding      bool
	PFCTLEnabled      bool
	DHCPRunning       bool
}

// GetStatus returns current NAT status
func (m *Manager) GetStatus() (*Status, error) {
	connections, _ := m.GetActiveConnections()
	if connections == nil {
		connections = []Connection{}
	}

	isActive := m.IsActive()
	status := &Status{
		Active:            isActive,
		Running:           isActive, // Alias for backward compatibility
		ExternalIP:        "N/A",
		Uptime:            "N/A",
		ConnectedDevices:  []ConnectedDevice{},
		ActiveConnections: connections,
		BytesIn:           0,
		BytesOut:          0,
		IPForwarding:      isActive,
		PFCTLEnabled:      isActive,
		DHCPRunning:       isActive,
	}

	if m.config == nil {
		return status, nil
	}

	// Try to get external IP
	if m.config.ExternalInterface != "" {
		cmd := exec.Command("ifconfig", m.config.ExternalInterface)
		if output, err := cmd.Output(); err == nil {
			re := regexp.MustCompile(`inet (\d+\.\d+\.\d+\.\d+)`)
			if matches := re.FindStringSubmatch(string(output)); len(matches) > 1 {
				status.ExternalIP = matches[1]
			}
		}
	}

	return status, nil
}

// getInterfaceType determines the type of network interface
func getInterfaceType(name string) string {
	if strings.HasPrefix(name, "en") {
		return "Ethernet"
	} else if strings.HasPrefix(name, "wi") || strings.HasPrefix(name, "wlan") {
		return "WiFi"
	} else if strings.HasPrefix(name, "bridge") {
		return "Bridge"
	} else if strings.HasPrefix(name, "lo") {
		return "Loopback"
	}
	return "Other"
}

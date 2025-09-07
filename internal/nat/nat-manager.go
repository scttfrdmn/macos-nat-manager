package nat

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/scttfrdmn/macos-nat-manager/internal/config"
)

// Manager handles NAT operations
type Manager struct {
	config *config.Config
}

// NewManager creates a new NAT manager
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		config: cfg,
	}
}

// Status represents the current NAT status
type Status struct {
	Running           bool                `json:"running"`
	Config            *config.Config      `json:"config"`
	ExternalIP        string              `json:"external_ip"`
	IPForwarding      bool                `json:"ip_forwarding"`
	PFCTLEnabled      bool                `json:"pfctl_enabled"`
	DHCPRunning       bool                `json:"dhcp_running"`
	ConnectedDevices  []ConnectedDevice   `json:"connected_devices"`
	ActiveConnections []ActiveConnection  `json:"active_connections"`
	Uptime            string              `json:"uptime"`
	BytesIn           uint64              `json:"bytes_in"`
	BytesOut          uint64              `json:"bytes_out"`
}

// ConnectedDevice represents a device connected to the internal network
type ConnectedDevice struct {
	IP        string `json:"ip"`
	MAC       string `json:"mac"`
	Hostname  string `json:"hostname"`
	LeaseTime string `json:"lease_time"`
}

// ActiveConnection represents an active network connection
type ActiveConnection struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Protocol    string `json:"protocol"`
	State       string `json:"state"`
}

// NetworkInterface represents a network interface
type NetworkInterface struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Status string `json:"status"`
	IP     string `json:"ip"`
}

// Start initiates the NAT service
func (m *Manager) Start() error {
	// Validate configuration
	if err := m.config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Check if already running
	if running, err := m.IsRunning(); err != nil {
		return fmt.Errorf("failed to check running status: %w", err)
	} else if running {
		return fmt.Errorf("NAT is already running")
	}

	// Enable IP forwarding
	if err := m.enableIPForwarding(); err != nil {
		return fmt.Errorf("failed to enable IP forwarding: %w", err)
	}

	// Setup internal interface
	if err := m.setupInternalInterface(); err != nil {
		m.cleanup() // Cleanup on failure
		return fmt.Errorf("failed to setup internal interface: %w", err)
	}

	// Setup NAT rules
	if err := m.setupNATRules(); err != nil {
		m.cleanup() // Cleanup on failure
		return fmt.Errorf("failed to setup NAT rules: %w", err)
	}

	// Start DHCP server
	if err := m.startDHCPServer(); err != nil {
		m.cleanup() // Cleanup on failure
		return fmt.Errorf("failed to start DHCP server: %w", err)
	}

	// Save state
	if err := m.saveState(); err != nil {
		// Don't fail startup for state save issues
		fmt.Printf("Warning: failed to save state: %v\n", err)
	}

	return nil
}

// Stop terminates the NAT service
func (m *Manager) Stop() error {
	var errors []string

	// Stop DHCP server
	if err := m.stopDHCPServer(); err != nil {
		errors = append(errors, fmt.Sprintf("DHCP server: %v", err))
	}

	// Remove NAT rules
	if err := m.removeNATRules(); err != nil {
		errors = append(errors, fmt.Sprintf("NAT rules: %v", err))
	}

	// Cleanup internal interface
	if err := m.cleanupInternalInterface(); err != nil {
		errors = append(errors, fmt.Sprintf("interface cleanup: %v", err))
	}

	// Disable IP forwarding
	if err := m.disableIPForwarding(); err != nil {
		errors = append(errors, fmt.Sprintf("IP forwarding: %v", err))
	}

	// Remove state
	if err := m.removeState(); err != nil {
		// Don't include in errors as it's not critical
		fmt.Printf("Warning: failed to remove state: %v\n", err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// IsRunning checks if NAT is currently active
func (m *Manager) IsRunning() (bool, error) {
	// Check multiple indicators
	checks := []func() bool{
		m.isPFCTLEnabled,
		m.isDHCPRunning,
		m.isIPForwardingEnabled,
	}

	activeCount := 0
	for _, check := range checks {
		if check() {
			activeCount++
		}
	}

	// Consider it running if at least 2/3 checks pass
	return activeCount >= 2, nil
}

// GetStatus returns the current NAT status
func (m *Manager) GetStatus() (*Status, error) {
	status := &Status{
		Config: m.config,
	}

	var err error
	status.Running, err = m.IsRunning()
	if err != nil {
		return nil, fmt.Errorf("failed to check running status: %w", err)
	}

	status.ExternalIP = m.getExternalIP()
	status.IPForwarding = m.isIPForwardingEnabled()
	status.PFCTLEnabled = m.isPFCTLEnabled()
	status.DHCPRunning = m.isDHCPRunning()

	if status.Running {
		status.ConnectedDevices = m.getConnectedDevices()
		status.ActiveConnections = m.getActiveConnections()
		status.Uptime = m.getUptime()
		status.BytesIn, status.BytesOut = m.getTrafficStats()
	}

	return status, nil
}

// cleanup performs cleanup operations
func (m *Manager) cleanup() {
	// Best effort cleanup - don't return errors
	m.stopDHCPServer()
	m.removeNATRules()
	m.cleanupInternalInterface()
	m.disableIPForwarding()
}

// enableIPForwarding enables IP packet forwarding
func (m *Manager) enableIPForwarding() error {
	return exec.Command("sysctl", "-w", "net.inet.ip.forwarding=1").Run()
}

// disableIPForwarding disables IP packet forwarding
func (m *Manager) disableIPForwarding() error {
	return exec.Command("sysctl", "-w", "net.inet.ip.forwarding=0").Run()
}

// isIPForwardingEnabled checks if IP forwarding is enabled
func (m *Manager) isIPForwardingEnabled() bool {
	cmd := exec.Command("sysctl", "net.inet.ip.forwarding")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "net.inet.ip.forwarding: 1")
}

// setupInternalInterface configures the internal network interface
func (m *Manager) setupInternalInterface() error {
	iface := m.config.InternalInterface

	// If it's a bridge interface, create it
	if strings.HasPrefix(iface, "bridge") {
		// Remove existing bridge if it exists
		exec.Command("ifconfig", iface, "destroy").Run()

		// Create new bridge
		if err := exec.Command("ifconfig", iface, "create").Run(); err != nil {
			return fmt.Errorf("failed to create bridge interface: %w", err)
		}
	}

	// Configure interface with IP address
	gatewayIP := m.config.GetGatewayIP() + "/24"
	if err := exec.Command("ifconfig", iface, gatewayIP, "up").Run(); err != nil {
		return fmt.Errorf("failed to configure interface IP: %w", err)
	}

	return nil
}

// cleanupInternalInterface removes the internal interface
func (m *Manager) cleanupInternalInterface() error {
	iface := m.config.InternalInterface

	// Only destroy bridge interfaces we created
	if strings.HasPrefix(iface, "bridge") {
		return exec.Command("ifconfig", iface, "destroy").Run()
	}

	return nil
}

// setupNATRules configures pfctl NAT rules
func (m *Manager) setupNATRules() error {
	natRules := fmt.Sprintf(`nat on %s from %s to any -> (%s)
pass from %s to any keep state
pass to %s keep state`,
		m.config.ExternalInterface,
		m.config.GetInternalCIDR(),
		m.config.ExternalInterface,
		m.config.GetInternalCIDR(),
		m.config.GetInternalCIDR())

	// Write rules to temporary file
	rulesFile := "/tmp/nat_rules_" + strconv.FormatInt(time.Now().Unix(), 10) + ".conf"
	if err := os.WriteFile(rulesFile, []byte(natRules), 0644); err != nil {
		return fmt.Errorf("failed to write NAT rules: %w", err)
	}

	// Load pfctl rules
	if err := exec.Command("pfctl", "-f", rulesFile).Run(); err != nil {
		os.Remove(rulesFile)
		return fmt.Errorf("failed to load pfctl rules: %w", err)
	}

	// Enable pfctl
	if err := exec.Command("pfctl", "-e").Run(); err != nil {
		os.Remove(rulesFile)
		return fmt.Errorf("failed to enable pfctl: %w", err)
	}

	// Clean up temporary file
	os.Remove(rulesFile)

	return nil
}

// removeNATRules removes pfctl NAT rules
func (m *Manager) removeNATRules() error {
	return exec.Command("pfctl", "-d").Run()
}

// isPFCTLEnabled checks if pfctl is enabled with NAT rules
func (m *Manager) isPFCTLEnabled() bool {
	cmd := exec.Command("pfctl", "-s", "info")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "Status: Enabled")
}

// startDHCPServer starts the DHCP server
func (m *Manager) startDHCPServer() error {
	// Check if dnsmasq is available
	if _, err := exec.LookPath("dnsmasq"); err != nil {
		return fmt.Errorf("dnsmasq not found. Install with: brew install dnsmasq")
	}

	// Stop any existing dnsmasq processes
	exec.Command("killall", "dnsmasq").Run()

	// Start dnsmasq with configuration
	args := []string{
		fmt.Sprintf("--interface=%s", m.config.InternalInterface),
		fmt.Sprintf("--dhcp-range=%s,%s,%s", m.config.DHCPRange.Start, m.config.DHCPRange.End, m.config.DHCPRange.Lease),
		fmt.Sprintf("--dhcp-option=3,%s", m.config.GetGatewayIP()), // Gateway
		fmt.Sprintf("--dhcp-option=6,%s", strings.Join(m.config.DNSServers, ",")), // DNS
		"--bind-interfaces",
		"--except-interface=lo0",
		"--no-daemon",
	}

	cmd := exec.Command("dnsmasq", args...)
	return cmd.Start()
}

// stopDHCPServer stops the DHCP server
func (m *Manager) stopDHCPServer() error {
	return exec.Command("killall", "dnsmasq").Run()
}

// isDHCPRunning checks if DHCP server is running
func (m *Manager) isDHCPRunning() bool {
	cmd := exec.Command("pgrep", "dnsmasq")
	return cmd.Run() == nil
}

// getExternalIP gets the IP address of the external interface
func (m *Manager) getExternalIP() string {
	iface, err := net.InterfaceByName(m.config.ExternalInterface)
	if err != nil {
		return "N/A"
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return "N/A"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return "N/A"
}

// getConnectedDevices returns list of connected devices
func (m *Manager) getConnectedDevices() []ConnectedDevice {
	// This would typically parse DHCP lease file
	// For now, return empty list
	return []ConnectedDevice{}
}

// getActiveConnections returns list of active connections
func (m *Manager) getActiveConnections() []ActiveConnection {
	cmd := exec.Command("netstat", "-n")
	output, err := cmd.Output()
	if err != nil {
		return []ActiveConnection{}
	}

	var connections []ActiveConnection
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	re := regexp.MustCompile(`^(tcp|udp)\s+\d+\s+\d+\s+(\S+)\s+(\S+)\s+(\S+)`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if len(matches) == 5 {
			connections = append(connections, ActiveConnection{
				Protocol:    strings.ToUpper(matches[1]),
				Source:      matches[2],
				Destination: matches[3],
				State:       matches[4],
			})
		}
	}

	return connections
}

// getUptime returns NAT service uptime
func (m *Manager) getUptime() string {
	// This would typically be calculated from startup time
	// For now, return placeholder
	return "Unknown"
}

// getTrafficStats returns traffic statistics
func (m *Manager) getTrafficStats() (uint64, uint64) {
	// This would typically parse interface statistics
	// For now, return zeros
	return 0, 0
}

// saveState saves current state to file
func (m *Manager) saveState() error {
	stateFile, err := config.GetStateFilePath()
	if err != nil {
		return err
	}

	// Create a simple state file indicating NAT is running
	state := fmt.Sprintf("running: true\nstarted: %s\nconfig: %s\n", 
		time.Now().Format(time.RFC3339),
		m.config.ExternalInterface+"->"+m.config.InternalInterface)

	return os.WriteFile(stateFile, []byte(state), 0644)
}

// removeState removes the state file
func (m *Manager) removeState() error {
	stateFile, err := config.GetStateFilePath()
	if err != nil {
		return err
	}

	return os.Remove(stateFile)
}
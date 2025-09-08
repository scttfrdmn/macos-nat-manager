package nat

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	config := &Config{
		ExternalInterface: "en0",
		InternalInterface: "bridge100",
		InternalNetwork:   "192.168.100",
		DHCPRange: DHCPRange{
			Start: "192.168.100.100",
			End:   "192.168.100.200",
			Lease: "12h",
		},
		DNSServers: []string{"8.8.8.8", "8.8.4.4"},
	}

	manager := NewManager(config)

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	if manager.config != config {
		t.Error("Manager config not set correctly")
	}

	if manager.IsActive() {
		t.Error("New manager should not be active")
	}
}

func TestManagerWithNilConfig(t *testing.T) {
	manager := NewManager(nil)

	if manager == nil {
		t.Fatal("NewManager returned nil with nil config")
	}

	if manager.IsActive() {
		t.Error("Manager with nil config should not be active")
	}
}

func TestGetNetworkInterfaces(t *testing.T) {
	manager := NewManager(nil)

	interfaces, err := manager.GetNetworkInterfaces()
	if err != nil {
		t.Errorf("GetNetworkInterfaces failed: %v", err)
	}

	// Should have at least a loopback interface
	if len(interfaces) == 0 {
		t.Error("Expected at least one network interface")
	}

	// Check that interfaces have required fields
	for _, iface := range interfaces {
		if iface.Name == "" {
			t.Error("Interface name should not be empty")
		}
		if iface.Status == "" {
			t.Error("Interface status should not be empty")
		}
		if iface.Type == "" {
			t.Error("Interface type should not be empty")
		}
	}
}

func TestGetStatus(t *testing.T) {
	config := &Config{
		ExternalInterface: "en0",
		InternalInterface: "bridge100",
		InternalNetwork:   "192.168.100",
		DHCPRange: DHCPRange{
			Start: "192.168.100.100",
			End:   "192.168.100.200",
			Lease: "12h",
		},
		DNSServers: []string{"8.8.8.8"},
	}

	manager := NewManager(config)

	status, err := manager.GetStatus()
	if err != nil {
		t.Errorf("GetStatus failed: %v", err)
	}

	if status == nil {
		t.Fatal("GetStatus returned nil status")
	}

	if status.Active {
		t.Error("New manager status should not be active")
	}

	if status.Running {
		t.Error("New manager status should not be running")
	}

	// Status should have empty but initialized fields
	if status.ConnectedDevices == nil {
		t.Error("ConnectedDevices should be initialized")
	}

	if status.ActiveConnections == nil {
		t.Error("ActiveConnections should be initialized")
	}
}

func TestGetActiveConnections(t *testing.T) {
	manager := NewManager(nil)

	connections, err := manager.GetActiveConnections()
	if err != nil {
		t.Errorf("GetActiveConnections failed: %v", err)
	}

	// Should return a slice (may be empty)
	if connections == nil {
		t.Errorf("GetActiveConnections should return non-nil slice, got nil")
	}

	t.Logf("GetActiveConnections returned %d connections", len(connections))
}

func TestGetConfig(t *testing.T) {
	config := &Config{
		ExternalInterface: "en0",
		InternalInterface: "bridge100",
		InternalNetwork:   "192.168.100",
	}

	manager := NewManager(config)
	retrievedConfig := manager.GetConfig()

	if retrievedConfig != config {
		t.Error("GetConfig should return the same config instance")
	}

	// Test with nil config
	nilManager := NewManager(nil)
	if nilManager.GetConfig() != nil {
		t.Error("GetConfig should return nil when manager has nil config")
	}
}

func TestManagerCleanup(t *testing.T) {
	manager := NewManager(nil)

	// Cleanup should not panic even if nothing was set up
	manager.Cleanup()

	// Test with config
	config := &Config{
		ExternalInterface: "en0",
		InternalInterface: "bridge100",
	}
	manager2 := NewManager(config)
	manager2.Cleanup()
}

func TestStartNATWithNilConfig(t *testing.T) {
	manager := NewManager(nil)

	err := manager.StartNAT()
	if err == nil {
		t.Error("StartNAT should fail with nil config")
	}

	expectedErr := "NAT config is nil"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestStopNATWithNilConfig(t *testing.T) {
	manager := NewManager(nil)

	err := manager.StopNAT()
	if err == nil {
		t.Error("StopNAT should fail with nil config")
	}

	expectedErr := "NAT config is nil"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestGetInterfaceType(t *testing.T) {
	testCases := []struct {
		name     string
		expected string
	}{
		{"en0", "Ethernet"},
		{"en1", "Ethernet"},
		{"wi0", "WiFi"},
		{"wlan0", "WiFi"},
		{"bridge100", "Bridge"},
		{"bridge101", "Bridge"},
		{"lo0", "Loopback"},
		{"lo", "Loopback"},
		{"gif0", "Other"},
		{"stf0", "Other"},
		{"unknown", "Other"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getInterfaceType(tc.name)
			if result != tc.expected {
				t.Errorf("getInterfaceType(%s) = %s, expected %s", tc.name, result, tc.expected)
			}
		})
	}
}

func TestDHCPRange(t *testing.T) {
	dhcp := DHCPRange{
		Start: "192.168.100.100",
		End:   "192.168.100.200",
		Lease: "12h",
	}

	if dhcp.Start == "" || dhcp.End == "" || dhcp.Lease == "" {
		t.Error("DHCPRange fields should be properly initialized")
	}
}

func TestNetworkInterface(t *testing.T) {
	iface := NetworkInterface{
		Name:   "en0",
		Type:   "Ethernet",
		Status: "up",
		IP:     "192.168.1.100",
	}

	if iface.Name != "en0" {
		t.Error("NetworkInterface Name not set correctly")
	}
	if iface.Type != "Ethernet" {
		t.Error("NetworkInterface Type not set correctly")
	}
	if iface.Status != "up" {
		t.Error("NetworkInterface Status not set correctly")
	}
	if iface.IP != "192.168.1.100" {
		t.Error("NetworkInterface IP not set correctly")
	}
}

func TestConnection(t *testing.T) {
	conn := Connection{
		Source:      "192.168.100.10:8080",
		Destination: "8.8.8.8:53",
		Protocol:    "TCP",
		State:       "ESTABLISHED",
	}

	if conn.Source != "192.168.100.10:8080" {
		t.Error("Connection Source not set correctly")
	}
	if conn.Destination != "8.8.8.8:53" {
		t.Error("Connection Destination not set correctly")
	}
	if conn.Protocol != "TCP" {
		t.Error("Connection Protocol not set correctly")
	}
	if conn.State != "ESTABLISHED" {
		t.Error("Connection State not set correctly")
	}
}

func TestConnectedDevice(t *testing.T) {
	device := ConnectedDevice{
		IP:        "192.168.100.10",
		MAC:       "aa:bb:cc:dd:ee:ff",
		Hostname:  "test-device",
		LeaseTime: "11h59m",
	}

	if device.IP != "192.168.100.10" {
		t.Error("ConnectedDevice IP not set correctly")
	}
	if device.MAC != "aa:bb:cc:dd:ee:ff" {
		t.Error("ConnectedDevice MAC not set correctly")
	}
	if device.Hostname != "test-device" {
		t.Error("ConnectedDevice Hostname not set correctly")
	}
	if device.LeaseTime != "11h59m" {
		t.Error("ConnectedDevice LeaseTime not set correctly")
	}
}

func TestStatus(t *testing.T) {
	status := &Status{
		Active:            true,
		Running:           true,
		ExternalIP:        "203.0.113.1",
		Uptime:            "2h30m",
		ConnectedDevices:  []ConnectedDevice{},
		ActiveConnections: []Connection{},
		BytesIn:           1024,
		BytesOut:          2048,
		IPForwarding:      true,
		PFCTLEnabled:      true,
		DHCPRunning:       true,
	}

	if !status.Active {
		t.Error("Status Active should be true")
	}
	if !status.Running {
		t.Error("Status Running should be true")
	}
	if status.ExternalIP != "203.0.113.1" {
		t.Error("Status ExternalIP not set correctly")
	}
	if status.Uptime != "2h30m" {
		t.Error("Status Uptime not set correctly")
	}
	if status.ConnectedDevices == nil {
		t.Error("Status ConnectedDevices should be initialized")
	}
	if status.ActiveConnections == nil {
		t.Error("Status ActiveConnections should be initialized")
	}
	if status.BytesIn != 1024 {
		t.Error("Status BytesIn not set correctly")
	}
	if status.BytesOut != 2048 {
		t.Error("Status BytesOut not set correctly")
	}
}

package nat

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	config := &NATConfig{
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
	config := &NATConfig{
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

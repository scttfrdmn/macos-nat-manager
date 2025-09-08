// Package integration provides end-to-end integration tests for NAT manager
// These tests require root privileges and should be run manually or in dedicated test environments
package integration

import (
	"os"
	"os/user"
	"testing"
	"time"

	"github.com/scttfrdmn/macos-nat-manager/internal/config"
	"github.com/scttfrdmn/macos-nat-manager/internal/nat"
)

// TestMain checks if we're running as root before running integration tests
func TestMain(m *testing.M) {
	if !isRoot() {
		println("⚠️  Integration tests require root privileges")
		println("   Run with: sudo go test ./test/integration/...")
		println("   Skipping integration tests in non-root environment")
		os.Exit(0)
	}

	println("🔒 Running integration tests with root privileges...")
	println("⚠️  These tests will modify system network configuration")

	// Run tests
	code := m.Run()

	// Cleanup any remaining NAT configuration
	cleanup()

	os.Exit(code)
}

func isRoot() bool {
	currentUser, err := user.Current()
	if err != nil {
		return false
	}
	return currentUser.Uid == "0"
}

func cleanup() {
	// Best effort cleanup of any lingering NAT configuration
	cfg := &nat.Config{
		ExternalInterface: "en0",
		InternalInterface: "bridge200", // Use test-specific bridge
		InternalNetwork:   "192.168.200",
	}

	manager := nat.NewManager(cfg)
	manager.Cleanup()
}

// TestNATFullLifecycle tests the complete NAT lifecycle with real network operations
func TestNATFullLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	manager := createTestManager(t)
	defer cleanupTestManager(t, manager)

	// Run each phase of the lifecycle test
	t.Run("StartNAT", func(t *testing.T) {
		testStartNAT(t, manager)
	})

	t.Run("VerifyNetworkConfiguration", func(t *testing.T) {
		testNetworkConfiguration(t, manager)
	})

	t.Run("VerifyNATStatus", func(t *testing.T) {
		testNATStatus(t, manager)
	})

	t.Run("TestConnections", func(t *testing.T) {
		testConnections(t, manager)
	})

	t.Run("StopNAT", func(t *testing.T) {
		testStopNAT(t, manager)
	})
}

// Helper functions to reduce complexity

func createTestManager(t *testing.T) *nat.Manager {
	testConfig := &nat.Config{
		ExternalInterface: "en0",       // Assume primary ethernet
		InternalInterface: "bridge200", // Test-specific bridge
		InternalNetwork:   "192.168.200",
		DHCPRange: nat.DHCPRange{
			Start: "192.168.200.100",
			End:   "192.168.200.199",
			Lease: "1h",
		},
		DNSServers: []string{"8.8.8.8", "1.1.1.1"},
	}
	return nat.NewManager(testConfig)
}

func cleanupTestManager(t *testing.T, manager *nat.Manager) {
	t.Log("Cleaning up NAT configuration...")
	err := manager.StopNAT()
	if err != nil {
		t.Logf("Cleanup error (non-fatal): %v", err)
	}
	manager.Cleanup()
}

func testStartNAT(t *testing.T, manager *nat.Manager) {
	t.Log("Starting NAT with real network configuration...")
	err := manager.StartNAT()
	if err != nil {
		t.Fatalf("Failed to start NAT: %v", err)
	}

	if !manager.IsActive() {
		t.Error("Manager should report as active after StartNAT")
	}
}

func testNetworkConfiguration(t *testing.T, manager *nat.Manager) {
	// Allow time for network configuration to settle
	time.Sleep(2 * time.Second)

	interfaces, err := manager.GetNetworkInterfaces()
	if err != nil {
		t.Fatalf("Failed to get network interfaces: %v", err)
	}

	testConfig := manager.GetConfig()
	foundBridge := false
	for _, iface := range interfaces {
		if iface.Name == testConfig.InternalInterface {
			foundBridge = true
			if iface.Status != "up" {
				t.Errorf("Bridge interface should be up, got: %s", iface.Status)
			}
			t.Logf("Found bridge interface: %s (IP: %s)", iface.Name, iface.IP)
			break
		}
	}

	if !foundBridge {
		t.Error("Bridge interface was not created or not found")
	}
}

func testNATStatus(t *testing.T, manager *nat.Manager) {
	status, err := manager.GetStatus()
	if err != nil {
		t.Fatalf("Failed to get NAT status: %v", err)
	}

	if !status.Active {
		t.Error("NAT status should show as active")
	}

	if !status.IPForwarding {
		t.Error("IP forwarding should be enabled")
	}

	if !status.PFCTLEnabled {
		t.Error("pfctl should be enabled")
	}

	t.Logf("NAT Status - Active: %t, External IP: %s", status.Active, status.ExternalIP)
}

func testConnections(t *testing.T, manager *nat.Manager) {
	connections, err := manager.GetActiveConnections()
	if err != nil {
		t.Errorf("Failed to get active connections: %v", err)
	}

	t.Logf("Found %d active connections", len(connections))

	// Log first few connections for debugging
	maxLog := 5
	if len(connections) < maxLog {
		maxLog = len(connections)
	}
	for i := 0; i < maxLog; i++ {
		conn := connections[i]
		t.Logf("Connection %d: %s %s -> %s (%s)",
			i+1, conn.Protocol, conn.Source, conn.Destination, conn.State)
	}
}

func testStopNAT(t *testing.T, manager *nat.Manager) {
	t.Log("Stopping NAT configuration...")
	err := manager.StopNAT()
	if err != nil {
		t.Fatalf("Failed to stop NAT: %v", err)
	}

	if manager.IsActive() {
		t.Error("Manager should not report as active after StopNAT")
	}
}

// TestConfigurationPersistence tests saving and loading configuration
func TestConfigurationPersistence(t *testing.T) {
	tempDir := t.TempDir()
	configPath := tempDir + "/test-config.yaml"

	originalConfig := &config.Config{
		ExternalInterface: "en0",
		InternalInterface: "bridge200",
		InternalNetwork:   "192.168.200",
		DHCPRange: config.DHCPRange{
			Start: "192.168.200.100",
			End:   "192.168.200.199",
			Lease: "1h",
		},
		DNSServers: []string{"8.8.8.8", "1.1.1.1"},
		Active:     false,
	}

	t.Run("SaveConfiguration", func(t *testing.T) {
		err := originalConfig.SaveTo(configPath)
		if err != nil {
			t.Fatalf("Failed to save configuration: %v", err)
		}

		// Verify file was created
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Fatal("Configuration file was not created")
		}
	})

	t.Run("LoadConfiguration", func(t *testing.T) {
		loadedConfig, err := config.LoadFrom(configPath)
		if err != nil {
			t.Fatalf("Failed to load configuration: %v", err)
		}

		// Verify all fields match
		if loadedConfig.ExternalInterface != originalConfig.ExternalInterface {
			t.Errorf("External interface mismatch: got %s, want %s",
				loadedConfig.ExternalInterface, originalConfig.ExternalInterface)
		}

		if loadedConfig.InternalNetwork != originalConfig.InternalNetwork {
			t.Errorf("Internal network mismatch: got %s, want %s",
				loadedConfig.InternalNetwork, originalConfig.InternalNetwork)
		}

		// Test validation
		if err := loadedConfig.Validate(); err != nil {
			t.Errorf("Loaded configuration should be valid: %v", err)
		}
	})
}

// TestNetworkInterfaceDiscovery tests real network interface detection
func TestNetworkInterfaceDiscovery(t *testing.T) {
	manager := nat.NewManager(nil)

	interfaces, err := manager.GetNetworkInterfaces()
	if err != nil {
		t.Fatalf("Failed to get network interfaces: %v", err)
	}

	if len(interfaces) == 0 {
		t.Fatal("Should find at least one network interface")
	}

	t.Logf("Found %d network interfaces:", len(interfaces))

	// Verify interface properties and log for manual verification
	hasLoopback := false
	hasEthernet := false

	for _, iface := range interfaces {
		t.Logf("- %s (%s): %s [%s]", iface.Name, iface.Type, iface.Status, iface.IP)

		// Basic validation
		if iface.Name == "" {
			t.Error("Interface name should not be empty")
		}
		if iface.Type == "" {
			t.Error("Interface type should not be empty")
		}
		if iface.Status == "" {
			t.Error("Interface status should not be empty")
		}

		// Check for expected interface types
		if iface.Type == "Loopback" {
			hasLoopback = true
		}
		if iface.Type == "Ethernet" || iface.Type == "WiFi" {
			hasEthernet = true
		}
	}

	if !hasLoopback {
		t.Error("Should find at least one loopback interface")
	}

	// Note: Don't require ethernet as systems might be WiFi-only
	t.Logf("Found ethernet/wifi interfaces: %t", hasEthernet)
}

// TestSecurityValidation tests security-related functionality
func TestSecurityValidation(t *testing.T) {
	t.Run("RejectInvalidConfigurations", func(t *testing.T) {
		// Test various invalid configurations
		invalidConfigs := []*nat.Config{
			{},                               // Empty config
			{ExternalInterface: "en0"},       // Missing internal interface
			{InternalInterface: "bridge100"}, // Missing external interface
			{
				ExternalInterface: "nonexistent99",
				InternalInterface: "bridge100",
				InternalNetwork:   "192.168.100",
			}, // Nonexistent external interface
		}

		for i, cfg := range invalidConfigs {
			manager := nat.NewManager(cfg)
			err := manager.StartNAT()
			if err == nil {
				t.Errorf("Config %d should have failed validation but didn't", i)
			} else {
				t.Logf("Config %d correctly rejected: %v", i, err)
			}
		}
	})

	t.Run("CleanupOnFailure", func(t *testing.T) {
		// Test that failed operations don't leave system in bad state
		badConfig := &nat.Config{
			ExternalInterface: "nonexistent99",
			InternalInterface: "bridge299",
			InternalNetwork:   "192.168.299",
		}

		manager := nat.NewManager(badConfig)

		// This should fail
		err := manager.StartNAT()
		if err == nil {
			t.Fatal("Expected StartNAT to fail with invalid config")
		}

		// Cleanup should not panic or error
		manager.Cleanup()

		// System should be in clean state
		if manager.IsActive() {
			t.Error("Manager should not report as active after failed start and cleanup")
		}
	})
}

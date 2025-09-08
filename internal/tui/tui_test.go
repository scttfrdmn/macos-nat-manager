package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/scttfrdmn/macos-nat-manager/internal/config"
	"github.com/scttfrdmn/macos-nat-manager/internal/nat"
)

func TestNewApp(t *testing.T) {
	cfg := &config.Config{
		ExternalInterface: "en0",
		InternalInterface: "bridge100",
		InternalNetwork:   "192.168.100",
		DHCPRange: config.DHCPRange{
			Start: "192.168.100.100",
			End:   "192.168.100.200",
			Lease: "12h",
		},
		DNSServers: []string{"8.8.8.8", "8.8.4.4"},
	}

	app := NewApp(cfg)

	if app == nil {
		t.Fatal("NewApp returned nil")
	}

	if app.config != cfg {
		t.Error("App config not set correctly")
	}

	if app.manager == nil {
		t.Error("App manager should be initialized")
	}
}

func TestInitialModel(t *testing.T) {
	cfg := &config.Config{
		ExternalInterface: "en0",
		InternalInterface: "bridge100",
	}

	app := NewApp(cfg)
	model := app.initialModel()

	if model.currentView != "menu" {
		t.Errorf("Expected initial view to be 'menu', got '%s'", model.currentView)
	}

	if model.app != app {
		t.Error("Model app reference not set correctly")
	}

	if model.config != cfg {
		t.Error("Model config reference not set correctly")
	}

	if model.manager == nil {
		t.Error("Model manager should be initialized")
	}
}

func TestModelInit(t *testing.T) {
	cfg := &config.Config{ExternalInterface: "en0"}
	app := NewApp(cfg)
	model := app.initialModel()

	cmd := model.Init()
	if cmd == nil {
		t.Error("Init should return a command")
	}
}

func TestModelHandleWindowSize(t *testing.T) {
	cfg := &config.Config{ExternalInterface: "en0"}
	app := NewApp(cfg)
	model := app.initialModel()

	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModelInterface, cmd := model.handleWindowSize(msg)
	newModel := newModelInterface.(Model)

	if newModel.width != 80 {
		t.Errorf("Expected width 80, got %d", newModel.width)
	}
	if newModel.height != 24 {
		t.Errorf("Expected height 24, got %d", newModel.height)
	}
	if cmd != nil {
		t.Error("handleWindowSize should return nil command")
	}
}

func TestModelHandleInterfaces(t *testing.T) {
	cfg := &config.Config{ExternalInterface: "en0"}
	app := NewApp(cfg)
	model := app.initialModel()

	interfaces := []nat.NetworkInterface{
		{Name: "en0", Type: "Ethernet", Status: "up", IP: "192.168.1.100"},
		{Name: "bridge100", Type: "Bridge", Status: "up", IP: "192.168.100.1"},
	}

	msg := interfacesMsg{interfaces: interfaces}
	newModelInterface, cmd := model.handleInterfaces(msg)
	newModel := newModelInterface.(Model)

	if len(newModel.interfaces) != 2 {
		t.Errorf("Expected 2 interfaces, got %d", len(newModel.interfaces))
	}

	if newModel.interfaces[0].Name != "en0" {
		t.Errorf("Expected first interface to be 'en0', got '%s'", newModel.interfaces[0].Name)
	}

	if cmd != nil {
		t.Error("handleInterfaces should return nil command")
	}
}

func TestModelHandleConnections(t *testing.T) {
	cfg := &config.Config{ExternalInterface: "en0"}
	app := NewApp(cfg)
	model := app.initialModel()

	connections := []nat.Connection{
		{Source: "192.168.100.10:8080", Destination: "8.8.8.8:53", Protocol: "TCP", State: "ESTABLISHED"},
		{Source: "192.168.100.11:443", Destination: "1.1.1.1:53", Protocol: "UDP", State: "ESTABLISHED"},
	}

	msg := connectionsMsg{connections: connections}
	newModelInterface, cmd := model.handleConnections(msg)
	newModel := newModelInterface.(Model)

	if len(newModel.connections) != 2 {
		t.Errorf("Expected 2 connections, got %d", len(newModel.connections))
	}

	if newModel.connections[0].Source != "192.168.100.10:8080" {
		t.Errorf("Expected first connection source '192.168.100.10:8080', got '%s'",
			newModel.connections[0].Source)
	}

	// Check that table rows were set
	if len(newModel.table.Rows()) < 2 {
		t.Error("Table should have connection rows")
	}

	if cmd != nil {
		t.Error("handleConnections should return nil command")
	}
}

func TestModelHandleNATResult(t *testing.T) {
	cfg := &config.Config{ExternalInterface: "en0"}
	app := NewApp(cfg)
	model := app.initialModel()

	// Test successful result
	successMsg := natResultMsg{success: true, err: nil}
	newModelInterface, cmd := model.handleNATResult(successMsg)
	newModel := newModelInterface.(Model)

	if newModel.err != nil {
		t.Error("Error should be nil for successful result")
	}
	if cmd != nil {
		t.Error("handleNATResult should return nil command")
	}
}

func TestModelHandleTick(t *testing.T) {
	cfg := &config.Config{ExternalInterface: "en0"}
	app := NewApp(cfg)
	model := app.initialModel()

	// Test with inactive NAT
	_, cmd := model.handleTick()
	if cmd == nil {
		t.Error("handleTick should return a tick command")
	}
}

func TestInterfaceItem(t *testing.T) {
	iface := nat.NetworkInterface{
		Name:   "en0",
		Type:   "Ethernet",
		Status: "up",
		IP:     "192.168.1.100",
	}

	item := interfaceItem{iface}

	title := item.Title()
	if title != "en0" {
		t.Errorf("Expected title 'en0', got '%s'", title)
	}

	description := item.Description()
	if !contains(description, "Ethernet") {
		t.Error("Description should contain interface type")
	}

	filterValue := item.FilterValue()
	if filterValue != "en0" {
		t.Errorf("Expected filter value 'en0', got '%s'", filterValue)
	}
}

func TestTickMsg(t *testing.T) {
	// Test that tick function returns a tickMsg
	cmd := tick()
	if cmd == nil {
		t.Error("tick() should return a command")
	}
}

func TestGetConfigValue(t *testing.T) {
	// Test with non-empty value
	result := getConfigValue("en0", "N/A")
	if result != "en0" {
		t.Errorf("Expected 'en0', got '%s'", result)
	}

	// Test with empty value
	result2 := getConfigValue("", "N/A")
	if result2 != "N/A" {
		t.Errorf("Expected 'N/A', got '%s'", result2)
	}

	// Test with different default
	result3 := getConfigValue("", "Default")
	if result3 != "Default" {
		t.Errorf("Expected 'Default', got '%s'", result3)
	}
}

// Mock manager for testing
type mockManager struct {
	active bool
}

func (m *mockManager) IsActive() bool {
	return m.active
}

func (m *mockManager) GetNetworkInterfaces() ([]nat.NetworkInterface, error) {
	return []nat.NetworkInterface{
		{Name: "en0", Type: "Ethernet", Status: "up", IP: "192.168.1.100"},
	}, nil
}

func (m *mockManager) StartNAT() error {
	m.active = true
	return nil
}

func (m *mockManager) StopNAT() error {
	m.active = false
	return nil
}

func (m *mockManager) GetActiveConnections() ([]nat.Connection, error) {
	return []nat.Connection{}, nil
}

func (m *mockManager) GetStatus() (*nat.Status, error) {
	return &nat.Status{Active: m.active}, nil
}

func (m *mockManager) GetConfig() *nat.Config {
	return &nat.Config{ExternalInterface: "en0"}
}

func (m *mockManager) Cleanup() {}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

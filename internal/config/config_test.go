package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg == nil {
		t.Fatal("Default() returned nil")
	}

	// Check default values
	if cfg.InternalInterface != "bridge100" {
		t.Errorf("Expected default internal interface 'bridge100', got '%s'", cfg.InternalInterface)
	}

	if cfg.InternalNetwork != "192.168.100" {
		t.Errorf("Expected default internal network '192.168.100', got '%s'", cfg.InternalNetwork)
	}

	if len(cfg.DNSServers) != 2 {
		t.Errorf("Expected 2 default DNS servers, got %d", len(cfg.DNSServers))
	}

	if cfg.DHCPRange.Start != "192.168.100.100" {
		t.Errorf("Expected default DHCP start '192.168.100.100', got '%s'", cfg.DHCPRange.Start)
	}

	if cfg.DHCPRange.End != "192.168.100.200" {
		t.Errorf("Expected default DHCP end '192.168.100.200', got '%s'", cfg.DHCPRange.End)
	}

	if cfg.DHCPRange.Lease != "12h" {
		t.Errorf("Expected default DHCP lease '12h', got '%s'", cfg.DHCPRange.Lease)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				ExternalInterface: "en0",
				InternalInterface: "bridge100",
				InternalNetwork:   "192.168.100",
				DHCPRange: DHCPRange{
					Start: "192.168.100.100",
					End:   "192.168.100.200",
					Lease: "12h",
				},
				DNSServers: []string{"8.8.8.8"},
			},
			wantErr: false,
		},
		{
			name: "missing external interface",
			config: &Config{
				InternalInterface: "bridge100",
				InternalNetwork:   "192.168.100",
				DHCPRange: DHCPRange{
					Start: "192.168.100.100",
					End:   "192.168.100.200",
					Lease: "12h",
				},
			},
			wantErr: true,
		},
		{
			name: "missing internal interface",
			config: &Config{
				ExternalInterface: "en0",
				InternalNetwork:   "192.168.100",
				DHCPRange: DHCPRange{
					Start: "192.168.100.100",
					End:   "192.168.100.200",
					Lease: "12h",
				},
			},
			wantErr: true,
		},
		{
			name: "missing internal network",
			config: &Config{
				ExternalInterface: "en0",
				InternalInterface: "bridge100",
				DHCPRange: DHCPRange{
					Start: "192.168.100.100",
					End:   "192.168.100.200",
					Lease: "12h",
				},
			},
			wantErr: true,
		},
		{
			name: "missing DHCP start",
			config: &Config{
				ExternalInterface: "en0",
				InternalInterface: "bridge100",
				InternalNetwork:   "192.168.100",
				DHCPRange: DHCPRange{
					End:   "192.168.100.200",
					Lease: "12h",
				},
			},
			wantErr: true,
		},
		{
			name: "missing DHCP end",
			config: &Config{
				ExternalInterface: "en0",
				InternalInterface: "bridge100",
				InternalNetwork:   "192.168.100",
				DHCPRange: DHCPRange{
					Start: "192.168.100.100",
					Lease: "12h",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetGatewayIP(t *testing.T) {
	cfg := &Config{
		InternalNetwork: "192.168.100",
	}

	gatewayIP := cfg.GetGatewayIP()
	expected := "192.168.100.1"

	if gatewayIP != expected {
		t.Errorf("GetGatewayIP() = %s, want %s", gatewayIP, expected)
	}
}

func TestGetInternalCIDR(t *testing.T) {
	cfg := &Config{
		InternalNetwork: "192.168.100",
	}

	cidr := cfg.GetInternalCIDR()
	expected := "192.168.100.0/24"

	if cidr != expected {
		t.Errorf("GetInternalCIDR() = %s, want %s", cidr, expected)
	}
}

func TestSaveToAndLoad(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	// Create test config
	originalConfig := &Config{
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

	// Save config
	err := originalConfig.SaveTo(configPath)
	if err != nil {
		t.Fatalf("SaveTo() failed: %v", err)
	}

	// Check file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Load config
	loadedConfig, err := LoadFrom(configPath)
	if err != nil {
		t.Fatalf("LoadFrom() failed: %v", err)
	}

	// Compare configs
	if loadedConfig.ExternalInterface != originalConfig.ExternalInterface {
		t.Errorf("ExternalInterface mismatch: got %s, want %s", 
			loadedConfig.ExternalInterface, originalConfig.ExternalInterface)
	}

	if loadedConfig.InternalInterface != originalConfig.InternalInterface {
		t.Errorf("InternalInterface mismatch: got %s, want %s", 
			loadedConfig.InternalInterface, originalConfig.InternalInterface)
	}

	if loadedConfig.InternalNetwork != originalConfig.InternalNetwork {
		t.Errorf("InternalNetwork mismatch: got %s, want %s", 
			loadedConfig.InternalNetwork, originalConfig.InternalNetwork)
	}

	if len(loadedConfig.DNSServers) != len(originalConfig.DNSServers) {
		t.Errorf("DNSServers length mismatch: got %d, want %d", 
			len(loadedConfig.DNSServers), len(originalConfig.DNSServers))
	}
}
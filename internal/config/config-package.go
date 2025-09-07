package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the NAT manager configuration
type Config struct {
	ExternalInterface string    `yaml:"external_interface" json:"external_interface"`
	InternalInterface string    `yaml:"internal_interface" json:"internal_interface"`
	InternalNetwork   string    `yaml:"internal_network" json:"internal_network"`
	DHCPRange         DHCPRange `yaml:"dhcp_range" json:"dhcp_range"`
	DNSServers        []string  `yaml:"dns_servers" json:"dns_servers"`
	
	// Runtime fields (not saved to config)
	Active bool `yaml:"-" json:"active"`
}

// DHCPRange represents the DHCP IP range configuration
type DHCPRange struct {
	Start string `yaml:"start" json:"start"`
	End   string `yaml:"end" json:"end"`
	Lease string `yaml:"lease" json:"lease"`
}

// Default returns a default configuration
func Default() *Config {
	return &Config{
		ExternalInterface: "",
		InternalInterface: "bridge100",
		InternalNetwork:   "192.168.100",
		DHCPRange: DHCPRange{
			Start: "192.168.100.100",
			End:   "192.168.100.200",
			Lease: "12h",
		},
		DNSServers: []string{"8.8.8.8", "8.8.4.4"},
	}
}

// Load reads configuration from the default location
func Load() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	return LoadFrom(configPath)
}

// LoadFrom reads configuration from the specified path
func LoadFrom(path string) (*Config, error) {
	// If file doesn't exist, return default config
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return Default(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate and set defaults for missing fields
	if config.InternalNetwork == "" {
		config.InternalNetwork = "192.168.100"
	}
	if config.DHCPRange.Start == "" {
		config.DHCPRange.Start = fmt.Sprintf("%s.100", config.InternalNetwork)
	}
	if config.DHCPRange.End == "" {
		config.DHCPRange.End = fmt.Sprintf("%s.200", config.InternalNetwork)
	}
	if config.DHCPRange.Lease == "" {
		config.DHCPRange.Lease = "12h"
	}
	if len(config.DNSServers) == 0 {
		config.DNSServers = []string{"8.8.8.8", "8.8.4.4"}
	}

	return &config, nil
}

// Save writes the configuration to the default location
func (c *Config) Save() error {
	configPath, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	return c.SaveTo(configPath)
}

// SaveTo writes the configuration to the specified path
func (c *Config) SaveTo(path string) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.ExternalInterface == "" {
		return fmt.Errorf("external interface is required")
	}

	if c.InternalInterface == "" {
		return fmt.Errorf("internal interface is required")
	}

	if c.InternalNetwork == "" {
		return fmt.Errorf("internal network is required")
	}

	if c.DHCPRange.Start == "" {
		return fmt.Errorf("DHCP start address is required")
	}

	if c.DHCPRange.End == "" {
		return fmt.Errorf("DHCP end address is required")
	}

	return nil
}

// GetGatewayIP returns the gateway IP for the internal network
func (c *Config) GetGatewayIP() string {
	return fmt.Sprintf("%s.1", c.InternalNetwork)
}

// GetInternalCIDR returns the internal network in CIDR notation
func (c *Config) GetInternalCIDR() string {
	return fmt.Sprintf("%s.0/24", c.InternalNetwork)
}

// getConfigPath returns the default configuration file path
func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".config", "nat-manager", "config.yaml"), nil
}

// GetStateFilePath returns the path for runtime state file
func GetStateFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".config", "nat-manager", "state.yaml"), nil
}
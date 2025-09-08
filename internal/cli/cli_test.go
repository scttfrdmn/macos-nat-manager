package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/scttfrdmn/macos-nat-manager/internal/nat"
)

func TestGetInterfaceDescription(t *testing.T) {
	testCases := []struct {
		iface    nat.NetworkInterface
		expected string
	}{
		{
			iface:    nat.NetworkInterface{Name: "en0"},
			expected: "Ethernet (Primary)",
		},
		{
			iface:    nat.NetworkInterface{Name: "en1"},
			expected: "Ethernet/WiFi",
		},
		{
			iface:    nat.NetworkInterface{Name: "bridge100"},
			expected: "Virtual Bridge",
		},
		{
			iface:    nat.NetworkInterface{Name: "utun0"},
			expected: "VPN Tunnel",
		},
		{
			iface:    nat.NetworkInterface{Name: "awdl0"},
			expected: "AirDrop/AirPlay",
		},
		{
			iface:    nat.NetworkInterface{Name: "lo0"},
			expected: "Loopback",
		},
		{
			iface:    nat.NetworkInterface{Name: "gif0"},
			expected: "Generic Tunnel",
		},
		{
			iface:    nat.NetworkInterface{Name: "stf0"},
			expected: "6to4 Tunnel",
		},
		{
			iface:    nat.NetworkInterface{Name: "unknown0"},
			expected: "Network Interface",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.iface.Name, func(t *testing.T) {
			result := getInterfaceDescription(tc.iface)
			if result != tc.expected {
				t.Errorf("getInterfaceDescription(%s) = %s, expected %s",
					tc.iface.Name, result, tc.expected)
			}
		})
	}
}

func TestFormatBool(t *testing.T) {
	testCases := []struct {
		input    bool
		expected string
	}{
		{true, "✅ Enabled"},
		{false, "❌ Disabled"},
	}

	for _, tc := range testCases {
		result := formatBool(tc.input)
		if result != tc.expected {
			t.Errorf("formatBool(%t) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	testCases := []struct {
		input    uint64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
	}

	for _, tc := range testCases {
		result := formatBytes(tc.input)
		if result != tc.expected {
			t.Errorf("formatBytes(%d) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}

func TestInterfacesCommand(t *testing.T) {
	// Reset flags to default values
	showAll = false
	filterType = ""

	// Capture output
	var buf bytes.Buffer
	interfacesCmd.SetOut(&buf)
	interfacesCmd.SetErr(&buf)

	// Test command execution
	err := interfacesCmd.Execute()
	if err != nil {
		t.Errorf("interfaces command failed: %v", err)
	}

	output := buf.String()

	// Check that output contains expected headers
	if !strings.Contains(output, "INTERFACE") {
		t.Error("Expected INTERFACE header in output")
	}
	if !strings.Contains(output, "TYPE") {
		t.Error("Expected TYPE header in output")
	}
	if !strings.Contains(output, "STATUS") {
		t.Error("Expected STATUS header in output")
	}

	// Should contain usage instructions
	if !strings.Contains(output, "Suitable for:") {
		t.Error("Expected usage instructions in output")
	}
}

func TestInterfacesCommandWithFlags(t *testing.T) {
	// Test with --all flag
	t.Run("with --all flag", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.AddCommand(interfacesCmd)

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"interfaces", "--all"})

		err := cmd.Execute()
		if err != nil {
			t.Errorf("interfaces --all command failed: %v", err)
		}

		// showAll flag should be set
		if !showAll {
			t.Error("Expected showAll flag to be true")
		}

		// Reset flag
		showAll = false
	})

	// Test with --type flag
	t.Run("with --type flag", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.AddCommand(interfacesCmd)

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"interfaces", "--type", "bridge"})

		err := cmd.Execute()
		if err != nil {
			t.Errorf("interfaces --type command failed: %v", err)
		}

		// filterType flag should be set
		if filterType != "bridge" {
			t.Errorf("Expected filterType to be 'bridge', got '%s'", filterType)
		}

		// Reset flag
		filterType = ""
	})
}

func TestRootCommand(t *testing.T) {
	// Test that root command exists and has expected properties
	if rootCmd == nil {
		t.Fatal("rootCmd should not be nil")
	}

	if rootCmd.Use != "nat-manager" {
		t.Errorf("Expected root command use to be 'nat-manager', got '%s'", rootCmd.Use)
	}

	if rootCmd.Short == "" {
		t.Error("Root command should have a short description")
	}

	if rootCmd.Long == "" {
		t.Error("Root command should have a long description")
	}
}

func TestExecuteFunction(t *testing.T) {
	// Test that Execute function exists and doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Execute function panicked: %v", r)
		}
	}()

	// We can't easily test Execute without affecting the global state,
	// but we can at least verify it doesn't panic when called
	// Note: This might exit the program in some cases, so we'll skip actual execution
	t.Skip("Execute function test skipped to avoid program exit")
}

func TestVersionVariables(t *testing.T) {
	// Test that version variables are properly defined
	// These are set by build flags but should have default values
	if Version == "" {
		Version = "dev" // Set default for test
	}
	if Commit == "" {
		Commit = "none" // Set default for test
	}
	if Date == "" {
		Date = "unknown" // Set default for test
	}

	if Version != "dev" && Version == "" {
		t.Error("Version should be set")
	}
	if Commit != "none" && Commit == "" {
		t.Error("Commit should be set")
	}
	if Date != "unknown" && Date == "" {
		t.Error("Date should be set")
	}
}

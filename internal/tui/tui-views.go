package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Margin(1, 0)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Margin(1, 0)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true)

	statusStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))
)

// View renders the current view
func (m Model) View() string {
	switch m.currentView {
	case "menu":
		return m.menuView()
	case "interfaces":
		return m.interfacesView()
	case "config":
		return m.configView()
	case "monitor":
		return m.monitorView()
	case "input":
		return m.inputView()
	default:
		return m.menuView()
	}
}

func (m Model) menuView() string {
	var status string
	if running, _ := m.manager.IsRunning(); running {
		status = successStyle.Render("🟢 NAT Active")
	} else {
		status = errorStyle.Render("🔴 NAT Inactive")
	}

	content := titleStyle.Render("macOS NAT Manager") + "\n\n"
	content += statusStyle.Render(status) + "\n\n"

	if m.config.ExternalInterface != "" && m.config.InternalInterface != "" {
		content += fmt.Sprintf("External: %s → Internal: %s\n", m.config.ExternalInterface, m.config.InternalInterface)
		content += fmt.Sprintf("Network: %s.0/24\n\n", m.config.InternalNetwork)
	} else {
		content += "⚠️  Please configure interfaces before starting NAT\n\n"
	}

	content += "1. Configure Interfaces\n"
	content += "2. Configure NAT Settings\n"
	content += "3. Start NAT\n"
	content += "4. Monitor Connections\n"
	content += "5. Stop NAT\n\n"

	if m.err != nil {
		content += errorStyle.Render(fmt.Sprintf("Error: %s", m.err)) + "\n\n"
		m.err = nil
	}

	content += helpStyle.Render("Press number to select, 'q' to quit")
	return content
}

func (m Model) interfacesView() string {
	content := titleStyle.Render("Network Interfaces") + "\n\n"
	
	if m.config.ExternalInterface != "" || m.config.InternalInterface != "" {
		content += fmt.Sprintf("Current selection - External: %s | Internal: %s\n\n", 
			m.config.ExternalInterface, m.config.InternalInterface)
	}
	
	content += m.list.View() + "\n\n"
	
	// Show interface recommendations
	content += "💡 Recommendations:\n"
	content += "   External: Use active interfaces with internet (en0, en1)\n"
	content += "   Internal: Use bridge interfaces (bridge100, bridge101)\n\n"
	
	content += helpStyle.Render("'e' set external, 'i' set internal, 'r' refresh, 'esc' back")
	return content
}

func (m Model) configView() string {
	content := titleStyle.Render("NAT Configuration") + "\n\n"
	
	// Interface configuration
	content += "🔌 Interfaces:\n"
	content += fmt.Sprintf("   External: %s\n", getConfigValue(m.config.ExternalInterface, "Not set"))
	content += fmt.Sprintf("   Internal: %s\n\n", getConfigValue(m.config.InternalInterface, "Not set"))
	
	// Network configuration
	content += "🌐 Network Settings:\n"
	content += fmt.Sprintf("1. Internal Network: %s.0/24\n", m.config.InternalNetwork)
	content += fmt.Sprintf("2. DHCP Start: %s\n", m.config.DHCPRange.Start)
	content += fmt.Sprintf("3. DHCP End: %s\n", m.config.DHCPRange.End)
	content += fmt.Sprintf("   DHCP Lease: %s\n", m.config.DHCPRange.Lease)
	content += fmt.Sprintf("   DNS Servers: %s\n\n", strings.Join(m.config.DNSServers, ", "))
	
	// Status
	if m.config.ExternalInterface != "" && m.config.InternalInterface != "" {
		content += successStyle.Render("✅ Configuration ready") + "\n\n"
	} else {
		content += errorStyle.Render("❌ Missing interface configuration") + "\n\n"
	}
	
	content += helpStyle.Render("Press number to edit, 'esc' to go back")
	return content
}

func (m Model) monitorView() string {
	content := titleStyle.Render("Connection Monitor") + "\n\n"
	
	// Show current configuration
	content += fmt.Sprintf("🔗 %s (%s) → %s (%s.1/24)\n\n",
		m.config.ExternalInterface,
		getExternalIP(m.manager),
		m.config.InternalInterface,
		m.config.InternalNetwork)
	
	// Connection count
	content += fmt.Sprintf("📊 Active connections: %d\n\n", len(m.connections))
	
	// Connections table
	if len(m.connections) > 0 {
		content += m.table.View() + "\n\n"
	} else {
		content += "No active connections\n\n"
	}
	
	// Statistics
	if status, err := m.manager.GetStatus(); err == nil {
		content += fmt.Sprintf("📈 Uptime: %s\n", status.Uptime)
		content += fmt.Sprintf("📱 Connected devices: %d\n\n", len(status.ConnectedDevices))
	}
	
	content += helpStyle.Render("'r' refresh, 'esc' back")
	return content
}

func (m Model) inputView() string {
	content := titleStyle.Render("Edit Configuration") + "\n\n"
	
	fieldName := ""
	fieldDescription := ""
	
	switch m.inputField {
	case "network":
		fieldName = "Internal Network"
		fieldDescription = "Network prefix for internal devices (e.g., 192.168.100)"
	case "dhcp_start":
		fieldName = "DHCP Range Start"
		fieldDescription = "First IP address in DHCP range (e.g., 192.168.100.100)"
	case "dhcp_end":
		fieldName = "DHCP Range End"
		fieldDescription = "Last IP address in DHCP range (e.g., 192.168.100.200)"
	}
	
	content += fmt.Sprintf("Field: %s\n", fieldName)
	content += fmt.Sprintf("Description: %s\n\n", fieldDescription)
	content += m.textInput.View() + "\n\n"
	content += helpStyle.Render("Enter to save, Esc to cancel")
	return content
}

// Helper functions
func getConfigValue(value, defaultText string) string {
	if value == "" {
		return errorStyle.Render(defaultText)
	}
	return successStyle.Render(value)
}

func getExternalIP(manager *nat.Manager) string {
	if status, err := manager.GetStatus(); err == nil {
		return status.ExternalIP
	}
	return "N/A"
}
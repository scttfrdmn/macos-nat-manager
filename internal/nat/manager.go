package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// NAT Configuration
type NATConfig struct {
	ExternalInterface string
	InternalInterface string
	InternalNetwork   string
	DHCPRange         DHCPRange
	DNSServers        []string
	Active            bool
}

type DHCPRange struct {
	Start string
	End   string
	Lease string
}

// Network Interface
type NetworkInterface struct {
	Name   string
	Type   string
	Status string
	IP     string
}

// Connection info for monitoring
type Connection struct {
	Source      string
	Destination string
	Protocol    string
	State       string
}

// Application state
type model struct {
	state           string
	interfaces      []NetworkInterface
	natConfig       *NATConfig
	connections     []Connection
	list            list.Model
	table           table.Model
	textInput       textinput.Model
	err             error
	width           int
	height          int
	currentView     string
	inputField      string
	dhcpPid         int
}

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

func main() {
	// Check if running as root
	if os.Geteuid() != 0 {
		fmt.Println("This tool requires root privileges. Please run with sudo.")
		os.Exit(1)
	}

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	
	// Handle cleanup on interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		p.Kill()
		cleanup()
		os.Exit(0)
	}()

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func initialModel() model {
	// Initialize text input
	ti := textinput.New()
	ti.Placeholder = "Enter value..."
	ti.CharLimit = 50
	ti.Width = 30

	// Initialize list
	items := []list.Item{}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Network Interfaces"

	// Initialize table
	columns := []table.Column{
		{Title: "Source", Width: 20},
		{Title: "Destination", Width: 20},
		{Title: "Protocol", Width: 10},
		{Title: "State", Width: 12},
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	return model{
		state:       "menu",
		currentView: "menu",
		list:        l,
		table:       t,
		textInput:   ti,
		natConfig: &NATConfig{
			InternalNetwork: "192.168.100",
			DHCPRange: DHCPRange{
				Start: "192.168.100.100",
				End:   "192.168.100.200",
				Lease: "12h",
			},
			DNSServers: []string{"8.8.8.8", "8.8.4.4"},
		},
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		getInterfaces,
		tick(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width-4, msg.Height-10)
		return m, nil

	case interfacesMsg:
		m.interfaces = msg.interfaces
		items := make([]list.Item, len(m.interfaces))
		for i, iface := range m.interfaces {
			items[i] = interfaceItem{iface}
		}
		m.list.SetItems(items)
		return m, nil

	case connectionsMsg:
		m.connections = msg.connections
		rows := make([]table.Row, len(m.connections))
		for i, conn := range m.connections {
			rows[i] = table.Row{conn.Source, conn.Destination, conn.Protocol, conn.State}
		}
		m.table.SetRows(rows)
		return m, nil

	case tickMsg:
		if m.natConfig.Active {
			cmds = append(cmds, getConnections, tick())
		} else {
			cmds = append(cmds, tick())
		}

	case tea.KeyMsg:
		switch m.currentView {
		case "menu":
			return m.handleMenuKeys(msg)
		case "interfaces":
			return m.handleInterfaceKeys(msg)
		case "config":
			return m.handleConfigKeys(msg)
		case "monitor":
			return m.handleMonitorKeys(msg)
		case "input":
			return m.handleInputKeys(msg)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) handleMenuKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		cleanup()
		return m, tea.Quit
	case "1":
		m.currentView = "interfaces"
		return m, getInterfaces
	case "2":
		m.currentView = "config"
		return m, nil
	case "3":
		if m.natConfig.ExternalInterface != "" && m.natConfig.InternalInterface != "" {
			return m, setupNAT(m.natConfig)
		}
		m.err = fmt.Errorf("please configure interfaces first")
		return m, nil
	case "4":
		if m.natConfig.Active {
			m.currentView = "monitor"
			return m, getConnections
		}
		m.err = fmt.Errorf("NAT is not active")
		return m, nil
	case "5":
		if m.natConfig.Active {
			return m, teardownNAT(m.natConfig)
		}
		m.err = fmt.Errorf("NAT is not active")
		return m, nil
	}
	return m, nil
}

func (m model) handleInterfaceKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.currentView = "menu"
		return m, nil
	case "e":
		if len(m.interfaces) > 0 {
			selected := m.list.SelectedItem().(interfaceItem)
			m.natConfig.ExternalInterface = selected.iface.Name
		}
		return m, nil
	case "i":
		if len(m.interfaces) > 0 {
			selected := m.list.SelectedItem().(interfaceItem)
			m.natConfig.InternalInterface = selected.iface.Name
		}
		return m, nil
	case "r":
		return m, getInterfaces
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) handleConfigKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.currentView = "menu"
		return m, nil
	case "1":
		m.currentView = "input"
		m.inputField = "network"
		m.textInput.SetValue(m.natConfig.InternalNetwork)
		m.textInput.Focus()
		return m, nil
	case "2":
		m.currentView = "input"
		m.inputField = "dhcp_start"
		m.textInput.SetValue(m.natConfig.DHCPRange.Start)
		m.textInput.Focus()
		return m, nil
	case "3":
		m.currentView = "input"
		m.inputField = "dhcp_end"
		m.textInput.SetValue(m.natConfig.DHCPRange.End)
		m.textInput.Focus()
		return m, nil
	}
	return m, nil
}

func (m model) handleMonitorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.currentView = "menu"
		return m, nil
	case "r":
		return m, getConnections
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) handleInputKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		value := m.textInput.Value()
		switch m.inputField {
		case "network":
			m.natConfig.InternalNetwork = value
		case "dhcp_start":
			m.natConfig.DHCPRange.Start = value
		case "dhcp_end":
			m.natConfig.DHCPRange.End = value
		}
		m.textInput.Blur()
		m.textInput.SetValue("")
		m.currentView = "config"
		return m, nil
	case "esc":
		m.textInput.Blur()
		m.textInput.SetValue("")
		m.currentView = "config"
		return m, nil
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
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

func (m model) menuView() string {
	var status string
	if m.natConfig.Active {
		status = successStyle.Render("🟢 NAT Active")
	} else {
		status = errorStyle.Render("🔴 NAT Inactive")
	}

	content := titleStyle.Render("macOS NAT Manager") + "\n\n"
	content += statusStyle.Render(status) + "\n\n"

	if m.natConfig.ExternalInterface != "" && m.natConfig.InternalInterface != "" {
		content += fmt.Sprintf("External: %s → Internal: %s\n", m.natConfig.ExternalInterface, m.natConfig.InternalInterface)
		content += fmt.Sprintf("Network: %s.0/24\n\n", m.natConfig.InternalNetwork)
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

func (m model) interfacesView() string {
	content := titleStyle.Render("Network Interfaces") + "\n\n"
	content += fmt.Sprintf("External: %s | Internal: %s\n\n", m.natConfig.ExternalInterface, m.natConfig.InternalInterface)
	content += m.list.View() + "\n\n"
	content += helpStyle.Render("'e' set external, 'i' set internal, 'r' refresh, 'esc' back")
	return content
}

func (m model) configView() string {
	content := titleStyle.Render("NAT Configuration") + "\n\n"
	content += fmt.Sprintf("1. Internal Network: %s.0/24\n", m.natConfig.InternalNetwork)
	content += fmt.Sprintf("2. DHCP Start: %s\n", m.natConfig.DHCPRange.Start)
	content += fmt.Sprintf("3. DHCP End: %s\n", m.natConfig.DHCPRange.End)
	content += fmt.Sprintf("   DHCP Lease: %s\n", m.natConfig.DHCPRange.Lease)
	content += fmt.Sprintf("   DNS Servers: %s\n\n", strings.Join(m.natConfig.DNSServers, ", "))
	content += helpStyle.Render("Press number to edit, 'esc' to go back")
	return content
}

func (m model) monitorView() string {
	content := titleStyle.Render("Connection Monitor") + "\n\n"
	content += fmt.Sprintf("Active connections through NAT (%d total):\n\n", len(m.connections))
	content += m.table.View() + "\n\n"
	content += helpStyle.Render("'r' refresh, 'esc' back")
	return content
}

func (m model) inputView() string {
	content := titleStyle.Render("Enter Value") + "\n\n"
	content += fmt.Sprintf("Field: %s\n\n", m.inputField)
	content += m.textInput.View() + "\n\n"
	content += helpStyle.Render("Enter to save, Esc to cancel")
	return content
}

// Interface item for list
type interfaceItem struct {
	iface NetworkInterface
}

func (i interfaceItem) Title() string       { return i.iface.Name }
func (i interfaceItem) Description() string { 
	return fmt.Sprintf("%s - %s (%s)", i.iface.Type, i.iface.IP, i.iface.Status) 
}
func (i interfaceItem) FilterValue() string { return i.iface.Name }

// Messages
type interfacesMsg struct {
	interfaces []NetworkInterface
}

type connectionsMsg struct {
	connections []Connection
}

type tickMsg time.Time

type natResultMsg struct {
	success bool
	err     error
}

// Commands
func tick() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func getInterfaces() tea.Msg {
	interfaces, err := listNetworkInterfaces()
	if err != nil {
		return interfacesMsg{interfaces: []NetworkInterface{}}
	}
	return interfacesMsg{interfaces: interfaces}
}

func getConnections() tea.Msg {
	connections, _ := getActiveConnections()
	return connectionsMsg{connections: connections}
}

func setupNAT(config *NATConfig) tea.Cmd {
	return func() tea.Msg {
		err := startNAT(config)
		if err != nil {
			return natResultMsg{success: false, err: err}
		}
		config.Active = true
		return natResultMsg{success: true, err: nil}
	}
}

func teardownNAT(config *NATConfig) tea.Cmd {
	return func() tea.Msg {
		err := stopNAT(config)
		if err != nil {
			return natResultMsg{success: false, err: err}
		}
		config.Active = false
		return natResultMsg{success: true, err: nil}
	}
}

// System functions
func listNetworkInterfaces() ([]NetworkInterface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var result []NetworkInterface
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, _ := iface.Addrs()
		var ip string
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ip = ipnet.IP.String()
					break
				}
			}
		}

		status := "Down"
		if iface.Flags&net.FlagUp != 0 {
			status = "Up"
		}

		ifaceType := "Ethernet"
		if strings.Contains(iface.Name, "en") {
			ifaceType = "Ethernet"
		} else if strings.Contains(iface.Name, "bridge") {
			ifaceType = "Bridge"
		} else if strings.Contains(iface.Name, "utun") {
			ifaceType = "VPN"
		}

		result = append(result, NetworkInterface{
			Name:   iface.Name,
			Type:   ifaceType,
			Status: status,
			IP:     ip,
		})
	}

	return result, nil
}

func startNAT(config *NATConfig) error {
	// Enable IP forwarding
	if err := exec.Command("sysctl", "-w", "net.inet.ip.forwarding=1").Run(); err != nil {
		return fmt.Errorf("failed to enable IP forwarding: %w", err)
	}

	// Create bridge interface if it doesn't exist
	if strings.HasPrefix(config.InternalInterface, "bridge") {
		exec.Command("ifconfig", config.InternalInterface, "destroy").Run() // Clean up if exists
		if err := exec.Command("ifconfig", config.InternalInterface, "create").Run(); err != nil {
			return fmt.Errorf("failed to create bridge interface: %w", err)
		}

		// Configure bridge interface
		bridgeIP := fmt.Sprintf("%s.1/24", config.InternalNetwork)
		if err := exec.Command("ifconfig", config.InternalInterface, bridgeIP, "up").Run(); err != nil {
			return fmt.Errorf("failed to configure bridge interface: %w", err)
		}
	}

	// Create pfctl NAT rules
	natRules := fmt.Sprintf(`nat on %s from %s.0/24 to any -> (%s)
pass from %s.0/24 to any keep state
pass to %s.0/24 keep state`,
		config.ExternalInterface,
		config.InternalNetwork,
		config.ExternalInterface,
		config.InternalNetwork,
		config.InternalNetwork)

	// Write rules to temporary file
	rulesFile := "/tmp/nat_rules.conf"
	if err := os.WriteFile(rulesFile, []byte(natRules), 0644); err != nil {
		return fmt.Errorf("failed to write NAT rules: %w", err)
	}

	// Load pfctl rules
	if err := exec.Command("pfctl", "-f", rulesFile).Run(); err != nil {
		return fmt.Errorf("failed to load pfctl rules: %w", err)
	}

	// Enable pfctl
	if err := exec.Command("pfctl", "-e").Run(); err != nil {
		return fmt.Errorf("failed to enable pfctl: %w", err)
	}

	// Start DHCP server using dnsmasq
	return startDHCPServer(config)
}

func startDHCPServer(config *NATConfig) error {
	// Check if dnsmasq is available
	if _, err := exec.LookPath("dnsmasq"); err != nil {
		return fmt.Errorf("dnsmasq not found. Install with: brew install dnsmasq")
	}

	// Kill any existing dnsmasq processes
	exec.Command("killall", "dnsmasq").Run()

	// Start dnsmasq
	args := []string{
		fmt.Sprintf("--interface=%s", config.InternalInterface),
		fmt.Sprintf("--dhcp-range=%s,%s,%s", config.DHCPRange.Start, config.DHCPRange.End, config.DHCPRange.Lease),
		fmt.Sprintf("--dhcp-option=3,%s.1", config.InternalNetwork), // Gateway
		fmt.Sprintf("--dhcp-option=6,%s", strings.Join(config.DNSServers, ",")), // DNS
		"--bind-interfaces",
		"--except-interface=lo0",
		"--no-daemon",
	}

	cmd := exec.Command("dnsmasq", args...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start DHCP server: %w", err)
	}

	return nil
}

func stopNAT(config *NATConfig) error {
	// Disable pfctl
	exec.Command("pfctl", "-d").Run()

	// Destroy bridge interface if we created it
	if strings.HasPrefix(config.InternalInterface, "bridge") {
		exec.Command("ifconfig", config.InternalInterface, "destroy").Run()
	}

	// Stop DHCP server
	exec.Command("killall", "dnsmasq").Run()

	// Disable IP forwarding
	exec.Command("sysctl", "-w", "net.inet.ip.forwarding=0").Run()

	return nil
}

func getActiveConnections() ([]Connection, error) {
	cmd := exec.Command("netstat", "-n")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var connections []Connection
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	re := regexp.MustCompile(`^(tcp|udp)\s+\d+\s+\d+\s+(\S+)\s+(\S+)\s+(\S+)`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if len(matches) == 5 {
			connections = append(connections, Connection{
				Protocol:    strings.ToUpper(matches[1]),
				Source:      matches[2],
				Destination: matches[3],
				State:       matches[4],
			})
		}
	}

	return connections, nil
}

func cleanup() {
	// Clean up any running processes
	exec.Command("pfctl", "-d").Run()
	exec.Command("killall", "dnsmasq").Run()
	exec.Command("sysctl", "-w", "net.inet.ip.forwarding=0").Run()
}
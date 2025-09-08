package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/scttfrdmn/macos-nat-manager/internal/config"
	"github.com/scttfrdmn/macos-nat-manager/internal/nat"
)

// Model represents the TUI application model
type Model struct {
	app         *App
	config      *config.Config
	manager     *nat.Manager
	state       string
	interfaces  []nat.NetworkInterface
	connections []nat.Connection
	list        list.Model
	table       table.Model
	textInput   textinput.Model
	err         error
	width       int
	height      int
	currentView string
	inputField  string
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		getInterfaces(m.manager),
		tick(),
	)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	case natResultMsg:
		if msg.success {
			m.err = nil
		} else {
			m.err = msg.err
		}
		return m, nil

	case tickMsg:
		if m.manager.IsActive() {
			cmds = append(cmds, getConnections(m.manager), tick())
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

func (m Model) handleMenuKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		m.app.cleanup()
		return m, tea.Quit
	case "1":
		m.currentView = "interfaces"
		return m, getInterfaces(m.manager)
	case "2":
		m.currentView = "config"
		return m, nil
	case "3":
		if m.config.ExternalInterface != "" && m.config.InternalInterface != "" {
			return m, setupNAT(m.manager)
		}
		m.err = fmt.Errorf("please configure interfaces first")
		return m, nil
	case "4":
		if m.manager.IsActive() {
			m.currentView = "monitor"
			return m, getConnections(m.manager)
		}
		m.err = fmt.Errorf("NAT is not active")
		return m, nil
	case "5":
		if m.manager.IsActive() {
			return m, teardownNAT(m.manager)
		}
		m.err = fmt.Errorf("NAT is not active")
		return m, nil
	}
	return m, nil
}

func (m Model) handleInterfaceKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.currentView = "menu"
		return m, nil
	case "e":
		if len(m.interfaces) > 0 {
			selected := m.list.SelectedItem().(interfaceItem)
			m.config.ExternalInterface = selected.iface.Name
		}
		return m, nil
	case "i":
		if len(m.interfaces) > 0 {
			selected := m.list.SelectedItem().(interfaceItem)
			m.config.InternalInterface = selected.iface.Name
		}
		return m, nil
	case "r":
		return m, getInterfaces(m.manager)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) handleConfigKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.currentView = "menu"
		return m, nil
	case "1":
		m.currentView = "input"
		m.inputField = "network"
		m.textInput.SetValue(m.config.InternalNetwork)
		m.textInput.Focus()
		return m, nil
	case "2":
		m.currentView = "input"
		m.inputField = "dhcp_start"
		m.textInput.SetValue(m.config.DHCPRange.Start)
		m.textInput.Focus()
		return m, nil
	case "3":
		m.currentView = "input"
		m.inputField = "dhcp_end"
		m.textInput.SetValue(m.config.DHCPRange.End)
		m.textInput.Focus()
		return m, nil
	}
	return m, nil
}

func (m Model) handleMonitorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.currentView = "menu"
		return m, nil
	case "r":
		return m, getConnections(m.manager)
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m Model) handleInputKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		value := m.textInput.Value()
		switch m.inputField {
		case "network":
			m.config.InternalNetwork = value
		case "dhcp_start":
			m.config.DHCPRange.Start = value
		case "dhcp_end":
			m.config.DHCPRange.End = value
		}
		m.textInput.Blur()
		m.textInput.SetValue("")
		m.currentView = "config"
		
		// Save configuration
		if err := m.config.Save(); err != nil {
			m.err = fmt.Errorf("failed to save config: %w", err)
		}
		
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

// Interface item for list
type interfaceItem struct {
	iface nat.NetworkInterface
}

func (i interfaceItem) Title() string { 
	return i.iface.Name 
}

func (i interfaceItem) Description() string { 
	return fmt.Sprintf("%s - %s (%s)", i.iface.Type, i.iface.IP, i.iface.Status) 
}

func (i interfaceItem) FilterValue() string { 
	return i.iface.Name 
}
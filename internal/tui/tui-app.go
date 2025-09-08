// Package tui provides the terminal user interface for the NAT manager
package tui

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/scttfrdmn/macos-nat-manager/internal/config"
	"github.com/scttfrdmn/macos-nat-manager/internal/nat"
)

// App represents the TUI application
type App struct {
	config  *config.Config
	manager *nat.Manager
}

// NewApp creates a new TUI application
func NewApp(cfg *config.Config) *App {
	// Convert config.Config to nat.NATConfig
	natConfig := &nat.NATConfig{
		ExternalInterface: cfg.ExternalInterface,
		InternalInterface: cfg.InternalInterface,
		InternalNetwork:   cfg.InternalNetwork,
		DHCPRange: nat.DHCPRange{
			Start: cfg.DHCPRange.Start,
			End:   cfg.DHCPRange.End,
			Lease: cfg.DHCPRange.Lease,
		},
		DNSServers: cfg.DNSServers,
		Active:     cfg.Active,
	}

	return &App{
		config:  cfg,
		manager: nat.NewManager(natConfig),
	}
}

// Run starts the TUI application
func (a *App) Run() error {
	p := tea.NewProgram(a.initialModel(), tea.WithAltScreen())

	// Handle cleanup on interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		p.Kill()
		a.cleanup()
		os.Exit(0)
	}()

	_, err := p.Run()
	return err
}

func (a *App) initialModel() Model {
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

	return Model{
		app:         a,
		config:      a.config,
		manager:     a.manager,
		state:       "menu",
		currentView: "menu",
		list:        l,
		table:       t,
		textInput:   ti,
	}
}

func (a *App) cleanup() {
	// Attempt to stop NAT service if running
	if a.manager.IsActive() {
		log.Println("Stopping NAT service...")
		if err := a.manager.StopNAT(); err != nil {
			log.Printf("Warning: failed to stop NAT: %v", err)
		}
	}
	a.manager.Cleanup()
}

// Messages for the TUI
type tickMsg time.Time
type interfacesMsg struct {
	interfaces []nat.NetworkInterface
}
type connectionsMsg struct {
	connections []nat.Connection
}
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

func getInterfaces(manager *nat.Manager) tea.Cmd {
	return func() tea.Msg {
		interfaces, err := manager.GetNetworkInterfaces()
		if err != nil {
			return interfacesMsg{interfaces: []nat.NetworkInterface{}}
		}
		return interfacesMsg{interfaces: interfaces}
	}
}

func getConnections(manager *nat.Manager) tea.Cmd {
	return func() tea.Msg {
		connections, err := manager.GetActiveConnections()
		if err != nil {
			return connectionsMsg{connections: []nat.Connection{}}
		}
		return connectionsMsg{connections: connections}
	}
}

func setupNAT(manager *nat.Manager) tea.Cmd {
	return func() tea.Msg {
		err := manager.StartNAT()
		if err != nil {
			return natResultMsg{success: false, err: err}
		}
		return natResultMsg{success: true, err: nil}
	}
}

func teardownNAT(manager *nat.Manager) tea.Cmd {
	return func() tea.Msg {
		err := manager.StopNAT()
		if err != nil {
			return natResultMsg{success: false, err: err}
		}
		return natResultMsg{success: true, err: nil}
	}
}

package cli

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/macos-nat-manager/internal/config"
	"github.com/scttfrdmn/macos-nat-manager/internal/tui"
)

var (
	// Version is the application version, set at build time
	Version = "dev"
	// Commit is the git commit hash, set at build time
	Commit  = "none"
	// Date is the build date, set at build time
	Date    = "unknown"
)

var (
	cfgFile     string
	verbose     bool
	configPath  string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "nat-manager",
	Short: "macOS NAT Manager - True NAT with address translation",
	Long: `macOS NAT Manager provides true Network Address Translation (NAT) 
functionality for macOS, unlike the built-in Internet Sharing which operates 
as a bridge. This tool creates proper address translation, hiding internal 
devices from the upstream network.

Features:
- True NAT implementation using pfctl
- Internal DHCP server with dnsmasq  
- Interactive TUI and CLI interfaces
- Real-time connection monitoring
- Clean setup and teardown
- Network isolation and privacy`,
	Version: fmt.Sprintf("%s (%s) built on %s", Version, Commit, Date),
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is provided, launch TUI
		if len(args) == 0 {
			launchTUI()
		} else {
			_ = cmd.Help()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.nat-manager.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVar(&configPath, "config-path", "", "path to store configuration")

	// Bind flags to viper
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("config-path", rootCmd.PersistentFlags().Lookup("config-path"))
}

// initConfig reads in config file and ENV variables.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".nat-manager" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".nat-manager")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	// Validate we're on macOS
	if runtime.GOOS != "darwin" {
		fmt.Fprintf(os.Stderr, "Error: This tool only works on macOS, detected: %s\n", runtime.GOOS)
		os.Exit(1)
	}

	// Check for root privileges
	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "Error: This tool requires root privileges. Please run with sudo.")
		os.Exit(1)
	}
}

func launchTUI() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	app := tui.NewApp(cfg)
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		os.Exit(1)
	}
}
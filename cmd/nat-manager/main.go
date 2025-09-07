package main

import (
	"fmt"
	"os"

	"github.com/scttfrdmn/macos-nat-manager/internal/cli"
)

// Version information (set by build flags)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Set version info for CLI
	cli.Version = version
	cli.Commit = commit
	cli.Date = date

	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
// Command mtox is a full-featured Tox Messenger terminal user interface.
package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	toxclient "github.com/opd-ai/mtox/internal/tox"
	"github.com/opd-ai/mtox/internal/tui"
	"github.com/opd-ai/mtox/internal/version"
)

// cliFlags holds parsed command-line flags.
type cliFlags struct {
	showHelp    bool
	showVersion bool
	anonOnly    bool
	noTor       bool
	noI2P       bool
	profilePath string
}

// parseFlags parses command-line flags and returns a cliFlags struct.
func parseFlags() cliFlags {
	f := cliFlags{}
	flag.BoolVar(&f.showHelp, "help", false, "Show help message and exit")
	flag.BoolVar(&f.showVersion, "version", false, "Show version and exit")
	flag.BoolVar(&f.anonOnly, "anon-only", false, "Enable anon-only mode (Tor + I2P, no clearnet)")
	flag.BoolVar(&f.noTor, "no-tor", false, "Disable Tor support")
	flag.BoolVar(&f.noI2P, "no-i2p", false, "Disable I2P support")
	flag.StringVar(&f.profilePath, "profile", "", "Custom profile path (default: ~/.config/mtox/profile.tox)")

	flag.Usage = printUsage
	flag.Parse()
	return f
}

// printUsage prints the usage information.
func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: mtox [options]\n\n")
	fmt.Fprintf(os.Stderr, "mtox is a full-featured Tox Messenger terminal user interface.\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
	fmt.Fprintf(os.Stderr, "  MTOX_ANON_ONLY=1    Same as --anon-only\n")
	fmt.Fprintf(os.Stderr, "  MTOX_DISABLE_TOR=1  Same as --no-tor\n")
	fmt.Fprintf(os.Stderr, "  MTOX_DISABLE_I2P=1  Same as --no-i2p\n")
}

// applyFlags sets environment variables based on parsed flags.
func applyFlags(f cliFlags) {
	if f.anonOnly {
		os.Setenv("MTOX_ANON_ONLY", "1")
	}
	if f.noTor {
		os.Setenv("MTOX_DISABLE_TOR", "1")
	}
	if f.noI2P {
		os.Setenv("MTOX_DISABLE_I2P", "1")
	}
	if f.profilePath != "" {
		os.Setenv("MTOX_PROFILE_PATH", f.profilePath)
	}
}

func main() {
	flags := parseFlags()

	if flags.showHelp {
		flag.Usage()
		os.Exit(0)
	}
	if flags.showVersion {
		fmt.Printf("mtox v%s\n", version.Version)
		os.Exit(0)
	}

	applyFlags(flags)

	client, err := toxclient.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initialising Tox: %v\n", err)
		os.Exit(1)
	}

	client.Start()

	app := tui.New(client)
	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running mtox: %v\n", err)
		client.Stop()
		os.Exit(1)
	}
}

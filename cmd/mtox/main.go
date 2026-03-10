// Command mtox is a full-featured Tox Messenger terminal user interface.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	toxclient "github.com/opd-ai/mtox/internal/tox"
	"github.com/opd-ai/mtox/internal/tui"
)

func main() {
	client, err := toxclient.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initialising Tox: %v\n", err)
		os.Exit(1)
	}

	client.Start()

	app := tui.New(client)

	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running mtox: %v\n", err)
		// Ensure cleanup.
		client.Stop()
		os.Exit(1)
	}
}

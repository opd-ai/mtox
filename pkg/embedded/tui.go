// Package embedded provides an embeddable mtox TUI runtime for host applications.
package embedded

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	toxclient "github.com/opd-ai/mtox/internal/tox"
	"github.com/opd-ai/mtox/internal/tui"
)

// TUI wraps the mtox client and Bubble Tea program so it can be embedded and reused.
type TUI struct {
	client  *toxclient.Client
	program *tea.Program
}

// New creates a new embeddable mtox TUI with the default terminal options.
// Additional Bubble Tea program options are appended after defaults.
func New(programOptions ...tea.ProgramOption) (*TUI, error) {
	client, err := toxclient.NewClient()
	if err != nil {
		return nil, fmt.Errorf("initializing tox client: %w", err)
	}
	return newWithClient(client, programOptions...), nil
}

// newWithClient creates a new embeddable mtox TUI from an existing tox client.
// Additional Bubble Tea program options are appended after defaults.
func newWithClient(client *toxclient.Client, programOptions ...tea.ProgramOption) *TUI {
	defaultOptions := []tea.ProgramOption{
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	}
	options := append(defaultOptions, programOptions...)
	return &TUI{
		client:  client,
		program: tea.NewProgram(tui.New(client), options...),
	}
}

// Run starts the tox client and runs the TUI program.
func (t *TUI) Run() error {
	if t == nil || t.client == nil || t.program == nil {
		return fmt.Errorf("embedded tui is not initialized")
	}

	t.client.Start()
	defer t.client.Stop()

	if _, err := t.program.Run(); err != nil {
		return fmt.Errorf("running mtox tui: %w", err)
	}
	return nil
}

// Program returns the underlying Bubble Tea program for advanced integrations
// such as sending custom messages or using Bubble Tea program APIs directly.
func (t *TUI) Program() *tea.Program {
	if t == nil {
		return nil
	}
	return t.program
}

// Stop stops the underlying tox client.
// It is safe to call multiple times.
func (t *TUI) Stop() {
	if t == nil || t.client == nil {
		return
	}
	t.client.Stop()
}

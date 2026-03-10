// Package tui implements the terminal user interface for mtox.
package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Panel borders and backgrounds.
	activePanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))

	inactivePanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))

	// Contact list item styles.
	contactSelected = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			PaddingLeft(1).
			PaddingRight(1)

	contactNormal = lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1)

	// Status indicator colours.
	colorOnline  = lipgloss.Color("76")
	colorAway    = lipgloss.Color("214")
	colorBusy    = lipgloss.Color("196")
	colorOffline = lipgloss.Color("240")

	// Status bar styles.
	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("252")).
			PaddingLeft(1).
			PaddingRight(1)

	statusConnected = lipgloss.NewStyle().
			Foreground(lipgloss.Color("76")).
			Bold(true)

	statusDisconnected = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true)

	// Chat panel styles.
	chatHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("229")).
			PaddingLeft(1)

	messageTimestamp = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243"))

	messageSelf = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))

	messagePeer = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	messageAction = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Italic(true)

	typingIndicator = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")).
			Italic(true)

	// Modal dialog style.
	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2).
			Background(lipgloss.Color("235"))

	// Unread badge style.
	unreadBadge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("235")).
			Background(lipgloss.Color("205")).
			Bold(true).
			PaddingLeft(1).
			PaddingRight(1)

	// Panel title.
	panelTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("229")).
			PaddingLeft(1).
			Underline(true)
)

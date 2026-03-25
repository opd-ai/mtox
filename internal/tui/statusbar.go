package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	toxclient "github.com/opd-ai/mtox/internal/tox"
	"github.com/opd-ai/mtox/internal/version"
	"github.com/opd-ai/toxcore"
)

// statusBar renders the bottom status bar.
type statusBar struct {
	connectionStatus toxcore.ConnectionStatus
	selfAddress      string
	width            int
	torStatus        toxclient.AnonymityStatus
	i2pStatus        toxclient.AnonymityStatus
}

// newStatusBar creates a new status bar with the given width.
func newStatusBar(width int) statusBar {
	return statusBar{width: width}
}

// view renders the status bar.
func (s statusBar) view() string {
	connStr := s.connectionString()
	anonStr := s.anonymityString()
	addrStr := s.addressString()
	versionStr := "mtox v" + version.Version

	left := statusBarStyle.Render(connStr)
	right := statusBarStyle.Render(versionStr)

	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)

	// Only include anon segment if there's content to show
	var anon string
	var anonWidth int
	if anonStr != "" {
		anon = statusBarStyle.Render(anonStr)
		anonWidth = lipgloss.Width(anon)
	}

	midWidth := s.width - leftWidth - anonWidth - rightWidth
	if midWidth < 0 {
		midWidth = 0
	}

	mid := statusBarStyle.Width(midWidth).Render(addrStr)

	var bar string
	if anonStr != "" {
		bar = lipgloss.JoinHorizontal(lipgloss.Top, left, anon, mid, right)
	} else {
		bar = lipgloss.JoinHorizontal(lipgloss.Top, left, mid, right)
	}

	// Pad to full width.
	if lipgloss.Width(bar) < s.width {
		bar += statusBarStyle.Width(s.width - lipgloss.Width(bar)).Render("")
	}

	return bar
}

// connectionString returns a formatted connection status string.
func (s statusBar) connectionString() string {
	switch s.connectionStatus {
	case toxcore.ConnectionUDP:
		return statusConnected.Render("🟢 Connected (UDP)")
	case toxcore.ConnectionTCP:
		return statusConnected.Render("🟡 Connected (TCP)")
	default:
		return statusDisconnected.Render("🔴 Disconnected")
	}
}

// anonymityString returns Tor/I2P status indicators.
func (s statusBar) anonymityString() string {
	var parts []string

	// Tor status indicator
	switch s.torStatus {
	case toxclient.AnonymityAvailable:
		parts = append(parts, "🧅Tor")
	case toxclient.AnonymityConnecting:
		parts = append(parts, "🧅…")
	}

	// I2P status indicator
	switch s.i2pStatus {
	case toxclient.AnonymityAvailable:
		parts = append(parts, "🧄I2P")
	case toxclient.AnonymityConnecting:
		parts = append(parts, "🧄…")
	}

	if len(parts) == 0 {
		return ""
	}
	return " " + strings.Join(parts, " ") + " "
}

// addressString returns a formatted Tox address string.
func (s statusBar) addressString() string {
	if s.selfAddress == "" {
		return "My ID: (loading...)"
	}
	// Show first 16 chars of address.
	addr := s.selfAddress
	if len(addr) > 16 {
		addr = addr[:16] + "..."
	}
	return fmt.Sprintf("My ID: %s", strings.ToUpper(addr))
}

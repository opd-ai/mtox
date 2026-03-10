package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/opd-ai/toxcore"
)

// statusBar renders the bottom status bar.
type statusBar struct {
	connectionStatus toxcore.ConnectionStatus
	selfAddress      string
	width            int
}

func newStatusBar(width int) statusBar {
	return statusBar{width: width}
}

func (s statusBar) view() string {
	connStr := s.connectionString()
	addrStr := s.addressString()
	versionStr := "mtox v0.1"

	left := statusBarStyle.Render(connStr)
	mid := statusBarStyle.Render(addrStr)
	right := statusBarStyle.Render(versionStr)

	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	midWidth := s.width - leftWidth - rightWidth
	if midWidth < 0 {
		midWidth = 0
	}

	mid = statusBarStyle.Width(midWidth).Render(addrStr)

	bar := lipgloss.JoinHorizontal(lipgloss.Top, left, mid, right)

	// Pad to full width.
	if lipgloss.Width(bar) < s.width {
		bar += statusBarStyle.Width(s.width - lipgloss.Width(bar)).Render("")
	}

	return bar
}

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

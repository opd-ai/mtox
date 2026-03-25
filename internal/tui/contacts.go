package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/opd-ai/toxcore"
)

// contactEntry holds display info for a single contact.
type contactEntry struct {
	FriendID         uint32
	Name             string
	StatusMessage    string
	ConnectionStatus toxcore.ConnectionStatus
	FriendStatus     toxcore.FriendStatus
	UnreadCount      int
}

// contactsPanel is the left contacts list panel.
type contactsPanel struct {
	contacts []contactEntry
	selected int
	focused  bool
	width    int
	height   int
}

// newContactsPanel creates a new contacts panel with the given dimensions.
func newContactsPanel(width, height int) contactsPanel {
	return contactsPanel{
		width:  width,
		height: height,
	}
}

// setContacts updates the contact list from the tox friend list.
func (p *contactsPanel) setContacts(friends map[uint32]*toxcore.Friend) {
	// Preserve unread counts across updates.
	unread := make(map[uint32]int, len(p.contacts))
	for _, c := range p.contacts {
		unread[c.FriendID] = c.UnreadCount
	}

	p.contacts = make([]contactEntry, 0, len(friends))
	for id, f := range friends {
		name := f.Name
		if name == "" {
			name = fmt.Sprintf("Friend %d", id)
		}
		p.contacts = append(p.contacts, contactEntry{
			FriendID:         id,
			Name:             name,
			StatusMessage:    f.StatusMessage,
			ConnectionStatus: f.ConnectionStatus,
			FriendStatus:     f.Status,
			UnreadCount:      unread[id],
		})
	}

	// Sort by name for stable ordering.
	sortContacts(p.contacts)

	if p.selected >= len(p.contacts) && len(p.contacts) > 0 {
		p.selected = len(p.contacts) - 1
	}
}

// sortContacts sorts contacts by name using a simple insertion sort to avoid importing sort.
func sortContacts(cs []contactEntry) {
	for i := 1; i < len(cs); i++ {
		for j := i; j > 0 && cs[j].Name < cs[j-1].Name; j-- {
			cs[j], cs[j-1] = cs[j-1], cs[j]
		}
	}
}

// incrementUnread increases the unread count for a contact.
func (p *contactsPanel) incrementUnread(friendID uint32) {
	for i := range p.contacts {
		if p.contacts[i].FriendID == friendID {
			p.contacts[i].UnreadCount++
			return
		}
	}
}

// clearUnread resets the unread count for a contact to zero.
func (p *contactsPanel) clearUnread(friendID uint32) {
	for i := range p.contacts {
		if p.contacts[i].FriendID == friendID {
			p.contacts[i].UnreadCount = 0
			return
		}
	}
}

// updateConnectionStatus updates the connection status for a contact.
func (p *contactsPanel) updateConnectionStatus(friendID uint32, status toxcore.ConnectionStatus) {
	for i := range p.contacts {
		if p.contacts[i].FriendID == friendID {
			p.contacts[i].ConnectionStatus = status
			return
		}
	}
}

// updateName updates the display name for a contact.
func (p *contactsPanel) updateName(friendID uint32, name string) {
	for i := range p.contacts {
		if p.contacts[i].FriendID == friendID {
			p.contacts[i].Name = name
			return
		}
	}
}

// selectedFriendID returns the ID of the currently selected contact, or 0 if none.
func (p *contactsPanel) selectedFriendID() (uint32, bool) {
	if len(p.contacts) == 0 {
		return 0, false
	}
	return p.contacts[p.selected].FriendID, true
}

// update handles keyboard and mouse events for the contacts panel.
func (p *contactsPanel) update(msg tea.Msg) (bool, uint32) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		return p.handleKeyInput(m)
	case tea.MouseMsg:
		return p.handleMouseInput(m)
	}
	return false, 0
}

// handleKeyInput processes keyboard navigation in the contacts panel.
func (p *contactsPanel) handleKeyInput(m tea.KeyMsg) (bool, uint32) {
	if !p.focused {
		return false, 0
	}
	switch m.String() {
	case "up", "k":
		p.moveSelectionUp()
	case "down", "j":
		p.moveSelectionDown()
	case "enter":
		if id, ok := p.selectedFriendID(); ok {
			return true, id
		}
	}
	return false, 0
}

// handleMouseInput processes mouse clicks in the contacts panel.
func (p *contactsPanel) handleMouseInput(m tea.MouseMsg) (bool, uint32) {
	if m.Action != tea.MouseActionPress || m.Button != tea.MouseButtonLeft {
		return false, 0
	}
	// Adjust for border (1) + title (2) = row offset 3.
	row := m.Y - 3
	if row >= 0 && row < len(p.contacts) {
		p.selected = row
		return true, p.contacts[row].FriendID
	}
	return false, 0
}

// moveSelectionUp moves the selection cursor up one position.
func (p *contactsPanel) moveSelectionUp() {
	if p.selected > 0 {
		p.selected--
	}
}

// moveSelectionDown moves the selection cursor down one position.
func (p *contactsPanel) moveSelectionDown() {
	if p.selected < len(p.contacts)-1 {
		p.selected++
	}
}

// view renders the contacts panel.
func (p contactsPanel) view() string {
	inner := strings.Builder{}

	title := panelTitle.Render("Contacts")
	inner.WriteString(title)
	inner.WriteString("\n")
	inner.WriteString(strings.Repeat("─", max(0, p.width-4)))
	inner.WriteString("\n")

	availHeight := p.height - 5 // account for border (2), title (1), divider (1), padding (1)
	start := 0
	if p.selected >= availHeight {
		start = p.selected - availHeight + 1
	}

	for i := start; i < len(p.contacts) && i-start < availHeight; i++ {
		c := p.contacts[i]
		line := p.renderContact(c, i == p.selected)
		inner.WriteString(line)
		inner.WriteString("\n")
	}

	if len(p.contacts) == 0 {
		inner.WriteString(contactNormal.Render("(no contacts)"))
		inner.WriteString("\n")
	}

	content := inner.String()
	innerWidth := p.width - 2

	style := inactivePanel
	if p.focused {
		style = activePanel
	}

	return style.Width(innerWidth).Height(p.height - 2).Render(content)
}

// renderContact formats a single contact entry for display.
func (p contactsPanel) renderContact(c contactEntry, selected bool) string {
	indicator := statusIndicator(c.ConnectionStatus, c.FriendStatus)
	name := c.Name
	if len(name) > p.width-8 {
		name = name[:p.width-11] + "..."
	}

	line := fmt.Sprintf("%s %s", indicator, name)

	if c.UnreadCount > 0 {
		badge := unreadBadge.Render(fmt.Sprintf("%d", c.UnreadCount))
		// Pad the line and append badge at right.
		maxNameLen := p.width - 10
		if len(line) > maxNameLen {
			line = line[:maxNameLen]
		}
		line = fmt.Sprintf("%-*s %s", maxNameLen, line, badge)
	}

	if selected {
		return contactSelected.Width(p.width - 4).Render(line)
	}
	return contactNormal.Width(p.width - 4).Render(line)
}

// statusIndicator returns a coloured status indicator character.
// It considers both connection status (online/offline) and friend status (away/busy).
func statusIndicator(connStatus toxcore.ConnectionStatus, friendStatus toxcore.FriendStatus) string {
	if connStatus == toxcore.ConnectionNone {
		return lipgloss.NewStyle().Foreground(colorOffline).Render("○")
	}
	// Online - check friend status for away/busy
	switch friendStatus {
	case toxcore.FriendStatusAway:
		return lipgloss.NewStyle().Foreground(colorAway).Render("◌")
	case toxcore.FriendStatusBusy:
		return lipgloss.NewStyle().Foreground(colorBusy).Render("◉")
	default:
		return lipgloss.NewStyle().Foreground(colorOnline).Render("●")
	}
}

// max returns the larger of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

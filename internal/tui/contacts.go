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

func newContactsPanel(width, height int) contactsPanel {
	return contactsPanel{
		width:  width,
		height: height,
	}
}

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

func (p *contactsPanel) incrementUnread(friendID uint32) {
	for i := range p.contacts {
		if p.contacts[i].FriendID == friendID {
			p.contacts[i].UnreadCount++
			return
		}
	}
}

func (p *contactsPanel) clearUnread(friendID uint32) {
	for i := range p.contacts {
		if p.contacts[i].FriendID == friendID {
			p.contacts[i].UnreadCount = 0
			return
		}
	}
}

func (p *contactsPanel) updateConnectionStatus(friendID uint32, status toxcore.ConnectionStatus) {
	for i := range p.contacts {
		if p.contacts[i].FriendID == friendID {
			p.contacts[i].ConnectionStatus = status
			return
		}
	}
}

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
		if !p.focused {
			return false, 0
		}
		switch m.String() {
		case "up", "k":
			if p.selected > 0 {
				p.selected--
			}
		case "down", "j":
			if p.selected < len(p.contacts)-1 {
				p.selected++
			}
		case "enter":
			if id, ok := p.selectedFriendID(); ok {
				return true, id
			}
		}
	case tea.MouseMsg:
		if m.Action == tea.MouseActionPress && m.Button == tea.MouseButtonLeft {
			// Adjust for border (1) + title (2) = row offset 3.
			row := m.Y - 3
			if row >= 0 && row < len(p.contacts) {
				p.selected = row
				return true, p.contacts[row].FriendID
			}
		}
	}
	return false, 0
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

func (p contactsPanel) renderContact(c contactEntry, selected bool) string {
	indicator := statusIndicator(c.ConnectionStatus)
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
func statusIndicator(status toxcore.ConnectionStatus) string {
	switch status {
	case toxcore.ConnectionUDP, toxcore.ConnectionTCP:
		return lipgloss.NewStyle().Foreground(colorOnline).Render("●")
	default:
		return lipgloss.NewStyle().Foreground(colorOffline).Render("○")
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

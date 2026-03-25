package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/opd-ai/toxcore"
)

func TestSortContacts(t *testing.T) {
	contacts := []contactEntry{
		{FriendID: 3, Name: "Charlie"},
		{FriendID: 1, Name: "Alice"},
		{FriendID: 2, Name: "Bob"},
	}

	sortContacts(contacts)

	expected := []string{"Alice", "Bob", "Charlie"}
	for i, name := range expected {
		if contacts[i].Name != name {
			t.Errorf("sortContacts: index %d = %q, want %q", i, contacts[i].Name, name)
		}
	}
}

func TestSortContacts_Empty(t *testing.T) {
	var contacts []contactEntry
	sortContacts(contacts)
	if len(contacts) != 0 {
		t.Error("sortContacts should handle empty slice")
	}
}

func TestSortContacts_Single(t *testing.T) {
	contacts := []contactEntry{{FriendID: 1, Name: "Solo"}}
	sortContacts(contacts)
	if contacts[0].Name != "Solo" {
		t.Error("sortContacts should handle single element")
	}
}

func TestContactsPanel_IncrementUnread(t *testing.T) {
	p := newContactsPanel(20, 10)
	p.contacts = []contactEntry{
		{FriendID: 1, Name: "Alice", UnreadCount: 0},
		{FriendID: 2, Name: "Bob", UnreadCount: 5},
	}

	p.incrementUnread(1)
	if p.contacts[0].UnreadCount != 1 {
		t.Errorf("incrementUnread(1): UnreadCount = %d, want 1", p.contacts[0].UnreadCount)
	}

	p.incrementUnread(2)
	if p.contacts[1].UnreadCount != 6 {
		t.Errorf("incrementUnread(2): UnreadCount = %d, want 6", p.contacts[1].UnreadCount)
	}

	// Increment non-existent friend should be no-op.
	p.incrementUnread(999)
}

func TestContactsPanel_ClearUnread(t *testing.T) {
	p := newContactsPanel(20, 10)
	p.contacts = []contactEntry{
		{FriendID: 1, Name: "Alice", UnreadCount: 3},
	}

	p.clearUnread(1)
	if p.contacts[0].UnreadCount != 0 {
		t.Errorf("clearUnread(1): UnreadCount = %d, want 0", p.contacts[0].UnreadCount)
	}

	// Clear non-existent friend should be no-op.
	p.clearUnread(999)
}

func TestContactsPanel_UpdateConnectionStatus(t *testing.T) {
	p := newContactsPanel(20, 10)
	p.contacts = []contactEntry{
		{FriendID: 1, Name: "Alice", ConnectionStatus: toxcore.ConnectionNone},
	}

	p.updateConnectionStatus(1, toxcore.ConnectionUDP)
	if p.contacts[0].ConnectionStatus != toxcore.ConnectionUDP {
		t.Errorf("updateConnectionStatus: got %v, want ConnectionUDP", p.contacts[0].ConnectionStatus)
	}
}

func TestContactsPanel_UpdateName(t *testing.T) {
	p := newContactsPanel(20, 10)
	p.contacts = []contactEntry{
		{FriendID: 1, Name: "OldName"},
	}

	p.updateName(1, "NewName")
	if p.contacts[0].Name != "NewName" {
		t.Errorf("updateName: got %q, want %q", p.contacts[0].Name, "NewName")
	}
}

func TestContactsPanel_SelectedFriendID(t *testing.T) {
	p := newContactsPanel(20, 10)

	// Empty contacts.
	_, ok := p.selectedFriendID()
	if ok {
		t.Error("selectedFriendID should return false for empty contacts")
	}

	p.contacts = []contactEntry{
		{FriendID: 5, Name: "Eve"},
		{FriendID: 7, Name: "Grace"},
	}
	p.selected = 1

	id, ok := p.selectedFriendID()
	if !ok {
		t.Error("selectedFriendID should return true when contacts exist")
	}
	if id != 7 {
		t.Errorf("selectedFriendID: got %d, want 7", id)
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{1, 2, 2},
		{5, 3, 5},
		{-1, -5, -1},
		{0, 0, 0},
	}

	for _, tt := range tests {
		got := max(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("max(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestContactsPanel_UpdateKeyNav(t *testing.T) {
	p := newContactsPanel(20, 10)
	p.focused = true
	p.contacts = []contactEntry{
		{FriendID: 1, Name: "Alice"},
		{FriendID: 2, Name: "Bob"},
		{FriendID: 3, Name: "Charlie"},
	}
	p.selected = 0

	// Down navigation.
	_, _ = p.update(tea.KeyMsg{Type: tea.KeyDown})
	if p.selected != 1 {
		t.Errorf("down nav: selected = %d, want 1", p.selected)
	}

	// Down with 'j'.
	p.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if p.selected != 2 {
		t.Errorf("j nav: selected = %d, want 2", p.selected)
	}

	// Up navigation.
	p.update(tea.KeyMsg{Type: tea.KeyUp})
	if p.selected != 1 {
		t.Errorf("up nav: selected = %d, want 1", p.selected)
	}

	// Up with 'k'.
	p.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if p.selected != 0 {
		t.Errorf("k nav: selected = %d, want 0", p.selected)
	}

	// Up at top should stay at 0.
	p.update(tea.KeyMsg{Type: tea.KeyUp})
	if p.selected != 0 {
		t.Errorf("up at top: selected = %d, want 0", p.selected)
	}
}

func TestContactsPanel_UpdateKeyNav_Bottom(t *testing.T) {
	p := newContactsPanel(20, 10)
	p.focused = true
	p.contacts = []contactEntry{
		{FriendID: 1, Name: "Alice"},
		{FriendID: 2, Name: "Bob"},
	}
	p.selected = 1 // at bottom

	// Down at bottom should stay at bottom.
	p.update(tea.KeyMsg{Type: tea.KeyDown})
	if p.selected != 1 {
		t.Errorf("down at bottom: selected = %d, want 1", p.selected)
	}
}

func TestContactsPanel_UpdateNotFocused(t *testing.T) {
	p := newContactsPanel(20, 10)
	p.focused = false
	p.contacts = []contactEntry{
		{FriendID: 1, Name: "Alice"},
	}
	p.selected = 0

	// Key events should be ignored when not focused.
	selected, _ := p.update(tea.KeyMsg{Type: tea.KeyDown})
	if selected {
		t.Error("update should return false when not focused")
	}
}

func TestContactsPanel_UpdateEnter(t *testing.T) {
	p := newContactsPanel(20, 10)
	p.focused = true
	p.contacts = []contactEntry{
		{FriendID: 5, Name: "Eve"},
	}
	p.selected = 0

	selected, friendID := p.update(tea.KeyMsg{Type: tea.KeyEnter})
	if !selected {
		t.Error("update with Enter should return true")
	}
	if friendID != 5 {
		t.Errorf("update with Enter: friendID = %d, want 5", friendID)
	}
}

func TestContactsPanel_UpdateEnterEmpty(t *testing.T) {
	p := newContactsPanel(20, 10)
	p.focused = true
	// No contacts.

	selected, _ := p.update(tea.KeyMsg{Type: tea.KeyEnter})
	if selected {
		t.Error("update with Enter on empty contacts should return false")
	}
}

func TestContactsPanel_ScrollingView(t *testing.T) {
	p := newContactsPanel(20, 10)
	// Create more contacts than can fit in view.
	p.contacts = make([]contactEntry, 20)
	for i := 0; i < 20; i++ {
		p.contacts[i] = contactEntry{
			FriendID: uint32(i),
			Name:     string(rune('A' + i)),
		}
	}
	p.selected = 15 // Force scrolling.

	view := p.view()
	if view == "" {
		t.Error("contactsPanel.view() with scrolling returned empty")
	}
}

func TestContactsPanel_MouseClick(t *testing.T) {
	p := newContactsPanel(20, 10)
	p.focused = true
	p.contacts = []contactEntry{
		{FriendID: 1, Name: "Alice"},
		{FriendID: 2, Name: "Bob"},
		{FriendID: 3, Name: "Charlie"},
	}
	p.selected = 0

	// Simulate mouse click on second contact (row 4, accounting for border + title)
	selected, friendID := p.update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		Y:      4, // Border(1) + Title(2) + row offset = 4 for second contact
	})

	if !selected {
		t.Error("mouse click should select contact")
	}
	if friendID != 2 {
		t.Errorf("mouse click selected friendID = %d, want 2", friendID)
	}
	if p.selected != 1 {
		t.Errorf("selected index = %d, want 1", p.selected)
	}
}

func TestContactsPanel_MouseClickOutOfBounds(t *testing.T) {
	p := newContactsPanel(20, 10)
	p.focused = true
	p.contacts = []contactEntry{
		{FriendID: 1, Name: "Alice"},
	}
	p.selected = 0

	// Simulate mouse click outside contact area
	selected, _ := p.update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		Y:      100, // Way outside
	})

	if selected {
		t.Error("mouse click outside contacts should not select")
	}
}

func TestContactsPanel_MouseNonLeftClick(t *testing.T) {
	p := newContactsPanel(20, 10)
	p.focused = true
	p.contacts = []contactEntry{
		{FriendID: 1, Name: "Alice"},
	}

	// Right click should not select
	selected, _ := p.update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonRight,
		Y:      3,
	})

	if selected {
		t.Error("right click should not select contact")
	}
}

func TestContactsPanel_FocusBlur(t *testing.T) {
	p := newContactsPanel(20, 10)

	p.focused = true
	if !p.focused {
		t.Error("focused should be true after setting")
	}

	p.focused = false
	if p.focused {
		t.Error("focused should be false after unsetting")
	}
}

func TestContactsPanel_UpdateNonKeyMsg(t *testing.T) {
	p := newContactsPanel(20, 10)
	p.focused = true
	p.contacts = []contactEntry{
		{FriendID: 1, Name: "Alice"},
	}

	// Window size message should be ignored
	selected, _ := p.update(tea.WindowSizeMsg{Width: 100, Height: 50})
	if selected {
		t.Error("non-key message should not select")
	}
}

func TestSortContacts_AlreadySorted(t *testing.T) {
	contacts := []contactEntry{
		{FriendID: 1, Name: "Alice"},
		{FriendID: 2, Name: "Bob"},
		{FriendID: 3, Name: "Charlie"},
	}

	sortContacts(contacts)

	expected := []string{"Alice", "Bob", "Charlie"}
	for i, name := range expected {
		if contacts[i].Name != name {
			t.Errorf("sortContacts: index %d = %q, want %q", i, contacts[i].Name, name)
		}
	}
}

func TestSortContacts_ReverseSorted(t *testing.T) {
	contacts := []contactEntry{
		{FriendID: 3, Name: "Zoe"},
		{FriendID: 2, Name: "Mike"},
		{FriendID: 1, Name: "Alice"},
	}

	sortContacts(contacts)

	expected := []string{"Alice", "Mike", "Zoe"}
	for i, name := range expected {
		if contacts[i].Name != name {
			t.Errorf("sortContacts: index %d = %q, want %q", i, contacts[i].Name, name)
		}
	}
}

func TestContactEntry_AllFields(t *testing.T) {
	entry := contactEntry{
		FriendID:         100,
		Name:             "TestName",
		StatusMessage:    "Testing status",
		ConnectionStatus: toxcore.ConnectionTCP,
		FriendStatus:     toxcore.FriendStatusBusy,
		UnreadCount:      10,
	}

	if entry.FriendID != 100 {
		t.Error("FriendID mismatch")
	}
	if entry.Name != "TestName" {
		t.Error("Name mismatch")
	}
	if entry.StatusMessage != "Testing status" {
		t.Error("StatusMessage mismatch")
	}
	if entry.ConnectionStatus != toxcore.ConnectionTCP {
		t.Error("ConnectionStatus mismatch")
	}
	if entry.FriendStatus != toxcore.FriendStatusBusy {
		t.Error("FriendStatus mismatch")
	}
	if entry.UnreadCount != 10 {
		t.Error("UnreadCount mismatch")
	}
}

func TestStatusIndicator_AllStatuses(t *testing.T) {
	tests := []struct {
		conn   toxcore.ConnectionStatus
		friend toxcore.FriendStatus
		desc   string
	}{
		{toxcore.ConnectionNone, toxcore.FriendStatusNone, "offline"},
		{toxcore.ConnectionUDP, toxcore.FriendStatusNone, "online UDP"},
		{toxcore.ConnectionTCP, toxcore.FriendStatusNone, "online TCP"},
		{toxcore.ConnectionUDP, toxcore.FriendStatusAway, "away"},
		{toxcore.ConnectionTCP, toxcore.FriendStatusBusy, "busy"},
	}

	for _, tt := range tests {
		got := statusIndicator(tt.conn, tt.friend)
		if got == "" {
			t.Errorf("statusIndicator for %s returned empty", tt.desc)
		}
	}
}

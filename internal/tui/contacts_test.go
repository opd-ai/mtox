package tui

import (
	"testing"

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

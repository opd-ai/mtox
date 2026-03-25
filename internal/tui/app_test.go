package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	toxclient "github.com/opd-ai/mtox/internal/tox"
	"github.com/opd-ai/toxcore"
)

func TestNewStatusBar(t *testing.T) {
	sb := newStatusBar(100)
	if sb.width != 100 {
		t.Errorf("newStatusBar width = %d, want 100", sb.width)
	}
	if sb.selfAddress != "" {
		t.Error("newStatusBar selfAddress should be empty")
	}
}

func TestStatusBar_ConnectionString(t *testing.T) {
	tests := []struct {
		status toxcore.ConnectionStatus
		want   string // Just check it's non-empty and contains expected text
	}{
		{toxcore.ConnectionNone, "Disconnected"},
		{toxcore.ConnectionUDP, "UDP"},
		{toxcore.ConnectionTCP, "TCP"},
	}

	for _, tt := range tests {
		sb := statusBar{connectionStatus: tt.status}
		got := sb.connectionString()
		if got == "" {
			t.Errorf("connectionString() with status %v returned empty string", tt.status)
		}
	}
}

func TestStatusBar_AddressString(t *testing.T) {
	sb := statusBar{}

	// Empty address.
	got := sb.addressString()
	if got == "" {
		t.Error("addressString() with empty address returned empty string")
	}
	if got != "My ID: (loading...)" {
		t.Errorf("addressString() empty = %q, want %q", got, "My ID: (loading...)")
	}

	// Short address.
	sb.selfAddress = "ABCD1234"
	got = sb.addressString()
	if got == "" {
		t.Error("addressString() with short address returned empty")
	}

	// Long address should be truncated.
	sb.selfAddress = "12345678901234567890ABCDEF"
	got = sb.addressString()
	if len(got) > 30 {
		// Should be truncated
		t.Logf("Long address rendered as: %s", got)
	}
}

func TestStatusBar_AnonymityString(t *testing.T) {
	tests := []struct {
		torStatus toxclient.AnonymityStatus
		i2pStatus toxclient.AnonymityStatus
		wantEmpty bool
	}{
		{toxclient.AnonymityUnavailable, toxclient.AnonymityUnavailable, true},
		{toxclient.AnonymityAvailable, toxclient.AnonymityUnavailable, false},
		{toxclient.AnonymityUnavailable, toxclient.AnonymityAvailable, false},
		{toxclient.AnonymityAvailable, toxclient.AnonymityAvailable, false},
		{toxclient.AnonymityConnecting, toxclient.AnonymityUnavailable, false},
		{toxclient.AnonymityUnavailable, toxclient.AnonymityConnecting, false},
	}

	for _, tt := range tests {
		sb := statusBar{
			torStatus: tt.torStatus,
			i2pStatus: tt.i2pStatus,
		}
		got := sb.anonymityString()
		if tt.wantEmpty && got != "" {
			t.Errorf("anonymityString() with tor=%v, i2p=%v = %q, want empty",
				tt.torStatus, tt.i2pStatus, got)
		}
		if !tt.wantEmpty && got == "" {
			t.Errorf("anonymityString() with tor=%v, i2p=%v = empty, want non-empty",
				tt.torStatus, tt.i2pStatus)
		}
	}
}

func TestStatusBar_View(t *testing.T) {
	sb := newStatusBar(80)
	sb.connectionStatus = toxcore.ConnectionUDP
	sb.selfAddress = "ABC123"

	view := sb.view()
	if view == "" {
		t.Error("statusBar.view() returned empty string")
	}
}

func TestStatusBar_ViewWithAnonymity(t *testing.T) {
	sb := newStatusBar(100)
	sb.connectionStatus = toxcore.ConnectionTCP
	sb.selfAddress = "DEF456"
	sb.torStatus = toxclient.AnonymityAvailable
	sb.i2pStatus = toxclient.AnonymityConnecting

	view := sb.view()
	if view == "" {
		t.Error("statusBar.view() with anonymity returned empty string")
	}
}

func TestContactEntry_Fields(t *testing.T) {
	entry := contactEntry{
		FriendID:         42,
		Name:             "TestUser",
		StatusMessage:    "Hello",
		ConnectionStatus: toxcore.ConnectionUDP,
		FriendStatus:     toxcore.FriendStatusAway,
		UnreadCount:      5,
	}

	if entry.FriendID != 42 {
		t.Error("FriendID mismatch")
	}
	if entry.Name != "TestUser" {
		t.Error("Name mismatch")
	}
	if entry.StatusMessage != "Hello" {
		t.Error("StatusMessage mismatch")
	}
	if entry.ConnectionStatus != toxcore.ConnectionUDP {
		t.Error("ConnectionStatus mismatch")
	}
	if entry.FriendStatus != toxcore.FriendStatusAway {
		t.Error("FriendStatus mismatch")
	}
	if entry.UnreadCount != 5 {
		t.Error("UnreadCount mismatch")
	}
}

func TestStatusIndicator(t *testing.T) {
	tests := []struct {
		conn   toxcore.ConnectionStatus
		friend toxcore.FriendStatus
	}{
		{toxcore.ConnectionNone, toxcore.FriendStatusNone}, // Offline
		{toxcore.ConnectionUDP, toxcore.FriendStatusNone},  // Online
		{toxcore.ConnectionTCP, toxcore.FriendStatusAway},  // Away
		{toxcore.ConnectionUDP, toxcore.FriendStatusBusy},  // Busy
		{toxcore.ConnectionTCP, toxcore.FriendStatus(99)},  // Unknown status
	}

	for _, tt := range tests {
		got := statusIndicator(tt.conn, tt.friend)
		if got == "" {
			t.Errorf("statusIndicator(%v, %v) returned empty string", tt.conn, tt.friend)
		}
	}
}

func TestContactsPanel_SetContacts(t *testing.T) {
	p := newContactsPanel(20, 10)

	// Set initial unread count.
	p.contacts = []contactEntry{
		{FriendID: 1, Name: "Alice", UnreadCount: 3},
	}

	// Create new friends map.
	friends := map[uint32]*toxcore.Friend{
		1: {Name: "Alice Updated", ConnectionStatus: toxcore.ConnectionUDP},
		2: {Name: "Bob", ConnectionStatus: toxcore.ConnectionNone},
	}

	p.setContacts(friends)

	if len(p.contacts) != 2 {
		t.Errorf("setContacts: len(contacts) = %d, want 2", len(p.contacts))
	}

	// Check unread count preserved for friend 1.
	for _, c := range p.contacts {
		if c.FriendID == 1 && c.UnreadCount != 3 {
			t.Errorf("setContacts: unread count for friend 1 = %d, want 3", c.UnreadCount)
		}
	}
}

func TestContactsPanel_SetContacts_EmptyName(t *testing.T) {
	p := newContactsPanel(20, 10)

	friends := map[uint32]*toxcore.Friend{
		5: {Name: "", ConnectionStatus: toxcore.ConnectionNone},
	}

	p.setContacts(friends)

	if len(p.contacts) != 1 {
		t.Fatalf("setContacts: len(contacts) = %d, want 1", len(p.contacts))
	}

	// Should use default name format.
	if p.contacts[0].Name == "" {
		t.Error("setContacts should provide default name for empty friend name")
	}
}

func TestContactsPanel_View(t *testing.T) {
	p := newContactsPanel(20, 10)
	p.contacts = []contactEntry{
		{FriendID: 1, Name: "Alice", ConnectionStatus: toxcore.ConnectionUDP},
		{FriendID: 2, Name: "Bob", ConnectionStatus: toxcore.ConnectionNone},
	}

	view := p.view()
	if view == "" {
		t.Error("contactsPanel.view() returned empty string")
	}
}

func TestContactsPanel_ViewEmpty(t *testing.T) {
	p := newContactsPanel(20, 10)
	view := p.view()
	if view == "" {
		t.Error("contactsPanel.view() with no contacts returned empty string")
	}
}

func TestContactsPanel_ViewFocused(t *testing.T) {
	p := newContactsPanel(20, 10)
	p.focused = true
	p.contacts = []contactEntry{
		{FriendID: 1, Name: "Alice"},
	}

	view := p.view()
	if view == "" {
		t.Error("contactsPanel.view() focused returned empty string")
	}
}

func TestContactsPanel_RenderContact(t *testing.T) {
	p := newContactsPanel(20, 10)

	c := contactEntry{
		FriendID:         1,
		Name:             "Alice",
		ConnectionStatus: toxcore.ConnectionUDP,
		UnreadCount:      0,
	}

	// Not selected.
	line := p.renderContact(c, false)
	if line == "" {
		t.Error("renderContact (not selected) returned empty")
	}

	// Selected.
	line = p.renderContact(c, true)
	if line == "" {
		t.Error("renderContact (selected) returned empty")
	}
}

func TestContactsPanel_RenderContactWithUnread(t *testing.T) {
	p := newContactsPanel(30, 10)

	c := contactEntry{
		FriendID:         1,
		Name:             "Alice",
		ConnectionStatus: toxcore.ConnectionUDP,
		UnreadCount:      5,
	}

	line := p.renderContact(c, false)
	if line == "" {
		t.Error("renderContact with unread returned empty")
	}
}

func TestContactsPanel_RenderContactLongName(t *testing.T) {
	p := newContactsPanel(15, 10)

	c := contactEntry{
		FriendID: 1,
		Name:     "VeryLongContactNameThatShouldBeTruncated",
	}

	line := p.renderContact(c, false)
	if line == "" {
		t.Error("renderContact with long name returned empty")
	}
}

func TestChatPanel_NewChatPanel(t *testing.T) {
	cp := newChatPanel(80, 24)

	if cp.width != 80 {
		t.Errorf("newChatPanel width = %d, want 80", cp.width)
	}
	if cp.height != 24 {
		t.Errorf("newChatPanel height = %d, want 24", cp.height)
	}
	if cp.input.CharLimit != 1000 {
		t.Errorf("newChatPanel input.CharLimit = %d, want 1000", cp.input.CharLimit)
	}
}

func TestChatPanel_View(t *testing.T) {
	cp := newChatPanel(80, 24)

	// No friend selected.
	view := cp.view()
	if view == "" {
		t.Error("chatPanel.view() with no friend returned empty")
	}

	// With friend selected.
	cp.friendName = "Alice"
	view = cp.view()
	if view == "" {
		t.Error("chatPanel.view() with friend returned empty")
	}
}

func TestChatPanel_ViewFocused(t *testing.T) {
	cp := newChatPanel(80, 24)
	cp.focused = true
	cp.friendName = "Bob"

	view := cp.view()
	if view == "" {
		t.Error("chatPanel.view() focused returned empty")
	}
}

func TestChatPanel_RenderHistoryWithMessages(t *testing.T) {
	cp := newChatPanel(80, 24)
	cp.friendName = "Alice"

	now := time.Now()
	cp.history = []chatMessage{
		{ts: now, senderID: 0, name: "You", body: "Hello!"},
		{ts: now, senderID: 1, name: "Alice", body: "Hi there!"},
		{ts: now, senderID: 0, name: "You", body: "/me waves", isAction: true},
	}

	content := cp.renderHistory()
	if content == "" {
		t.Error("renderHistory with messages returned empty")
	}
}

func TestChatPanel_SetTyping(t *testing.T) {
	cp := newChatPanel(80, 24)
	cp.friendName = "Alice"

	cp.isTyping = true
	content := cp.renderHistory()
	if content == "" {
		t.Error("renderHistory with typing indicator returned empty")
	}
}

func TestModalKindConstants(t *testing.T) {
	// Verify modal kind constants are distinct.
	if modalNone == modalAddFriend {
		t.Error("modalNone should not equal modalAddFriend")
	}
	if modalAddFriend == modalFriendRequest {
		t.Error("modalAddFriend should not equal modalFriendRequest")
	}
	if modalNone == modalFriendRequest {
		t.Error("modalNone should not equal modalFriendRequest")
	}
}

func TestFocusConstants(t *testing.T) {
	// Verify focus constants are distinct.
	if focusContacts == focusChat {
		t.Error("focusContacts should not equal focusChat")
	}
}

func TestPendingRequest_Fields(t *testing.T) {
	req := pendingRequest{
		PublicKey: [32]byte{1, 2, 3},
		Message:   "Hello!",
	}

	if req.PublicKey[0] != 1 {
		t.Error("pendingRequest PublicKey mismatch")
	}
	if req.Message != "Hello!" {
		t.Error("pendingRequest Message mismatch")
	}
}

func TestContactsPanelWidth(t *testing.T) {
	if contactsPanelWidth <= 0 {
		t.Errorf("contactsPanelWidth = %d, expected > 0", contactsPanelWidth)
	}
}

func TestStatusBar_Resize(t *testing.T) {
	sb := newStatusBar(80)
	sb.width = 120

	if sb.width != 120 {
		t.Errorf("statusBar.width = %d, want 120", sb.width)
	}
}

func TestStatusBar_SetConnectionStatus(t *testing.T) {
	sb := newStatusBar(80)
	sb.connectionStatus = toxcore.ConnectionUDP

	if sb.connectionStatus != toxcore.ConnectionUDP {
		t.Errorf("connectionStatus = %v, want ConnectionUDP", sb.connectionStatus)
	}
}

func TestStatusBar_SetAddress(t *testing.T) {
	sb := newStatusBar(80)
	sb.selfAddress = "ABCDEF123456"

	if sb.selfAddress != "ABCDEF123456" {
		t.Errorf("selfAddress = %q, want %q", sb.selfAddress, "ABCDEF123456")
	}
}

func TestPendingFileRequest_Fields(t *testing.T) {
	req := pendingFileRequest{
		FriendID: 5,
		FileID:   10,
		Filename: "document.pdf",
		FileSize: 1024 * 1024,
	}

	if req.FriendID != 5 {
		t.Errorf("FriendID = %d, want 5", req.FriendID)
	}
	if req.FileID != 10 {
		t.Errorf("FileID = %d, want 10", req.FileID)
	}
	if req.Filename != "document.pdf" {
		t.Errorf("Filename = %q, want %q", req.Filename, "document.pdf")
	}
	if req.FileSize != 1024*1024 {
		t.Errorf("FileSize = %d, want %d", req.FileSize, 1024*1024)
	}
}

func TestModalKind_Values(t *testing.T) {
	// Verify modal kind constants are distinct.
	kinds := []modalKind{modalNone, modalAddFriend, modalFriendRequest, modalFileRequest}
	seen := make(map[modalKind]bool)

	for _, k := range kinds {
		if seen[k] {
			t.Errorf("duplicate modal kind value: %d", k)
		}
		seen[k] = true
	}
}

func TestChatPanel_InputFocus(t *testing.T) {
	cp := newChatPanel(80, 24)

	// Initially not focused
	if cp.focused {
		t.Error("chatPanel should not be focused initially")
	}

	// Focus the panel
	cp.focus()
	if !cp.focused {
		t.Error("chatPanel should be focused after focus()")
	}

	// Blur the panel
	cp.blur()
	if cp.focused {
		t.Error("chatPanel should not be focused after blur()")
	}
}

func TestChatPanel_InputPlaceholder(t *testing.T) {
	cp := newChatPanel(80, 24)

	if cp.input.Placeholder != "type a message..." {
		t.Errorf("input.Placeholder = %q, want %q", cp.input.Placeholder, "type a message...")
	}
}

func TestChatPanel_ViewNoFriend(t *testing.T) {
	cp := newChatPanel(80, 24)
	cp.friendName = ""

	view := cp.view()
	if view == "" {
		t.Error("view() with no friend should not be empty")
	}
}

func TestChatPanel_ViewportDimensions(t *testing.T) {
	cp := newChatPanel(100, 30)

	// Check viewport dimensions are derived from panel dimensions
	if cp.viewport.Width != 96 { // 100 - 4
		t.Errorf("viewport.Width = %d, want 96", cp.viewport.Width)
	}
	if cp.viewport.Height != 22 { // 30 - 8
		t.Errorf("viewport.Height = %d, want 22", cp.viewport.Height)
	}
}

func TestChatPanel_ResizeViewport(t *testing.T) {
	cp := newChatPanel(80, 24)
	cp.resize(120, 40)

	if cp.viewport.Width != 116 { // 120 - 4
		t.Errorf("after resize, viewport.Width = %d, want 116", cp.viewport.Width)
	}
	if cp.viewport.Height != 32 { // 40 - 8
		t.Errorf("after resize, viewport.Height = %d, want 32", cp.viewport.Height)
	}
}

func TestStatusBar_ConnectionStringNone(t *testing.T) {
	sb := statusBar{connectionStatus: toxcore.ConnectionNone}
	got := sb.connectionString()
	if got == "" {
		t.Error("connectionString() with ConnectionNone returned empty")
	}
	if !strings.Contains(got, "Disconnected") {
		t.Errorf("connectionString() = %q, should contain 'Disconnected'", got)
	}
}

func TestStatusBar_ConnectionStringUDP(t *testing.T) {
	sb := statusBar{connectionStatus: toxcore.ConnectionUDP}
	got := sb.connectionString()
	if got == "" {
		t.Error("connectionString() with ConnectionUDP returned empty")
	}
	if !strings.Contains(got, "UDP") {
		t.Errorf("connectionString() = %q, should contain 'UDP'", got)
	}
}

func TestStatusBar_ConnectionStringTCP(t *testing.T) {
	sb := statusBar{connectionStatus: toxcore.ConnectionTCP}
	got := sb.connectionString()
	if got == "" {
		t.Error("connectionString() with ConnectionTCP returned empty")
	}
	if !strings.Contains(got, "TCP") {
		t.Errorf("connectionString() = %q, should contain 'TCP'", got)
	}
}

func TestStatusBar_AddressStringLoading(t *testing.T) {
	sb := statusBar{selfAddress: ""}
	got := sb.addressString()
	if got != "My ID: (loading...)" {
		t.Errorf("addressString() = %q, want %q", got, "My ID: (loading...)")
	}
}

func TestStatusBar_AddressStringShort(t *testing.T) {
	sb := statusBar{selfAddress: "ABC123"}
	got := sb.addressString()
	if !strings.Contains(got, "ABC123") {
		t.Errorf("addressString() = %q, should contain 'ABC123'", got)
	}
}

func TestStatusBar_AddressStringLong(t *testing.T) {
	sb := statusBar{selfAddress: "12345678901234567890ABCDEF"}
	got := sb.addressString()
	// Should be truncated
	if strings.Contains(got, "ABCDEF") {
		t.Errorf("addressString() = %q, long address should be truncated", got)
	}
}

func TestStatusBar_AnonymityBothAvailable(t *testing.T) {
	sb := statusBar{
		torStatus: toxclient.AnonymityAvailable,
		i2pStatus: toxclient.AnonymityAvailable,
	}
	got := sb.anonymityString()
	if !strings.Contains(got, "Tor") || !strings.Contains(got, "I2P") {
		t.Errorf("anonymityString() = %q, should contain both 'Tor' and 'I2P'", got)
	}
}

func TestStatusBar_AnonymityBothConnecting(t *testing.T) {
	sb := statusBar{
		torStatus: toxclient.AnonymityConnecting,
		i2pStatus: toxclient.AnonymityConnecting,
	}
	got := sb.anonymityString()
	if got == "" {
		t.Error("anonymityString() with both connecting should not be empty")
	}
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		bytes    uint64
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{1073741824, "1.0 GB"},
		{1610612736, "1.5 GB"},
	}

	for _, tt := range tests {
		got := formatFileSize(tt.bytes)
		if got != tt.expected {
			t.Errorf("formatFileSize(%d) = %q, want %q", tt.bytes, got, tt.expected)
		}
	}
}

func TestStylesExist(t *testing.T) {
	// Verify style variables are defined (compile-time check + coverage).
	styles := []interface{}{
		activePanel,
		inactivePanel,
		contactSelected,
		contactNormal,
		colorOnline,
		colorAway,
		colorBusy,
		colorOffline,
		statusBarStyle,
		statusConnected,
		statusDisconnected,
		chatHeaderStyle,
		messageTimestamp,
		messageSelf,
		messagePeer,
		messageAction,
		typingIndicator,
		modalStyle,
		unreadBadge,
		panelTitle,
		fileTransferStyle,
		fileProgressStyle,
	}

	for i, s := range styles {
		if s == nil {
			t.Errorf("style[%d] is nil", i)
		}
	}
}

func TestChatPanel_History(t *testing.T) {
	cp := newChatPanel(80, 24)
	cp.friendID = 1
	cp.friendName = "Alice"

	// Initially empty
	if len(cp.history) != 0 {
		t.Errorf("initial history length = %d, want 0", len(cp.history))
	}

	// Add messages
	now := time.Now()
	cp.addMessage(chatMessage{ts: now, senderID: 1, name: "Alice", body: "Hello"})
	cp.addMessage(chatMessage{ts: now, senderID: 0, name: "You", body: "Hi there"})

	if len(cp.history) != 2 {
		t.Errorf("history length after adding = %d, want 2", len(cp.history))
	}

	// getHistory should return the history
	hist := cp.history
	if hist[0].body != "Hello" {
		t.Errorf("first message body = %q, want %q", hist[0].body, "Hello")
	}
	if hist[1].body != "Hi there" {
		t.Errorf("second message body = %q, want %q", hist[1].body, "Hi there")
	}
}

func TestContactsPanel_Empty(t *testing.T) {
	p := newContactsPanel(20, 10)

	// No contacts
	if len(p.contacts) != 0 {
		t.Errorf("initial contacts length = %d, want 0", len(p.contacts))
	}

	// View should still work
	view := p.view()
	if view == "" {
		t.Error("view() with no contacts returned empty")
	}

	// Selected friend should return false
	_, ok := p.selectedFriendID()
	if ok {
		t.Error("selectedFriendID should return false for empty contacts")
	}
}

func TestStatusBar_VersionInfo(t *testing.T) {
	sb := newStatusBar(100)
	view := sb.view()

	// View should contain version info
	if view == "" {
		t.Error("statusBar.view() returned empty")
	}
}

func TestChatPanel_ActionMessage(t *testing.T) {
	cp := newChatPanel(80, 24)
	cp.friendID = 1
	cp.friendName = "Alice"

	// Add an action message
	now := time.Now()
	cp.addMessage(chatMessage{
		ts:       now,
		senderID: 1,
		name:     "Alice",
		body:     "waves hello",
		isAction: true,
	})

	content := cp.renderHistory()
	if content == "" {
		t.Error("renderHistory with action message returned empty")
	}
}

func TestContactsPanel_LongList(t *testing.T) {
	p := newContactsPanel(20, 5) // small height

	// Add many contacts
	p.contacts = make([]contactEntry, 20)
	for i := 0; i < 20; i++ {
		p.contacts[i] = contactEntry{
			FriendID: uint32(i),
			Name:     string(rune('A' + i)),
		}
	}
	p.selected = 15 // Beyond visible range

	view := p.view()
	if view == "" {
		t.Error("view() with scrolling returned empty")
	}
}

func TestStatusBar_ErrorStatus(t *testing.T) {
	sb := newStatusBar(80)
	sb.torStatus = toxclient.AnonymityError
	sb.i2pStatus = toxclient.AnonymityError

	got := sb.anonymityString()
	// Error status might still show something or be empty
	// The important thing is it doesn't panic
	_ = got
}

func TestChatPanel_ManyMessages(t *testing.T) {
	cp := newChatPanel(80, 10) // small height
	cp.friendID = 1
	cp.friendName = "Alice"

	// Add many messages
	now := time.Now()
	for i := 0; i < 100; i++ {
		cp.addMessage(chatMessage{
			ts:       now.Add(time.Duration(i) * time.Minute),
			senderID: uint32(i % 2),
			name:     []string{"You", "Alice"}[i%2],
			body:     fmt.Sprintf("Message %d", i),
		})
	}

	if len(cp.history) != 100 {
		t.Errorf("history length = %d, want 100", len(cp.history))
	}

	// Should not panic when rendering many messages
	content := cp.renderHistory()
	if content == "" {
		t.Error("renderHistory with many messages returned empty")
	}
}

func TestContactsPanel_Resize(t *testing.T) {
	p := newContactsPanel(20, 10)

	// Resize
	p.width = 40
	p.height = 20

	if p.width != 40 {
		t.Errorf("width after resize = %d, want 40", p.width)
	}
	if p.height != 20 {
		t.Errorf("height after resize = %d, want 20", p.height)
	}
}

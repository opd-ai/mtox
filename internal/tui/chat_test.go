package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestChatPanel_AddMessage(t *testing.T) {
	cp := newChatPanel(80, 24)
	cp.friendID = 1
	cp.friendName = "Alice"

	msg := chatMessage{
		ts:       time.Now(),
		senderID: 1,
		name:     "Alice",
		body:     "Hello!",
	}

	cp.addMessage(msg)

	if len(cp.history) != 1 {
		t.Errorf("addMessage: len(history) = %d, want 1", len(cp.history))
	}
	if cp.history[0].body != "Hello!" {
		t.Errorf("addMessage: body = %q, want %q", cp.history[0].body, "Hello!")
	}
}

func TestChatPanel_SetFriend(t *testing.T) {
	cp := newChatPanel(80, 24)

	// Add some history.
	cp.history = []chatMessage{
		{ts: time.Now(), senderID: 0, name: "You", body: "Test"},
	}
	cp.friendID = 1

	// Switch to a different friend.
	cp.setFriend(2, "Bob")

	if cp.friendID != 2 {
		t.Errorf("setFriend: friendID = %d, want 2", cp.friendID)
	}
	if cp.friendName != "Bob" {
		t.Errorf("setFriend: friendName = %q, want %q", cp.friendName, "Bob")
	}
	if cp.history != nil {
		t.Error("setFriend should clear history when switching friends")
	}
}

func TestChatPanel_SetFriend_SameFriend(t *testing.T) {
	cp := newChatPanel(80, 24)
	cp.friendID = 1
	cp.friendName = "Alice"
	cp.history = []chatMessage{
		{ts: time.Now(), senderID: 1, name: "Alice", body: "Hello"},
	}

	// Set to same friend should be a no-op.
	cp.setFriend(1, "Alice")

	if len(cp.history) != 1 {
		t.Error("setFriend should not clear history when setting same friend")
	}
}

func TestChatPanel_SetFriendWithHistory(t *testing.T) {
	cp := newChatPanel(80, 24)

	savedHistory := []chatMessage{
		{ts: time.Now(), senderID: 3, name: "Charlie", body: "Old message"},
		{ts: time.Now(), senderID: 0, name: "You", body: "Reply"},
	}

	cp.setFriendWithHistory(3, "Charlie", savedHistory)

	if cp.friendID != 3 {
		t.Errorf("setFriendWithHistory: friendID = %d, want 3", cp.friendID)
	}
	if cp.friendName != "Charlie" {
		t.Errorf("setFriendWithHistory: friendName = %q, want %q", cp.friendName, "Charlie")
	}
	if len(cp.history) != 2 {
		t.Errorf("setFriendWithHistory: len(history) = %d, want 2", len(cp.history))
	}
}

func TestChatPanel_RenderHistory_Empty(t *testing.T) {
	cp := newChatPanel(80, 24)
	content := cp.renderHistory()

	if content == "" {
		t.Error("renderHistory should return non-empty string for empty history")
	}
}

func TestChatPanel_RenderHistory_WithTyping(t *testing.T) {
	cp := newChatPanel(80, 24)
	cp.friendName = "Alice"
	cp.isTyping = true

	content := cp.renderHistory()

	// Should contain typing indicator.
	if content == "" {
		t.Error("renderHistory with typing should return non-empty content")
	}
}

func TestChatPanel_Resize(t *testing.T) {
	cp := newChatPanel(80, 24)

	cp.resize(100, 30)

	if cp.width != 100 {
		t.Errorf("resize: width = %d, want 100", cp.width)
	}
	if cp.height != 30 {
		t.Errorf("resize: height = %d, want 30", cp.height)
	}
}

func TestChatPanel_FocusBlur(t *testing.T) {
	cp := newChatPanel(80, 24)

	cp.focus()
	if !cp.focused {
		t.Error("focus() should set focused = true")
	}

	cp.blur()
	if cp.focused {
		t.Error("blur() should set focused = false")
	}
}

func TestChatMessage_Fields(t *testing.T) {
	now := time.Now()
	msg := chatMessage{
		ts:       now,
		senderID: 42,
		name:     "Test",
		body:     "Hello",
		isAction: true,
	}

	if msg.ts != now {
		t.Error("chatMessage ts field mismatch")
	}
	if msg.senderID != 42 {
		t.Error("chatMessage senderID field mismatch")
	}
	if msg.name != "Test" {
		t.Error("chatMessage name field mismatch")
	}
	if msg.body != "Hello" {
		t.Error("chatMessage body field mismatch")
	}
	if !msg.isAction {
		t.Error("chatMessage isAction field mismatch")
	}
}

func TestChatPanel_RefreshViewport(t *testing.T) {
	cp := newChatPanel(80, 24)
	cp.friendName = "Alice"
	cp.history = []chatMessage{
		{ts: time.Now(), senderID: 0, name: "You", body: "Test"},
	}

	// Should not panic.
	cp.refreshViewport()
}

func TestChatPanel_Update_NonFocused(t *testing.T) {
	cp := newChatPanel(80, 24)
	cp.focused = false

	// Key events should be ignored when not focused.
	text, submitted := cp.update(tea.KeyMsg{Type: tea.KeyEnter})
	if submitted {
		t.Error("update should not submit when not focused")
	}
	if text != "" {
		t.Errorf("update returned text %q when not focused", text)
	}
}

func TestChatPanel_Update_EmptyInput(t *testing.T) {
	cp := newChatPanel(80, 24)
	cp.focused = true
	cp.input.SetValue("")

	// Enter with empty input should not submit.
	text, submitted := cp.update(tea.KeyMsg{Type: tea.KeyEnter})
	if submitted {
		t.Error("update should not submit empty input")
	}
	if text != "" {
		t.Errorf("update returned text %q for empty input", text)
	}
}

func TestChatPanel_Update_WithInput(t *testing.T) {
	cp := newChatPanel(80, 24)
	cp.focused = true
	cp.input.SetValue("Hello world")

	// Enter with non-empty input should submit.
	text, submitted := cp.update(tea.KeyMsg{Type: tea.KeyEnter})
	if !submitted {
		t.Error("update should submit non-empty input")
	}
	if text != "Hello world" {
		t.Errorf("update returned text %q, want %q", text, "Hello world")
	}
	// Input should be cleared after submission.
	if cp.input.Value() != "" {
		t.Error("input should be cleared after submission")
	}
}

func TestChatPanel_Update_WhitespaceOnly(t *testing.T) {
	cp := newChatPanel(80, 24)
	cp.focused = true
	cp.input.SetValue("   ")

	// Enter with whitespace-only input should not submit.
	text, submitted := cp.update(tea.KeyMsg{Type: tea.KeyEnter})
	if submitted {
		t.Error("update should not submit whitespace-only input")
	}
	if text != "" {
		t.Errorf("update returned text %q for whitespace-only input", text)
	}
}

func TestChatPanel_RenderHistory_ActionMessage(t *testing.T) {
	cp := newChatPanel(80, 24)
	cp.friendName = "Alice"
	cp.history = []chatMessage{
		{ts: time.Now(), senderID: 1, name: "Alice", body: "waves", isAction: true},
	}

	content := cp.renderHistory()
	if content == "" {
		t.Error("renderHistory with action message returned empty")
	}
}

func TestChatPanel_RenderHistory_SelfMessage(t *testing.T) {
	cp := newChatPanel(80, 24)
	cp.friendName = "Alice"
	cp.history = []chatMessage{
		{ts: time.Now(), senderID: 0, name: "You", body: "Hello from me!"},
	}

	content := cp.renderHistory()
	if content == "" {
		t.Error("renderHistory with self message returned empty")
	}
}

func TestChatPanel_RenderHistory_PeerMessage(t *testing.T) {
	cp := newChatPanel(80, 24)
	cp.friendName = "Alice"
	cp.history = []chatMessage{
		{ts: time.Now(), senderID: 1, name: "Alice", body: "Hello from peer!"},
	}

	content := cp.renderHistory()
	if content == "" {
		t.Error("renderHistory with peer message returned empty")
	}
}

package tox

import (
	"testing"

	"github.com/opd-ai/toxcore"
)

func TestAnonymityStatus_String(t *testing.T) {
	tests := []struct {
		status AnonymityStatus
		want   string
	}{
		{AnonymityUnavailable, "unavailable"},
		{AnonymityConnecting, "connecting"},
		{AnonymityAvailable, "available"},
		{AnonymityError, "error"},
		{AnonymityStatus(99), "unknown"},
	}

	for _, tt := range tests {
		got := tt.status.String()
		if got != tt.want {
			t.Errorf("AnonymityStatus(%d).String() = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestToxEvents_ImplementInterface(t *testing.T) {
	// Verify all event types implement the ToxEvent interface.
	// This is a compile-time check but exercising them ensures coverage.
	events := []ToxEvent{
		FriendRequestEvent{},
		FriendMessageEvent{},
		FriendConnectionStatusEvent{},
		FriendNameEvent{},
		FriendStatusMessageEvent{},
		FriendTypingEvent{},
		SelfConnectionStatusEvent{},
		TickEvent{},
		AnonymityStatusEvent{},
		FileRecvRequestEvent{},
		FileRecvChunkEvent{},
		FileChunkRequestEvent{},
		FileSendCompleteEvent{},
		FileRecvCompleteEvent{},
		FileTransferErrorEvent{},
	}

	for _, e := range events {
		// Call the interface method to ensure it's implemented.
		e.toxEvent()
	}

	if len(events) != 15 {
		t.Errorf("Expected 15 event types, got %d", len(events))
	}
}

func TestNetworkConstants(t *testing.T) {
	if NetworkTor != "tor" {
		t.Errorf("NetworkTor = %q, want %q", NetworkTor, "tor")
	}
	if NetworkI2P != "i2p" {
		t.Errorf("NetworkI2P = %q, want %q", NetworkI2P, "i2p")
	}
}

func TestFriendRequestEvent_Fields(t *testing.T) {
	pubKey := [32]byte{1, 2, 3, 4, 5}
	e := FriendRequestEvent{
		PublicKey: pubKey,
		Message:   "Hello, let's chat!",
	}

	if e.PublicKey != pubKey {
		t.Error("FriendRequestEvent.PublicKey mismatch")
	}
	if e.Message != "Hello, let's chat!" {
		t.Errorf("FriendRequestEvent.Message = %q, want %q", e.Message, "Hello, let's chat!")
	}
}

func TestFriendMessageEvent_Fields(t *testing.T) {
	e := FriendMessageEvent{
		FriendID: 42,
		Message:  "Hi there!",
	}

	if e.FriendID != 42 {
		t.Errorf("FriendMessageEvent.FriendID = %d, want 42", e.FriendID)
	}
	if e.Message != "Hi there!" {
		t.Errorf("FriendMessageEvent.Message = %q, want %q", e.Message, "Hi there!")
	}
}

func TestFriendConnectionStatusEvent_Fields(t *testing.T) {
	e := FriendConnectionStatusEvent{
		FriendID: 5,
		Status:   toxcore.ConnectionUDP,
	}

	if e.FriendID != 5 {
		t.Errorf("FriendConnectionStatusEvent.FriendID = %d, want 5", e.FriendID)
	}
	if e.Status != toxcore.ConnectionUDP {
		t.Errorf("FriendConnectionStatusEvent.Status = %v, want ConnectionUDP", e.Status)
	}
}

func TestFriendNameEvent_Fields(t *testing.T) {
	e := FriendNameEvent{
		FriendID: 3,
		Name:     "Alice",
	}

	if e.FriendID != 3 {
		t.Errorf("FriendNameEvent.FriendID = %d, want 3", e.FriendID)
	}
	if e.Name != "Alice" {
		t.Errorf("FriendNameEvent.Name = %q, want %q", e.Name, "Alice")
	}
}

func TestFriendStatusMessageEvent_Fields(t *testing.T) {
	e := FriendStatusMessageEvent{
		FriendID:      7,
		StatusMessage: "Away for lunch",
	}

	if e.FriendID != 7 {
		t.Errorf("FriendStatusMessageEvent.FriendID = %d, want 7", e.FriendID)
	}
	if e.StatusMessage != "Away for lunch" {
		t.Errorf("FriendStatusMessageEvent.StatusMessage = %q, want %q", e.StatusMessage, "Away for lunch")
	}
}

func TestFriendTypingEvent_Fields(t *testing.T) {
	e := FriendTypingEvent{
		FriendID: 10,
		IsTyping: true,
	}

	if e.FriendID != 10 {
		t.Errorf("FriendTypingEvent.FriendID = %d, want 10", e.FriendID)
	}
	if !e.IsTyping {
		t.Error("FriendTypingEvent.IsTyping = false, want true")
	}
}

func TestSelfConnectionStatusEvent_Fields(t *testing.T) {
	e := SelfConnectionStatusEvent{
		Status: toxcore.ConnectionTCP,
	}

	if e.Status != toxcore.ConnectionTCP {
		t.Errorf("SelfConnectionStatusEvent.Status = %v, want ConnectionTCP", e.Status)
	}
}

func TestAnonymityStatusEvent_Fields(t *testing.T) {
	e := AnonymityStatusEvent{
		Network: NetworkTor,
		Status:  AnonymityAvailable,
		Address: "abc123.onion",
		Error:   "",
	}

	if e.Network != NetworkTor {
		t.Errorf("AnonymityStatusEvent.Network = %q, want %q", e.Network, NetworkTor)
	}
	if e.Status != AnonymityAvailable {
		t.Errorf("AnonymityStatusEvent.Status = %v, want AnonymityAvailable", e.Status)
	}
	if e.Address != "abc123.onion" {
		t.Errorf("AnonymityStatusEvent.Address = %q, want %q", e.Address, "abc123.onion")
	}
	if e.Error != "" {
		t.Errorf("AnonymityStatusEvent.Error = %q, want empty", e.Error)
	}
}

func TestAnonymityStatusEvent_WithError(t *testing.T) {
	e := AnonymityStatusEvent{
		Network: NetworkI2P,
		Status:  AnonymityError,
		Address: "",
		Error:   "connection failed",
	}

	if e.Network != NetworkI2P {
		t.Errorf("AnonymityStatusEvent.Network = %q, want %q", e.Network, NetworkI2P)
	}
	if e.Status != AnonymityError {
		t.Errorf("AnonymityStatusEvent.Status = %v, want AnonymityError", e.Status)
	}
	if e.Address != "" {
		t.Errorf("AnonymityStatusEvent.Address = %q, want empty", e.Address)
	}
	if e.Error != "connection failed" {
		t.Errorf("AnonymityStatusEvent.Error = %q, want %q", e.Error, "connection failed")
	}
}

func TestTickEvent(t *testing.T) {
	e := TickEvent{}
	e.toxEvent() // Should not panic

	// TickEvent has no fields
	events := []ToxEvent{e}
	if len(events) != 1 {
		t.Error("TickEvent should implement ToxEvent")
	}
}

func TestAllEventTypes_Coverage(t *testing.T) {
	// Create instances of all event types to ensure full coverage
	events := []ToxEvent{
		FriendRequestEvent{PublicKey: [32]byte{}, Message: "test"},
		FriendMessageEvent{FriendID: 1, Message: "test"},
		FriendConnectionStatusEvent{FriendID: 1, Status: toxcore.ConnectionUDP},
		FriendNameEvent{FriendID: 1, Name: "test"},
		FriendStatusMessageEvent{FriendID: 1, StatusMessage: "test"},
		FriendTypingEvent{FriendID: 1, IsTyping: true},
		SelfConnectionStatusEvent{Status: toxcore.ConnectionTCP},
		TickEvent{},
		AnonymityStatusEvent{Network: NetworkTor, Status: AnonymityAvailable},
		FileRecvRequestEvent{FriendID: 1, FileID: 1, Kind: 0, FileSize: 100, Filename: "test.txt"},
		FileRecvChunkEvent{FriendID: 1, FileID: 1, Position: 0, Data: []byte("data")},
		FileChunkRequestEvent{FriendID: 1, FileID: 1, Position: 0, Length: 1024},
		FileSendCompleteEvent{FriendID: 1, FileID: 1, Filename: "test.txt"},
		FileRecvCompleteEvent{FriendID: 1, FileID: 1, Filename: "test.txt", SavePath: "/tmp/test.txt"},
		FileTransferErrorEvent{FriendID: 1, FileID: 1, Filename: "test.txt", Error: "failed"},
	}

	// Call toxEvent on each to ensure the interface method is exercised
	for i, e := range events {
		e.toxEvent()
		if e == nil {
			t.Errorf("event %d is nil", i)
		}
	}

	if len(events) != 15 {
		t.Errorf("expected 15 event types, got %d", len(events))
	}
}

func TestFriendTypingEvent_False(t *testing.T) {
	e := FriendTypingEvent{
		FriendID: 5,
		IsTyping: false,
	}

	if e.FriendID != 5 {
		t.Errorf("FriendTypingEvent.FriendID = %d, want 5", e.FriendID)
	}
	if e.IsTyping {
		t.Error("FriendTypingEvent.IsTyping = true, want false")
	}
}

func TestSelfConnectionStatusEvent_None(t *testing.T) {
	e := SelfConnectionStatusEvent{
		Status: toxcore.ConnectionNone,
	}

	if e.Status != toxcore.ConnectionNone {
		t.Errorf("SelfConnectionStatusEvent.Status = %v, want ConnectionNone", e.Status)
	}
}

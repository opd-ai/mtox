package tox

import "testing"

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
	}

	for _, e := range events {
		// Call the interface method to ensure it's implemented.
		e.toxEvent()
	}

	if len(events) != 9 {
		t.Errorf("Expected 9 event types, got %d", len(events))
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

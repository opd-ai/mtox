// Package tox provides types for bridging toxcore callbacks to bubbletea messages.
package tox

import "github.com/opd-ai/toxcore"

// ToxEvent is the interface implemented by all events emitted by the tox client.
type ToxEvent interface {
	toxEvent()
}

// FriendRequestEvent is fired when a friend request is received.
type FriendRequestEvent struct {
	PublicKey [32]byte
	Message   string
}

func (FriendRequestEvent) toxEvent() {}

// FriendMessageEvent is fired when a message is received from a friend.
type FriendMessageEvent struct {
	FriendID uint32
	Message  string
}

func (FriendMessageEvent) toxEvent() {}

// FriendConnectionStatusEvent is fired when a friend's connection status changes.
type FriendConnectionStatusEvent struct {
	FriendID uint32
	Status   toxcore.ConnectionStatus
}

func (FriendConnectionStatusEvent) toxEvent() {}

// FriendNameEvent is fired when a friend updates their name.
type FriendNameEvent struct {
	FriendID uint32
	Name     string
}

func (FriendNameEvent) toxEvent() {}

// FriendStatusMessageEvent is fired when a friend updates their status message.
type FriendStatusMessageEvent struct {
	FriendID      uint32
	StatusMessage string
}

func (FriendStatusMessageEvent) toxEvent() {}

// FriendTypingEvent is fired when a friend starts or stops typing.
type FriendTypingEvent struct {
	FriendID uint32
	IsTyping bool
}

func (FriendTypingEvent) toxEvent() {}

// SelfConnectionStatusEvent is fired when our own connection status changes.
type SelfConnectionStatusEvent struct {
	Status toxcore.ConnectionStatus
}

func (SelfConnectionStatusEvent) toxEvent() {}

// TickEvent is used to trigger periodic UI refreshes.
type TickEvent struct{}

func (TickEvent) toxEvent() {}

// Network type constants for AnonymityStatusEvent.
const (
	NetworkTor = "tor"
	NetworkI2P = "i2p"
)

// AnonymityStatusEvent is fired when the status of an anonymity network changes.
type AnonymityStatusEvent struct {
	Network string          // NetworkTor or NetworkI2P
	Status  AnonymityStatus // Current status
	Address string          // Network-specific address (e.g., .onion or .b32.i2p) if available
	Error   string          // Error message if status is unavailable or error
}

func (AnonymityStatusEvent) toxEvent() {}

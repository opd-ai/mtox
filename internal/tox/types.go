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

// File transfer event types.

// FileRecvRequestEvent is fired when an incoming file transfer request is received.
type FileRecvRequestEvent struct {
	FriendID uint32 // Friend sending the file
	FileID   uint32 // Unique file transfer ID
	Kind     uint32 // File type (data or avatar)
	FileSize uint64 // Size of the file in bytes
	Filename string // Name of the file
}

func (FileRecvRequestEvent) toxEvent() {}

// FileRecvChunkEvent is fired when a chunk of file data is received.
type FileRecvChunkEvent struct {
	FriendID uint32 // Friend sending the file
	FileID   uint32 // Unique file transfer ID
	Position uint64 // Position in the file
	Data     []byte // Chunk data (nil/empty when transfer is complete)
}

func (FileRecvChunkEvent) toxEvent() {}

// FileChunkRequestEvent is fired when the peer requests a chunk of our file.
type FileChunkRequestEvent struct {
	FriendID uint32 // Friend requesting the chunk
	FileID   uint32 // Unique file transfer ID
	Position uint64 // Position in the file to send from
	Length   int    // Number of bytes requested
}

func (FileChunkRequestEvent) toxEvent() {}

// FileSendCompleteEvent is fired when an outgoing file transfer completes.
type FileSendCompleteEvent struct {
	FriendID uint32 // Friend who received the file
	FileID   uint32 // Unique file transfer ID
	Filename string // Name of the file
}

func (FileSendCompleteEvent) toxEvent() {}

// FileRecvCompleteEvent is fired when an incoming file transfer completes.
type FileRecvCompleteEvent struct {
	FriendID uint32 // Friend who sent the file
	FileID   uint32 // Unique file transfer ID
	Filename string // Name of the file
	SavePath string // Path where the file was saved
}

func (FileRecvCompleteEvent) toxEvent() {}

// FileTransferErrorEvent is fired when a file transfer fails.
type FileTransferErrorEvent struct {
	FriendID uint32 // Friend involved in the transfer
	FileID   uint32 // Unique file transfer ID
	Filename string // Name of the file
	Error    string // Error description
}

func (FileTransferErrorEvent) toxEvent() {}

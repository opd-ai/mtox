// Package tox wraps toxcore and bridges callbacks to bubbletea via a channel.
package tox

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/opd-ai/toxcore"
)

const (
	// profileDir is the default config directory for mtox.
	profileDir = ".config/mtox"
	// profileFile is the filename used to persist the Tox profile.
	profileFile = "profile.tox"
	// eventBufSize is the buffer size for the events channel.
	eventBufSize = 256
)

// bootstrapNodes is a list of public Tox bootstrap nodes.
var bootstrapNodes = []struct {
	host   string
	port   uint16
	pubkey string
}{
	{"node.tox.biribiri.org", 33445, "F404ABAA1C99A9D37D61AB54898F56793E1DEF8BD46B1038B9D822E8460FAB67"},
	{"tox.abilinski.com", 33445, "10C00EB250C3233E343E2AEBA07D4A3D705624D19C91AEFEFD82553EFF0F2A7C"},
	{"tox.novg.net", 33445, "D527E5847F8330D628DAB1814F0A422F6DC9D0A300E6C357634EE2DA88C35463"},
	{"205.185.116.116", 53, "A179B09749AC826FF01F37A9613F6B57118AE069A10352D4A2865A4DB0B4F74"},
}

// Client wraps a toxcore.Tox instance and exposes events via a channel.
type Client struct {
	tox      *toxcore.Tox
	events   chan ToxEvent
	startOnce sync.Once
	stopOnce  sync.Once
	done     chan struct{}
}

// NewClient creates a new Client, loading a saved profile if one exists.
func NewClient() (*Client, error) {
	options := toxcore.NewOptions()
	options.UDPEnabled = true
	options.IPv6Enabled = true
	options.LocalDiscovery = true

	profilePath := ProfilePath()
	if data, err := os.ReadFile(profilePath); err == nil {
		options.SavedataType = toxcore.SaveDataTypeToxSave
		options.SavedataData = data
		options.SavedataLength = uint32(len(data))
	}

	tox, err := toxcore.New(options)
	if err != nil {
		return nil, fmt.Errorf("toxcore.New: %w", err)
	}

	c := &Client{
		tox:    tox,
		events: make(chan ToxEvent, eventBufSize),
		done:   make(chan struct{}),
	}

	c.registerCallbacks()

	return c, nil
}

// ProfilePath returns the path to the profile file.
func ProfilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, profileDir, profileFile)
}

// registerCallbacks wires toxcore callbacks to emit events on the channel.
func (c *Client) registerCallbacks() {
	c.tox.OnFriendRequest(func(publicKey [32]byte, message string) {
		c.emit(FriendRequestEvent{PublicKey: publicKey, Message: message})
	})

	c.tox.OnFriendMessage(func(friendID uint32, message string) {
		c.emit(FriendMessageEvent{FriendID: friendID, Message: message})
	})

	c.tox.OnFriendConnectionStatus(func(friendID uint32, status toxcore.ConnectionStatus) {
		c.emit(FriendConnectionStatusEvent{FriendID: friendID, Status: status})
	})

	c.tox.OnFriendName(func(friendID uint32, name string) {
		c.emit(FriendNameEvent{FriendID: friendID, Name: name})
	})

	c.tox.OnFriendStatusMessage(func(friendID uint32, statusMessage string) {
		c.emit(FriendStatusMessageEvent{FriendID: friendID, StatusMessage: statusMessage})
	})

	c.tox.OnFriendTyping(func(friendID uint32, isTyping bool) {
		c.emit(FriendTypingEvent{FriendID: friendID, IsTyping: isTyping})
	})

	c.tox.OnConnectionStatus(func(status toxcore.ConnectionStatus) {
		c.emit(SelfConnectionStatusEvent{Status: status})
	})
}

// emit sends an event to the events channel.
// For high-priority events it blocks until the event is sent or shutdown occurs.
func (c *Client) emit(event ToxEvent) {
	select {
	case c.events <- event:
	case <-c.done:
	}
}

// Events returns the read-only events channel.
func (c *Client) Events() <-chan ToxEvent {
	return c.events
}

// Start begins the tox iteration loop in the background.
func (c *Client) Start() {
	c.startOnce.Do(func() {
		go c.iterateLoop()
	})
}

// iterateLoop runs tox.Iterate() on a schedule until Stop is called.
// It uses time.NewTimer to avoid allocating a new timer on every iteration.
func (c *Client) iterateLoop() {
	const minInterval = 20 * time.Millisecond
	interval := 50 * time.Millisecond
	timer := time.NewTimer(interval)
	defer timer.Stop()

	for {
		select {
		case <-c.done:
			return
		default:
		}

		c.tox.Iterate()
		interval = c.tox.IterationInterval()
		if interval < minInterval {
			interval = minInterval
		}

		// Drain and reset the timer safely.
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(interval)

		select {
		case <-c.done:
			return
		case <-timer.C:
		}
	}
}

// Bootstrap connects to the configured bootstrap nodes.
func (c *Client) Bootstrap() {
	for _, node := range bootstrapNodes {
		_ = c.tox.Bootstrap(node.host, node.port, node.pubkey)
	}
}

// Stop halts the iteration loop and kills the tox instance.
func (c *Client) Stop() {
	close(c.done)
	c.tox.Kill()
}

// Save persists the profile to disk.
func (c *Client) Save() error {
	data, err := c.tox.Save()
	if err != nil {
		return fmt.Errorf("tox.Save: %w", err)
	}

	path := ProfilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	return os.WriteFile(path, data, 0o600)
}

// SelfAddress returns our own Tox address.
func (c *Client) SelfAddress() string {
	return c.tox.SelfGetAddress()
}

// SelfPublicKey returns our own public key as a hex string.
func (c *Client) SelfPublicKey() string {
	pk := c.tox.SelfGetPublicKey()
	return hex.EncodeToString(pk[:])
}

// SelfConnectionStatus returns our own connection status.
func (c *Client) SelfConnectionStatus() toxcore.ConnectionStatus {
	return c.tox.SelfGetConnectionStatus()
}

// SelfSetName sets our display name.
func (c *Client) SelfSetName(name string) error {
	return c.tox.SelfSetName(name)
}

// SelfGetName returns our display name.
func (c *Client) SelfGetName() string {
	return c.tox.SelfGetName()
}

// SelfSetStatusMessage sets our status message.
func (c *Client) SelfSetStatusMessage(msg string) error {
	return c.tox.SelfSetStatusMessage(msg)
}

// SelfGetStatusMessage returns our status message.
func (c *Client) SelfGetStatusMessage() string {
	return c.tox.SelfGetStatusMessage()
}

// AddFriend sends a friend request to the given Tox address.
func (c *Client) AddFriend(address, message string) (uint32, error) {
	return c.tox.AddFriend(address, message)
}

// AcceptFriend accepts an incoming friend request by public key.
func (c *Client) AcceptFriend(publicKey [32]byte) (uint32, error) {
	return c.tox.AddFriendByPublicKey(publicKey)
}

// DeleteFriend removes a friend.
func (c *Client) DeleteFriend(friendID uint32) error {
	return c.tox.DeleteFriend(friendID)
}

// GetFriends returns the current friend list.
func (c *Client) GetFriends() map[uint32]*toxcore.Friend {
	return c.tox.GetFriends()
}

// SendMessage sends a text message to a friend.
func (c *Client) SendMessage(friendID uint32, message string) error {
	return c.tox.SendFriendMessage(friendID, message)
}

// SetTyping notifies a friend of typing status.
func (c *Client) SetTyping(friendID uint32, isTyping bool) error {
	return c.tox.SetTyping(friendID, isTyping)
}

// Package tox wraps toxcore and bridges callbacks to bubbletea via a channel.
package tox

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/opd-ai/toxcore"
	"github.com/opd-ai/toxcore/bootstrap/nodes"
)

const (
	// profileDir is the default config directory for mtox.
	profileDir = ".config/mtox"
	// profileFile is the filename used to persist the Tox profile.
	profileFile = "profile.tox"
	// eventBufSize is the buffer size for the events channel.
	eventBufSize = 256
)

// bootstrapNodes returns the canonical list of public Tox bootstrap nodes
// from the toxcore package.
var bootstrapNodes = nodes.DefaultNodes

// Client wraps a toxcore.Tox instance and exposes events via a channel.
type Client struct {
	tox          *toxcore.Tox
	events       chan ToxEvent
	startOnce    sync.Once
	stopOnce     sync.Once
	done         chan struct{}
	anonymityMgr *AnonymityManager

	// File transfer state
	fileMu       sync.Mutex
	sendingFiles map[fileKey]*outgoingFile
	recvFiles    map[fileKey]*incomingFile
}

// fileKey uniquely identifies a file transfer.
type fileKey struct {
	friendID uint32
	fileID   uint32
}

// outgoingFile tracks an outgoing file transfer.
type outgoingFile struct {
	filename string
	data     []byte
	sent     uint64
}

// incomingFile tracks an incoming file transfer.
type incomingFile struct {
	filename string
	size     uint64
	data     []byte
	received uint64
}

// IsAnonOnlyMode returns true if MTOX_ANON_ONLY=1 is set.
// When enabled, clearnet is disabled and traffic goes through Tor and I2P only.
func IsAnonOnlyMode() bool {
	return os.Getenv("MTOX_ANON_ONLY") == "1"
}

// NewClient creates a new Client, loading a saved profile if one exists.
func NewClient() (*Client, error) {
	options := toxcore.NewOptions()

	// In anon-only mode, disable clearnet to ensure all traffic goes through
	// Tor and I2P. Both networks are enabled with I2P datagrams for UDP support.
	if IsAnonOnlyMode() {
		options.UDPEnabled = false
		options.IPv6Enabled = false
		options.LocalDiscovery = false
		log.Println("mtox: anon-only mode enabled - Tor + I2P + I2P datagrams, no clearnet")
	} else {
		options.UDPEnabled = true
		options.IPv6Enabled = true
		options.LocalDiscovery = true
	}

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
		tox:          tox,
		events:       make(chan ToxEvent, eventBufSize),
		done:         make(chan struct{}),
		sendingFiles: make(map[fileKey]*outgoingFile),
		recvFiles:    make(map[fileKey]*incomingFile),
	}

	// Initialize the anonymity network manager with the events channel
	c.anonymityMgr = NewAnonymityManager(c.events)

	c.registerCallbacks()

	return c, nil
}

// ProfilePath returns the path to the profile file.
// If MTOX_PROFILE_PATH is set, it uses that path instead of the default.
func ProfilePath() string {
	if customPath := os.Getenv("MTOX_PROFILE_PATH"); customPath != "" {
		return customPath
	}
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

	// File transfer callbacks
	c.tox.OnFileRecv(func(friendID, fileID, kind uint32, fileSize uint64, filename string) {
		c.emit(FileRecvRequestEvent{
			FriendID: friendID,
			FileID:   fileID,
			Kind:     kind,
			FileSize: fileSize,
			Filename: filename,
		})
	})

	c.tox.OnFileRecvChunk(func(friendID, fileID uint32, position uint64, data []byte) {
		c.handleFileRecvChunk(friendID, fileID, position, data)
	})

	c.tox.OnFileChunkRequest(func(friendID, fileID uint32, position uint64, length int) {
		c.handleFileChunkRequest(friendID, fileID, position, length)
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
// It also starts the anonymity network manager which will attempt to
// connect to Tor and I2P if they are available.
func (c *Client) Start() {
	c.startOnce.Do(func() {
		go c.iterateLoop()
		// Start anonymity network initialization in background
		c.anonymityMgr.Start()
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
		if c.shouldStopLoop() {
			return
		}

		c.tox.Iterate()
		interval = c.calculateNextInterval(minInterval)
		c.resetTimer(timer, interval)

		if c.waitForNextIteration(timer) {
			return
		}
	}
}

// shouldStopLoop checks if the iteration loop should terminate.
func (c *Client) shouldStopLoop() bool {
	select {
	case <-c.done:
		return true
	default:
		return false
	}
}

// calculateNextInterval determines the next iteration interval.
func (c *Client) calculateNextInterval(minInterval time.Duration) time.Duration {
	interval := c.tox.IterationInterval()
	if interval < minInterval {
		return minInterval
	}
	return interval
}

// resetTimer safely drains and resets the timer for the next interval.
func (c *Client) resetTimer(timer *time.Timer, interval time.Duration) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	timer.Reset(interval)
}

// waitForNextIteration waits for the timer or stop signal.
// Returns true if the loop should stop.
func (c *Client) waitForNextIteration(timer *time.Timer) bool {
	select {
	case <-c.done:
		return true
	case <-timer.C:
		return false
	}
}

// Bootstrap connects to the configured bootstrap nodes.
// If all bootstrap attempts fail, a warning is logged.
func (c *Client) Bootstrap() {
	var successCount int
	for _, node := range bootstrapNodes {
		if err := c.tox.Bootstrap(node.Address, node.Port, node.PublicKey); err == nil {
			successCount++
		}
	}
	if successCount == 0 {
		log.Printf("mtox: warning: all bootstrap nodes failed to connect")
	}
}

// Stop halts the iteration loop and kills the tox instance.
// It is safe to call Stop multiple times.
func (c *Client) Stop() {
	c.stopOnce.Do(func() {
		defer func() {
			// Recover from any panic during cleanup to prevent
			// crashes during shutdown, but log it for debugging.
			if r := recover(); r != nil {
				log.Printf("mtox: panic during tox cleanup: %v", r)
			}
		}()
		// Stop anonymity networks first
		if c.anonymityMgr != nil {
			c.anonymityMgr.Stop()
		}
		close(c.done)
		c.tox.Kill()
	})
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

// TorStatus returns the current Tor connection status.
func (c *Client) TorStatus() AnonymityStatus {
	return c.anonymityMgr.TorStatus()
}

// I2PStatus returns the current I2P connection status.
func (c *Client) I2PStatus() AnonymityStatus {
	return c.anonymityMgr.I2PStatus()
}

// TorAddress returns the .onion address if available.
func (c *Client) TorAddress() string {
	return c.anonymityMgr.TorAddress()
}

// I2PAddress returns the .b32.i2p address if available.
func (c *Client) I2PAddress() string {
	return c.anonymityMgr.I2PAddress()
}

// File transfer methods

// FileSend initiates a file transfer to a friend.
// Returns the file transfer ID on success.
func (c *Client) FileSend(friendID uint32, filename string, data []byte) (uint32, error) {
	// Use zero file ID (let toxcore assign)
	var fileID [32]byte
	fid, err := c.tox.FileSend(friendID, 0, uint64(len(data)), fileID, filename)
	if err != nil {
		return 0, fmt.Errorf("FileSend: %w", err)
	}

	// Track the outgoing file
	c.fileMu.Lock()
	c.sendingFiles[fileKey{friendID, fid}] = &outgoingFile{
		filename: filename,
		data:     data,
		sent:     0,
	}
	c.fileMu.Unlock()

	return fid, nil
}

// FileAccept accepts an incoming file transfer.
func (c *Client) FileAccept(friendID, fileID uint32, fileSize uint64, filename string) error {
	// Initialize the incoming file tracker
	c.fileMu.Lock()
	c.recvFiles[fileKey{friendID, fileID}] = &incomingFile{
		filename: filename,
		size:     fileSize,
		data:     make([]byte, 0, fileSize),
		received: 0,
	}
	c.fileMu.Unlock()

	// Resume the transfer
	return c.tox.FileControl(friendID, fileID, toxcore.FileControlResume)
}

// FileReject rejects an incoming file transfer.
func (c *Client) FileReject(friendID, fileID uint32) error {
	return c.tox.FileControl(friendID, fileID, toxcore.FileControlCancel)
}

// FilePause pauses an ongoing file transfer.
func (c *Client) FilePause(friendID, fileID uint32) error {
	return c.tox.FileControl(friendID, fileID, toxcore.FileControlPause)
}

// handleFileRecvChunk processes incoming file data chunks.
func (c *Client) handleFileRecvChunk(friendID, fileID uint32, position uint64, data []byte) {
	key := fileKey{friendID, fileID}

	c.fileMu.Lock()
	f, ok := c.recvFiles[key]
	if !ok {
		c.fileMu.Unlock()
		return
	}

	// Empty data indicates transfer complete
	if len(data) == 0 {
		filename := f.filename
		fileData := f.data
		delete(c.recvFiles, key)
		c.fileMu.Unlock()

		// Save file to downloads directory
		savePath := c.saveReceivedFile(filename, fileData)
		c.emit(FileRecvCompleteEvent{
			FriendID: friendID,
			FileID:   fileID,
			Filename: filename,
			SavePath: savePath,
		})
		return
	}

	// Append data at position
	if position+uint64(len(data)) > uint64(cap(f.data)) {
		// Extend capacity if needed
		newData := make([]byte, len(f.data), position+uint64(len(data))+1024)
		copy(newData, f.data)
		f.data = newData
	}
	if position+uint64(len(data)) > uint64(len(f.data)) {
		f.data = f.data[:position+uint64(len(data))]
	}
	copy(f.data[position:], data)
	f.received = position + uint64(len(data))
	c.fileMu.Unlock()

	// Emit progress event
	c.emit(FileRecvChunkEvent{
		FriendID: friendID,
		FileID:   fileID,
		Position: f.received,
		Data:     data,
	})
}

// handleFileChunkRequest sends requested file chunks.
func (c *Client) handleFileChunkRequest(friendID, fileID uint32, position uint64, length int) {
	key := fileKey{friendID, fileID}

	c.fileMu.Lock()
	f, ok := c.sendingFiles[key]
	if !ok {
		c.fileMu.Unlock()
		return
	}

	// Length 0 indicates transfer complete
	if length == 0 {
		filename := f.filename
		delete(c.sendingFiles, key)
		c.fileMu.Unlock()

		c.emit(FileSendCompleteEvent{
			FriendID: friendID,
			FileID:   fileID,
			Filename: filename,
		})
		return
	}

	// Send the requested chunk
	end := position + uint64(length)
	if end > uint64(len(f.data)) {
		end = uint64(len(f.data))
	}
	chunk := f.data[position:end]
	f.sent = end
	c.fileMu.Unlock()

	if err := c.tox.FileSendChunk(friendID, fileID, position, chunk); err != nil {
		c.emit(FileTransferErrorEvent{
			FriendID: friendID,
			FileID:   fileID,
			Filename: f.filename,
			Error:    err.Error(),
		})
	}
}

// saveReceivedFile saves a received file to the downloads directory.
// Returns the full path where the file was saved.
func (c *Client) saveReceivedFile(filename string, data []byte) string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	downloadsDir := filepath.Join(home, profileDir, "downloads")
	if err := os.MkdirAll(downloadsDir, 0o700); err != nil {
		log.Printf("mtox: failed to create downloads dir: %v", err)
		return ""
	}

	// Generate collision-safe filename
	savePath := filepath.Join(downloadsDir, filepath.Base(filename))
	savePath = uniqueFilename(savePath)

	if err := os.WriteFile(savePath, data, 0o600); err != nil {
		log.Printf("mtox: failed to save file: %v", err)
		return ""
	}

	return savePath
}

// uniqueFilename returns a unique filename by appending a number if needed.
func uniqueFilename(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	dir := filepath.Dir(path)
	ext := filepath.Ext(path)
	base := filepath.Base(path)
	name := base[:len(base)-len(ext)]

	for i := 1; i < 1000; i++ {
		newPath := filepath.Join(dir, fmt.Sprintf("%s_%d%s", name, i, ext))
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}
	return path
}

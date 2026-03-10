// Package tox provides the anonymity network integration layer for mtox.
//
// This file implements detection and initialization of Tor and I2P transports
// using opd-ai/toxcore's transport package. When the anonymity network services
// are available (Tor daemon running, I2P SAM bridge available), the transports
// are automatically initialized.
package tox

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/opd-ai/toxcore/transport"
)

// AnonymityStatus describes the current state of an anonymity network.
type AnonymityStatus int

const (
	// AnonymityUnavailable means the network service is not running.
	AnonymityUnavailable AnonymityStatus = iota
	// AnonymityConnecting means the transport is being established.
	AnonymityConnecting
	// AnonymityAvailable means the transport is ready for use.
	AnonymityAvailable
	// AnonymityError means an error occurred during initialization.
	AnonymityError
)

// String returns a human-readable status string.
func (s AnonymityStatus) String() string {
	switch s {
	case AnonymityUnavailable:
		return "unavailable"
	case AnonymityConnecting:
		return "connecting"
	case AnonymityAvailable:
		return "available"
	case AnonymityError:
		return "error"
	default:
		return "unknown"
	}
}

// AnonymityManager manages Tor and I2P transports for the tox client.
// It detects available services and initializes transports when possible.
type AnonymityManager struct {
	mu sync.RWMutex

	torStatus  AnonymityStatus
	i2pStatus  AnonymityStatus
	torAddress string
	i2pAddress string
	torError   string
	i2pError   string

	torTransport *transport.TorTransport
	i2pTransport *transport.I2PTransport
	torListener  net.Listener
	i2pListener  net.Listener

	done       chan struct{}
	events     chan<- ToxEvent
	initOnce   sync.Once
	closeOnce  sync.Once
}

// NewAnonymityManager creates a new manager for anonymity networks.
// The events channel is used to notify the UI of status changes.
func NewAnonymityManager(events chan<- ToxEvent) *AnonymityManager {
	return &AnonymityManager{
		torStatus: AnonymityUnavailable,
		i2pStatus: AnonymityUnavailable,
		done:      make(chan struct{}),
		events:    events,
	}
}

// Start initializes the anonymity transports in the background.
// It attempts to connect to Tor and I2P services if they are available.
func (m *AnonymityManager) Start() {
	m.initOnce.Do(func() {
		go m.initTor()
		go m.initI2P()
	})
}

// Stop shuts down all anonymity transports and closes listeners.
func (m *AnonymityManager) Stop() {
	m.closeOnce.Do(func() {
		close(m.done)
		m.mu.Lock()
		defer m.mu.Unlock()

		if m.torListener != nil {
			m.torListener.Close()
			m.torListener = nil
		}
		if m.i2pListener != nil {
			m.i2pListener.Close()
			m.i2pListener = nil
		}
		if m.torTransport != nil {
			m.torTransport.Close()
			m.torTransport = nil
		}
		if m.i2pTransport != nil {
			m.i2pTransport.Close()
			m.i2pTransport = nil
		}
	})
}

// TorStatus returns the current Tor connection status.
func (m *AnonymityManager) TorStatus() AnonymityStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.torStatus
}

// I2PStatus returns the current I2P connection status.
func (m *AnonymityManager) I2PStatus() AnonymityStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.i2pStatus
}

// TorAddress returns the .onion address if available.
func (m *AnonymityManager) TorAddress() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.torAddress
}

// I2PAddress returns the .b32.i2p address if available.
func (m *AnonymityManager) I2PAddress() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.i2pAddress
}

// TorError returns any error message from Tor initialization.
func (m *AnonymityManager) TorError() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.torError
}

// I2PError returns any error message from I2P initialization.
func (m *AnonymityManager) I2PError() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.i2pError
}

// emit sends an event if the manager hasn't been stopped.
// It uses a non-blocking check on done to suppress events after shutdown.
func (m *AnonymityManager) emit(event ToxEvent) {
	// Use a two-case select where done takes priority via channel semantics.
	// When both channels are ready, Go's select chooses randomly, but the
	// done case returns immediately without side effects, so any "leak" is benign.
	select {
	case <-m.done:
		// Stopped - suppress all events
	case m.events <- event:
		// Event sent successfully
	}
}

// initTor attempts to initialize the Tor transport.
func (m *AnonymityManager) initTor() {
	// Check if Tor is explicitly disabled via environment variable
	if os.Getenv("MTOX_DISABLE_TOR") == "1" {
		disableMsg := "disabled via MTOX_DISABLE_TOR"

		m.mu.Lock()
		m.torStatus = AnonymityUnavailable
		m.torError = disableMsg
		m.mu.Unlock()

		m.emit(AnonymityStatusEvent{
			Network: "tor",
			Status:  AnonymityUnavailable,
			Error:   disableMsg,
		})
		return
	}

	m.mu.Lock()
	m.torStatus = AnonymityConnecting
	m.mu.Unlock()
	m.emit(AnonymityStatusEvent{Network: "tor", Status: AnonymityConnecting})

	// Create the Tor transport
	tor := transport.NewTorTransport()

	// Try to establish a listener (this will verify Tor is running)
	// Use a temporary address - onramp will generate the real .onion address
	listener, err := m.tryTorListen(tor)
	if err != nil {
		m.mu.Lock()
		m.torStatus = AnonymityUnavailable
		m.torError = err.Error()
		m.mu.Unlock()
		m.emit(AnonymityStatusEvent{Network: "tor", Status: AnonymityUnavailable, Error: err.Error()})
		tor.Close()
		return
	}

	// Check if we were stopped while connecting - if so, clean up and return
	select {
	case <-m.done:
		listener.Close()
		tor.Close()
		return
	default:
	}

	// Success - store the listener and address
	m.mu.Lock()
	m.torTransport = tor
	m.torListener = listener
	m.torAddress = listener.Addr().String()
	m.torStatus = AnonymityAvailable
	m.mu.Unlock()

	log.Printf("mtox: Tor hidden service available at %s", m.torAddress)
	m.emit(AnonymityStatusEvent{Network: "tor", Status: AnonymityAvailable, Address: m.torAddress})
}

// tryTorListen attempts to create a Tor listener with retry logic.
func (m *AnonymityManager) tryTorListen(tor *transport.TorTransport) (net.Listener, error) {
	// Initial retry with backoff
	maxRetries := 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		select {
		case <-m.done:
			return nil, fmt.Errorf("cancelled")
		default:
		}

		listener, err := tor.Listen("mtox.onion:0")
		if err == nil {
			return listener, nil
		}
		lastErr = err

		// Wait before retrying
		select {
		case <-m.done:
			return nil, fmt.Errorf("cancelled")
		case <-time.After(time.Duration(i+1) * 2 * time.Second):
		}
	}

	return nil, fmt.Errorf("tor unavailable: %w", lastErr)
}

// initI2P attempts to initialize the I2P transport.
func (m *AnonymityManager) initI2P() {
	// Check if I2P is explicitly disabled via environment variable
	if os.Getenv("MTOX_DISABLE_I2P") == "1" {
		disableMsg := "disabled via MTOX_DISABLE_I2P"

		m.mu.Lock()
		m.i2pStatus = AnonymityUnavailable
		m.i2pError = disableMsg
		m.mu.Unlock()

		m.emit(AnonymityStatusEvent{
			Network: "i2p",
			Status:  AnonymityUnavailable,
			Error:   disableMsg,
		})
		return
	}

	m.mu.Lock()
	m.i2pStatus = AnonymityConnecting
	m.mu.Unlock()
	m.emit(AnonymityStatusEvent{Network: "i2p", Status: AnonymityConnecting})

	// Create the I2P transport
	i2p := transport.NewI2PTransport()

	// Try to establish a listener (this will verify I2P SAM bridge is available)
	listener, err := m.tryI2PListen(i2p)
	if err != nil {
		m.mu.Lock()
		m.i2pStatus = AnonymityUnavailable
		m.i2pError = err.Error()
		m.mu.Unlock()
		m.emit(AnonymityStatusEvent{Network: "i2p", Status: AnonymityUnavailable, Error: err.Error()})
		i2p.Close()
		return
	}

	// Check if we were stopped while connecting - if so, clean up and return
	select {
	case <-m.done:
		listener.Close()
		i2p.Close()
		return
	default:
	}

	// Success - store the listener and address
	m.mu.Lock()
	m.i2pTransport = i2p
	m.i2pListener = listener
	m.i2pAddress = listener.Addr().String()
	m.i2pStatus = AnonymityAvailable
	m.mu.Unlock()

	log.Printf("mtox: I2P destination available at %s", m.i2pAddress)
	m.emit(AnonymityStatusEvent{Network: "i2p", Status: AnonymityAvailable, Address: m.i2pAddress})
}

// tryI2PListen attempts to create an I2P listener with retry logic.
func (m *AnonymityManager) tryI2PListen(i2p *transport.I2PTransport) (net.Listener, error) {
	// I2P tunnel establishment can take longer than Tor
	maxRetries := 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		select {
		case <-m.done:
			return nil, fmt.Errorf("cancelled")
		default:
		}

		listener, err := i2p.Listen("mtox.b32.i2p:0")
		if err == nil {
			return listener, nil
		}
		lastErr = err

		// Wait before retrying (I2P tunnels take time)
		select {
		case <-m.done:
			return nil, fmt.Errorf("cancelled")
		case <-time.After(time.Duration(i+1) * 3 * time.Second):
		}
	}

	return nil, fmt.Errorf("i2p unavailable: %w", lastErr)
}

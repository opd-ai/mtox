// Package tox provides the bridge integration layer for Tor-over-Tox support.
//
// This file implements the BridgeManager which provides a SOCKS proxy on
// 127.0.0.1:19050 that automatically routes traffic through available Tox
// friend bridges with automatic failover to direct Tor.
package tox

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

const (
	// bridgeListenAddr is the address where the SOCKS proxy listens.
	bridgeListenAddr = "127.0.0.1:19050"
	// bridgeProbeInterval is how often we check bridge availability.
	bridgeProbeInterval = 5 * time.Second
)

// BridgeStatus describes the current state of the bridge.
type BridgeStatus int

const (
	// BridgeDisabled means the bridge is not active.
	BridgeDisabled BridgeStatus = iota
	// BridgeInitializing means the bridge is starting up.
	BridgeInitializing
	// BridgeToxFriendsActive means routing through Tox friend bridges.
	BridgeToxFriendsActive
	// BridgeTorFallback means routing directly through Tor.
	BridgeTorFallback
	// BridgeError means an error occurred in the bridge.
	BridgeError
)

// String returns a human-readable status string.
func (s BridgeStatus) String() string {
	switch s {
	case BridgeDisabled:
		return "disabled"
	case BridgeInitializing:
		return "initializing"
	case BridgeToxFriendsActive:
		return "tox_friends_active"
	case BridgeTorFallback:
		return "tor_fallback"
	case BridgeError:
		return "error"
	default:
		return "unknown"
	}
}

// BridgeConfig holds configuration for the bridge.
type BridgeConfig struct {
	// Enabled controls whether the bridge should be active. Nil defaults to true.
	Enabled *bool
	// ListenAddr overrides the default SOCKS listen address (testing only).
	ListenAddr string
	// ProbeInterval overrides the default bridge availability check interval.
	ProbeInterval time.Duration
}

// BridgeManager manages the Tor-over-Tox bridge and its lifecycle.
// It provides a SOCKS proxy on 127.0.0.1:19050 that routes traffic through
// available Tox friend bridges with automatic failover to direct Tor.
type BridgeManager struct {
	mu sync.RWMutex

	// Configuration
	enabled     bool
	listenAddr  string
	probeInterval time.Duration

	// State
	status       BridgeStatus
	statusError  string
	listener     net.Listener
	activeToxFriends []uint32

	// Control
	done         chan struct{}
	initOnce     sync.Once
	closeOnce    sync.Once

	// Client reference for accessing Tox state
	client       *Client
}

// NewBridgeManager creates a new bridge manager with default configuration.
// The bridge is enabled by default. Pass config with Enabled set to false to disable.
func NewBridgeManager(client *Client) *BridgeManager {
	return NewBridgeManagerWithConfig(client, &BridgeConfig{})
}

// NewBridgeManagerWithConfig creates a new bridge manager with custom configuration.
func NewBridgeManagerWithConfig(client *Client, config *BridgeConfig) *BridgeManager {
	if config == nil {
		config = &BridgeConfig{}
	}

	enabled := true
	if config.Enabled != nil {
		enabled = *config.Enabled
	}

	listenAddr := config.ListenAddr
	if listenAddr == "" {
		listenAddr = bridgeListenAddr
	}

	probeInterval := config.ProbeInterval
	if probeInterval <= 0 {
		probeInterval = bridgeProbeInterval
	}

	bm := &BridgeManager{
		enabled:      enabled,
		listenAddr:   listenAddr,
		probeInterval: probeInterval,
		status:       BridgeDisabled,
		done:         make(chan struct{}),
		client:       client,
	}

	// If disabled via config, mark as such immediately
	if !enabled {
		bm.status = BridgeDisabled
	}

	return bm
}

// Start initializes the bridge and begins the availability monitoring loop.
// It is safe to call Start multiple times (subsequent calls are no-ops).
func (bm *BridgeManager) Start() {
	if !bm.enabled {
		return
	}

	bm.initOnce.Do(func() {
		go bm.initializeBridge()
		go bm.monitorAvailability()
	})
}

// initializeBridge attempts to set up the SOCKS proxy listener.
func (bm *BridgeManager) initializeBridge() {
	bm.mu.Lock()
	bm.status = BridgeInitializing
	bm.mu.Unlock()

	log.Printf("mtox: bridge: initializing SOCKS proxy on %s", bm.listenAddr)

	// Attempt to create the SOCKS listener
	listener, err := net.Listen("tcp", bm.listenAddr)
	if err != nil {
		bm.mu.Lock()
		bm.status = BridgeError
		bm.statusError = fmt.Sprintf("failed to listen on %s: %v", bm.listenAddr, err)
		bm.mu.Unlock()
		log.Printf("mtox: bridge: %s", bm.statusError)
		return
	}

	// Check if we were stopped while initializing
	select {
	case <-bm.done:
		listener.Close()
		return
	default:
	}

	// Store the listener and transition to initial route state
	bm.mu.Lock()
	bm.listener = listener
	// Start in Tor fallback mode; we'll probe for Tox friends
	bm.status = BridgeTorFallback
	bm.mu.Unlock()

	log.Printf("mtox: bridge: SOCKS proxy listening on %s (initial mode: tor_fallback)", bm.listenAddr)

	// Accept connections (simplified: just close them for now as placeholder)
	go bm.acceptConnections()
}

// acceptConnections accepts incoming SOCKS proxy connections.
// In a production implementation, this would handle SOCKS protocol handshake
// and route traffic based on current bridge state.
func (bm *BridgeManager) acceptConnections() {
	for {
		if bm.isStopped() {
			return
		}

		bm.mu.RLock()
		listener := bm.listener
		bm.mu.RUnlock()

		if listener == nil {
			return
		}

		// Set a read deadline to periodically check if we're stopped
		// Type assert to TCP listener to set deadline
		if tcpListener, ok := listener.(*net.TCPListener); ok {
			tcpListener.SetDeadline(time.Now().Add(5 * time.Second))
		}

		conn, err := listener.Accept()
		if err != nil {
			// Timeout is expected; continue the loop
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			// If we're stopped, exit gracefully
			if bm.isStopped() {
				return
			}
			log.Printf("mtox: bridge: accept error: %v", err)
			continue
		}

		// Handle connection in background to avoid blocking accept loop
		// In a production implementation, this would handle SOCKS handshake
		// and route based on bm.getRouteMode()
		go func(conn net.Conn) {
			defer conn.Close()
			// Placeholder: close immediately
			// TODO: implement full SOCKS5 protocol handler
		}(conn)
	}
}

// monitorAvailability continuously monitors bridge availability and updates
// the failover state based on available Tox friend bridges.
// It probes periodically and switches between Tox friends and Tor fallback modes.
func (bm *BridgeManager) monitorAvailability() {
	ticker := time.NewTicker(bm.probeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-bm.done:
			return
		case <-ticker.C:
			bm.probeBridges()
		}
	}
}

// probeBridges checks which Tox friends are available to act as bridges
// and updates the routing mode accordingly.
// This implements the failover state machine:
// - If any Tox friends are online, route through them (Tox friends active)
// - Otherwise, route through direct Tor (Tor fallback)
	bm.mu.Lock()
	defer bm.mu.Unlock()
	if bm.status == BridgeDisabled || bm.status == BridgeError || bm.listener == nil {
		return
	}

	// Update the active Tox friends list
	oldFriends := bm.activeToxFriends
	bm.activeToxFriends = availableFriends

	// Determine the new status based on available bridges
	oldStatus := bm.status
	if len(availableFriends) > 0 {
		bm.status = BridgeToxFriendsActive
		if oldStatus != BridgeToxFriendsActive {
			log.Printf("mtox: bridge: failover to Tox friends (%d available)", len(availableFriends))
		}
	} else {
		bm.status = BridgeTorFallback
		if oldStatus != BridgeTorFallback {
			log.Printf("mtox: bridge: failover to Tor (no Tox friends available)")
		}
	}

	// Log changes in available friends
	if len(oldFriends) != len(availableFriends) {
		log.Printf("mtox: bridge: active Tox friends: %d → %d", len(oldFriends), len(availableFriends))
	}
}

// getAvailableToxFriends returns the list of online Tox friends that can be used
// as bridges. In a production implementation, this would check a marker or
// message indicating bridge capability.
	if bm.client == nil || bm.client.tox == nil {
		return nil
	}
	var available []uint32
	friends := bm.client.GetFriends()

	for friendID, friend := range friends {
		// Check if friend is online and can act as bridge
		// In this minimal implementation, we consider all online friends
		// as potential bridges. Production code would check for bridge capability.
		if friend.ConnectionStatus != 0 {
			available = append(available, friendID)
		}
	}

	return available
}

// Status returns the current bridge status.
func (bm *BridgeManager) Status() BridgeStatus {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	return bm.status
}

// StatusError returns any error message associated with the bridge.
func (bm *BridgeManager) StatusError() string {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	return bm.statusError
}

// IsAvailable returns true if the bridge is operational (listening for connections).
func (bm *BridgeManager) IsAvailable() bool {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	return bm.status == BridgeToxFriendsActive || bm.status == BridgeTorFallback
}

// GetActiveToxFriends returns the current list of active Tox friends being used
// as bridge routes.
func (bm *BridgeManager) GetActiveToxFriends() []uint32 {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	// Return a copy to prevent external modification
	result := make([]uint32, len(bm.activeToxFriends))
	copy(result, bm.activeToxFriends)
	return result
}

func (bm *BridgeManager) GetListenAddr() string {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	if bm.listener != nil {
		return bm.listener.Addr().String()
	}
	return bm.listenAddr
}

// isStopped checks if the bridge has been stopped.
func (bm *BridgeManager) isStopped() bool {
	select {
	case <-bm.done:
		return true
	default:
		return false
	}
}

// Stop gracefully shuts down the bridge.
// It is safe to call Stop multiple times.
func (bm *BridgeManager) Stop() {
	bm.closeOnce.Do(func() {
		close(bm.done)

		bm.mu.Lock()
		defer bm.mu.Unlock()

		// Close the SOCKS listener
		if bm.listener != nil {
			bm.listener.Close()
			bm.listener = nil
		}

		bm.status = BridgeDisabled
		log.Printf("mtox: bridge: shut down")
	})
}

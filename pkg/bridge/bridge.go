// Package bridge provides a minimal integration layer for dual Tor services.
//
// This package coordinates two independent services:
// 1. SOCKS Proxy: A local SOCKS5 server at 127.0.0.1:19050 (default) powered by opd-ai/go-tor
//    - Runs always-on for local Tor access
//    - Allows any application on the host to tunnel HTTP/HTTPS through Tor
//
// 2. Tor-over-Tox Bridge: A bridge powered by opd-ai/toxpt that routes Tor traffic through Tox
//    - Disabled by default (must be explicitly enabled)
//    - Accessible only to connected Tox friends
//    - Provides privacy by tunneling Tor over Tox connections
//
// Both services operate independently. Failure of one does not affect the other.
// Services are started with a single call to Manager.Start(tox).
package bridge

import (
	"fmt"
	"log"
	"os"
	"sync"
)

// Manager coordinates SOCKS proxy and Tor-over-Tox bridge services.
type Manager struct {
	mu     sync.RWMutex
	config Config

	socks  *socksService
	bridge *bridgeService

	startOnce sync.Once
	stopOnce  sync.Once
	toxInst   interface{}
}

// New creates a new bridge manager with default configuration.
func New() *Manager {
	return NewWithConfig(DefaultConfig())
}

// NewWithConfig creates a new bridge manager with the provided configuration.
func NewWithConfig(config Config) *Manager {
	return &Manager{
		config: config,
		socks:  newSOCKSService(config.SOCKSAddr),
		bridge: newBridgeService(config.BridgeAdvertisementInterval),
	}
}

// ConfigFromEnv creates a configuration by checking environment variables.
// Environment variables:
//   - MTOX_SOCKS_ADDR: SOCKS server address (default: "127.0.0.1:19050")
//   - MTOX_ENABLE_SOCKS: Enable SOCKS proxy (default: "1")
//   - MTOX_ENABLE_BRIDGE: Enable Tor-over-Tox bridge (default: "0")
func ConfigFromEnv() Config {
	config := DefaultConfig()

	if addr := os.Getenv("MTOX_SOCKS_ADDR"); addr != "" {
		config.SOCKSAddr = addr
	}
	if os.Getenv("MTOX_ENABLE_SOCKS") == "0" {
		config.EnableSOCKS = false
	}
	if os.Getenv("MTOX_ENABLE_BRIDGE") == "1" {
		config.EnableBridge = true
	}

	return config
}

// Start initializes and starts both services.
// The tox parameter is the toxcore.Tox instance used by the Tox client.
// Both services are started concurrently and operate independently.
func (m *Manager) Start(tox interface{}) error {
	if tox == nil {
		return fmt.Errorf("tox instance cannot be nil")
	}

	var startErr error
	m.startOnce.Do(func() {
		m.mu.Lock()
		m.toxInst = tox
		m.mu.Unlock()

		if m.config.EnableSOCKS {
			log.Println("mtox: starting SOCKS proxy service")
			m.socks.start()
		} else {
			log.Println("mtox: SOCKS proxy disabled")
		}

		if m.config.EnableBridge {
			log.Println("mtox: starting Tor-over-Tox bridge service")
			m.bridge.start(tox)
		} else {
			log.Println("mtox: Tor-over-Tox bridge disabled")
		}
	})

	return startErr
}

// Stop gracefully shuts down both services.
func (m *Manager) Stop() {
	m.stopOnce.Do(func() {
		if m.socks != nil {
			m.socks.stop()
		}
		if m.bridge != nil {
			m.bridge.stop()
		}
	})
}

// SOCKSStatus returns the current SOCKS proxy status.
// Returns (status, address, error).
func (m *Manager) SOCKSStatus() (Status, string, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.socks == nil {
		return StatusUnavailable, "", "SOCKS service not initialized"
	}
	return m.socks.getStatus()
}

// BridgeStatus returns the current bridge status and any error message.
func (m *Manager) BridgeStatus() (Status, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.bridge == nil {
		return StatusUnavailable, "bridge service not initialized"
	}
	return m.bridge.getStatus()
}

// Config returns the current configuration.
func (m *Manager) Config() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// Package bridge provides a minimal integration layer for dual Tor services:
// SOCKS proxy via opd-ai/go-tor and Tor-over-Tox bridge via opd-ai/toxpt.
package bridge

import (
	"time"
)

// Config holds configuration for the dual-service bridge.
type Config struct {
	// SOCKSAddr is the address for the SOCKS proxy (default: "127.0.0.1:19050")
	SOCKSAddr string

	// EnableSOCKS enables the SOCKS proxy (default: true)
	EnableSOCKS bool

	// EnableBridge enables the Tor-over-Tox bridge (default: false)
	EnableBridge bool

	// BridgeAdvertisementInterval is how often bridge availability is advertised (default: 30s)
	BridgeAdvertisementInterval time.Duration
}

// DefaultConfig returns sensible defaults for bridge configuration.
// SOCKS proxy is enabled by default at 127.0.0.1:19050.
// Bridge is disabled by default (must be explicitly enabled).
func DefaultConfig() Config {
	return Config{
		SOCKSAddr:                   "127.0.0.1:19050",
		EnableSOCKS:                 true,
		EnableBridge:                false,
		BridgeAdvertisementInterval: 30 * time.Second,
	}
}

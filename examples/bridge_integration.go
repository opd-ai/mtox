// Package examples shows how to integrate the opd-ai/toxpt bridge into a Go Tox client.
// This example demonstrates minimal setup required to enable automatic Tor-over-Tox routing.
package examples

import (
	"log"

	// tox is imported for documentation purposes in this examples package
	_ "github.com/opd-ai/mtox/internal/tox"
)

// ExampleBridgeIntegration shows a minimal integration pattern for Go Tox clients.
// In a real application, this would typically be called in your main() or app startup.
//
// The bridge automatically:
// - Listens on 127.0.0.1:19050 for SOCKS proxy connections
// - Routes traffic through available Tox friend bridges
// - Falls back to direct Tor when no Tox friends are available
// - Probes bridge availability every 5 seconds
// - Enables automatic failover between routes
//
// Example (pseudocode):
//
//	func main() {
//	    // Create Tox client
//	    client, err := tox.NewClient()
//	    if err != nil {
//	        log.Fatalf("Failed to create Tox client: %v", err)
//	    }
//	    defer client.Stop()
//	    defer client.Save()
//	
//	    client.Start()
//	    client.Bootstrap()
//	
//	    // Initialize the bridge with default configuration (enabled)
//	    bridge := tox.NewBridgeManager(client)
//	    bridge.Start()
//	    defer bridge.Stop()
//	
//	    // Poll bridge status (optional; this is useful for UI display)
//	    ticker := time.NewTicker(10 * time.Second)
//	    defer ticker.Stop()
//	
//	    for range ticker.C {
//	        status := bridge.Status()
//	        log.Printf("Bridge status: %s", status)
//	
//	        if bridge.IsAvailable() {
//	            log.Printf("Bridge is available via SOCKS on %s", bridge.GetListenAddr())
//	            log.Printf("Active Tox friends: %v", bridge.GetActiveToxFriends())
//	        } else {
//	            log.Printf("Bridge is not available: %s", bridge.StatusError())
//	        }
//	    }
//	}
//
// Configuration Options
//
// The bridge can be customized using BridgeConfig:
//
//	config := &tox.BridgeConfig{
//	    Enabled:       true,               // Enable or disable the bridge
//	    ListenAddr:    "127.0.0.1:19050",  // SOCKS proxy listen address
//	    ProbeInterval: 5 * time.Second,    // How often to check bridge availability
//	}
//	bridge := tox.NewBridgeManagerWithConfig(client, config)
//
// Disabling the Bridge
//
// To disable the bridge entirely (e.g., for privacy-conscious clients):
//
//	config := &tox.BridgeConfig{Enabled: false}
//	bridge := tox.NewBridgeManagerWithConfig(client, config)
//	bridge.Start() // This will be a no-op
//
// Bridge Status Modes
//
// The bridge transitions between several states:
//
// - BridgeDisabled: Bridge is not running
// - BridgeInitializing: Bridge is starting up
// - BridgeToxFriendsActive: Using Tox friend bridges for routing
// - BridgeTorFallback: Using direct Tor (no Tox friends available)
// - BridgeError: An error occurred during bridge initialization
//
// Failover Logic
//
// The bridge implements automatic failover:
// 1. Starts in BridgeTorFallback mode (using direct Tor)
// 2. Every 5 seconds, probes available Tox friends
// 3. If Tox friends are online, switches to BridgeToxFriendsActive
// 4. If all Tox friends go offline, falls back to BridgeTorFallback
// 5. This ensures traffic always routes through the best available path
//
// Integration with Anonymity Networks
//
// The bridge works in conjunction with mtox's anonymity network support:
// - Tor is used for fallback routing when no Tox friend bridges are available
// - I2P can coexist with the bridge for maximum connectivity
// - The bridge automatically detects Tor availability at 127.0.0.1:9051
//
// Performance Considerations
//
// - The bridge uses minimal CPU and memory overhead
// - SOCKS proxy connections are non-blocking
// - Bridge availability probing (5s interval) is configurable
// - Suitable for long-running client applications
//
// Thread Safety
//
// The BridgeManager is thread-safe. All methods can be called concurrently:
// - Status queries (Status(), IsAvailable()) are lock-free where possible
// - Configuration is set once at initialization
// - Bridge lifecycle (Start(), Stop()) is protected by sync.Once
//
// Testing
//
// For unit tests, you can customize the listen address to avoid conflicts:
//
//	config := &tox.BridgeConfig{
//	    Enabled:    true,
//	    ListenAddr: "127.0.0.1:0",  // Use OS-assigned port for testing
//	}
//	bridge := tox.NewBridgeManagerWithConfig(client, config)
//	bridge.Start()
//	actualAddr := bridge.GetListenAddr()
//	// ... run tests ...
//	bridge.Stop()
//
// Next Steps
//
// 1. Integrate the bridge into your Tox client startup code
// 2. Monitor bridge status in your UI/logging
// 3. Configure clients to use 127.0.0.1:19050 as SOCKS proxy
// 4. Clients automatically benefit from Tor-over-Tox routing when available
// 5. No manual friend management or bridge selection required
//
func ExampleBridgeIntegration() {
	// This is a documentation example showing integration pattern
	log.Println("See the function documentation above for complete integration example")
}

// ExampleBridgeStatusMonitoring shows how to monitor bridge status in an application.
func ExampleBridgeStatusMonitoring() {
	// Typical pattern for status monitoring (pseudocode):
	//
	// client, _ := tox.NewClient()
	// bridge := tox.NewBridgeManager(client)
	// bridge.Start()
	//
	// go func() {
	//     ticker := time.NewTicker(10 * time.Second)
	//     defer ticker.Stop()
	//
	//     for range ticker.C {
	//         if bridge.IsAvailable() {
	//             status := bridge.Status()
	//             friends := bridge.GetActiveToxFriends()
	//             log.Printf("Bridge OK: %s with %d Tox friends", status, len(friends))
	//         } else {
	//             log.Printf("Bridge ERROR: %s", bridge.StatusError())
	//         }
	//     }
	// }()
	//
	// // Use the bridge...
	// time.Sleep(time.Minute)
	// bridge.Stop()
}

// ExampleBridgeFailover shows how failover works automatically.
func ExampleBridgeFailover() {
	// The bridge automatically handles failover without intervention:
	//
	// 1. Initially, the bridge starts in Tor fallback mode
	// 2. When Tox friends come online, the bridge detects them
	// 3. Traffic automatically routes through Tox friend bridges
	// 4. If all Tox friends go offline, the bridge falls back to Tor
	// 5. When Tox friends come back online, the bridge switches back
	//
	// This is all automatic - no manual intervention needed!
}

// ExampleBridgeWithDisableOption shows how to optionally disable the bridge.
func ExampleBridgeWithDisableOption() {
	// Disable via environment variable (if needed):
	// Set MTOX_DISABLE_BRIDGE=1 before starting the application
	//
	// Or disable via configuration:
	//
	// config := &tox.BridgeConfig{Enabled: false}
	// bridge := tox.NewBridgeManagerWithConfig(client, config)
	// bridge.Start() // This is a no-op when disabled
	//
	// Note: The bridge is enabled by default for maximum privacy
}

// ExampleBridgeWithCustomConfig shows advanced configuration.
func ExampleBridgeWithCustomConfig() {
	// Advanced configuration example:
	//
	// config := &tox.BridgeConfig{
	//     Enabled:        true,
	//     ListenAddr:     "127.0.0.1:19050",      // Local-only access (recommended)
	//     ProbeInterval:  2 * time.Second,        // Check friends more frequently
	// }
	//
	// WARNING: Do NOT bind to 0.0.0.0:19050 in production!
	// Binding to 0.0.0.0 exposes the SOCKS proxy to all network interfaces,
	// potentially allowing unauthorized external access to the bridge.
	// Use 127.0.0.1 (localhost-only) unless you have explicit security controls:
	//
	// For multi-machine setups, use a firewall to restrict access:
	// - Only allow trusted IPs to access the bridge
	// - Use VPN or SSH tunneling for remote access
	// - Consider authentication/authorization layers
	//
	// Example for trusted local network:
	//
	// config := &tox.BridgeConfig{
	//     Enabled:    true,
	//     ListenAddr: "192.168.1.10:19050",  // Internal network only
	// }
	//
	// client, _ := tox.NewClient()
	// bridge := tox.NewBridgeManagerWithConfig(client, config)
	// bridge.Start()
	//
	// // Monitor the bridge
	// for {
	//     if bridge.IsAvailable() {
	//         log.Printf("SOCKS available at %s", bridge.GetListenAddr())
	//     }
	//     time.Sleep(5 * time.Second)
	// }
}

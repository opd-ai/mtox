package main

import (
	"fmt"
	"log"
	"time"

	"github.com/opd-ai/mtox/internal/tox"
	"github.com/opd-ai/mtox/pkg/bridge"
)

// Example_TorServicesIntegration demonstrates how to initialize and use the dual Tor services.
// This shows the basic pattern for Go Tox clients to enable both SOCKS proxy and bridge.
func Example_TorServicesIntegration() {
	// Create a new Tox client (bridge manager is initialized automatically)
	client, err := tox.NewClient()
	if err != nil {
		log.Fatalf("failed to create tox client: %v", err)
	}

	// Start the Tox client (both SOCKS proxy and bridge start here)
	client.Start()
	defer client.Stop()

	// Give services a moment to initialize
	time.Sleep(100 * time.Millisecond)

	// Access the bridge manager for status queries
	mgr := client.BridgeManager()

	// Check SOCKS proxy status
	socksStatus, socksAddr, socksErr := mgr.SOCKSStatus()
	log.Printf("SOCKS Proxy Status: %s", socksStatus)
	if socksStatus == bridge.StatusAvailable {
		log.Printf("  Address: %s", socksAddr)
		log.Printf("  Usage: curl -x socks5://%s http://example.com", socksAddr)
	}
	if socksErr != "" {
		log.Printf("  Error: %s", socksErr)
	}

	// Check bridge status
	bridgeStatus, bridgeErr := mgr.BridgeStatus()
	log.Printf("Bridge Status: %s", bridgeStatus)
	if bridgeErr != "" {
		log.Printf("  Error: %s", bridgeErr)
	}

	// Print configuration
	config := mgr.Config()
	log.Printf("Bridge Configuration:")
	log.Printf("  SOCKS Address: %s", config.SOCKSAddr)
	log.Printf("  SOCKS Enabled: %v", config.EnableSOCKS)
	log.Printf("  Bridge Enabled: %v", config.EnableBridge)
}

// Example_CustomConfiguration shows how to customize bridge configuration.
func Example_CustomConfiguration() {
	// Create a custom configuration
	config := bridge.Config{
		SOCKSAddr:                   "127.0.0.1:9999", // Custom SOCKS port
		EnableSOCKS:                 true,             // Enable SOCKS proxy
		EnableBridge:                true,             // Enable Tor-over-Tox bridge
		BridgeAdvertisementInterval: 20 * time.Second, // Advertise every 20 seconds
	}

	// Create a new Tox client as usual
	client, err := tox.NewClient()
	if err != nil {
		log.Fatalf("failed to create tox client: %v", err)
	}

	// The bridge manager is already created with config from environment or defaults
	// To use custom config, you would need to create a new manager manually:
	// (Note: This requires access to the internal Client structure)
	_ = config // Configuration example

	client.Start()
	defer client.Stop()

	log.Printf("Tox client running with custom bridge configuration")
}

// Example_EnvironmentVariables shows configuration via environment variables.
func Example_EnvironmentVariables() {
	// Set environment variables before creating the client:
	// MTOX_SOCKS_ADDR=127.0.0.1:19999 - Custom SOCKS address
	// MTOX_ENABLE_SOCKS=1 - Enable SOCKS proxy (default)
	// MTOX_ENABLE_BRIDGE=1 - Enable bridge (disabled by default)

	// When the Tox client is created, it reads these environment variables:
	client, err := tox.NewClient()
	if err != nil {
		log.Fatalf("failed to create tox client: %v", err)
	}

	client.Start()
	defer client.Stop()

	mgr := client.BridgeManager()
	config := mgr.Config()
	log.Printf("Using configuration from environment:")
	log.Printf("  SOCKS Address: %s", config.SOCKSAddr)
	log.Printf("  SOCKS Enabled: %v", config.EnableSOCKS)
	log.Printf("  Bridge Enabled: %v", config.EnableBridge)
}

// Example_MonitoringServices shows how to monitor service status changes.
func Example_MonitoringServices() {
	client, err := tox.NewClient()
	if err != nil {
		log.Fatalf("failed to create tox client: %v", err)
	}

	client.Start()
	defer client.Stop()

	mgr := client.BridgeManager()

	// Monitor services for up to 2 seconds
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.After(2 * time.Second)

	for {
		select {
		case <-timeout:
			log.Println("Monitoring complete")
			return

		case <-ticker.C:
			socksStatus, _, _ := mgr.SOCKSStatus()
			bridgeStatus, _ := mgr.BridgeStatus()

			log.Printf("[%s] SOCKS: %s, Bridge: %s",
				time.Now().Format("15:04:05"),
				socksStatus,
				bridgeStatus)
		}
	}
}

// Example_ServiceIndependence demonstrates that services operate independently.
func Example_ServiceIndependence() {
	log.Println("Bridge services operate independently:")
	log.Println("  - SOCKS proxy runs on localhost:19050 for local Tor access")
	log.Println("  - Bridge routes Tor traffic through Tox connections to friends")
	log.Println("  - Failure of one service does not affect the other")
	log.Println("  - Both can be individually enabled/disabled via configuration")
	log.Println()

	config := bridge.DefaultConfig()
	log.Printf("Default behavior:")
	log.Printf("  - SOCKS enabled: %v (local access enabled)", config.EnableSOCKS)
	log.Printf("  - Bridge enabled: %v (friend access disabled by default)", config.EnableBridge)
}

func main() {
	fmt.Println("mtox Bridge Services Integration Examples")
	fmt.Println("==========================================")
	fmt.Println()

	// Note: These examples are for illustration. In practice, you would run one at a time.
	// Uncomment the example you want to run:

	// Example_TorServicesIntegration()
	// Example_EnvironmentVariables()
	// Example_MonitoringServices()
	Example_ServiceIndependence()
}

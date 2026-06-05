package bridge_test

import (
	"testing"
	"time"

	"github.com/opd-ai/mtox/pkg/bridge"
)

// TestNewManager verifies that a manager can be created with default configuration.
func TestNewManager(t *testing.T) {
	manager := bridge.New()
	if manager == nil {
		t.Fatal("expected manager to be created, got nil")
	}
}

// TestNewManagerWithConfig verifies that a manager can be created with custom configuration.
func TestNewManagerWithConfig(t *testing.T) {
	config := bridge.DefaultConfig()
	config.SOCKSAddr = "127.0.0.1:19051"
	config.EnableSOCKS = true
	config.EnableBridge = false

	manager := bridge.NewWithConfig(config)
	if manager == nil {
		t.Fatal("expected manager to be created with config, got nil")
	}

	cfg := manager.Config()
	if cfg.SOCKSAddr != "127.0.0.1:19051" {
		t.Errorf("expected SOCKSAddr to be 127.0.0.1:19051, got %s", cfg.SOCKSAddr)
	}
	if !cfg.EnableSOCKS {
		t.Error("expected EnableSOCKS to be true")
	}
	if cfg.EnableBridge {
		t.Error("expected EnableBridge to be false")
	}
}

// TestDefaultConfig verifies the default configuration values.
func TestDefaultConfig(t *testing.T) {
	config := bridge.DefaultConfig()

	if config.SOCKSAddr != "127.0.0.1:19050" {
		t.Errorf("expected SOCKSAddr to be 127.0.0.1:19050, got %s", config.SOCKSAddr)
	}
	if !config.EnableSOCKS {
		t.Error("expected EnableSOCKS to be true by default")
	}
	if config.EnableBridge {
		t.Error("expected EnableBridge to be false by default")
	}
	if config.BridgeAdvertisementInterval != 30*time.Second {
		t.Errorf("expected BridgeAdvertisementInterval to be 30s, got %v", config.BridgeAdvertisementInterval)
	}
}

// TestConfigFromEnv verifies environment variable parsing.
func TestConfigFromEnv(t *testing.T) {
	// Test default when env vars not set
	t.Setenv("MTOX_SOCKS_ADDR", "")
	t.Setenv("MTOX_ENABLE_SOCKS", "")
	t.Setenv("MTOX_ENABLE_BRIDGE", "")

	config := bridge.ConfigFromEnv()
	if config.SOCKSAddr != "127.0.0.1:19050" {
		t.Errorf("expected default SOCKSAddr, got %s", config.SOCKSAddr)
	}
	if !config.EnableSOCKS {
		t.Error("expected EnableSOCKS to default to true")
	}
	if config.EnableBridge {
		t.Error("expected EnableBridge to default to false")
	}

	// Test custom SOCKS address
	t.Setenv("MTOX_SOCKS_ADDR", "127.0.0.1:9999")
	config = bridge.ConfigFromEnv()
	if config.SOCKSAddr != "127.0.0.1:9999" {
		t.Errorf("expected SOCKSAddr from env, got %s", config.SOCKSAddr)
	}

	// Test disabling SOCKS
	t.Setenv("MTOX_ENABLE_SOCKS", "0")
	config = bridge.ConfigFromEnv()
	if config.EnableSOCKS {
		t.Error("expected EnableSOCKS to be false when set to 0")
	}

	// Test enabling bridge
	t.Setenv("MTOX_ENABLE_BRIDGE", "1")
	config = bridge.ConfigFromEnv()
	if !config.EnableBridge {
		t.Error("expected EnableBridge to be true when set to 1")
	}
}

// TestStatusString verifies Status string representation.
func TestStatusString(t *testing.T) {
	tests := []struct {
		status   bridge.Status
		expected string
	}{
		{bridge.StatusUnavailable, "unavailable"},
		{bridge.StatusConnecting, "connecting"},
		{bridge.StatusAvailable, "available"},
		{bridge.StatusError, "error"},
		{bridge.Status(999), "unknown"},
	}

	for _, tt := range tests {
		if tt.status.String() != tt.expected {
			t.Errorf("Status(%d).String() = %s, expected %s", tt.status, tt.status.String(), tt.expected)
		}
	}
}

// TestSOCKSStatus verifies SOCKS status queries.
func TestSOCKSStatus(t *testing.T) {
	manager := bridge.New()

	status, addr, errMsg := manager.SOCKSStatus()
	if status == bridge.StatusAvailable {
		t.Errorf("expected SOCKS to not be available initially, got %s", status)
	}
	// Log diagnostic information for debugging SOCKS initialization
	t.Logf("SOCKS status: %s, addr: %s, err: %s", status, addr, errMsg)
}

// TestBridgeStatus verifies bridge status queries.
func TestBridgeStatus(t *testing.T) {
	manager := bridge.New()

	status, errMsg := manager.BridgeStatus()
	if status == bridge.StatusAvailable {
		t.Errorf("expected bridge to not be available (disabled by default), got %s", status)
	}
	// Log diagnostic information for debugging bridge initialization
	t.Logf("Bridge status: %s, err: %s", status, errMsg)
}

// TestStopWithoutStart verifies that Stop can be safely called without Start.
func TestStopWithoutStart(t *testing.T) {
	manager := bridge.New()
	// Should not panic
	manager.Stop()
	manager.Stop() // Calling twice should also be safe
}

// TestManagerStart verifies that Start handles nil tox parameter.
func TestManagerStartNilTox(t *testing.T) {
	manager := bridge.New()
	err := manager.Start(nil)
	if err == nil {
		t.Error("expected Start(nil) to return an error")
	}
}

// TestConfigCopy verifies that Config() returns a copy of the configuration.
func TestConfigCopy(t *testing.T) {
	config := bridge.Config{
		SOCKSAddr:                   "127.0.0.1:19050",
		EnableSOCKS:                 true,
		EnableBridge:                false,
		BridgeAdvertisementInterval: 30 * time.Second,
	}

	manager := bridge.NewWithConfig(config)
	retrieved := manager.Config()

	// Verify the configuration matches
	if retrieved.SOCKSAddr != config.SOCKSAddr {
		t.Errorf("SOCKSAddr mismatch: %s != %s", retrieved.SOCKSAddr, config.SOCKSAddr)
	}
	if retrieved.EnableSOCKS != config.EnableSOCKS {
		t.Errorf("EnableSOCKS mismatch: %v != %v", retrieved.EnableSOCKS, config.EnableSOCKS)
	}
}

// TestConcurrentStatusChecks verifies thread safety of status methods.
func TestConcurrentStatusChecks(t *testing.T) {
	manager := bridge.New()

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, _, _ = manager.SOCKSStatus()
			_, _ = manager.BridgeStatus()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestStatusEvents verifies that events can be created and used.
func TestStatusEvents(t *testing.T) {
	socksEvent := bridge.SOCKSStatusEvent{
		Status: bridge.StatusAvailable,
		Addr:   "127.0.0.1:19050",
		Error:  "",
	}

	if socksEvent.Status != bridge.StatusAvailable {
		t.Errorf("expected SOCKSStatusEvent Status to be available")
	}
	if socksEvent.Addr != "127.0.0.1:19050" {
		t.Errorf("expected SOCKSStatusEvent Addr to be 127.0.0.1:19050, got %s", socksEvent.Addr)
	}

	bridgeEvent := bridge.BridgeStatusEvent{
		Status: bridge.StatusAvailable,
		Error:  "",
	}

	if bridgeEvent.Status != bridge.StatusAvailable {
		t.Errorf("expected BridgeStatusEvent Status to be available")
	}
}

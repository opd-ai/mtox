package tox

import (
	"testing"
	"time"
)

// TestBridgeStatus tests the BridgeStatus String method.
func TestBridgeStatus_String(t *testing.T) {
	tests := []struct {
		status   BridgeStatus
		expected string
	}{
		{BridgeDisabled, "disabled"},
		{BridgeInitializing, "initializing"},
		{BridgeToxFriendsActive, "tox_friends_active"},
		{BridgeTorFallback, "tor_fallback"},
		{BridgeError, "error"},
	}

	for _, tt := range tests {
		if got := tt.status.String(); got != tt.expected {
			t.Errorf("BridgeStatus.String() = %s, want %s", got, tt.expected)
		}
	}
}

// TestNewBridgeManager tests default bridge manager creation.
func TestNewBridgeManager(t *testing.T) {
	// Create a minimal client for testing
	client := &Client{
		tox: nil,
	}

	bm := NewBridgeManager(client)

	if bm.client != client {
		t.Errorf("NewBridgeManager() did not set client correctly")
	}
	if !bm.enabled {
		t.Errorf("NewBridgeManager() should enable bridge by default")
	}
	if bm.listenAddr != bridgeListenAddr {
		t.Errorf("NewBridgeManager() listenAddr = %s, want %s", bm.listenAddr, bridgeListenAddr)
	}
	if bm.probeInterval != bridgeProbeInterval {
		t.Errorf("NewBridgeManager() probeInterval = %v, want %v", bm.probeInterval, bridgeProbeInterval)
	}
	if bm.status != BridgeDisabled {
		t.Errorf("NewBridgeManager() initial status = %s, want disabled", bm.status)
	}
}

// TestNewBridgeManagerWithConfig_Disabled tests bridge creation with disabled config.
func TestNewBridgeManagerWithConfig_Disabled(t *testing.T) {
	client := &Client{}
	config := &BridgeConfig{Enabled: false}

	bm := NewBridgeManagerWithConfig(client, config)

	if bm.enabled {
		t.Errorf("NewBridgeManagerWithConfig() with Enabled: false should disable bridge")
	}
	if bm.status != BridgeDisabled {
		t.Errorf("NewBridgeManagerWithConfig() with Enabled: false should set status to disabled")
	}
}

// TestNewBridgeManagerWithConfig_CustomListenAddr tests custom listen address.
func TestNewBridgeManagerWithConfig_CustomListenAddr(t *testing.T) {
	client := &Client{}
	customAddr := "127.0.0.1:9999"
	config := &BridgeConfig{
		Enabled:    true,
		ListenAddr: customAddr,
	}

	bm := NewBridgeManagerWithConfig(client, config)

	if bm.listenAddr != customAddr {
		t.Errorf("NewBridgeManagerWithConfig() listenAddr = %s, want %s", bm.listenAddr, customAddr)
	}
}

// TestNewBridgeManagerWithConfig_CustomProbeInterval tests custom probe interval.
func TestNewBridgeManagerWithConfig_CustomProbeInterval(t *testing.T) {
	client := &Client{}
	customInterval := 2 * time.Second
	config := &BridgeConfig{
		Enabled:       true,
		ProbeInterval: customInterval,
	}

	bm := NewBridgeManagerWithConfig(client, config)

	if bm.probeInterval != customInterval {
		t.Errorf("NewBridgeManagerWithConfig() probeInterval = %v, want %v", bm.probeInterval, customInterval)
	}
}

// TestNewBridgeManagerWithConfig_NilConfig tests nil config handling.
func TestNewBridgeManagerWithConfig_NilConfig(t *testing.T) {
	client := &Client{}

	bm := NewBridgeManagerWithConfig(client, nil)

	if !bm.enabled {
		t.Errorf("NewBridgeManagerWithConfig(nil) should enable bridge by default")
	}
	if bm.listenAddr != bridgeListenAddr {
		t.Errorf("NewBridgeManagerWithConfig(nil) should use default listenAddr")
	}
}

// TestBridgeManager_Status tests status getter.
func TestBridgeManager_Status(t *testing.T) {
	client := &Client{}
	bm := NewBridgeManager(client)

	status := bm.Status()
	if status != BridgeDisabled {
		t.Errorf("Status() = %s, want disabled", status)
	}
}

// TestBridgeManager_StatusError tests status error getter.
func TestBridgeManager_StatusError(t *testing.T) {
	client := &Client{}
	config := &BridgeConfig{Enabled: true, ListenAddr: "invalid:invalid"}
	bm := NewBridgeManagerWithConfig(client, config)

	// Manually set an error for testing
	bm.mu.Lock()
	bm.statusError = "test error"
	bm.mu.Unlock()

	err := bm.StatusError()
	if err != "test error" {
		t.Errorf("StatusError() = %s, want 'test error'", err)
	}
}

// TestBridgeManager_IsAvailable tests availability check.
func TestBridgeManager_IsAvailable(t *testing.T) {
	client := &Client{}
	bm := NewBridgeManager(client)

	// Initially should not be available
	if bm.IsAvailable() {
		t.Errorf("IsAvailable() should return false initially")
	}

	// Manually set to available status
	bm.mu.Lock()
	bm.status = BridgeTorFallback
	bm.mu.Unlock()

	if !bm.IsAvailable() {
		t.Errorf("IsAvailable() should return true when status is BridgeTorFallback")
	}

	// Set to Tox friends active
	bm.mu.Lock()
	bm.status = BridgeToxFriendsActive
	bm.mu.Unlock()

	if !bm.IsAvailable() {
		t.Errorf("IsAvailable() should return true when status is BridgeToxFriendsActive")
	}

	// Set to error
	bm.mu.Lock()
	bm.status = BridgeError
	bm.mu.Unlock()

	if bm.IsAvailable() {
		t.Errorf("IsAvailable() should return false when status is BridgeError")
	}
}

// TestBridgeManager_GetActiveToxFriends tests getting active Tox friends.
func TestBridgeManager_GetActiveToxFriends(t *testing.T) {
	client := &Client{}
	bm := NewBridgeManager(client)

	// Initially empty
	friends := bm.GetActiveToxFriends()
	if len(friends) != 0 {
		t.Errorf("GetActiveToxFriends() initially = %v, want empty slice", friends)
	}

	// Manually set some friends
	bm.mu.Lock()
	bm.activeToxFriends = []uint32{1, 2, 3}
	bm.mu.Unlock()

	friends = bm.GetActiveToxFriends()
	if len(friends) != 3 {
		t.Errorf("GetActiveToxFriends() = %d items, want 3", len(friends))
	}
	if friends[0] != 1 || friends[1] != 2 || friends[2] != 3 {
		t.Errorf("GetActiveToxFriends() = %v, want [1 2 3]", friends)
	}

	// Verify it's a copy (modifying result shouldn't affect manager)
	friends[0] = 999
	friends2 := bm.GetActiveToxFriends()
	if friends2[0] != 1 {
		t.Errorf("GetActiveToxFriends() should return a copy, not the internal slice")
	}
}

// TestBridgeManager_GetListenAddr tests getting listen address.
func TestBridgeManager_GetListenAddr(t *testing.T) {
	customAddr := "127.0.0.1:8888"
	client := &Client{}
	config := &BridgeConfig{Enabled: true, ListenAddr: customAddr}
	bm := NewBridgeManagerWithConfig(client, config)

	addr := bm.GetListenAddr()
	if addr != customAddr {
		t.Errorf("GetListenAddr() = %s, want %s", addr, customAddr)
	}
}

// TestBridgeManager_StartIdempotent tests that Start() is idempotent.
func TestBridgeManager_StartIdempotent(t *testing.T) {
	client := &Client{}
	bm := NewBridgeManager(client)

	// Call Start multiple times
	bm.Start()
	bm.Start()
	bm.Start()

	// Give goroutines a chance to run
	time.Sleep(100 * time.Millisecond)

	// Should not panic and should only initialize once
	if bm.Status() == BridgeDisabled {
		t.Errorf("Start() should have triggered initialization")
	}

	bm.Stop()
}

// TestBridgeManager_StartDisabled tests that Start() does nothing when disabled.
func TestBridgeManager_StartDisabled(t *testing.T) {
	client := &Client{}
	config := &BridgeConfig{Enabled: false}
	bm := NewBridgeManagerWithConfig(client, config)

	bm.Start()
	time.Sleep(100 * time.Millisecond)

	if bm.Status() != BridgeDisabled {
		t.Errorf("Start() on disabled bridge should not change status")
	}
}

// TestBridgeManager_Stop tests stopping the bridge.
func TestBridgeManager_Stop(t *testing.T) {
	client := &Client{}
	bm := NewBridgeManager(client)

	// Start the bridge
	bm.Start()
	time.Sleep(50 * time.Millisecond)

	// Stop it
	bm.Stop()

	// Verify stopped state
	if bm.Status() != BridgeDisabled {
		t.Errorf("Stop() should set status to BridgeDisabled")
	}

	// Verify done channel is closed
	if !bm.isStopped() {
		t.Errorf("Stop() should close done channel")
	}
}

// TestBridgeManager_StopIdempotent tests that Stop() is idempotent.
func TestBridgeManager_StopIdempotent(t *testing.T) {
	client := &Client{}
	bm := NewBridgeManager(client)

	bm.Start()
	time.Sleep(50 * time.Millisecond)

	// Call Stop multiple times
	bm.Stop()
	bm.Stop()
	bm.Stop()

	// Should not panic
}

// TestBridgeManager_IsStopped tests the isStopped helper.
func TestBridgeManager_IsStopped(t *testing.T) {
	client := &Client{}
	bm := NewBridgeManager(client)

	if bm.isStopped() {
		t.Errorf("isStopped() should return false initially")
	}

	bm.Stop()

	if !bm.isStopped() {
		t.Errorf("isStopped() should return true after Stop()")
	}
}

// TestBridgeManager_GetAvailableToxFriends tests getting available Tox friends.
func TestBridgeManager_GetAvailableToxFriends(t *testing.T) {
	client := &Client{}
	bm := NewBridgeManager(client)

	// With nil client, should return empty list
	bm.client = nil
	friends := bm.getAvailableToxFriends()
	if len(friends) != 0 {
		t.Errorf("getAvailableToxFriends() with nil client should return empty list")
	}

	// The function attempts to call GetFriends() which requires a valid tox instance.
	// Testing with actual Tox interactions would require integration tests.
	// The nil client case above sufficiently validates the nil check.
}

// TestBridgeManager_ProbeBridges tests the bridge probing logic.
func TestBridgeManager_ProbeBridges(t *testing.T) {
	client := &Client{}
	bm := NewBridgeManager(client)

	// Set client to nil to avoid GetFriends() panic
	bm.client = nil

	// Manually set status to something other than disabled
	bm.mu.Lock()
	bm.status = BridgeTorFallback
	bm.mu.Unlock()

	// Run probe (should remain in TorFallback since no friends available)
	bm.probeBridges()

	if bm.Status() != BridgeTorFallback {
		t.Errorf("probeBridges() with no friends should result in BridgeTorFallback")
	}
}

// TestBridgeConfig_Defaults tests that BridgeConfig uses appropriate defaults.
func TestBridgeConfig_Defaults(t *testing.T) {
	client := &Client{}

	// With empty config
	config := &BridgeConfig{}
	bm := NewBridgeManagerWithConfig(client, config)

	if bm.listenAddr != bridgeListenAddr {
		t.Errorf("Empty config should use default listenAddr")
	}
	if bm.probeInterval != bridgeProbeInterval {
		t.Errorf("Empty config should use default probeInterval")
	}
}

// TestBridgeManager_ConcurrentStatusAccess tests thread-safe status access.
func TestBridgeManager_ConcurrentStatusAccess(t *testing.T) {
	client := &Client{}
	bm := NewBridgeManager(client)

	// Set to a known state
	bm.mu.Lock()
	bm.status = BridgeToxFriendsActive
	bm.mu.Unlock()

	// Multiple goroutines reading status concurrently
	done := make(chan bool, 3)
	for i := 0; i < 3; i++ {
		go func() {
			status := bm.Status()
			if status != BridgeToxFriendsActive {
				t.Errorf("Concurrent read got unexpected status")
			}
			done <- true
		}()
	}

	// Wait for all reads to complete
	for i := 0; i < 3; i++ {
		<-done
	}
}

// TestBridgeManager_ListenerAddressFormat tests that listen address is properly formatted.
func TestBridgeManager_ListenerAddressFormat(t *testing.T) {
	tests := []struct {
		addr string
	}{
		{"127.0.0.1:19050"},
		{"localhost:9999"},
		{"0.0.0.0:8000"},
	}

	for _, tt := range tests {
		client := &Client{}
		config := &BridgeConfig{Enabled: true, ListenAddr: tt.addr}
		bm := NewBridgeManagerWithConfig(client, config)

		if bm.GetListenAddr() != tt.addr {
			t.Errorf("GetListenAddr() = %s, want %s", bm.GetListenAddr(), tt.addr)
		}
	}
}

package tox

import (
	"net"
	"os"
	"testing"
	"time"
)

func TestExtractHost(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"abc123.onion:1234", "abc123.onion"},
		{"xyz789.b32.i2p:5678", "xyz789.b32.i2p"},
		{"localhost:8080", "localhost"},
		{"example.onion", "example.onion"}, // no port
		{"plain.b32.i2p", "plain.b32.i2p"}, // no port
		{"192.168.1.1:443", "192.168.1.1"},
		{"[::1]:8080", "::1"}, // IPv6 with port
	}

	for _, tt := range tests {
		got := extractHost(tt.input)
		if got != tt.want {
			t.Errorf("extractHost(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestIsAnonOnlyMode(t *testing.T) {
	// Save original env value.
	original, hadOriginal := os.LookupEnv("MTOX_ANON_ONLY")
	defer func() {
		if hadOriginal {
			os.Setenv("MTOX_ANON_ONLY", original)
		} else {
			os.Unsetenv("MTOX_ANON_ONLY")
		}
	}()

	os.Unsetenv("MTOX_ANON_ONLY")
	if IsAnonOnlyMode() {
		t.Error("IsAnonOnlyMode() = true when MTOX_ANON_ONLY is unset")
	}

	os.Setenv("MTOX_ANON_ONLY", "0")
	if IsAnonOnlyMode() {
		t.Error("IsAnonOnlyMode() = true when MTOX_ANON_ONLY=0")
	}

	os.Setenv("MTOX_ANON_ONLY", "1")
	if !IsAnonOnlyMode() {
		t.Error("IsAnonOnlyMode() = false when MTOX_ANON_ONLY=1")
	}
}

func TestProfilePath(t *testing.T) {
	path := ProfilePath()
	if path == "" {
		t.Error("ProfilePath() returned empty string")
	}
	// Path should end with profile.tox
	if len(path) < 11 || path[len(path)-11:] != "profile.tox" {
		t.Errorf("ProfilePath() = %q, expected to end with 'profile.tox'", path)
	}
}

func TestProfilePath_CustomPath(t *testing.T) {
	// Save original env value.
	original, hadOriginal := os.LookupEnv("MTOX_PROFILE_PATH")
	defer func() {
		if hadOriginal {
			os.Setenv("MTOX_PROFILE_PATH", original)
		} else {
			os.Unsetenv("MTOX_PROFILE_PATH")
		}
	}()

	customPath := "/tmp/custom/profile.tox"
	os.Setenv("MTOX_PROFILE_PATH", customPath)

	got := ProfilePath()
	if got != customPath {
		t.Errorf("ProfilePath() with MTOX_PROFILE_PATH = %q, want %q", got, customPath)
	}
}

func TestNewAnonymityManager(t *testing.T) {
	events := make(chan ToxEvent, 10)
	mgr := NewAnonymityManager(events)

	if mgr == nil {
		t.Fatal("NewAnonymityManager returned nil")
	}
	if mgr.torStatus != AnonymityUnavailable {
		t.Errorf("initial torStatus = %v, want AnonymityUnavailable", mgr.torStatus)
	}
	if mgr.i2pStatus != AnonymityUnavailable {
		t.Errorf("initial i2pStatus = %v, want AnonymityUnavailable", mgr.i2pStatus)
	}
	if mgr.done == nil {
		t.Error("done channel is nil")
	}
}

func TestAnonymityManager_Stop(t *testing.T) {
	events := make(chan ToxEvent, 10)
	mgr := NewAnonymityManager(events)

	// Stop should be safe to call multiple times.
	mgr.Stop()
	mgr.Stop()
	mgr.Stop()

	// done channel should be closed.
	select {
	case <-mgr.done:
		// expected
	default:
		t.Error("done channel not closed after Stop()")
	}
}

func TestAnonymityManager_StatusGetters(t *testing.T) {
	events := make(chan ToxEvent, 10)
	mgr := NewAnonymityManager(events)

	// Set some values directly for testing.
	mgr.mu.Lock()
	mgr.torStatus = AnonymityAvailable
	mgr.torAddress = "test.onion"
	mgr.torError = ""
	mgr.i2pStatus = AnonymityConnecting
	mgr.i2pAddress = "test.b32.i2p"
	mgr.i2pError = "connecting..."
	mgr.mu.Unlock()

	if s := mgr.TorStatus(); s != AnonymityAvailable {
		t.Errorf("TorStatus() = %v, want AnonymityAvailable", s)
	}
	if s := mgr.I2PStatus(); s != AnonymityConnecting {
		t.Errorf("I2PStatus() = %v, want AnonymityConnecting", s)
	}
	if a := mgr.TorAddress(); a != "test.onion" {
		t.Errorf("TorAddress() = %q, want %q", a, "test.onion")
	}
	if a := mgr.I2PAddress(); a != "test.b32.i2p" {
		t.Errorf("I2PAddress() = %q, want %q", a, "test.b32.i2p")
	}
	if e := mgr.TorError(); e != "" {
		t.Errorf("TorError() = %q, want empty", e)
	}
	if e := mgr.I2PError(); e != "connecting..." {
		t.Errorf("I2PError() = %q, want %q", e, "connecting...")
	}
}

func TestAnonymityManager_EmitSuppression(t *testing.T) {
	events := make(chan ToxEvent, 1)
	mgr := NewAnonymityManager(events)

	// Emit before stop should work.
	mgr.emit(TickEvent{})
	select {
	case <-events:
		// expected
	case <-time.After(100 * time.Millisecond):
		t.Error("event not received before Stop")
	}

	// Stop the manager.
	mgr.Stop()

	// Emit after stop should be suppressed.
	mgr.emit(TickEvent{})
	select {
	case <-events:
		t.Error("event received after Stop - should be suppressed")
	case <-time.After(100 * time.Millisecond):
		// expected - no event should arrive
	}
}

func TestAnonymityManager_EmitChannelFull(t *testing.T) {
	// Create a channel that's already full.
	events := make(chan ToxEvent, 1)
	events <- TickEvent{} // fill the channel

	mgr := NewAnonymityManager(events)

	// This emit should not block (drops the event if channel is full).
	done := make(chan struct{})
	go func() {
		mgr.emit(TickEvent{})
		close(done)
	}()

	select {
	case <-done:
		// emit completed without blocking
	case <-time.After(500 * time.Millisecond):
		t.Error("emit blocked on full channel")
	}
}

func TestAnonymityManager_StartIdempotent(t *testing.T) {
	events := make(chan ToxEvent, 10)
	mgr := NewAnonymityManager(events)
	defer mgr.Stop()

	// Multiple Start() calls should be safe (only first one takes effect).
	mgr.Start()
	mgr.Start()
	mgr.Start()
	// No panic = success
}

func TestAnonymityManager_DisabledViaEnvTor(t *testing.T) {
	original, hadOriginal := os.LookupEnv("MTOX_DISABLE_TOR")
	os.Setenv("MTOX_DISABLE_TOR", "1")
	defer func() {
		if hadOriginal {
			os.Setenv("MTOX_DISABLE_TOR", original)
		} else {
			os.Unsetenv("MTOX_DISABLE_TOR")
		}
	}()

	events := make(chan ToxEvent, 10)
	mgr := NewAnonymityManager(events)

	// Run initTor directly to test disabled path.
	mgr.initTor()

	if mgr.TorStatus() != AnonymityUnavailable {
		t.Errorf("TorStatus() = %v after disable, want AnonymityUnavailable", mgr.TorStatus())
	}
	if mgr.TorError() == "" {
		t.Error("TorError() should indicate disabled")
	}
}

func TestAnonymityManager_DisabledViaEnvI2P(t *testing.T) {
	original, hadOriginal := os.LookupEnv("MTOX_DISABLE_I2P")
	os.Setenv("MTOX_DISABLE_I2P", "1")
	defer func() {
		if hadOriginal {
			os.Setenv("MTOX_DISABLE_I2P", original)
		} else {
			os.Unsetenv("MTOX_DISABLE_I2P")
		}
	}()

	events := make(chan ToxEvent, 10)
	mgr := NewAnonymityManager(events)

	// Run initI2P directly to test disabled path.
	mgr.initI2P()

	if mgr.I2PStatus() != AnonymityUnavailable {
		t.Errorf("I2PStatus() = %v after disable, want AnonymityUnavailable", mgr.I2PStatus())
	}
	if mgr.I2PError() == "" {
		t.Error("I2PError() should indicate disabled")
	}
}

func TestBootstrapNodes(t *testing.T) {
	if len(bootstrapNodes) == 0 {
		t.Fatal("bootstrapNodes is empty")
	}

	for i, node := range bootstrapNodes {
		if node.Address == "" {
			t.Errorf("bootstrapNodes[%d].Address is empty", i)
		}
		if node.Port == 0 {
			t.Errorf("bootstrapNodes[%d].Port is 0", i)
		}
		if len(node.PublicKey) != 64 {
			t.Errorf("bootstrapNodes[%d].PublicKey length = %d, want 64", i, len(node.PublicKey))
		}
	}
}

func TestEventBufSize(t *testing.T) {
	if eventBufSize < 1 {
		t.Errorf("eventBufSize = %d, expected >= 1", eventBufSize)
	}
}

func TestProfileDirAndFile(t *testing.T) {
	if profileDir == "" {
		t.Error("profileDir is empty")
	}
	if profileFile == "" {
		t.Error("profileFile is empty")
	}
}

func TestAnonymityManager_RetryWithBackoff_Cancelled(t *testing.T) {
	events := make(chan ToxEvent, 10)
	mgr := NewAnonymityManager(events)

	// Close done immediately to simulate cancellation.
	close(mgr.done)

	attemptCount := 0
	_, err := mgr.retryWithBackoff(10*time.Millisecond, func() (net.Listener, error) {
		attemptCount++
		return nil, os.ErrNotExist // Simulate failure
	})

	if err == nil || err.Error() != "cancelled" {
		t.Errorf("retryWithBackoff should return 'cancelled' error, got: %v", err)
	}
	// Should not have attempted (or at most one attempt before seeing cancelled).
	if attemptCount > 1 {
		t.Errorf("retryWithBackoff made %d attempts after cancelled", attemptCount)
	}
}

func TestAnonymityManager_RetryWithBackoff_Success(t *testing.T) {
	events := make(chan ToxEvent, 10)
	mgr := NewAnonymityManager(events)
	defer mgr.Stop()

	attemptCount := 0
	mockListener := &mockNetListener{}

	listener, err := mgr.retryWithBackoff(10*time.Millisecond, func() (net.Listener, error) {
		attemptCount++
		if attemptCount < 3 {
			return nil, os.ErrNotExist // Fail first two attempts
		}
		return mockListener, nil // Succeed on third attempt
	})
	if err != nil {
		t.Errorf("retryWithBackoff should succeed, got error: %v", err)
	}
	if listener == nil {
		t.Error("retryWithBackoff should return non-nil listener on success")
	}
	if attemptCount != 3 {
		t.Errorf("retryWithBackoff made %d attempts, expected 3", attemptCount)
	}
}

// mockNetListener is a minimal implementation of net.Listener for testing.
type mockNetListener struct{}

func (m *mockNetListener) Accept() (net.Conn, error) { return nil, nil }
func (m *mockNetListener) Close() error              { return nil }
func (m *mockNetListener) Addr() net.Addr            { return mockAddr{} }

type mockAddr struct{}

func (mockAddr) Network() string { return "mock" }
func (mockAddr) String() string  { return "mock:1234" }

func TestAnonymityManager_RetryWithBackoff_CancelledDuringWait(t *testing.T) {
	events := make(chan ToxEvent, 10)
	mgr := NewAnonymityManager(events)

	// Cancel after a short delay to test cancellation during backoff wait.
	go func() {
		time.Sleep(50 * time.Millisecond)
		mgr.Stop()
	}()

	attemptCount := 0
	_, err := mgr.retryWithBackoff(100*time.Millisecond, func() (net.Listener, error) {
		attemptCount++
		return nil, os.ErrNotExist // Always fail
	})

	if err == nil || err.Error() != "cancelled" {
		t.Errorf("retryWithBackoff should return 'cancelled' error, got: %v", err)
	}
}

func TestAnonymityStatus_AllValues(t *testing.T) {
	tests := []struct {
		status   AnonymityStatus
		expected string
	}{
		{AnonymityUnavailable, "unavailable"},
		{AnonymityConnecting, "connecting"},
		{AnonymityAvailable, "available"},
		{AnonymityError, "error"},
		{AnonymityStatus(100), "unknown"},
		{AnonymityStatus(-1), "unknown"},
	}

	for _, tt := range tests {
		got := tt.status.String()
		if got != tt.expected {
			t.Errorf("AnonymityStatus(%d).String() = %q, want %q", tt.status, got, tt.expected)
		}
	}
}

func TestAnonymityManager_InitialState(t *testing.T) {
	events := make(chan ToxEvent, 10)
	mgr := NewAnonymityManager(events)

	if mgr.TorStatus() != AnonymityUnavailable {
		t.Errorf("initial TorStatus = %v, want AnonymityUnavailable", mgr.TorStatus())
	}
	if mgr.I2PStatus() != AnonymityUnavailable {
		t.Errorf("initial I2PStatus = %v, want AnonymityUnavailable", mgr.I2PStatus())
	}
	if mgr.TorAddress() != "" {
		t.Errorf("initial TorAddress = %q, want empty", mgr.TorAddress())
	}
	if mgr.I2PAddress() != "" {
		t.Errorf("initial I2PAddress = %q, want empty", mgr.I2PAddress())
	}
	if mgr.TorError() != "" {
		t.Errorf("initial TorError = %q, want empty", mgr.TorError())
	}
	if mgr.I2PError() != "" {
		t.Errorf("initial I2PError = %q, want empty", mgr.I2PError())
	}
}

func TestAnonymityManager_StopClosesDone(t *testing.T) {
	events := make(chan ToxEvent, 10)
	mgr := NewAnonymityManager(events)

	// Done should not be closed initially.
	select {
	case <-mgr.done:
		t.Fatal("done channel should not be closed before Stop()")
	default:
	}

	mgr.Stop()

	// Done should be closed after Stop().
	select {
	case <-mgr.done:
		// Expected
	default:
		t.Error("done channel should be closed after Stop()")
	}
}

func TestExtractHost_IPv6(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"[2001:db8::1]:8080", "2001:db8::1"},
		{"[::1]:443", "::1"},
		{"[fe80::1%eth0]:80", "fe80::1%eth0"},
	}

	for _, tt := range tests {
		got := extractHost(tt.input)
		if got != tt.want {
			t.Errorf("extractHost(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestAnonymityManager_EmitAfterStop(t *testing.T) {
	events := make(chan ToxEvent, 10)
	mgr := NewAnonymityManager(events)

	mgr.Stop()

	// Emit after stop should be suppressed and not block.
	done := make(chan struct{})
	go func() {
		mgr.emit(TickEvent{})
		close(done)
	}()

	select {
	case <-done:
		// emit completed without blocking
	case <-time.After(100 * time.Millisecond):
		t.Error("emit blocked after Stop()")
	}

	// No event should be received.
	select {
	case <-events:
		t.Error("event should not be received after Stop()")
	default:
		// Expected
	}
}

func TestAnonymityManager_CalculateNextBackoff(t *testing.T) {
	events := make(chan ToxEvent, 10)
	mgr := NewAnonymityManager(events)
	defer mgr.Stop()

	tests := []struct {
		current time.Duration
		max     time.Duration
		want    time.Duration
	}{
		{100 * time.Millisecond, 1 * time.Second, 150 * time.Millisecond}, // 100 + 50 = 150
		{1 * time.Second, 5 * time.Second, 1500 * time.Millisecond},       // 1000 + 500 = 1500
		{4 * time.Second, 5 * time.Second, 5 * time.Second},               // 4000 + 2000 = 6000, capped at 5000
		{10 * time.Second, 5 * time.Second, 5 * time.Second},              // Already over max
	}

	for _, tt := range tests {
		got := mgr.calculateNextBackoff(tt.current, tt.max)
		if got != tt.want {
			t.Errorf("calculateNextBackoff(%v, %v) = %v, want %v", tt.current, tt.max, got, tt.want)
		}
	}
}

func TestAnonymityManager_IsStopped(t *testing.T) {
	events := make(chan ToxEvent, 10)
	mgr := NewAnonymityManager(events)

	// Initially not stopped
	if mgr.isStopped() {
		t.Error("isStopped() should return false before Stop()")
	}

	mgr.Stop()

	// After stop
	if !mgr.isStopped() {
		t.Error("isStopped() should return true after Stop()")
	}
}

func TestAnonymityManager_WaitWithBackoff_NotCancelled(t *testing.T) {
	events := make(chan ToxEvent, 10)
	mgr := NewAnonymityManager(events)
	defer mgr.Stop()

	// Short backoff should not be cancelled
	cancelled := mgr.waitWithBackoff(10 * time.Millisecond)
	if cancelled {
		t.Error("waitWithBackoff should not be cancelled with short duration")
	}
}

func TestAnonymityManager_WaitWithBackoff_Cancelled(t *testing.T) {
	events := make(chan ToxEvent, 10)
	mgr := NewAnonymityManager(events)

	// Stop immediately
	mgr.Stop()

	// Should be cancelled immediately
	cancelled := mgr.waitWithBackoff(1 * time.Hour)
	if !cancelled {
		t.Error("waitWithBackoff should be cancelled after Stop()")
	}
}

func TestAnonymityManager_StatusMethods(t *testing.T) {
	events := make(chan ToxEvent, 10)
	mgr := NewAnonymityManager(events)
	defer mgr.Stop()

	// Test all status methods with various values
	mgr.mu.Lock()
	mgr.torStatus = AnonymityAvailable
	mgr.torAddress = "example.onion"
	mgr.torError = "no error"
	mgr.i2pStatus = AnonymityConnecting
	mgr.i2pAddress = "example.b32.i2p"
	mgr.i2pError = "connecting"
	mgr.mu.Unlock()

	if mgr.TorStatus() != AnonymityAvailable {
		t.Errorf("TorStatus() = %v, want AnonymityAvailable", mgr.TorStatus())
	}
	if mgr.I2PStatus() != AnonymityConnecting {
		t.Errorf("I2PStatus() = %v, want AnonymityConnecting", mgr.I2PStatus())
	}
	if mgr.TorAddress() != "example.onion" {
		t.Errorf("TorAddress() = %q, want %q", mgr.TorAddress(), "example.onion")
	}
	if mgr.I2PAddress() != "example.b32.i2p" {
		t.Errorf("I2PAddress() = %q, want %q", mgr.I2PAddress(), "example.b32.i2p")
	}
	if mgr.TorError() != "no error" {
		t.Errorf("TorError() = %q, want %q", mgr.TorError(), "no error")
	}
	if mgr.I2PError() != "connecting" {
		t.Errorf("I2PError() = %q, want %q", mgr.I2PError(), "connecting")
	}
}

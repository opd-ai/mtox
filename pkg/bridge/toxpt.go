package bridge

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/opd-ai/toxpt"
)

// bridgeService manages the Tor-over-Tox bridge.
type bridgeService struct {
	mu       sync.RWMutex
	status   Status
	bridge   *toxpt.EmbeddableBridge
	done     chan struct{}
	ctx      context.Context
	cancel   context.CancelFunc
	interval time.Duration
	err      string
}

// newBridgeService creates a new bridge service.
func newBridgeService(interval time.Duration) *bridgeService {
	return &bridgeService{
		status:   StatusUnavailable,
		done:     make(chan struct{}),
		interval: interval,
	}
}

// start begins the bridge service.
// The tox parameter is the toxcore.Tox instance used by the Tox client.
func (b *bridgeService) start(toxInstance interface{}) {
	b.mu.Lock()
	b.status = StatusConnecting
	b.mu.Unlock()

	go b.startBridge(toxInstance)
}

// startBridge initializes the toxpt bridge and begins advertisement.
func (b *bridgeService) startBridge(toxInstance interface{}) {
	// Create a context for the bridge
	b.ctx, b.cancel = context.WithCancel(context.Background())

	// Create bridge configuration with the Tox instance
	cfg := toxpt.Config{}
	// Note: toxpt.Config should be populated with the Tox instance and other settings
	// This requires checking the toxpt package for the correct configuration fields

	bridge, err := toxpt.NewEmbeddableBridge(cfg)
	if err != nil {
		b.mu.Lock()
		b.status = StatusError
		b.err = fmt.Sprintf("bridge init: %v", err)
		b.mu.Unlock()

		log.Printf("mtox: Tor-over-Tox bridge init failed: %v", err)
		return
	}

	// Start the bridge
	if err := bridge.Start(b.ctx); err != nil {
		b.mu.Lock()
		b.status = StatusError
		b.err = fmt.Sprintf("bridge start: %v", err)
		b.mu.Unlock()

		log.Printf("mtox: Tor-over-Tox bridge start failed: %v", err)
		return
	}

	b.mu.Lock()
	b.bridge = bridge
	b.status = StatusAvailable
	b.mu.Unlock()

	log.Printf("mtox: Tor-over-Tox bridge started")

	// Wait until stopped
	<-b.done
}

// getStatus returns the current bridge service status.
func (b *bridgeService) getStatus() (Status, string) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.status, b.err
}

// stop shuts down the bridge service.
func (b *bridgeService) stop() {
	select {
	case <-b.done:
		return
	default:
		close(b.done)
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cancel != nil {
		b.cancel()
	}
	if b.bridge != nil {
		b.bridge = nil
	}

	b.status = StatusUnavailable
}

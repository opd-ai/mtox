package bridge

import (
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// socksService manages the SOCKS proxy server.
// It provides a local SOCKS5 proxy at the configured address for Tor access.
type socksService struct {
	mu       sync.RWMutex
	addr     string
	status   Status
	listener net.Listener
	done     chan struct{}
	err      string
}

// newSOCKSService creates a new SOCKS service.
func newSOCKSService(addr string) *socksService {
	return &socksService{
		addr:   addr,
		status: StatusUnavailable,
		done:   make(chan struct{}),
	}
}

// start begins the SOCKS proxy server with retries.
func (s *socksService) start() {
	s.mu.Lock()
	s.status = StatusConnecting
	s.mu.Unlock()

	go s.startWithRetry()
}

// startWithRetry attempts to start the SOCKS server with exponential backoff.
func (s *socksService) startWithRetry() {
	backoff := 500 * time.Millisecond
	maxBackoff := 1 * time.Minute

	for {
		select {
		case <-s.done:
			return
		default:
		}

		// Attempt to listen on the configured address
		listener, err := net.Listen("tcp", s.addr)
		if err == nil {
			s.mu.Lock()
			prevErr := s.err
			s.listener = listener
			s.status = StatusAvailable
			s.err = ""
			s.mu.Unlock()

			if prevErr != "" {
				log.Printf("mtox: SOCKS proxy recovered after error: %s", prevErr)
			}
			log.Printf("mtox: SOCKS proxy started at %s", s.addr)

			// Accept connections and handle them
			go s.acceptConnections(listener)
			return
		}

		s.mu.Lock()
		s.err = err.Error()
		s.mu.Unlock()

		select {
		case <-s.done:
			return
		case <-time.After(backoff):
			// Use 1.5x growth to keep retries smoother than 2x backoff
			// while still avoiding rapid retry storms.
			backoff = (backoff * 3) / 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}
}

// acceptConnections accepts incoming connections and handles them.
// For now, this is a placeholder that can be integrated with go-tor's client.
func (s *socksService) acceptConnections(listener net.Listener) {
	for {
		select {
		case <-s.done:
			return
		default:
		}

		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-s.done:
				return
			default:
			}
			continue
		}

		// Handle connection in background
		go s.handleConnection(conn)
	}
}

// handleConnection handles a single client connection.
// This is a placeholder for SOCKS5 protocol handling integrated with go-tor.
// TODO: Implement SOCKS5 protocol negotiation and route connections through go-tor's client.
// This will enable local applications to tunnel traffic through the Tor network via this SOCKS proxy.
func (s *socksService) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Minimal SOCKS5 negotiation: explicitly reject all methods so clients fail fast
	// instead of hanging on a half-open TCP connection.
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))

	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil {
		return
	}
	if header[0] != 0x05 {
		return
	}

	methodCount := int(header[1])
	if methodCount <= 0 {
		return
	}

	methods := make([]byte, methodCount)
	if _, err := io.ReadFull(conn, methods); err != nil {
		return
	}

	// 0xFF means "no acceptable authentication methods".
	_, _ = conn.Write([]byte{0x05, 0xFF})
}

// getStatus returns the current SOCKS service status.
func (s *socksService) getStatus() (Status, string, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	addr := ""
	if s.listener != nil {
		addr = s.listener.Addr().String()
	}
	return s.status, addr, s.err
}

// stop shuts down the SOCKS service.
func (s *socksService) stop() {
	select {
	case <-s.done:
		return
	default:
		close(s.done)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener != nil {
		s.listener.Close()
		s.listener = nil
	}

	s.status = StatusUnavailable
}

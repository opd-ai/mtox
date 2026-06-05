package bridge

// Status represents the operational state of a bridge service.
type Status int

const (
	// StatusUnavailable means the service is not running.
	StatusUnavailable Status = iota
	// StatusConnecting means the service is being established.
	StatusConnecting
	// StatusAvailable means the service is operational.
	StatusAvailable
	// StatusError means an error occurred.
	StatusError
)

// String returns a human-readable status string.
func (s Status) String() string {
	switch s {
	case StatusUnavailable:
		return "unavailable"
	case StatusConnecting:
		return "connecting"
	case StatusAvailable:
		return "available"
	case StatusError:
		return "error"
	default:
		return "unknown"
	}
}

// SOCKSStatusEvent describes SOCKS proxy status for bridge-level consumers.
type SOCKSStatusEvent struct {
	Status Status
	Addr   string // SOCKS server address if available (e.g., "127.0.0.1:19050")
	Error  string // Error message if status is Error
}

// BridgeStatusEvent describes Tor-over-Tox bridge status for bridge-level consumers.
type BridgeStatusEvent struct {
	Status Status
	Error  string // Error message if status is Error
}

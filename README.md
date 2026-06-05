# mtox

[![CI](https://github.com/opd-ai/mtox/actions/workflows/ci.yml/badge.svg)](https://github.com/opd-ai/mtox/actions/workflows/ci.yml)
[![Release](https://github.com/opd-ai/mtox/actions/workflows/release.yml/badge.svg)](https://github.com/opd-ai/mtox/actions/workflows/release.yml)

**mtox** is a full-featured Tox Messenger terminal user interface (TUI) written in Go.

It uses [`github.com/opd-ai/toxcore`](https://github.com/opd-ai/toxcore) as the networking backend and [`github.com/charmbracelet/bubbletea`](https://github.com/charmbracelet/bubbletea) as the TUI framework.

## UI

```
┌──────────────────┬───────────────────────────────┐
│  Contacts        │  Chat with: Alice              │
│  ──────────      │  ─────────────────────────     │
│ ● Alice (online) │  [10:30] Alice: Hey!           │
│ ○ Bob (offline)  │  [10:31] You: Hi there         │
│ ◌ Charlie (away) │  [10:32] Alice is typing...    │
│                  │                                 │
│                  │                                 │
│                  ├─────────────────────────────────│
│                  │ > type a message...             │
├──────────────────┴─────────────────────────────────┤
│ 🟢 Connected (UDP) 🧅Tor 🧄I2P │ My ID:... │ mtox v0.1 │
└────────────────────────────────────────────────────┘
```

## Features

- Two-pane TUI: contact list on the left, chat on the right
- Real-time friend status indicators (online / offline)
- Scrollable per-friend chat history with timestamps
- Typing indicators
- Incoming friend request dialog (accept / reject)
- Add friend dialog (`Ctrl+N`)
- Profile persistence at `~/.config/mtox/profile.tox`
- Mouse support (click to select contacts / focus chat)
- Graceful shutdown with auto-save
- **Automatic Tor and I2P support** when services are available

## Anonymity Network Support

mtox automatically enables **Tor** and **I2P** support when the respective services are detected. **Both networks can be used simultaneously**, providing maximum connectivity options for privacy-conscious users.

| Network | Detection | Status Indicator | Address Format | Transport |
|---------|-----------|------------------|----------------|-----------|
| **Tor** | Tor daemon on port 9051 | 🧅Tor | `*.onion` | TCP (stream) |
| **I2P** | SAM bridge on port 7656 | 🧄I2P | `*.b32.i2p` | TCP + UDP (datagrams) |

### Simultaneous Network Support

When both Tor and I2P services are available, mtox initializes both transports in parallel:
- **Tor** provides TCP-based hidden service connectivity
- **I2P** provides both stream (TCP-like) and datagram (UDP-like) connectivity via SAM

This allows you to communicate with contacts on either network while maintaining anonymity.

### Enabling Tor Support

1. Install Tor: `apt install tor` or `brew install tor`
2. Start the Tor service: `systemctl start tor` or `brew services start tor`
3. Launch mtox - Tor will be automatically detected

### Enabling I2P Support

1. Install I2P: `apt install i2pd` or follow [geti2p.net](https://geti2p.net/en/download)
2. Enable SAM in I2P router configuration (usually at http://127.0.0.1:7657/configclients)
3. Launch mtox - I2P will be automatically detected

### Environment Variables

| Variable | Description |
|----------|-------------|
| `MTOX_DISABLE_TOR=1` | Disable Tor even if service is available |
| `MTOX_DISABLE_I2P=1` | Disable I2P even if service is available |
| `MTOX_ANON_ONLY=1` | Anon-only mode: Tor + I2P + I2P datagrams, no clearnet |

### Anon-Only Mode

When `MTOX_ANON_ONLY=1` is set, mtox disables clearnet UDP/IPv6/local discovery and enables both Tor and I2P transports. This reduces clearnet exposure but **does not guarantee all traffic goes through anonymity networks** - toxcore may still make some clearnet TCP connections for DHT bootstrapping.

```bash
# Run mtox in anon-only mode
MTOX_ANON_ONLY=1 ./mtox
```

**What anon-only mode does:**
- ✅ Enables Tor hidden services (TCP)
- ✅ Enables I2P destinations (TCP)  
- ✅ Enables I2P datagrams (UDP-like)
- ❌ Disables clearnet UDP
- ❌ Disables IPv6
- ❌ Disables local discovery

**Note:** For complete anonymity guarantees, consider running mtox inside a network namespace or VM that blocks all non-Tor/I2P traffic.

## Downloads

Pre-built binaries are available from the [GitHub Releases](https://github.com/opd-ai/mtox/releases) page:

| Platform | Architecture | Download |
|----------|--------------|----------|
| Linux    | amd64        | [mtox-linux-amd64](https://github.com/opd-ai/mtox/releases/latest/download/mtox-linux-amd64) |
| Linux    | arm64        | [mtox-linux-arm64](https://github.com/opd-ai/mtox/releases/latest/download/mtox-linux-arm64) |
| macOS    | amd64        | [mtox-darwin-amd64](https://github.com/opd-ai/mtox/releases/latest/download/mtox-darwin-amd64) |
| macOS    | arm64        | [mtox-darwin-arm64](https://github.com/opd-ai/mtox/releases/latest/download/mtox-darwin-arm64) |
| Windows  | amd64        | [mtox-windows-amd64.exe](https://github.com/opd-ai/mtox/releases/latest/download/mtox-windows-amd64.exe) |
| Windows  | arm64        | [mtox-windows-arm64.exe](https://github.com/opd-ai/mtox/releases/latest/download/mtox-windows-arm64.exe) |

## Build

```bash
go build ./cmd/mtox
```

## Bridge Integration for Go Tox Clients

The mtox package includes a **Tor-over-Tox bridge module** (`internal/tox/BridgeManager`) that enables Go Tox clients to automatically leverage Tor-over-Tox routing with minimal integration effort.

### Bridge Features

- **SOCKS Proxy (stub)**: Listens on `127.0.0.1:19050` (SOCKS5 handler/routing not implemented yet)
- **Failover State Tracking**: Periodically probes friend availability and updates status (routing integration TODO)
- **Zero Configuration**: Enabled by default with sensible defaults; optional disable flag for privacy-conscious clients
- **Bridge Status Monitoring**: Simple query interface for bridge health and routing mode
- **Thread-Safe**: All operations are safe for concurrent access
- **Minimal Integration**: ~100 lines of code in typical client

### Quick Start

Initialize the bridge in your Tox client startup code:

```go
import "github.com/opd-ai/mtox/internal/tox"

func main() {
    // Create and start Tox client
    client, err := tox.NewClient()
    if err != nil {
        log.Fatal(err)
    }
    defer client.Stop()
    client.Start()
    client.Bootstrap()

    // Initialize bridge (enabled by default)
    bridge := tox.NewBridgeManager(client)
    bridge.Start()
    defer bridge.Stop()

    // Bridge is now listening on 127.0.0.1:19050
    // Configure your Tox-enabled applications to use this SOCKS proxy
}
```

### Bridge Status Modes

The bridge transitions between routing modes automatically:

| Status | Description |
|--------|-------------|
| `BridgeDisabled` | Bridge is not active |
| `BridgeInitializing` | Bridge is starting up |
| `BridgeToxFriendsActive` | Routing through available Tox friend bridges |
| `BridgeTorFallback` | Routing directly through Tor (no friends available) |
| `BridgeError` | An error occurred during initialization |

### Failover Logic

The bridge implements automatic, intelligent failover:

1. Starts in `BridgeTorFallback` mode (direct Tor routing)
2. Every 5 seconds, probes for available Tox friends
3. If friends are online → switches to `BridgeToxFriendsActive`
4. If all friends go offline → falls back to `BridgeTorFallback`
5. Traffic always routes through the best available path

### Configuration

Customize the bridge behavior:

```go
config := &tox.BridgeConfig{
    Enabled:       true,              // Enable/disable bridge
    ListenAddr:    "127.0.0.1:19050", // SOCKS proxy listen address
    ProbeInterval: 5 * time.Second,   // How often to check friend availability
}
bridge := tox.NewBridgeManagerWithConfig(client, config)
bridge.Start()
```

### Monitoring Bridge Status

```go
// Poll bridge status periodically
if bridge.IsAvailable() {
    status := bridge.Status()
    friends := bridge.GetActiveToxFriends()
    log.Printf("Bridge OK: %s with %d Tox friends", status, len(friends))
} else {
    log.Printf("Bridge Error: %s", bridge.StatusError())
}
```

### Disabling the Bridge

To disable the bridge entirely:

```go
config := &tox.BridgeConfig{Enabled: false}
bridge := tox.NewBridgeManagerWithConfig(client, config)
bridge.Start() // This is a no-op when disabled
```

### Integration Examples

See [examples/bridge_integration.go](examples/bridge_integration.go) for detailed integration patterns and advanced usage.

## Embedding

`mtox` now exposes an embeddable runtime at `github.com/opd-ai/mtox/pkg/embedded`.

Host applications can construct and run the reusable TUI via `embedded.New(...)` and `(*embedded.TUI).Run()`.

## Run

```bash
./mtox
```

On first launch a new Tox identity is generated and saved to `~/.config/mtox/profile.tox`. Subsequent launches reuse the same identity.

## Keyboard Shortcuts

| Key            | Action                                |
|----------------|---------------------------------------|
| `Tab`          | Switch focus between contacts / chat  |
| `↑` / `↓` / `j` / `k` | Navigate contacts list        |
| `Enter`        | Select contact / send message         |
| `Ctrl+N`       | Add friend (opens dialog)             |
| `Ctrl+S`       | Save profile                          |
| `Ctrl+C` / `Ctrl+Q` | Quit (auto-saves)                |
| `Esc`          | Cancel current dialog                 |
| `R`            | Reject a friend request               |

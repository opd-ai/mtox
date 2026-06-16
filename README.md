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
| `MTOX_STRICT_SECURITY=1` | Enable strict security mode (reject legacy clients without full post-quantum capabilities) |

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

### Legacy Client Support

By default, mtox enables **backward compatibility with legacy Tox clients and bootstrap servers**. This means mtox can gracefully downgrade its security capabilities to communicate with:
- **Legacy vanilla Tox clients** (without post-quantum cryptography)
- **Classical-only Tox nodes** (supporting only X3DH and header encryption)
- **Older Tox bootstrap servers**

The transport layer performs an automatic negotiation to use the intersection of security capabilities between both peers, ensuring maximum interoperability. This provides the best available security for each peer: full post-quantum protection with compatible peers, or classical encryption with legacy peers.

**Note:** This is the recommended default setting for maximum compatibility. If you want to enforce strict post-quantum security only, rejecting all legacy peers, set `MTOX_STRICT_SECURITY=1`:

```bash
# Enable strict security (reject legacy clients)
MTOX_STRICT_SECURITY=1 ./mtox
```

When strict security is enabled:
- ❌ Connections to legacy clients without full post-quantum support will fail
- ✅ All accepted connections use X3DH + header encryption + PQXDH (post-quantum)

## Tor Services Integration

mtox includes an integrated layer for dual Tor services powered by [opd-ai/go-tor](https://github.com/opd-ai/go-tor) and [opd-ai/toxpt](https://github.com/opd-ai/toxpt).
This integration is currently a scaffold (the service framework exists, but full traffic routing is not implemented yet).

### SOCKS Proxy (Always-On)
- **Address**: `127.0.0.1:19050` (default, configurable)
- **Purpose**: Local SOCKS endpoint reserved for upcoming Tor proxy integration
- **Enabled By Default**: Yes
- **Control**: `MTOX_ENABLE_SOCKS=0` to disable

Current behavior: connections are accepted and immediately closed with SOCKS5 "no acceptable methods" (`0xFF`) to prevent client hangs.

### Tor-over-Tox Bridge (Optional)
- **Purpose**: Route Tor traffic through Tox connections to friends
- **Access**: Only available to connected Tox friends, not exposed externally
- **Enabled By Default**: No (disabled for privacy)
- **Control**: `MTOX_ENABLE_BRIDGE=1` to enable

When enabled, mtox initializes bridge components, but full Tor-over-Tox traffic routing is not yet complete.

### Configuration

Both services are initialized automatically when the Tox client starts. Control them via environment variables:

```bash
# Custom SOCKS address
MTOX_SOCKS_ADDR=127.0.0.1:9999 ./mtox

# Disable SOCKS proxy (keep bridge ready)
MTOX_ENABLE_SOCKS=0 ./mtox

# Enable Tor-over-Tox bridge (for Tox friends)
MTOX_ENABLE_BRIDGE=1 ./mtox

# Both settings together
MTOX_ENABLE_SOCKS=1 MTOX_ENABLE_BRIDGE=1 ./mtox
```

### Programmatic Integration

The full client example below uses `internal/tox`, which is only importable from within this module.
External modules should use the public `pkg/bridge` package directly.

Within this repository, Tox clients can activate both services with a single call:

```go
import "github.com/opd-ai/mtox/internal/tox"
import "github.com/opd-ai/mtox/pkg/bridge"

// Create Tox client (includes bridge manager initialization)
client, err := tox.NewClient()
if err != nil {
    log.Fatal(err)
}

// Start both SOCKS proxy and bridge with one call
client.Start()

// Query service status
mgr := client.BridgeManager()
socksStatus, socksAddr, socksErr := mgr.SOCKSStatus()
bridgeStatus, bridgeErr := mgr.BridgeStatus()

// Graceful shutdown (both services stop cleanly)
client.Stop()
```

### Service Independence

Both services operate independently:
- **SOCKS proxy** failure does not affect the Tor-over-Tox bridge
- **Bridge** failure does not affect local SOCKS proxy
- Each service has its own retry and error handling logic
- Failure of one service is logged but doesn't crash the Tox client

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

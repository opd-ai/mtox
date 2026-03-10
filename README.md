# mtox

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
│ 🟢 Connected (UDP) 🧅Tor 🧄I2P │ My ID:... │ v0.1 │
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

mtox automatically enables **Tor** and **I2P** support when the respective services are detected:

| Network | Detection | Status Indicator | Address Format |
|---------|-----------|------------------|----------------|
| **Tor** | Tor daemon on port 9051 | 🧅Tor | `*.onion` |
| **I2P** | SAM bridge on port 7656 | 🧄I2P | `*.b32.i2p` |

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
| `TOR_CONTROL_ADDR` | Custom Tor control address (default: `127.0.0.1:9051`) |
| `I2P_SAM_ADDR` | Custom I2P SAM address (default: `127.0.0.1:7656`) |

## Build

```bash
go build ./cmd/mtox
```

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

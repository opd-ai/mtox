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
│ 🟢 Connected (UDP) │ My ID: ABC123... │ mtox v0.1  │
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

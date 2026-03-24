# Implementation Gaps — 2026-03-24

This document details the gaps between mtox's stated goals (from README.md and documentation) and its current implementation.

---

## Per-Friend Chat History Persistence

- **Stated Goal**: "Scrollable per-friend chat history with timestamps" — README.md implies chat history is maintained per friend.
- **Current State**: Chat history is stored in `chatPanel.history` (`internal/tui/chat.go:26`) but is **cleared on every contact switch** at line 58: `c.history = nil`. Messages exchanged with a contact are lost when switching to another contact and returning.
- **Impact**: **Severe usability impact**. Users cannot review previous conversations, making mtox impractical for actual messaging use cases. This is the single largest gap between stated functionality and user expectations.
- **Closing the Gap**:
  1. Add `historyByFriend map[uint32][]chatMessage` field to the `App` struct in `internal/tui/app.go`
  2. Modify `selectFriend()` (line 318) to save `a.chat.history` to the map before switching
  3. Modify `chatPanel.setFriend()` (line 52) to load history from the map instead of clearing
  4. Consider optional disk persistence to `~/.config/mtox/history/<friendID>.json` for cross-session history
  5. **Validation**: Send messages to multiple contacts, switch between them repeatedly, verify all messages remain visible

---

## Test Coverage

- **Stated Goal**: Implied by the existence of a ROADMAP.md and active development — maintainable, production-quality code requires tests.
- **Current State**: **0% test coverage**. Running `go test ./...` returns "no test files" for all three packages (`cmd/mtox`, `internal/tox`, `internal/tui`).
- **Impact**: **High risk for maintenance**. Without tests, refactoring is dangerous, regressions are undetectable, and contributors cannot verify their changes don't break existing functionality.
- **Closing the Gap**:
  1. Create `internal/tox/client_test.go`:
     - Test `NewClient()` initializes correctly
     - Test event emission via `emit()` reaches the events channel
     - Test callback registration wires up correctly
     - Mock toxcore.Tox interface for unit testing
  2. Create `internal/tox/anonymity_test.go`:
     - Test status transitions (Unavailable → Connecting → Available)
     - Test retry backoff logic timing
     - Test environment variable handling (MTOX_DISABLE_TOR, etc.)
  3. Create `internal/tui/app_test.go`:
     - Test `Update()` handles all ToxEvent types correctly
     - Test modal open/close state transitions
     - Test keyboard shortcuts trigger correct actions
  4. Create `internal/tui/contacts_test.go`:
     - Test sorting algorithm stability
     - Test unread count increment/clear
     - Test connection status updates
  5. **Validation**: `go test -cover ./...` should show >60% coverage on core logic

---

## Group Chat (Conference) Support

- **Stated Goal**: The README shows keyboard shortcut `Ctrl+G` but does not explicitly claim group chat functionality.
- **Current State**: `internal/tui/app.go:229-231` handles Ctrl+G by displaying "Group chat not yet supported." The Tox protocol supports conferences (group chats), and the underlying toxcore library likely has the API, but mtox does not implement it.
- **Impact**: **Medium limitation**. Users expecting Tox feature parity with clients like qTox will find this missing. Group communication is important for many use cases.
- **Closing the Gap**:
  1. Add conference event types to `internal/tox/types.go`:
     - `ConferenceInviteEvent`
     - `ConferenceMessageEvent`
     - `ConferencePeerJoinEvent`
     - `ConferencePeerLeaveEvent`
  2. Register conference callbacks in `client.go:registerCallbacks()`
  3. Add wrapper methods to Client:
     - `ConferenceNew() (uint32, error)`
     - `ConferenceInvite(conferenceID uint32, friendID uint32) error`
     - `ConferenceSendMessage(conferenceID uint32, message string) error`
  4. Update contacts panel to show conferences (separate section or mixed list)
  5. Implement Ctrl+G to create new conference with type selection dialog
  6. Handle incoming conference invitations in modal
  7. **Validation**: Create conference, invite a friend, exchange messages bidirectionally

---

## File Transfer Support

- **Stated Goal**: Not explicitly claimed in README.md.
- **Current State**: No file transfer implementation exists. The Tox protocol supports `FileSend()`, `FileControl()`, and related APIs for secure peer-to-peer file transfer.
- **Impact**: **Medium limitation**. File sharing is a common messaging feature. Privacy-conscious users may prefer Tox's encrypted P2P file transfer over third-party services.
- **Closing the Gap**:
  1. Add file transfer event types to `internal/tox/types.go`:
     - `FileRecvRequestEvent` (incoming file offer)
     - `FileRecvChunkEvent` (data chunk received)
     - `FileControlEvent` (accept/cancel/pause)
  2. Register file callbacks in `client.go:registerCallbacks()`
  3. Add wrapper methods to Client:
     - `FileSend(friendID uint32, filename string, data []byte) error`
     - `FileControl(friendID uint32, fileID uint32, control FileControlType) error`
  4. Implement `/file <path>` command in chat input
  5. Create file transfer progress indicator in status bar or chat
  6. Handle incoming files with accept/reject modal and save location prompt
  7. **Validation**: Send file to friend, receive file from friend, verify file integrity

---

## CI/CD Pipeline

- **Stated Goal**: Not explicitly stated, but expected for maintainable open-source projects.
- **Current State**: No `.github/workflows/`, Makefile, or CI configuration exists. Build and test verification is entirely manual.
- **Impact**: **Medium risk for contributions**. Pull requests cannot be automatically validated. Build breakages may go unnoticed.
- **Closing the Gap**:
  1. Create `.github/workflows/ci.yml`:
     ```yaml
     name: CI
     on: [push, pull_request]
     jobs:
       build:
         runs-on: ubuntu-latest
         steps:
           - uses: actions/checkout@v4
           - uses: actions/setup-go@v5
             with:
               go-version: '1.24'
           - run: go build ./cmd/mtox
           - run: go test -race ./...
           - run: go vet ./...
     ```
  2. Add build status badge to README.md
  3. Consider adding release workflow for binary artifacts
  4. **Validation**: Push to a branch, verify GitHub Actions runs successfully

---

## Voice/Video Call Support

- **Stated Goal**: Not claimed in README.md.
- **Current State**: No AV-related code exists. The Tox protocol supports `tox_av_*` APIs for real-time audio/video.
- **Impact**: **Low for TUI context**. Voice/video calls are less common in terminal applications. Audio-only might be feasible.
- **Closing the Gap**:
  1. Research opd-ai/toxcore AV API availability
  2. Implement audio-only call support first (more TUI-appropriate)
  3. Display call status in status bar
  4. Handle incoming calls with accept/reject modal
  5. Implement call audio routing (microphone → Tox → speaker)
  6. **Validation**: Initiate call, confirm audio flows bidirectionally

---

## Typing Notification Optimization

- **Stated Goal**: "Typing indicators" — README.md claims this feature.
- **Current State**: Feature works, but `SetTyping()` is called on **every keystroke** (`internal/tui/app.go:273-275`) rather than on state transitions. This generates unnecessary network traffic.
- **Impact**: **Minor inefficiency**. Does not affect functionality but wastes bandwidth and may cause performance issues on slow connections.
- **Closing the Gap**:
  1. Add `lastTypingState bool` field to `App` struct
  2. Modify `handleKey()` at line 273 to track state:
     ```go
     isTyping := len(a.chat.input.Value()) > 0
     if isTyping != a.lastTypingState {
         _ = a.client.SetTyping(a.activeFriendID, isTyping)
         a.lastTypingState = isTyping
     }
     ```
  3. **Validation**: Add debug logging, confirm SetTyping only called on transitions

---

## Command-Line Options

- **Stated Goal**: Not explicitly stated.
- **Current State**: The application has no `--help`, `-h`, or other command-line options. Running `./mtox --help` starts the application instead of showing usage information.
- **Impact**: **Low usability gap**. Users cannot discover options like `MTOX_ANON_ONLY` without reading documentation.
- **Closing the Gap**:
  1. Add `flag` package to `cmd/mtox/main.go`
  2. Implement `-h/--help` flag to print usage
  3. Consider adding flags for:
     - `--anon-only` (equivalent to MTOX_ANON_ONLY=1)
     - `--no-tor` (equivalent to MTOX_DISABLE_TOR=1)
     - `--no-i2p` (equivalent to MTOX_DISABLE_I2P=1)
     - `--profile <path>` (custom profile location)
  4. **Validation**: Run `./mtox --help` and confirm usage is displayed

---

## Summary

| Gap | Severity | User Impact | Effort |
|-----|----------|-------------|--------|
| Per-friend chat history | HIGH | Severe — messages lost | Medium |
| Test coverage | HIGH | Maintenance risk | High |
| Group chat | MEDIUM | Feature limitation | High |
| File transfer | MEDIUM | Feature limitation | High |
| CI/CD pipeline | MEDIUM | Contribution friction | Low |
| Typing optimization | LOW | Minor inefficiency | Low |
| Voice/video calls | LOW | Feature limitation | Very High |
| Command-line options | LOW | Discoverability | Low |

**Recommended priority order**: Chat history → Test coverage → CI/CD → Typing optimization → Group chat → File transfer → Command-line options → Voice/video

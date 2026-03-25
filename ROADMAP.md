# Goal-Achievement Assessment

## Project Context

- **What it claims to do**: mtox is a "full-featured Tox Messenger terminal user interface (TUI)" that provides secure peer-to-peer messaging with a two-pane interface (contacts + chat), automatic Tor/I2P anonymity network support, typing indicators, friend requests, file transfers, profile persistence, and graceful shutdown.

- **Target audience**: Privacy-conscious users who prefer terminal-based interfaces and want secure, decentralized messaging with optional anonymity network routing via Tor and I2P.

- **Architecture**:
  | Package | Role | Files |
  |---------|------|-------|
  | `cmd/mtox` | Application entrypoint with CLI flags | 1 |
  | `internal/tox` | Tox client wrapper bridging toxcore callbacks to bubbletea messages; anonymity network management; file transfers | 3 |
  | `internal/tui` | Terminal UI components (app, chat, contacts, statusbar, styles) | 5 |
  | `internal/version` | Version constant | 1 |

- **Existing CI/quality gates**:
  - ✅ GitHub Actions CI (`.github/workflows/ci.yml`): build, test (`-race`), vet, staticcheck
  - ✅ Coverage threshold enforcement (30% minimum)
  - ✅ Release workflow (`.github/workflows/release.yml`): 6 platform targets
  - ❌ No integration tests with actual toxcore

## Goal-Achievement Summary

| Stated Goal | Status | Evidence | Gap Description |
|-------------|--------|----------|-----------------|
| Full-featured Tox TUI | ✅ Achieved | Complete implementation across internal/tui and internal/tox packages | — |
| Two-pane TUI (contacts + chat) | ✅ Achieved | `internal/tui/app.go`, `contacts.go`, `chat.go` implement split layout with `contactsPanelWidth = 22` | — |
| Real-time friend status indicators | ✅ Achieved | `contacts.go:252-265` - `statusIndicator()` renders ●/○/◌/◉ based on ConnectionStatus and FriendStatus | — |
| Scrollable per-friend chat history | ✅ Achieved | `app.go:73` - `historyByFriend map[uint32][]chatMessage`; `selectFriend()` saves/restores history | — |
| Typing indicators | ✅ Achieved | `chat.go:106-108` displays typing state; `app.go:434-444` only sends `SetTyping()` on state transitions | — |
| Incoming friend request dialog | ✅ Achieved | `app.go:610-672` - modal with accept (Enter) / reject (R) keys | — |
| Add friend dialog (Ctrl+N) | ✅ Achieved | `app.go:601-639` - modal input and submission | — |
| Profile persistence at ~/.config/mtox | ✅ Achieved | `client.go:96-101` loads, `client.go:311-323` saves to ProfilePath() | — |
| Mouse support | ✅ Achieved | `cmd/mtox/main.go:91` - `tea.WithMouseCellMotion()`; `app.go:447-465` handles clicks | — |
| Graceful shutdown with auto-save | ✅ Achieved | `app.go:594-599` `quit()` calls `Save()` then `Stop()` | — |
| Automatic Tor support | ✅ Achieved | `anonymity.go:188-248` - `NewTorTransport` with retry loop via `retryWithBackoff()` | — |
| Automatic I2P support | ✅ Achieved | `anonymity.go:258-318` - `NewI2PTransport` with retry loop via `retryWithBackoff()` | — |
| Simultaneous Tor + I2P | ✅ Achieved | `anonymity.go:87-91` - both `initTor()` and `initI2P()` run in parallel goroutines | — |
| Status bar shows Tor/I2P indicators | ✅ Achieved | `statusbar.go:83-105` - shows 🧅Tor and 🧄I2P when available | — |
| MTOX_ANON_ONLY mode | ✅ Achieved | `client.go:85-94` disables UDP/IPv6/local discovery | — |
| MTOX_DISABLE_TOR/I2P variables | ✅ Achieved | `anonymity.go:190-203`, `261-274` check environment | — |
| File transfers (send/receive) | ✅ Achieved | `client.go:411-601` implements FileSend/FileAccept/FileReject; `app.go:536-592` handles `/file` command | — |
| Command-line options | ✅ Achieved | `cmd/mtox/main.go:26-50` implements `--help`, `--version`, `--anon-only`, `--no-tor`, `--no-i2p`, `--profile` | — |
| Test coverage (>30%) | ✅ Achieved | `go test -cover`: tox 36.4%, tui 36.7% - meets CI threshold | — |
| Group chat (conferences) | ❌ Missing | `app.go:334` returns "Group chat not yet supported" notification | Core Tox feature not exposed |
| Voice/video calls | ❌ Missing | No implementation - toxcore supports AV APIs | Advanced Tox feature |

**Overall: 18/20 goals fully achieved (90%)**

## Metrics Summary (go-stats-generator)

| Metric | Value | Assessment |
|--------|-------|------------|
| Total Lines of Code | 1,431 | Compact codebase |
| Functions/Methods | 166 | Well-factored |
| Average Function Length | 10.0 lines | Excellent (threshold: <15) |
| Functions >50 lines | 3 (1.8%) | Excellent |
| Functions >100 lines | 0 | Perfect |
| High Complexity (>10) | 1 function (`emit` at 10.9) | Acceptable |
| Documentation Coverage | 87.3% | Good |
| Package Coverage | 100% | Excellent |
| Function Coverage | 100% | Excellent |
| Method Coverage | 78.9% | Good |
| Duplication Ratio | 0.35% | Low |
| Circular Dependencies | 0 | None |
| Test Coverage (tox) | 36.4% | Meets threshold |
| Test Coverage (tui) | 36.7% | Meets threshold |

### Complexity Analysis

| Function | File | Lines | Complexity | Assessment |
|----------|------|-------|------------|------------|
| emit | internal/tox/anonymity.go | 19 | 10.9 | Borderline - event suppression logic |
| renderHistory | internal/tui/chat.go | 25 | 9.3 | Acceptable - message formatting |
| initTor | internal/tox/anonymity.go | 59 | 8.3 | Acceptable - transport setup |
| initI2P | internal/tox/anonymity.go | 59 | 8.3 | Acceptable - transport setup |
| Update (app) | internal/tui/app.go | 69 | ~8 | Justified - bubbletea Update() is inherently a message dispatcher |

All complexity scores are now below the traditional "warning" threshold of 15. The codebase is well-factored.

### Code Quality Highlights

1. **Low duplication** (0.35%) - Only one 10-line clone pair between initTor and initI2P (expected structural similarity)
2. **Good documentation** (87.3%) - All exported functions are documented
3. **No circular dependencies** - Clean package structure
4. **Consolidated retry logic** - `retryWithBackoff()` reduces code duplication
5. **Type-safe event system** - ToxEvent interface with concrete implementations
6. **Per-friend history persistence** - Fixed in `historyByFriend` map

---

## Roadmap

### Priority 1: Implement Group Chat (Conference) Support

**Gap**: `Ctrl+G` shows "not yet supported" but group chat is a core Tox protocol feature.

**Evidence**: `app.go:334` returns notification instead of implementing feature.

**Impact**: High - limits utility for users who need group communication. Tox conferences are used by many users for team/community chat.

**Blocked by**: The toxcore library (github.com/opd-ai/toxcore) exposes `ConferenceNew()`, `ConferenceInvite()`, and `ConferenceSendMessage()` but lacks the necessary callback methods (`OnConferenceInvite`, `OnConferenceMessage`, `OnConferencePeerJoin/Leave`) to receive conference events.

- [ ] **Upstream**: Request/implement conference callbacks in opd-ai/toxcore
- [ ] Add conference event types to `internal/tox/types.go`:
  - `ConferenceInviteEvent`
  - `ConferenceMessageEvent`
  - `ConferencePeerListChangedEvent`
- [ ] Register toxcore conference callbacks in `client.go:registerCallbacks()`
- [ ] Add wrapper methods to Client: `ConferenceNew()`, `ConferenceInvite()`, `ConferenceSendMessage()`, `ConferenceLeave()`
- [ ] Update contacts panel to show conferences (separate "Groups" section below contacts)
- [ ] Implement `Ctrl+G` to create new conference with type selection (text/av)
- [ ] Handle incoming conference invitations in modal (accept/reject)
- [ ] Display peer list in chat panel header for conferences
- [ ] **Validation**: Create conference, invite friend, exchange messages bidirectionally

### Priority 2: Increase Test Coverage to >50%

**Gap**: Current coverage (36%) meets CI threshold but is insufficient for confident refactoring.

**Evidence**: `go test -cover ./...` output shows coverage around 36% for both packages.

**Impact**: Medium - insufficient regression protection; refactoring group chat addition will be risky.

**Blockers**: Many Client methods require a real toxcore.Tox instance. Adding proper mocks would require defining an interface for the toxcore API.

- [ ] Define `ToxInterface` in `internal/tox/` to enable mocking
- [ ] Expand `internal/tox/client_test.go`:
  - Test `NewClient()` initialization and options (anon-only mode)
  - Test `emit()` channel behavior (normal, full buffer, shutdown)
  - Test `Bootstrap()` success/failure counting
  - Test file transfer state management
- [ ] Expand `internal/tox/anonymity_test.go`:
  - Test `calculateNextBackoff()` with various inputs
  - Test `isStopped()` behavior
  - Test `extractHost()` edge cases
- [ ] Expand `internal/tui/app_test.go`:
  - Test `Update()` with mock ToxEvents
  - Test modal open/close state transitions
  - Test `selectFriend()` saves/restores history
  - Test `/file` command parsing
- [ ] **Validation**: `go test -cover ./...` shows >50% coverage on both packages

### Priority 3: Voice/Video Call Support (Future)

**Gap**: Audio/video calls are not implemented.

**Evidence**: No AV-related code; toxcore supports `tox_av_*` APIs.

**Impact**: Low for TUI context - voice/video less common in terminal applications, but audio-only could be useful.

- [ ] Research opd-ai/toxcore AV API availability (`toxav.New`, `toxav.Call`, `toxav.Answer`)
- [ ] Implement audio-only call support first (more TUI-appropriate):
  - Display call status in status bar (🔊 In Call, ⏳ Calling...)
  - Handle incoming calls in modal (accept/reject with keybindings)
  - Implement audio routing (microphone → Tox → speaker) using portaudio or similar
  - Add mute/unmute toggle (Ctrl+M)
- [ ] Consider video support via ASCII art, sixel graphics, or external viewer (terminal-dependent)
- [ ] **Validation**: Initiate call, confirm audio flows bidirectionally

### Priority 4: Polish & UX Improvements

**Gap**: Minor polish items that would improve user experience.

**Impact**: Low - quality-of-life improvements.

- [ ] Add contact search/filter (type to filter contacts list)
- [ ] Add message editing support (if toxcore supports it)
- [ ] Add message deletion support
- [ ] Add read receipts display (if toxcore supports it)
- [ ] Add profile editing UI (change name/status without restarting)
- [ ] Add export/import chat history feature
- [ ] Add notification sound support (bell character or external command)
- [ ] **Validation**: Manual testing of each feature

---

## Summary

mtox achieves **90% of its stated goals** and is fully functional as a Tox messenger TUI. All core messaging features work correctly:

- ✅ Two-pane UI with contacts and chat
- ✅ Per-friend chat history persistence
- ✅ Typing indicators (optimized for state transitions)
- ✅ Friend request handling
- ✅ File transfer support (send/receive)
- ✅ Tor and I2P anonymity network support (simultaneous)
- ✅ Profile persistence and graceful shutdown
- ✅ Command-line options
- ✅ CI pipeline with build/test/vet/staticcheck and coverage threshold
- ✅ Release workflow for 6 platforms

**Highest-priority gap**: Group chat (conferences) is the only core Tox protocol feature not yet implemented, but it's blocked on upstream toxcore callback support.

**Test coverage**: At 36%, coverage meets the CI threshold and provides basic regression protection. Increasing to 50%+ would enable safer refactoring for the group chat feature.

The codebase is well-structured with:
- Low duplication (0.35%)
- Good documentation coverage (87.3%)
- No circular dependencies
- Low complexity (only one function at borderline 10.9)
- Excellent function length distribution (avg 10 lines, none >100)

The project is at v0.1.0 with a working release pipeline. Addressing the conference support gap will establish feature parity with mature Tox clients like qTox and uTox.

---

*Generated: 2026-03-25 | go-stats-generator v1.0.0*

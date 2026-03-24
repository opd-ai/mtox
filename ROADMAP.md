# Goal-Achievement Assessment

## Project Context

- **What it claims to do**: mtox is a "full-featured Tox Messenger terminal user interface (TUI)" that provides secure peer-to-peer messaging with a two-pane interface, automatic Tor/I2P anonymity network support, and standard Tox messenger features.

- **Target audience**: Privacy-conscious users who prefer terminal-based interfaces and want secure, decentralized messaging with optional anonymity network routing.

- **Architecture**:
  | Package | Role |
  |---------|------|
  | `cmd/mtox` | Application entrypoint |
  | `internal/tox` | Tox client wrapper bridging toxcore callbacks to bubbletea messages |
  | `internal/tui` | Terminal UI components (app, chat, contacts, statusbar, styles) |

- **Existing CI/quality gates**: None. No GitHub Actions workflows, Makefile, or CI configuration exists.

## Goal-Achievement Summary

| Stated Goal | Status | Evidence | Gap Description |
|-------------|--------|----------|-----------------|
| Two-pane TUI (contacts + chat) | ✅ Achieved | `internal/tui/app.go`, `contacts.go`, `chat.go` implement split layout | — |
| Real-time friend status indicators | ✅ Achieved | `contacts.go:221-228` - statusIndicator() renders ●/○ based on ConnectionStatus | — |
| Scrollable per-friend chat history | ⚠️ Partial | `chat.go` uses viewport for scrolling but history is **not persisted per-friend** - switching contacts clears history (`chat.go:56-59`) | History lost on contact switch |
| Typing indicators | ✅ Achieved | `chat.go:92-96`, `app.go:141-146` handle FriendTypingEvent | — |
| Incoming friend request dialog | ✅ Achieved | `app.go:398-404`, `449-459` - modal with accept/reject | — |
| Add friend dialog (Ctrl+N) | ✅ Achieved | `app.go:389-395`, `413-426` - modal input and submission | — |
| Profile persistence at ~/.config/mtox | ✅ Achieved | `client.go:70-75` loads, `client.go:227-239` saves to ProfilePath() | — |
| Mouse support | ✅ Achieved | `main.go:28` - `tea.WithMouseCellMotion()`, `app.go:282-302` handles clicks | — |
| Graceful shutdown with auto-save | ✅ Achieved | `app.go:382-386` quit() calls Save() then Stop() | — |
| Automatic Tor support | ✅ Achieved | `anonymity.go:187-248` - NewTorTransport with retry loop | — |
| Automatic I2P support | ✅ Achieved | `anonymity.go:283-344` - NewI2PTransport with retry loop | — |
| Simultaneous Tor + I2P | ✅ Achieved | Both initTor() and initI2P() run in parallel goroutines | — |
| Status bar shows Tor/I2P indicators | ✅ Achieved | `statusbar.go:78-101` - shows 🧅Tor and 🧄I2P when available | — |
| MTOX_ANON_ONLY mode | ✅ Achieved | `client.go:47-68` disables UDP/IPv6/local discovery | — |
| MTOX_DISABLE_TOR/I2P variables | ✅ Achieved | `anonymity.go:189-204`, `285-300` check environment | — |
| File transfers | ❌ Missing | No implementation - README doesn't claim this but Tox protocol supports it | Core Tox feature not exposed |
| Group chat | ❌ Missing | `app.go:230` returns "Group chat not yet supported" notification | Ctrl+G exists but non-functional |
| Voice/video calls | ❌ Missing | No implementation - README doesn't claim this but Tox protocol supports it | Core Tox feature not exposed |
| Test coverage | ❌ Missing | `go test ./...` shows "no test files" for all packages | 0% test coverage |

**Overall: 14/17 goals fully achieved (82%)**

## Metrics Summary (go-stats-generator)

| Metric | Value | Assessment |
|--------|-------|------------|
| Total Lines of Code | 1,039 | Compact codebase |
| Functions/Methods | 106 | Well-factored |
| Average Function Length | 11.4 lines | Good |
| Functions >50 lines | 5 (4.7%) | Acceptable |
| High Complexity (>10) | 3 functions | Minor concern |
| Documentation Coverage | 84.9% | Good |
| Package/Function Coverage | 100% | Excellent |
| Method Coverage | 76.5% | Adequate |
| Duplication Ratio | 1.63% | Low |
| Circular Dependencies | 0 | None |

### High-Risk Functions (Complexity >10)

| Function | File | Lines | Complexity |
|----------|------|-------|------------|
| iterateLoop | internal/tox/client.go | 33 | 18.4 |
| handleKey | internal/tui/app.go | 66 | 17.9 |
| overlayModal | internal/tui/app.go | 66 | 15.3 |
| tryI2PListenWithRetry | internal/tox/anonymity.go | 29 | 14.0 |
| tryTorListenWithRetry | internal/tox/anonymity.go | 28 | 14.0 |
| Update | internal/tui/app.go | 79 | 13.2 |

These functions are on critical paths (event loop, input handling) but their complexity is justified by the inherent state machine nature of TUI applications.

---

## Roadmap

### Priority 1: Per-Friend Chat History Persistence

**Gap**: Chat history is cleared when switching between contacts, making the application impractical for actual use.

**Evidence**: `chat.go:56-59` - `setFriend()` sets `c.history = nil` on every contact switch.

**Impact**: High - this breaks the core messaging experience.

- [ ] Add `historyByFriend map[uint32][]chatMessage` to `App` struct (`app.go`)
- [ ] Modify `selectFriend()` to save current history to map before switching
- [ ] Modify `chatPanel.setFriend()` to load history from map instead of clearing
- [ ] Consider persisting history to disk (e.g., `~/.config/mtox/history/`)
- [ ] **Validation**: Send messages to multiple contacts, switch between them, verify history is preserved

### Priority 2: Add Test Coverage

**Gap**: 0% test coverage across all packages.

**Evidence**: `go test ./...` returns "no test files" for all 3 packages.

**Impact**: High - no regression protection, makes refactoring risky.

- [ ] Create `internal/tox/client_test.go` - test event emission, callback wiring
- [ ] Create `internal/tox/anonymity_test.go` - test status transitions, retry backoff logic
- [ ] Create `internal/tui/app_test.go` - test Update() with mock messages
- [ ] Create `internal/tui/contacts_test.go` - test sorting, unread count logic
- [ ] Add `go test -race ./...` to CI when CI is added
- [ ] **Validation**: `go test -cover ./...` shows >60% coverage on core logic

### Priority 3: Implement Group Chat (Conference) Support

**Gap**: Ctrl+G shows "not yet supported" but group chat is a core Tox feature.

**Evidence**: `app.go:229-231` returns notification instead of implementing feature.

**Impact**: Medium - limits utility for users who need group communication.

- [ ] Add `ConferenceEvent` types to `types.go` (conference invite, message, peer join/leave)
- [ ] Register toxcore conference callbacks in `client.go:registerCallbacks()`
- [ ] Add `ConferenceNew()`, `ConferenceInvite()`, `ConferenceSendMessage()` wrapper methods to Client
- [ ] Create conference list UI in contacts panel (separate section or mixed with friends)
- [ ] Implement Ctrl+G to create new conference, prompt for conference type
- [ ] Handle incoming conference invitations in modal
- [ ] **Validation**: Create conference, invite friend, exchange messages

### Priority 4: Add CI/CD Pipeline

**Gap**: No automated quality gates exist.

**Evidence**: No `.github/workflows/`, `Makefile`, or CI configuration.

**Impact**: Medium - no automated build/test verification for contributions.

- [ ] Create `.github/workflows/ci.yml`:
  ```yaml
  - go build ./cmd/mtox
  - go test -race ./...
  - go vet ./...
  - staticcheck ./... (optional)
  ```
- [ ] Add build status badge to README.md
- [ ] Consider adding release workflow for binary artifacts
- [ ] **Validation**: Push to branch, verify CI runs successfully

### Priority 5: Implement File Transfer Support

**Gap**: No file transfer capability despite being a core Tox protocol feature.

**Evidence**: No file-related code in codebase; toxcore API supports `FileSend()`, `FileControl()`.

**Impact**: Medium - limits utility for users who need to share files securely.

- [ ] Add file transfer event types to `types.go` (FileChunk, FileControl, etc.)
- [ ] Register `OnFileRecv*` callbacks in `client.go`
- [ ] Add file send command (e.g., `/file <path>` in chat input)
- [ ] Create file transfer progress indicator in chat or status bar
- [ ] Handle incoming file requests in modal (accept/reject/save location)
- [ ] Implement file reception with progress tracking
- [ ] **Validation**: Send file to friend, receive file from friend, verify integrity

### Priority 6: Extract Duplicated Retry Logic

**Gap**: Tor and I2P retry loops are nearly identical (23 lines duplicated).

**Evidence**: go-stats-generator reports clone pair at `anonymity.go:257-279` and `anonymity.go:354-376`.

**Impact**: Low - maintainability concern, not user-facing.

- [ ] Extract generic `retryWithBackoff(ctx, attempt func() error, opts RetryOpts) error`
- [ ] Refactor `tryTorListenWithRetry()` and `tryI2PListenWithRetry()` to use shared function
- [ ] **Validation**: `go-stats-generator` duplication ratio decreases; behavior unchanged

### Priority 7: Add Voice/Video Call Support

**Gap**: Audio/video calls are not implemented.

**Evidence**: No AV-related code; toxcore supports `tox_av_*` APIs.

**Impact**: Low for TUI context - less common in terminal applications.

- [ ] Research toxcore AV API availability in opd-ai/toxcore
- [ ] Add audio-only call support first (more feasible in terminal)
- [ ] Display call status in status bar
- [ ] Handle incoming calls in modal
- [ ] **Validation**: Initiate call, audio flows bidirectionally

---

## Appendix: PR #1 Unchecked Items

The original PR #1 listed several fix items that appear to have been addressed or are minor:

| Item | Status |
|------|--------|
| Typing notification only on state transitions | ⚠️ Current code sends on every keystroke (`app.go:273-275`) |
| Stop() idempotent via sync.Once | ✅ Fixed in PR #2 |
| Per-friend chat history | ❌ Still not implemented (Priority 1 above) |
| Use time.Timer to avoid per-iteration allocs | ✅ Implemented (`client.go:166`) |
| Blocking emit with done-channel fallback | ✅ Implemented (`client.go:138-143`) |

The typing notification optimization is minor but should be addressed:

- [ ] Track `lastTypingState bool` in App
- [ ] Only call `SetTyping()` when state changes from false→true or true→false
- [ ] **Location**: `app.go:273-276`

---

## Summary

mtox achieves 82% of its stated goals and is functional as a basic Tox messenger TUI. The highest-priority gap is **per-friend chat history persistence**, which severely impacts usability. Adding **test coverage** is critical for maintainability as the project grows. Group chat and file transfer would elevate mtox to feature parity with mature Tox clients like qTox and uTox.

The codebase is well-structured with low duplication and good documentation coverage. The project is in early development (v0.1) with no releases yet, so addressing these gaps now will establish a solid foundation for future growth.

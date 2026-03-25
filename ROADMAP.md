# Goal-Achievement Assessment

## Project Context

- **What it claims to do**: mtox is a "full-featured Tox Messenger terminal user interface (TUI)" that provides secure peer-to-peer messaging with a two-pane interface (contacts + chat), automatic Tor/I2P anonymity network support, and standard Tox messenger features including typing indicators, friend requests, and profile persistence.

- **Target audience**: Privacy-conscious users who prefer terminal-based interfaces and want secure, decentralized messaging with optional anonymity network routing via Tor and I2P.

- **Architecture**:
  | Package | Role | Files |
  |---------|------|-------|
  | `cmd/mtox` | Application entrypoint | 1 |
  | `internal/tox` | Tox client wrapper bridging toxcore callbacks to bubbletea messages; anonymity network management | 3 |
  | `internal/tui` | Terminal UI components (app, chat, contacts, statusbar, styles) | 5 |

- **Existing CI/quality gates**:
  - ✅ GitHub Actions CI (`.github/workflows/ci.yml`): build, test (`-race`), vet
  - ❌ No staticcheck, golangci-lint, or code coverage enforcement
  - ❌ No release workflow for binary artifacts

## Goal-Achievement Summary

| Stated Goal | Status | Evidence | Gap Description |
|-------------|--------|----------|-----------------|
| Two-pane TUI (contacts + chat) | ✅ Achieved | `internal/tui/app.go`, `contacts.go`, `chat.go` implement split layout with `contactsPanelWidth = 22` | — |
| Real-time friend status indicators | ✅ Achieved | `contacts.go` - `statusIndicator()` renders ●/○/◌ based on ConnectionStatus | — |
| Scrollable per-friend chat history | ✅ Achieved | `app.go:58` - `historyByFriend map[uint32][]chatMessage`; `selectFriend()` saves/restores history | — |
| Typing indicators | ✅ Achieved | `chat.go:101-104` displays typing state; `app.go:277-282` only sends `SetTyping()` on state transitions | — |
| Incoming friend request dialog | ✅ Achieved | `app.go:414-420`, `445-461` - modal with accept/reject via Enter/R keys | — |
| Add friend dialog (Ctrl+N) | ✅ Achieved | `app.go:405-411`, `429-441` - modal input and submission | — |
| Profile persistence at ~/.config/mtox | ✅ Achieved | `client.go:70-75` loads, `client.go:234-245` saves to ProfilePath() | — |
| Mouse support | ✅ Achieved | `cmd/mtox/main.go` - `tea.WithMouseCellMotion()`; `app.go:290-308` handles clicks | — |
| Graceful shutdown with auto-save | ✅ Achieved | `app.go:398-401` `quit()` calls `Save()` then `Stop()` | — |
| Automatic Tor support | ✅ Achieved | `anonymity.go:187-248` - `NewTorTransport` with retry loop via `retryWithBackoff()` | — |
| Automatic I2P support | ✅ Achieved | `anonymity.go:259-318` - `NewI2PTransport` with retry loop via `retryWithBackoff()` | — |
| Simultaneous Tor + I2P | ✅ Achieved | `anonymity.go:86-91` - both `initTor()` and `initI2P()` run in parallel goroutines | — |
| Status bar shows Tor/I2P indicators | ✅ Achieved | `statusbar.go` - shows 🧅Tor and 🧄I2P when available | — |
| MTOX_ANON_ONLY mode | ✅ Achieved | `client.go:47-68` disables UDP/IPv6/local discovery | — |
| MTOX_DISABLE_TOR/I2P variables | ✅ Achieved | `anonymity.go:189-204`, `261-275` check environment | — |
| Test coverage | ⚠️ Partial | `go test -cover`: tox 5.9%, tui 14.8% - tests exist but coverage is low | Need >60% coverage on core logic |
| Group chat (conferences) | ❌ Missing | `app.go:234-235` returns "Group chat not yet supported" notification | Core Tox feature not exposed |
| File transfers | ❌ Missing | No implementation - toxcore supports `FileSend()`, `FileControl()` | Core Tox feature not exposed |
| Voice/video calls | ❌ Missing | No implementation - toxcore supports `tox_av_*` APIs | Core Tox feature not exposed |
| Command-line options | ❌ Missing | No `-h/--help` or flags; all config via env vars | Minor discoverability gap |

**Overall: 15/19 goals fully achieved (79%), 1 partial, 3 missing**

## Metrics Summary (go-stats-generator)

| Metric | Value | Assessment |
|--------|-------|------------|
| Total Lines of Code | 1,047 | Compact codebase |
| Functions/Methods | 108 | Well-factored |
| Average Function Length | 11.3 lines | Good (threshold: <15) |
| Functions >50 lines | 5 (4.6%) | Acceptable |
| High Complexity (>10) | 3 functions | Acceptable for TUI state machines |
| Documentation Coverage | 84.9% | Good |
| Package Coverage | 100% | Excellent |
| Function Coverage | 100% | Excellent |
| Method Coverage | 76.5% | Adequate |
| Duplication Ratio | 0.49% | Low |
| Circular Dependencies | 0 | None |
| Test Coverage (tox) | 5.9% | Needs improvement |
| Test Coverage (tui) | 14.8% | Needs improvement |

### High-Complexity Functions (Complexity >10)

| Function | File | Lines | Complexity | Assessment |
|----------|------|-------|------------|------------|
| handleKey | internal/tui/app.go | 69 | 19.7 | Justified - keyboard input state machine |
| iterateLoop | internal/tox/client.go | 33 | 18.4 | Justified - event loop with timer management |
| overlayModal | internal/tui/app.go | 66 | 15.3 | Justified - modal rendering logic |
| retryWithBackoff | internal/tox/anonymity.go | 28 | 14.0 | Good - consolidated retry logic |
| Update | internal/tui/app.go | 79 | 13.2 | Justified - bubbletea Update() is inherently complex |

These functions are on critical paths (event loop, input handling) but their complexity is justified by the inherent state machine nature of TUI applications. The refactoring of retry logic into `retryWithBackoff()` addressed previous duplication concerns.

### Recent Improvements Since Last Assessment

1. **Per-friend chat history persistence** - Fixed in `app.go:58` with `historyByFriend` map
2. **Typing notification optimization** - Now only calls `SetTyping()` on state transitions (`app.go:277-282`)
3. **Test files added** - `types_test.go`, `anonymity_test.go`, `chat_test.go`, `contacts_test.go`
4. **Retry logic consolidated** - `retryWithBackoff()` function reduced duplication
5. **CI pipeline added** - `.github/workflows/ci.yml` with build, test, vet

---

## Roadmap

### Priority 1: Increase Test Coverage to >60%

**Gap**: Current coverage is too low (tox: 5.9%, tui: 14.8%) for safe refactoring.

**Evidence**: `go test -cover ./...` output shows minimal coverage despite test files existing.

**Impact**: High - insufficient regression protection; refactoring is risky.

**Status**: Coverage improved to tox: 35.6%, tui: 36.9% (from 30.4%, 36.0%).

- [x] Expand `internal/tox/client_test.go`:
  - Test `NewClient()` initialization and options (anon-only mode)
  - Test `emit()` channel behavior (normal, full buffer, shutdown)
  - Test `Bootstrap()` success/failure counting
  - Mock toxcore.Tox interface for unit testing
- [x] Expand `internal/tox/anonymity_test.go`:
  - Test `AnonymityStatus.String()` for all values
  - Test `Start()` initialization idempotence
  - Test `Stop()` cleanup and double-stop safety
  - Test `retryWithBackoff()` with mock listener factory
- [x] Expand `internal/tui/chat_test.go`:
  - Test `setFriendWithHistory()` preserves messages
  - Test `addMessage()` appends correctly
  - Test `renderHistory()` formats timestamps correctly
  - Test typing indicator display
- [x] Expand `internal/tui/contacts_test.go`:
  - Test sorting by connection status then name
  - Test `incrementUnread()` / `clearUnread()` behavior
  - Test `updateConnectionStatus()` updates correctly
- [x] Add `internal/tui/app_test.go`:
  - Test `Update()` with mock ToxEvents
  - Test modal open/close state transitions
  - Test `selectFriend()` saves/restores history
- [ ] **Validation**: `go test -cover ./...` shows >60% coverage on `internal/tox` and `internal/tui`

### Priority 2: Implement Group Chat (Conference) Support

**Gap**: `Ctrl+G` shows "not yet supported" but group chat is a core Tox protocol feature.

**Evidence**: `app.go:234-235` returns notification instead of implementing feature.

**Impact**: Medium - limits utility for users who need group communication. Tox conferences are used by many users for team/community chat.

- [ ] Add conference event types to `internal/tox/types.go`:
  ```go
  type ConferenceInviteEvent struct { ... }
  type ConferenceMessageEvent struct { ... }
  type ConferencePeerJoinEvent struct { ... }
  type ConferencePeerLeaveEvent struct { ... }
  ```
- [ ] Register toxcore conference callbacks in `client.go:registerCallbacks()`
- [ ] Add wrapper methods to Client:
  - `ConferenceNew(kind ConferenceType) (uint32, error)`
  - `ConferenceInvite(conferenceID, friendID uint32) error`
  - `ConferenceSendMessage(conferenceID uint32, message string) error`
  - `ConferenceGetPeers(conferenceID uint32) []ConferencePeer`
- [ ] Update contacts panel to show conferences (separate "Groups" section)
- [ ] Implement `Ctrl+G` to create new conference with type selection (text/av)
- [ ] Handle incoming conference invitations in modal (accept/reject)
- [ ] Display peer list in chat panel header for conferences
- [ ] **Validation**: Create conference, invite friend, exchange messages bidirectionally

### Priority 3: Implement File Transfer Support

**Gap**: No file transfer capability despite being a core Tox protocol feature.

**Evidence**: No file-related code in codebase; toxcore API supports `FileSend()`, `FileControl()`.

**Impact**: Medium - limits utility for users who need to share files securely via P2P encryption.

- [ ] Add file transfer event types to `internal/tox/types.go`:
  ```go
  type FileRecvRequestEvent struct { FriendID uint32; FileID uint32; Filename string; Size uint64 }
  type FileRecvChunkEvent struct { FriendID uint32; FileID uint32; Data []byte; Position uint64 }
  type FileControlEvent struct { FriendID uint32; FileID uint32; Control FileControlType }
  ```
- [ ] Register `OnFileRecv*` callbacks in `client.go:registerCallbacks()`
- [ ] Add wrapper methods to Client:
  - `FileSend(friendID uint32, filename string, data []byte) (uint32, error)`
  - `FileControl(friendID, fileID uint32, control FileControlType) error`
- [ ] Implement `/file <path>` command in chat input parsing
- [ ] Create file transfer progress indicator (show in chat as special message type)
- [ ] Handle incoming file requests in modal (accept/reject, choose save location)
- [ ] Implement file reception with progress tracking and completion notification
- [ ] **Validation**: Send file to friend, receive file from friend, verify integrity with checksum

### Priority 4: Add Command-Line Options

**Gap**: No `-h/--help` or command-line flags exist; all configuration is via environment variables.

**Evidence**: Running `./mtox --help` starts the application instead of showing usage.

**Impact**: Low - minor discoverability gap; users must read documentation to learn about env vars.

- [ ] Add `flag` package to `cmd/mtox/main.go`
- [ ] Implement flags:
  - `-h/--help`: Print usage and exit
  - `--version`: Print version and exit
  - `--anon-only`: Equivalent to `MTOX_ANON_ONLY=1`
  - `--no-tor`: Equivalent to `MTOX_DISABLE_TOR=1`
  - `--no-i2p`: Equivalent to `MTOX_DISABLE_I2P=1`
  - `--profile <path>`: Custom profile location (override `~/.config/mtox/profile.tox`)
- [ ] Add version constant to package (e.g., `const Version = "0.2.0"`)
- [ ] **Validation**: Run `./mtox --help`, confirm usage is displayed; verify flags work correctly

### Priority 5: Add staticcheck/golangci-lint to CI

**Gap**: CI only runs `go vet`, missing additional static analysis.

**Evidence**: `.github/workflows/ci.yml` does not include `staticcheck` or `golangci-lint`.

**Impact**: Low - additional quality gates would catch more issues before merge.

- [ ] Add `staticcheck` step to CI:
  ```yaml
  - name: Install staticcheck
    run: go install honnef.co/go/tools/cmd/staticcheck@latest
  - name: Staticcheck
    run: staticcheck ./...
  ```
- [ ] Consider adding `golangci-lint` for comprehensive linting
- [ ] Add coverage threshold enforcement:
  ```yaml
  - name: Test with coverage
    run: go test -coverprofile=coverage.out ./...
  - name: Check coverage threshold
    run: |
      COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%')
      if (( $(echo "$COVERAGE < 40" | bc -l) )); then
        echo "Coverage $COVERAGE% is below 40% threshold"
        exit 1
      fi
  ```
- [ ] **Validation**: Push to branch, verify new CI steps pass

### Priority 6: Add Release Workflow

**Gap**: No automated release workflow for binary artifacts.

**Evidence**: No release workflow in `.github/workflows/`.

**Impact**: Low - users must build from source; providing binaries improves adoption.

- [ ] Create `.github/workflows/release.yml`:
  - Trigger on tag push (`v*.*.*`)
  - Build binaries for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
  - Create GitHub Release with artifacts
- [ ] Add build badge to README.md
- [ ] Consider using GoReleaser for cross-compilation
- [ ] **Validation**: Tag a release, verify binaries are published to GitHub Releases

### Priority 7: Voice/Video Call Support (Future)

**Gap**: Audio/video calls are not implemented.

**Evidence**: No AV-related code; toxcore supports `tox_av_*` APIs.

**Impact**: Low for TUI context - voice/video less common in terminal applications, but audio-only could be useful.

- [ ] Research opd-ai/toxcore AV API availability
- [ ] Implement audio-only call support first (more TUI-appropriate)
  - Display call status in status bar
  - Handle incoming calls in modal (accept/reject)
  - Implement audio routing (microphone → Tox → speaker)
- [ ] Consider video support via ASCII art or sixel graphics (terminal-dependent)
- [ ] **Validation**: Initiate call, confirm audio flows bidirectionally

---

## Summary

mtox achieves **79% of its stated goals** and is functional as a Tox messenger TUI. All core messaging features work correctly:

- ✅ Two-pane UI with contacts and chat
- ✅ Per-friend chat history persistence (fixed)
- ✅ Typing indicators (optimized)
- ✅ Friend request handling
- ✅ Tor and I2P anonymity network support
- ✅ Profile persistence and graceful shutdown
- ✅ CI pipeline with build/test/vet

**Highest-priority gap**: Test coverage is low (5.9-14.8%) despite test files existing. Expanding test coverage is critical before adding new features to ensure regression protection.

**Feature gaps**: Group chat (conferences) and file transfers are core Tox protocol features that would elevate mtox to feature parity with mature clients like qTox and uTox.

The codebase is well-structured with:
- Low duplication (0.49%)
- Good documentation coverage (84.9%)
- No circular dependencies
- Reasonable complexity for a TUI application

The project is in early development (v0.1) with no tagged releases yet. Addressing these gaps will establish a solid foundation for broader adoption.

---

*Generated: 2026-03-24 | go-stats-generator v1.0.0*

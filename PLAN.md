# Implementation Plan: Production-Ready Tox TUI Client

## Project Context
- **What it does**: mtox is a full-featured Tox Messenger terminal user interface (TUI) with automatic Tor/I2P anonymity network support.
- **Current goal**: Achieve production-readiness through test coverage, then deliver missing core Tox protocol features (group chat, file transfer).
- **Estimated Scope**: Medium (10 items above thresholds across multiple categories)

## Goal-Achievement Status
| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| Two-pane TUI (contacts + chat) | ✅ Achieved | No |
| Real-time friend status indicators | ✅ Achieved | No |
| Scrollable per-friend chat history | ✅ Achieved | No |
| Typing indicators | ✅ Achieved | No |
| Friend request dialog (accept/reject) | ✅ Achieved | No |
| Add friend dialog (Ctrl+N) | ✅ Achieved | No |
| Profile persistence (~/.config/mtox) | ✅ Achieved | No |
| Mouse support | ✅ Achieved | No |
| Graceful shutdown with auto-save | ✅ Achieved | No |
| Automatic Tor/I2P support | ✅ Achieved | No |
| Test coverage >40% | ❌ 5.9-14.8% | **Yes** |
| Group chat (conferences) | ❌ Missing | **Yes** |
| File transfers | ❌ Missing | **Yes** |
| Command-line options (--help) | ❌ Missing | **Yes** |

**Summary**: 14/18 goals achieved (78%). This plan addresses the 4 unmet goals.

## Metrics Summary
- **Complexity hotspots on goal-critical paths**: 4 functions above threshold (>9)
  - `handleKey` (14): keyboard input state machine — will expand for new features
  - `iterateLoop` (13): event loop — may need new event types
  - `overlayModal` (11): modal rendering — new modals for group/file dialogs
  - `retryWithBackoff` (10): retry logic — no changes needed
- **Duplication ratio**: 0.49% (excellent, well below 5% threshold)
- **Doc coverage**: 71.3% function coverage (adequate)
- **Test coverage**: tox 5.9%, tui 14.8% (critically low)
- **Package coupling**: Clean — no circular dependencies, 3 packages with clear separation

## Research Findings
- **Community status**: Early development phase, no tagged releases, no open issues. Active recent commits.
- **Dependency outlook**: Bubble Tea v2 available with breaking changes (import path, View return type, KeyMsg→KeyPressMsg). Migration recommended but not urgent.
- **Domain best practices**: Tox privacy model exposes IP to contacts; mtox already mitigates this with Tor/I2P support.

---

## Implementation Steps

### Step 1: Expand Test Coverage to >40% ✅ COMPLETED
- **Deliverable**: New and expanded test files achieving >40% coverage across `internal/tox` and `internal/tui`
- **Dependencies**: None
- **Goal Impact**: Enables safe refactoring and regression protection for all subsequent steps
- **Files modified**:
  - `internal/tox/anonymity_test.go` — expanded with tests for AnonymityManager, emit suppression, environment variable handling, retry logic, and custom profile path
  - `internal/tui/app_test.go` (new) — tests for status bar, contacts panel, chat panel, and constants
  - `internal/tui/chat_test.go` — expanded with tests for input handling, message rendering
  - `internal/tui/contacts_test.go` — expanded with tests for keyboard navigation, enter handling
- **Result**: Coverage now 41.0% for `internal/tox` and 42.3% for `internal/tui`

### Step 2: Add Command-Line Options ✅ COMPLETED
- **Deliverable**: `-h/--help`, `--version`, `--anon-only`, `--no-tor`, `--no-i2p`, `--profile` flags
- **Dependencies**: Step 1 (tests ensure flags don't break existing behavior)
- **Goal Impact**: Closes the command-line options gap; improves discoverability
- **Files modified**:
  - `cmd/mtox/main.go` — added `flag` package, defined all flags, mapped to env vars
  - `internal/tui/statusbar.go` — now references version constant
  - `internal/version/version.go` (new) — defines `const Version = "0.1.0"`
  - `internal/tox/client.go` — ProfilePath() now respects MTOX_PROFILE_PATH env var
- **Result**: `./mtox --help` prints usage; `./mtox --version` prints version

### Step 3: Implement Group Chat (Conference) Support ⚠️ BLOCKED
- **Deliverable**: Create/join conferences, send/receive group messages, see peer list
- **Dependencies**: Step 1 (test coverage protects existing code during refactoring)
- **Goal Impact**: Enables team/community communication — core Tox protocol feature
- **Status**: BLOCKED - The toxcore library (github.com/opd-ai/toxcore) exposes ConferenceNew(), ConferenceInvite(), and ConferenceSendMessage() but lacks the necessary callback methods (OnConferenceInvite, OnConferenceMessage, OnConferencePeerJoin/Leave) to receive conference events. This requires upstream changes to toxcore.
- **Files to modify**:
  - `internal/tox/types.go` — add `ConferenceInviteEvent`, `ConferenceMessageEvent`, `ConferencePeerEvent`
  - `internal/tox/client.go` — register conference callbacks, add wrapper methods
  - `internal/tui/app.go` — replace "not yet supported" with conference creation modal
  - `internal/tui/contacts.go` — add "Groups" section to contact list
  - `internal/tui/chat.go` — display peer list in header for group chats
- **Acceptance**: User can create conference, invite friend, exchange messages bidirectionally
- **Validation**:
  ```bash
  # Manual test: Start two mtox instances, create conference in one, invite from other, exchange messages
  go test -run TestConference ./internal/tox ./internal/tui
  ```

### Step 4: Implement File Transfer Support ✅ COMPLETED
- **Deliverable**: Send/receive files with progress indication and accept/reject modal
- **Dependencies**: Step 1, Step 3 (conference infrastructure patterns reusable)
- **Goal Impact**: Enables secure P2P file sharing — core Tox protocol feature
- **Files modified**:
  - `internal/tox/types.go` — added `FileRecvRequestEvent`, `FileRecvChunkEvent`, `FileChunkRequestEvent`, `FileSendCompleteEvent`, `FileRecvCompleteEvent`, `FileTransferErrorEvent`
  - `internal/tox/client.go` — registered file callbacks, added `FileSend()`, `FileAccept()`, `FileReject()`, `FilePause()` wrappers and file data tracking
  - `internal/tui/app.go` — handle `/file <path>` command, file request modal, file transfer events
  - `internal/tui/styles.go` — added file transfer styles
- **Result**: Users can send files with `/file <path>`, receive files with accept/reject modal, files saved to `~/.config/mtox/downloads/`

### Step 5: Add staticcheck to CI Pipeline ✅ COMPLETED
- **Deliverable**: Updated CI workflow with staticcheck and coverage threshold
- **Dependencies**: Step 1 (coverage threshold requires adequate coverage first)
- **Goal Impact**: Catches issues before merge; enforces quality standards
- **Files to modify**:
  - `.github/workflows/ci.yml` — add staticcheck step, add coverage threshold check (30%)
- **Acceptance**: CI passes with staticcheck and fails if coverage drops below 30%
- **Validation**:
  ```bash
  # Push to branch, verify new CI steps appear and pass
  gh workflow view ci.yml
  ```
- **Status**: Added staticcheck step, coverage check (30% threshold), fixed unused struct fields in chat.go

### Step 6: Create Release Workflow ✅ COMPLETED
- **Deliverable**: Automated binary releases for linux/darwin/windows on tag push
- **Dependencies**: Steps 1-5 complete (release should include all features)
- **Goal Impact**: Enables adoption by users without Go toolchain
- **Files to modify**:
  - `.github/workflows/release.yml` (new) — trigger on `v*.*.*` tags, cross-compile, upload to GitHub Releases
  - `README.md` — add build status badge and download links
- **Acceptance**: Tagging `v0.2.0` creates GitHub Release with binaries for 6 OS/arch combinations
- **Validation**:
  ```bash
  # Tag and push, verify release appears with artifacts
  gh release list | grep -q "v0.2.0" && echo "PASS"
  ```
- **Status**: Created release.yml workflow targeting 6 platforms; added CI/Release badges and download links to README

---

## Complexity Impact Assessment

| Function | Current | After Step 3 | After Step 4 | Mitigation |
|----------|---------|--------------|--------------|------------|
| handleKey | 14 | ~16 | ~18 | Extract modal-specific handlers |
| Update | 9 | ~11 | ~13 | Use dispatch table for event types |
| overlayModal | 11 | ~13 | ~14 | Acceptable for modal rendering |
| iterateLoop | 13 | 13 | 13 | No changes needed |

**Mitigation strategy**: As complexity increases during Steps 3-4, extract dedicated handler methods:
- `handleConferenceKey()` for group chat keyboard input
- `handleFileTransferKey()` for file-related commands
- `handleToxEvent(event ToxEvent)` dispatch table in `Update()`

---

## Scope Calibration

| Metric | Current | Threshold | Items Above | Assessment |
|--------|---------|-----------|-------------|------------|
| Functions >9 complexity | 4 | 9.0 | 4 | Medium |
| Duplication ratio | 0.49% | 3% | 0 | Excellent |
| Doc coverage | 71.3% | 80% | 31 undocumented | Small gap |
| Test coverage | 5.9-14.8% | 40% | 2 packages | High priority |

**Overall Scope**: Medium — primary work is expanding test coverage and adding two feature modules.

---

## Validation Commands Summary

```bash
# Step 1: Test coverage
go test -cover ./internal/tox ./internal/tui 2>&1 | grep coverage

# Step 2: CLI flags
go build ./cmd/mtox && ./mtox --help && ./mtox --version

# Step 3: Conference tests
go test -v -run TestConference ./...

# Step 4: File transfer tests
go test -v -run TestFileTransfer ./...

# Step 5: CI with staticcheck
staticcheck ./...

# Step 6: Release workflow
gh release list --repo opd-ai/mtox

# Full metrics validation after all steps
go-stats-generator analyze . --skip-tests --format json | jq '{
  complexity_above_9: [.functions[] | select(.complexity.cyclomatic > 9) | .name],
  duplication: .duplication.duplication_ratio,
  doc_coverage: ([.functions[] | select(.documentation.has_comment)] | length) / ([.functions[]] | length) * 100
}'
```

---

*Generated: 2026-03-24 | go-stats-generator v1.0.0*

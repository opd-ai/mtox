# Implementation Gaps — 2026-03-24

This document details the gaps between mtox's stated goals (from README.md) and its current implementation.

---

## Test Coverage Below Maintainability Threshold

- **Stated Goal**: Implicit — maintainable, production-quality code requires adequate test coverage for safe refactoring and regression prevention.
- **Current State**: Test coverage is critically low despite test files existing:
  - `cmd/mtox`: 0% (no tests, expected for entrypoint)
  - `internal/tox`: 5.9%
  - `internal/tui`: 14.8%
  
  Test files exist (`types_test.go`, `anonymity_test.go`, `chat_test.go`, `contacts_test.go`) but cover only basic utility functions and data structures.
- **Impact**: **High maintenance risk**. Refactoring is dangerous, regressions are hard to detect, and contributors cannot verify their changes don't break existing functionality.
- **Closing the Gap**:
  1. Expand `internal/tox/client_test.go`:
     - Test `emit()` channel behavior (normal, full buffer, shutdown)
     - Test `Bootstrap()` success/failure counting
     - Mock toxcore.Tox interface for unit testing event callbacks
  2. Expand `internal/tox/anonymity_test.go`:
     - Test `Start()` initialization idempotence (sync.Once behavior)
     - Test `Stop()` cleanup and double-stop safety
     - Test `retryWithBackoff()` with mock listener factory
  3. Create `internal/tui/app_test.go`:
     - Test `Update()` handles all ToxEvent types correctly
     - Test modal open/close state transitions
     - Test `selectFriend()` saves/restores history correctly
  4. Expand `internal/tui/chat_test.go`:
     - Test `update()` keyboard handling when focused vs unfocused
     - Test message submission clears input
  5. **Validation**: `go test -cover ./... | awk '/internal\/(tox|tui)/ {print}'` shows >40% coverage

---

## Group Chat (Conference) Support

- **Stated Goal**: The README shows keyboard shortcut `Ctrl+G` implying group chat functionality is planned. The original PR #1 specification includes full conference API usage.
- **Current State**: `internal/tui/app.go:234-235` handles Ctrl+G by displaying "Group chat not yet supported." The toxcore library supports conferences, but mtox does not implement them.
- **Impact**: **Medium feature limitation**. Users expecting Tox feature parity with clients like qTox will find group communication missing. This is common for team and community use cases.
- **Closing the Gap**:
  1. Add conference event types to `internal/tox/types.go`:
     ```go
     type ConferenceInviteEvent struct {
         FriendID     uint32
         ConferenceID uint32
         Kind         toxcore.ConferenceType
     }
     type ConferenceMessageEvent struct {
         ConferenceID uint32
         PeerID       uint32
         Message      string
     }
     ```
  2. Register toxcore conference callbacks in `client.go:registerCallbacks()`:
     - `tox.OnConferenceInvite()`
     - `tox.OnConferenceMessage()`
     - `tox.OnConferencePeerJoin()` / `OnConferencePeerLeave()`
  3. Add wrapper methods to Client:
     - `ConferenceNew(kind ConferenceType) (uint32, error)`
     - `ConferenceInvite(conferenceID, friendID uint32) error`
     - `ConferenceSendMessage(conferenceID uint32, message string) error`
     - `ConferenceGetPeers(conferenceID uint32) []ConferencePeer`
  4. Update contacts panel to show conferences in a separate "Groups" section
  5. Implement `Ctrl+G` to create new conference with type selection modal (text/av)
  6. Handle incoming conference invitations in modal (accept/reject)
  7. Display peer list in chat panel header when viewing conferences
  8. **Validation**: Create conference, invite friend, exchange messages bidirectionally

---

## File Transfer Support

- **Stated Goal**: Not explicitly claimed in README.md, but the original PR #1 specification includes `tox.FileSend()` and `tox.FileControl()` in the API surface to use.
- **Current State**: No file transfer implementation exists. The Tox protocol supports secure peer-to-peer encrypted file transfer via `FileSend()`, `FileControl()`, and related callbacks.
- **Impact**: **Medium feature limitation**. File sharing is a common messaging feature. Privacy-conscious users may prefer Tox's encrypted P2P transfer over third-party services.
- **Closing the Gap**:
  1. Add file transfer event types to `internal/tox/types.go`:
     ```go
     type FileRecvRequestEvent struct {
         FriendID uint32
         FileID   uint32
         Kind     uint32
         Size     uint64
         Filename string
     }
     type FileRecvChunkEvent struct {
         FriendID uint32
         FileID   uint32
         Position uint64
         Data     []byte
     }
     type FileControlEvent struct {
         FriendID uint32
         FileID   uint32
         Control  toxcore.FileControl
     }
     ```
  2. Register `OnFileRecv*` callbacks in `client.go:registerCallbacks()`
  3. Add wrapper methods to Client:
     - `FileSend(friendID uint32, filename string, data []byte) (uint32, error)`
     - `FileControl(friendID, fileID uint32, control FileControlType) error`
  4. Implement `/file <path>` command in chat input parsing
  5. Create file transfer progress indicator (show in chat as special message type or in status bar)
  6. Handle incoming file requests in modal (accept/reject, choose save location)
  7. Store received files to `~/.config/mtox/downloads/` with collision-safe naming
  8. **Validation**: Send file to friend, receive file from friend, verify file integrity with checksum

---

## Command-Line Options

- **Stated Goal**: Not explicitly stated, but standard practice for CLI applications. Users expect `-h/--help` to show usage information.
- **Current State**: `cmd/mtox/main.go` has no flag parsing. Running `./mtox --help` starts the application instead of showing usage. All configuration is via environment variables (`MTOX_ANON_ONLY`, `MTOX_DISABLE_TOR`, `MTOX_DISABLE_I2P`), which are documented in the README but not discoverable at runtime.
- **Impact**: **Low usability gap**. Users cannot discover configuration options without reading documentation.
- **Closing the Gap**:
  1. Add `flag` package to `cmd/mtox/main.go`
  2. Implement flags:
     ```go
     var (
         showHelp   = flag.Bool("help", false, "Show usage information")
         showVer    = flag.Bool("version", false, "Show version")
         anonOnly   = flag.Bool("anon-only", false, "Enable anon-only mode")
         noTor      = flag.Bool("no-tor", false, "Disable Tor support")
         noI2P      = flag.Bool("no-i2p", false, "Disable I2P support")
         profile    = flag.String("profile", "", "Custom profile path")
     )
     flag.Parse()
     ```
  3. Map flags to environment variables for consistency:
     - `--anon-only` → `MTOX_ANON_ONLY=1`
     - `--no-tor` → `MTOX_DISABLE_TOR=1`
     - `--no-i2p` → `MTOX_DISABLE_I2P=1`
  4. Add version constant to a `version.go` file
  5. **Validation**: Run `./mtox --help` and confirm usage is displayed; verify `./mtox --version` shows version

---

## Voice/Video Call Support

- **Stated Goal**: Not claimed in README.md.
- **Current State**: No AV-related code exists. The Tox protocol supports `tox_av_*` APIs for real-time audio and video communication.
- **Impact**: **Low for TUI context**. Voice/video calls are less common in terminal applications, and the TUI paradigm doesn't lend itself well to video. Audio-only calls might be feasible with terminal-only controls.
- **Closing the Gap**:
  1. Research opd-ai/toxcore AV API availability and maturity
  2. Implement audio-only call support first (more TUI-appropriate):
     - Call initiation command: `/call <friend>` or `Ctrl+O`
     - Call status in status bar: "🎤 In call with Alice (0:42)"
     - Handle incoming calls with accept/reject modal
     - Implement audio routing (microphone → Tox → speaker) using platform audio APIs
  3. Display call controls in chat area or dedicated panel
  4. Consider video support via ASCII art rendering or sixel graphics (terminal-dependent, very advanced)
  5. **Validation**: Initiate call, confirm audio flows bidirectionally

---

## staticcheck/golangci-lint in CI

- **Stated Goal**: Implicit — CI should catch more issues than just `go vet`.
- **Current State**: CI runs `go build`, `go test -race`, and `go vet`. No additional static analysis tools are configured.
- **Impact**: **Low quality risk**. Additional linters could catch issues like unused code, shadowed variables, and style inconsistencies before merge.
- **Closing the Gap**:
  1. Add `staticcheck` step to `.github/workflows/ci.yml`:
     ```yaml
     - name: Install staticcheck
       run: go install honnef.co/go/tools/cmd/staticcheck@latest
     - name: Staticcheck
       run: staticcheck ./...
     ```
  2. Consider adding `golangci-lint` for comprehensive multi-linter analysis
  3. Add coverage threshold enforcement:
     ```yaml
     - name: Test with coverage
       run: go test -coverprofile=coverage.out ./...
     - name: Check coverage threshold
       run: |
         COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%')
         if (( $(echo "$COVERAGE < 30" | bc -l) )); then
           echo "Coverage $COVERAGE% is below 30% threshold"
           exit 1
         fi
     ```
  4. **Validation**: Push to branch, verify new CI steps pass

---

## Release Workflow

- **Stated Goal**: Implied by having a version number (v0.1) — users should be able to download pre-built binaries.
- **Current State**: No release workflow in `.github/workflows/`. Users must build from source.
- **Impact**: **Low adoption friction**. Providing binaries improves accessibility for users who don't have Go installed.
- **Closing the Gap**:
  1. Create `.github/workflows/release.yml`:
     ```yaml
     name: Release
     on:
       push:
         tags: ['v*.*.*']
     jobs:
       build:
         strategy:
           matrix:
             os: [linux, darwin, windows]
             arch: [amd64, arm64]
         runs-on: ubuntu-latest
         steps:
           - uses: actions/checkout@v4
           - uses: actions/setup-go@v5
             with:
               go-version: '1.24'
           - run: GOOS=${{ matrix.os }} GOARCH=${{ matrix.arch }} go build -o mtox-${{ matrix.os }}-${{ matrix.arch }} ./cmd/mtox
           - uses: softprops/action-gh-release@v1
             with:
               files: mtox-*
     ```
  2. Consider using GoReleaser for more sophisticated release automation
  3. Add build status badge to README.md
  4. **Validation**: Tag a release (e.g., `v0.1.0`), verify binaries are published to GitHub Releases

---

## Summary

| Gap | Severity | User Impact | Effort | Priority |
|-----|----------|-------------|--------|----------|
| Test coverage | HIGH | Maintenance risk | Medium | 1 |
| Group chat | MEDIUM | Feature limitation | High | 2 |
| File transfer | MEDIUM | Feature limitation | High | 3 |
| Command-line options | LOW | Discoverability | Low | 4 |
| CI staticcheck | LOW | Quality gates | Low | 5 |
| Release workflow | LOW | Adoption | Low | 6 |
| Voice/video calls | LOW | Feature limitation | Very High | 7 |

**Recommended priority order**: Test coverage → CI improvements → Command-line options → Group chat → File transfer → Release workflow → Voice/video

The core messaging functionality (15/15 documented features) is fully implemented and working. The gaps identified are primarily around advanced Tox protocol features (group chat, file transfer, AV calls) that were not explicitly promised in the README but would bring mtox to feature parity with mature clients like qTox and uTox.

---

*Generated: 2026-03-24 | Based on go-stats-generator analysis and manual code review*

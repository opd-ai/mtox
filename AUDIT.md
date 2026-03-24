# AUDIT — 2026-03-24

## Project Goals

**What it claims to do**: mtox is a "full-featured Tox Messenger terminal user interface (TUI)" that provides secure peer-to-peer messaging with:
- Two-pane interface (contacts + chat)
- Automatic Tor and I2P anonymity network support
- Real-time status indicators
- Typing indicators
- Friend request handling
- Profile persistence

**Target audience**: Privacy-conscious users who prefer terminal-based interfaces and want secure, decentralized messaging with optional anonymity network routing.

**Documented features** (from README.md):
1. Two-pane TUI: contact list on the left, chat on the right
2. Real-time friend status indicators (online / offline)
3. Scrollable per-friend chat history with timestamps
4. Typing indicators
5. Incoming friend request dialog (accept / reject)
6. Add friend dialog (`Ctrl+N`)
7. Profile persistence at `~/.config/mtox/profile.tox`
8. Mouse support (click to select contacts / focus chat)
9. Graceful shutdown with auto-save
10. Automatic Tor and I2P support when services are available
11. Simultaneous Tor + I2P operation
12. Status bar shows Tor/I2P indicators
13. `MTOX_ANON_ONLY` mode for reduced clearnet exposure
14. `MTOX_DISABLE_TOR` and `MTOX_DISABLE_I2P` environment variables

## Goal-Achievement Summary

| Goal | Status | Evidence |
|------|--------|----------|
| Two-pane TUI (contacts + chat) | ✅ Achieved | `internal/tui/app.go:490-491` — `JoinHorizontal(contactsView, chatView)` |
| Real-time friend status indicators | ✅ Achieved | `internal/tui/contacts.go:220-235` — `statusIndicator()` renders ●/○/◌/◉ |
| Scrollable per-friend chat history | ✅ Achieved | `internal/tui/app.go:58` — `historyByFriend map[uint32][]chatMessage` |
| Typing indicators | ✅ Achieved | `internal/tui/chat.go:101-104` — displays typing state; `app.go:277-282` — optimized to state transitions only |
| Incoming friend request dialog | ✅ Achieved | `internal/tui/app.go:414-420`, `445-461` — modal with accept/reject via Enter/R keys |
| Add friend dialog (Ctrl+N) | ✅ Achieved | `internal/tui/app.go:405-411`, `429-441` — modal input and submission |
| Profile persistence at ~/.config/mtox | ✅ Achieved | `internal/tox/client.go:70-75` loads, `client.go:234-245` saves to ProfilePath() |
| Mouse support | ✅ Achieved | `cmd/mtox/main.go:27` — `tea.WithMouseCellMotion()`; `app.go:290-308` handles clicks |
| Graceful shutdown with auto-save | ✅ Achieved | `internal/tui/app.go:398-401` — `quit()` calls `Save()` then `Stop()` |
| Automatic Tor support | ✅ Achieved | `internal/tox/anonymity.go:187-248` — `initTor()` with retry loop |
| Automatic I2P support | ✅ Achieved | `internal/tox/anonymity.go:259-318` — `initI2P()` with retry loop |
| Simultaneous Tor + I2P | ✅ Achieved | `internal/tox/anonymity.go:86-91` — both run in parallel goroutines |
| Status bar shows Tor/I2P indicators | ✅ Achieved | `internal/tui/statusbar.go:78-101` — shows 🧅Tor and 🧄I2P |
| MTOX_ANON_ONLY mode | ✅ Achieved | `internal/tox/client.go:47-68` — disables UDP/IPv6/local discovery |
| MTOX_DISABLE_TOR/I2P variables | ✅ Achieved | `internal/tox/anonymity.go:189-204`, `261-275` — check environment |
| Group chat (conferences) | ❌ Missing | `internal/tui/app.go:234-235` — returns "not yet supported" notification |
| File transfers | ❌ Missing | No implementation found in codebase |
| Voice/video calls | ❌ Missing | No AV-related code in codebase |
| Command-line options (--help) | ❌ Missing | `cmd/mtox/main.go` has no flag parsing |

**Overall: 15/19 stated goals fully achieved (79%), 4 missing features (not claimed in README)**

## Findings

### CRITICAL

*None identified* — All documented features are functional.

### HIGH

- [ ] **LOW-H1: Test coverage below maintainability threshold** — `go test -cover ./...` — Test coverage is 0% for `cmd/mtox`, 5.9% for `internal/tox`, and 14.8% for `internal/tui`. While test files exist, coverage is insufficient for safe refactoring. — **Remediation:** Expand test suites to >40% coverage on critical paths. Add tests for `Client.emit()`, `App.Update()` ToxEvent handling, and modal state transitions. Validate with `go test -cover ./... | grep -E "tox|tui"`.

- [ ] **LOW-H2: Bootstrap nodes may be stale** — `internal/tox/client.go:26-35` — The hardcoded bootstrap nodes list has no mechanism for updates and nodes may become unavailable over time. — **Remediation:** Either (a) fetch a current bootstrap nodes list from `nodes.tox.chat/json` at startup, or (b) add a fallback mechanism that logs when all nodes fail and suggests manual intervention. Validate by temporarily corrupting node addresses and verifying warning is logged.

### MEDIUM

- [ ] **LOW-M1: Duplication in Tor/I2P initialization** — `internal/tox/anonymity.go:237-246` and `anonymity.go:308-317` — go-stats-generator detected 10 duplicate lines (0.49% ratio) in the success handling paths for both network initializers. — **Remediation:** Extract common success handling into a helper method like `setTransportAvailable(network, transport, listener, addr)`. Validate with `go-stats-generator analyze . --sections duplication | grep -i "clone"` showing 0 clones.

- [ ] **LOW-M2: High cyclomatic complexity in handleKey** — `internal/tui/app.go:217` — Complexity score 19.7 (threshold: 15) due to extensive switch statement handling keyboard input. — **Remediation:** This is justified for a TUI state machine but consider extracting modal-specific key handling into separate methods (`handleAddFriendKey`, `handleFriendRequestKey`) for improved testability. Validate function has <15 cyclomatic complexity after refactoring.

- [ ] **LOW-M3: iterateLoop complexity** — `internal/tox/client.go:163` — Complexity score 18.4 due to timer management and channel selection. — **Remediation:** Current implementation is justified and correctly uses `time.Timer` to avoid per-iteration allocations. No action required; complexity is inherent to the event loop pattern.

- [ ] **LOW-M4: Method documentation coverage gap** — `go-stats-generator` — Method coverage is 76.5% while overall documentation is 84.9%. 23.5% of methods lack documentation comments. — **Remediation:** Add doc comments to unexported methods in `internal/tui/` that implement complex logic (e.g., `renderContact`, `handleIncomingMessage`, `selectFriend`). Validate with `go-stats-generator analyze . --sections documentation`.

### LOW

- [ ] **LOW-L1: Magic numbers in styles.go** — `internal/tui/styles.go:10-92` — Color codes like "62", "240", "229" are used without named constants. — **Remediation:** Define named color constants at the top of styles.go for maintainability (e.g., `colorPanelBorder = lipgloss.Color("62")`). Low priority as this is stylistic.

- [ ] **LOW-L2: Naming convention: ToxEvent** — `internal/tox/types.go:7` — Type `ToxEvent` uses package name as prefix (stuttering: `tox.ToxEvent`). — **Remediation:** Consider renaming to `Event` since the package already provides namespace. Low priority as this is a breaking change.

- [ ] **LOW-L3: No version constant** — `cmd/mtox/main.go`, `internal/tui/statusbar.go:29` — Version "v0.1" is hardcoded in the status bar with no way to update programmatically. — **Remediation:** Define `const Version = "0.1.0"` in a `version.go` file and reference from statusbar. Enables future `--version` flag.

- [ ] **LOW-L4: Unused setFriend method** — `internal/tui/chat.go:52-60` — The `setFriend()` method clears history and is superseded by `setFriendWithHistory()`. Tests still use it but it's never called from production code. — **Remediation:** Remove `setFriend()` and update tests to use `setFriendWithHistory()` directly, or keep for testing convenience with a comment.

- [ ] **LOW-L5: Bubbletea/Lipgloss v1 in use** — `go.mod:8-11` — Project uses Bubble Tea v1.3.10 and Lipgloss v1.1.0. v2 releases are available with improved overlay compositing and I/O handling. — **Remediation:** Plan migration to v2 when stable. Current v1 versions are functional; this is a future consideration for improved TUI rendering.

## Metrics Snapshot

| Metric | Value | Assessment |
|--------|-------|------------|
| Total Lines of Code | 1,047 | Compact codebase |
| Functions/Methods | 108 | Well-factored |
| Average Function Length | 11.3 lines | Good (threshold: <15) |
| Functions >50 lines | 5 (4.6%) | Acceptable |
| High Complexity (>10) | 3 functions | Within acceptable range for TUI |
| Documentation Coverage | 84.9% | Good |
| Package Coverage | 100% | Excellent |
| Function Coverage | 100% | Excellent |
| Method Coverage | 76.5% | Adequate |
| Duplication Ratio | 0.49% | Excellent (threshold: <5%) |
| Circular Dependencies | 0 | None |
| Test Coverage (tox) | 5.9% | Needs improvement |
| Test Coverage (tui) | 14.8% | Needs improvement |
| Test Coverage (cmd) | 0% | Expected for entrypoint |
| Go Vet Issues | 0 | Clean |
| Race Conditions | 0 | Clean (tested with `-race`) |

### High-Complexity Functions

| Function | File | Lines | Complexity | Justification |
|----------|------|-------|------------|---------------|
| handleKey | internal/tui/app.go:217 | 69 | 19.7 | Keyboard input state machine — inherent complexity |
| iterateLoop | internal/tox/client.go:163 | 33 | 18.4 | Event loop with timer management — justified |
| overlayModal | internal/tui/app.go:519 | 66 | 15.3 | Modal rendering logic — justified |
| retryWithBackoff | internal/tox/anonymity.go:331 | 28 | 14.0 | Consolidated retry logic — well-structured |
| Update | internal/tui/app.go:103 | 79 | 13.2 | Bubbletea Update() — inherent complexity |

### Dependency Analysis

**Direct dependencies** (from go.mod):
- `github.com/charmbracelet/bubbles v1.0.0` — TUI components ✅ Maintained
- `github.com/charmbracelet/bubbletea v1.3.10` — TUI framework ✅ Maintained  
- `github.com/charmbracelet/lipgloss v1.1.0` — Styling ✅ Maintained
- `github.com/opd-ai/toxcore v0.0.0-20260306...` — Tox protocol ✅ Active

**Known issues with dependencies**:
- opd-ai/toxcore has scalability limitations (2,048 node DHT limit, single-threaded event loop) but these do not affect mtox's target use case as a single-user TUI client.
- Bubbletea/Lipgloss v2 offers improved overlay compositing but v1 is stable and functional.

### CI/CD Assessment

**Current CI** (`.github/workflows/ci.yml`):
- ✅ Build verification (`go build ./cmd/mtox`)
- ✅ Test with race detector (`go test -race ./...`)
- ✅ Static analysis (`go vet ./...`)

**Missing from CI**:
- ❌ Coverage threshold enforcement
- ❌ staticcheck/golangci-lint
- ❌ Release workflow for binaries

---

*Generated: 2026-03-24 | go-stats-generator v1.0.0 | Audit performed against commit on main branch*

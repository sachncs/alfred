# Changelog

All notable changes to Alfred are documented here.

## [0.3.0-phase3] - 2026-07-23

### Added
- SQLite-backed persistence via `modernc.org/sqlite` (pure Go, no CGO)
- Schema migrations with `0001_initial.sql` (threads, sessions, turns, events, usage, settings tables)
- Hybrid thread store combining SQLite index with JSONL body
- Thread, session, turn, event, and usage SQL row implementations
- Settings SQL row with namespace+key addressing
- Retention pruning by age and count
- JSONL-to-hybrid migration (`migrate_jsonl.go`)
- Legacy Kun config migration (`migrate_legacy_kun.go`)
- `--db-path` CLI flag for custom database location

### Changed
- Health endpoint consolidated to `GET /v1/health` (removed `/health` and `/healthz`)
- Replaced custom `itoa` with `strconv.Itoa` in compactor and history hygiene

### Fixed
- Health response docstring updated to reference `/v1/health`

## [0.2.0-phase2] - 2026-07-22

### Added
- HTTP/SSE runtime (`HTTPRuntime`, `LocalRuntime`, `AlfredRuntime`)
- Turn execution loop with tool dispatch (max 20 iterations)
- Compaction (normal/aggressive/force modes)
- Token economy and history hygiene (ANSI strip, long-result trim)
- Steering queue for safe-boundary draining
- Sub-agent delegation hooks
- Usage counter and cache telemetry
- Tool budget profiles (explanation/review/implementation/scientific/long)
- Prompt cache with immutable prefix verification
- Memory store (user/workspace/project scopes)
- Skills loader (SKILL.md parsing)
- All built-in tools: Read, Write, Edit, ApplyPatch, Bash, Grep, Find, Ls
- Code review service
- Fork and resume semantics
- Goals and todos per thread
- SSE event streaming with replay buffer
- Thread, session, turn, approval, user-input, memory, skills, attachment, workspace, usage routes
- File thread store with atomic writes
- File session store (JSONL events)
- Bearer auth middleware

## [0.1.0-phase1] - 2026-07-22

### Added
- Go module initialization with `go.mod`, `Makefile`, `.golangci.yml`
- Core contract types: `Thread`, `Turn`, `Item`, `Event`, `Error`
- Tool interface with `Context` and `Result`
- ReadTool implementation
- Agent interface with `AgentState`
- ChatAgent implementation
- MCPServer interface with stdio transport
- EchoServer hello-world MCP worker
- Capability interface with Broker
- `cmd/alfred` binary with SIGINT handling
- Test harness with testify
- Lint configuration (gofmt, govet, errcheck, staticcheck, unused, gosimple, revive)
- Architecture README

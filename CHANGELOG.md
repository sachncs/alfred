# Changelog

All notable changes to Alfred are documented here.

## [0.7.0-phase7] - 2026-09-11

### Added
- Research workers: `paper_radar`, `multi_agent`, `sci_modality`, `evidence_dag`, `project_dag`, `feedback_gateway`, `bgc_discovery`
- Content / media workers: `image_generation`, `scientific_plotting`, `visual_document`, `ppt_master`
- Workspace previews: `workspace_intel`, `workspace_bioimaging`, `workspace_deck`, `workspace_molecular`, `workspace_omics`, `workspace_sequence`, `workspace_spectra`, `workspace_tabular`
- System workers: `workflow`, `schedule`, `remote_executor`, `gui_owl`
- Phase 5 workers: `search`, `model_router`, `plan_gateway`, `write_assist`, `inspector`
- Capability broker (`internal/capability/broker.go`) and resource / file bindings
- Discovery + Observer for capability lifecycle
- `feedback_gateway` worker wired into the orchestrator
- `docs/ROADMAP.md` and `docs/workers/` per-worker reference

### Changed
- `model_router` truncates upstream error bodies to 256 bytes and logs the full body through the redacting writer
- `doGet` / `doPost` in `internal/search/providers` read the full HTTP body via `io.ReadAll` + `io.LimitReader`
- `BearerAuth` uses `crypto/subtle.ConstantTimeCompare`
- `generateID` returns a UUIDv4 sourced from `crypto/rand`
- Default `URLPolicy` is https-only; permissive variant is `NewURLPolicyPermissive`

### Fixed
- `ReadOnlyFileCapability.Invoke` and `Binding.Bind` forward the caller `ctx`
- `HybridThreadStore.AppendEvents` is now atomic via JSONL write-ahead
- `SettingsStore.Load` propagates parse errors and logs via `internal/log`
- SSE replay buffer is capped at 256 events per thread
- `cmd/alfred` startup log surfaces SQLite persistence failures as a `turn.persistence_error` SSE event

## [0.6.0-phase6] - 2026-09-04

### Added
- `AppSettingsV1` settings persistence via SQLite settings table
- `SettingsStore` with `Load`, `Save`, `Patch`, `Delete`
- `GET /v1/settings`, `POST /v1/settings`, `PATCH /v1/settings` routes

## [0.5.0-phase5] - 2026-09-01

### Added
- Capability broker with `Register`, `Dispatch`, `Get`, `List`
- `Function` capability (JSON-in / JSON-out adapter)
- Worker registration for `search`, `model_router`, `plan_gateway`, `write_assist`, `inspector`
- `GET /v1/capabilities` route

## [0.4.0-phase4] - 2026-08-15

### Added
- Web UI under `/` and `/design` (htmx + Go html/template)
- Sidebar SSE listener (`web/static/js/sidebar-sse.js`)
- Web handlers for threads, sessions, settings, design

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

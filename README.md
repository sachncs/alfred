# Alfred

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
[![CI](https://img.shields.io/github/actions/workflow/status/sachncs/alfred/ci.yml?branch=master&label=ci)](./.github/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/sachncs/alfred)](https://goreportcard.com/report/github.com/sachncs/alfred)

**A local-first research agent runtime for scientists, engineers, and tool builders.**

Alfred runs an HTTP/SSE turn loop on your machine, persists every thread to
SQLite, and ships with 25+ MCP workers — search, plotting, evidence DAG,
image generation, scientific computing, and more. No external services
required for the local runtime.

## Quick start

```bash
git clone https://github.com/sachncs/alfred
cd alfred
make build
./bin/alfred --port 8899
```

## What's inside

- **Turn loop** with structured tool calling and tool budgets.
- **Persistence** via SQLite + JSONL (hybrid thread store, WAL, foreign keys).
- **Web UI** under `/` (htmx + Go html/template).
- **Workers**: search, paper radar, multi-agent, image generation, evidence DAG, project DAG, scientific plotting, visual documents, schedule, workflow, and 15 more — see [`docs/workers/`](./docs/workers/README.md).

## Architecture

Alfred uses a four-level OOP hierarchy. Every interface has an abstract base (level 1), a package-level base that fills in shared wiring (level 2), one or more specializations (level 3), and one Alfred-specific implementation (level 4). Inheritance is via embedding (Go's composition), and polymorphism is via interface satisfaction.

```
level 1 (abstract)        Tool, Agent, MCPServer, Capability, Runtime
level 2 (base)            FileSystemTool, ChatAgent, WorkerServer, Function, HTTPRuntime
level 3 (specialization)  ReadTool,                                  ModelRouterWorker
level 4 (Alfred-specific)                                      AlfredRuntime
```

### Layering

```
contract    ← wire types (Thread, Turn, Item, Error, wire schemas)
interface   ← Tool, Agent, MCPServer, Capability, Runtime
abstract    ← FileSystemTool, ChatAgent, WorkerServer, Function
base        ← ReadTool,                                  HTTPRuntime
specialization
Alfred-specific
```

Each layer only depends on the layer(s) above it. `cmd/alfred/main.go` is the only file that imports every package.

### Package layout

```
alfred/
├── go.mod                       # module github.com/alfred/alfred (Go 1.22+)
├── go.sum
├── Makefile                     # build / test / vet / lint / run / clean
├── .golangci.yml                # golangci-lint v2 config
├── README.md
├── CHANGELOG.md
├── cmd/alfred/main.go           # orchestrator with SIGINT handling
├── internal/
│   ├── contract/                # shared wire types (Thread, Turn, Item, Error)
│   │   └── testutil/            # fake builders for tests
│   ├── tool/                    # Tool interface, Context, Result
│   ├── fs/                      # FileSystemTool + Read/Write/Edit/ApplyPatch/Ls
│   ├── exec/                    # BashTool (shell execution with timeout)
│   ├── search/                  # GrepTool, FindTool
│   ├── agent/                   # Agent interface, AgentState, ChatAgent
│   ├── model/                   # model.Client interface, StubClient
│   ├── mcp/                     # MCPServer, StdioTransport
│   │   └── worker/              # WorkerServer abstract, EchoServer concrete
│   ├── capability/              # Capability interface, Broker, Function
│   ├── runtime/                 # HTTPRuntime, LocalRuntime, AlfredRuntime, turn loop
│   ├── store/                   # SQLite + JSONL persistence, hybrid thread store
│   ├── config/                  # CLI flags + env var configuration
│   └── log/                     # Structured logging with secret redaction
└── bin/alfred                   # built binary (gitignored)
```

## Run

```bash
# from ./alfred/
make build    # produces ./bin/alfred
make test     # runs go test -race -count=1 ./...
make vet      # runs go vet ./...
make lint     # runs golangci-lint run ./...
make run      # build + run
make clean    # remove ./bin/
```

## Phase 7 acceptance (verified)

```bash
$ make build
Built bin/alfred

$ make test
... all packages OK under -race

$ make vet
go vet clean

$ make lint
0 issues.

$ ./bin/alfred --port 8899
alfred 0.7.0-phase7 starting on port 8899 (db=~/.alfred/alfred.db)
mcp worker ready: id=echo-worker tools=1
mcp worker ready: id=mcp.search tools=...
mcp worker ready: id=mcp.model_router tools=...
... (25+ workers)
hybrid store ready: sqlite=~/.alfred/alfred.db jsonl=~/.alfred/events
settings loaded: theme=dark model=claude-sonnet-4-20250514
chat agent ready: id=alfred.chat tools=3
runtime ready
web ui ready: / and /design
capability routes ready: GET /v1/capabilities
settings routes ready: GET/POST/PATCH /v1/settings
smoke test passed
ready (Ctrl-C to exit)
READY
```

### Health check

```bash
$ curl http://127.0.0.1:8899/v1/health
{"status":"ok","version":"0.7.0-phase7","timestamp":"..."}
```

### Create a thread

```bash
$ curl -X POST http://127.0.0.1:8899/v1/threads -d '{"title":"research"}'
{"thread":{"id":"...","title":"research","status":"idle",...}}
```

## Recipes

Each recipe combines 2+ MCP workers in a single end-to-end flow. Start
`./bin/alfred --port 8899` first and run the curl block from another shell.

### 1. Research a paper

Prompt: "Summarize arxiv:2401.00001 in 3 bullets and list its top 3 cited works."

Worker chain: `search` → `paper_radar` → `evidence_dag` → `assistant_text`.

```bash
curl -X POST http://127.0.0.1:8899/v1/threads \
  -H 'Content-Type: application/json' \
  -d '{"title":"paper-2401.00001"}'

THREAD=$(curl -s http://127.0.0.1:8899/v1/threads | jq -r '.threads[0].id')

curl -X POST "http://127.0.0.1:8899/v1/threads/$THREAD/turns" \
  -H 'Content-Type: application/json' \
  -d '{"text":"Summarize arxiv:2401.00001 in 3 bullets and list its top 3 cited works."}'
```

Expected output (truncated):

```text
{"event":"turn.completed","data":{"status":"completed","items":[
  {"type":"text","text":"- Bullet 1\n- Bullet 2\n- Bullet 3"},
  {"type":"tool","tool":"paper_radar.paper_search","output":"..."},
  {"type":"tool","tool":"evidence_dag.evidence_update","output":"..."}
]}}
```

### 2. Write a code review

Prompt: "Review `internal/runtime/loop.go` for race conditions and summarize."

Worker chain: `fs_read` → `fs_read` → `assistant_text`.

```bash
curl -X POST "http://127.0.0.1:8899/v1/threads/$THREAD/turns" \
  -H 'Content-Type: application/json' \
  -d '{"text":"Review internal/runtime/loop.go for race conditions."}'
```

### 3. Generate a figure

Prompt: "Plot a histogram of `x` from `data.csv`."

Worker chain: `fs_read` → `scientific_plotting` → `workspace_tabular`.

```bash
curl -X POST "http://127.0.0.1:8899/v1/threads/$THREAD/turns" \
  -H 'Content-Type: application/json' \
  -d '{"text":"Plot a histogram of column x from data.csv."}'
```

### 4. Draft a slide deck

Prompt: "Make a 5-slide deck on transformers for a research meeting."

Worker chain: `write_assist` → `ppt_master` → `workspace_deck`.

```bash
curl -X POST "http://127.0.0.1:8899/v1/threads/$THREAD/turns" \
  -H 'Content-Type: application/json' \
  -d '{"text":"Make a 5-slide deck on transformers."}'
```

## What's complete (through Phase 7)

- **Phase 1**: OOP hierarchy (Tool, Agent, MCPServer, Capability), ChatAgent, EchoServer, binary boot
- **Phase 2**: HTTP/SSE runtime, turn loop with tool dispatch, compaction, token economy, history hygiene, steering queue, sub-agent delegation, usage tracking, tool budgets, prompt cache, memory store, skills loader, built-in tools (Read/Write/Edit/ApplyPatch/Bash/Grep/Find/Ls), code review, fork/resume, goals/todos, all HTTP routes
- **Phase 3**: SQLite persistence (pure Go via modernc.org/sqlite), schema migrations, hybrid thread store (SQLite index + JSONL body), retention pruning, JSONL migration, legacy Kun config migration
- **Phase 4**: Web UI under `/` and `/design` (htmx + Go html/template)
- **Phase 5**: Capability broker + initial workers (`search`, `model_router`, `plan_gateway`, `write_assist`, `inspector`)
- **Phase 6**: Settings persistence (`AppSettingsV1`) via SQLite settings table
- **Phase 7**: Research / content / workspace workers (paper radar, multi-agent, evidence DAG, image generation, scientific plotting, PPT, schedule, workflow, remote executor, workspace previews)

## What's NOT done yet

Phase 8+ layers on:

- UX features — plan mode, write mode, anchored comments, multi-thread views — Phase 8
- Remote channel runtime, cutover from `SciForge` — Phase 9

See [`docs/ROADMAP.md`](./docs/ROADMAP.md) for the full plan and [`docs/workers/`](./docs/workers/README.md) for the per-worker reference.

# Alfred

A research workbench: a local agent runtime with structured tool calling, persistent threads, and a fleet of MCP workers (search, plotting, evidence DAG, image generation, and more).

Phase 1 ships the foundation: the OOP hierarchy (Tool, Agent, MCPServer, Capability), one hello-world concrete implementation of each, and a `cmd/alfred` binary that boots, runs, and shuts down cleanly.

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
├── cmd/alfred/main.go           # hello-world orchestrator
├── internal/
│   ├── contract/                # shared wire types (Thread, Turn, Item, Error)
│   │   └── testutil/            # fake builders for tests
│   ├── tool/                    # Tool interface, Context, Result
│   ├── fs/                      # FileSystemTool abstract, ReadTool concrete
│   ├── agent/                   # Agent interface, AgentState, ChatAgent
│   ├── model/                   # model.Client interface, StubClient
│   ├── mcp/                     # MCPServer, StdioTransport
│   │   └── worker/              # WorkerServer abstract, EchoServer concrete
│   └── capability/              # Capability interface, Broker, Function
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

## Phase 1 acceptance (verified)

```bash
$ make build
Built bin/alfred

$ ./bin/alfred --version
alfred 0.1.0-phase1

$ make test
... 9 packages, all OK under -race

$ make vet
go vet clean

$ make lint
0 issues.

$ ./bin/alfred
2026/07/22 11:20:11 alfred 0.1.0-phase1 starting on port 8899 (placeholder)
2026/07/22 11:20:11 mcp worker ready: id=echo-worker tools=1
2026/07/22 11:20:11 chat agent ready: id=alfred.chat tools=1
2026/07/22 11:20:11 smoke test passed: agent produced a turn end-to-end
2026/07/22 11:20:11 ready (Ctrl-C to exit)
READY
^C
2026/07/22 11:20:13 received signal interrupt; shutting down
$ echo $?
0
```

## What's NOT in Phase 1

Phase 1 is the foundation only. Phase 2+ layers on:

- HTTP/SSE runtime (currently just a blocking binary with SIGINT handling)
- SQLite + JSONL persistence (currently in-memory)
- Web UI (htmx + Go html/template)
- All 28 MCP workers (currently only EchoServer)
- All 7 neumorphic UI primitives (only available in the SciForge codebase)
- DuckDuckGo search, model-router, plan-gateway, write-assist, runtime-inspector
- Remote channel runtime, settings persistence, anchored comments, write mode, SDD

See `todo/` for the full plan.

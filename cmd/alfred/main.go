package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/alfred/alfred/internal/agent"
	"github.com/alfred/alfred/internal/capability"
	"github.com/alfred/alfred/internal/config"
	"github.com/alfred/alfred/internal/contract"
	"github.com/alfred/alfred/internal/fs"
	"github.com/alfred/alfred/internal/mcp/worker"
	modelrouter "github.com/alfred/alfred/internal/mcp/worker/model_router"
	plangateway "github.com/alfred/alfred/internal/mcp/worker/plan_gateway"
	runtimeinspector "github.com/alfred/alfred/internal/mcp/worker/runtime_inspector"
	searchworker "github.com/alfred/alfred/internal/mcp/worker/search"
	writeassist "github.com/alfred/alfred/internal/mcp/worker/write_assist"
	"github.com/alfred/alfred/internal/model"
	"github.com/alfred/alfred/internal/runtime"
	"github.com/alfred/alfred/internal/store"
	"github.com/alfred/alfred/web"
)

const version = "0.5.0-phase5"

func main() {
	cfg := config.Defaults()
	showVersion := config.RegisterFlags(&cfg)
	flag.Parse()

	if *showVersion {
		fmt.Printf("alfred %s\n", version)
		return
	}

	cfg.ApplyEnv()
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Printf("alfred %s starting on port %d (db=%s)", version, cfg.Port, cfg.DBPath)

	// 1. Capability broker
	broker := capability.NewBroker()

	// 2. Hello-world MCP worker (EchoServer).
	echoServer := worker.NewEchoServer()
	log.Printf("mcp worker ready: id=%s tools=%d", echoServer.ID(), len(echoServer.Tools()))

	broker.Register(capability.NewFunction(
		"mcp.echo.status",
		func(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
			out := map[string]any{
				"id":    echoServer.ID(),
				"tools": len(echoServer.Tools()),
			}
			b, _ := json.Marshal(out)
			return b, nil
		},
	))

	// 2b. Register 5 critical MCP workers with the broker.
	workers := map[string]*worker.WorkerServer{
		"mcp.search":       searchworker.NewSearchWorker(),
		"mcp.model_router": modelrouter.NewModelRouterWorker(),
		"mcp.plan_gateway": plangateway.NewPlanGatewayWorker(),
		"mcp.write_assist": writeassist.NewWriteAssistWorker(),
		"mcp.inspector":    runtimeinspector.NewInspectorWorker(),
	}
	for name, ws := range workers {
		ws := ws
		log.Printf("mcp worker ready: id=%s tools=%d", ws.ID(), len(ws.Tools()))
		broker.Register(capability.NewFunction(
			capability.ID(name),
			func(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
				out := map[string]any{
					"id":    ws.ID(),
					"tools": len(ws.Tools()),
				}
				b, _ := json.Marshal(out)
				return b, nil
			},
		))
	}

	// 2c. Discovery + Observer for capability lifecycle.
	discovery := capability.NewDiscovery(broker)
	observer := capability.NewObserver()
	_ = observer

	// 3. Hybrid (SQLite + JSONL) thread store.
	eventsDir := filepath.Join(filepath.Dir(cfg.DBPath), "events")
	hybrid, err := store.NewHybridThreadStore(cfg.DBPath, eventsDir)
	if err != nil {
		log.Fatalf("hybrid store: %v", err)
	}
	defer func() { _ = hybrid.Close() }()
	sqlitePath, jsonlDir := hybrid.Addr()
	log.Printf("hybrid store ready: sqlite=%s jsonl=%s", sqlitePath, jsonlDir)

	// 4. Migration: legacy JSONL + ~/.kun.
	if err := store.MigrateFromJSONL(hybrid, cfg.WorkspaceRoot); err != nil {
		log.Printf("jsonl migration warning: %v", err)
	}
	if home, err := os.UserHomeDir(); err == nil {
		if err := store.MigrateLegacyKun(hybrid.SQLite(), home); err != nil {
			log.Printf("kun migration: %v", err)
		}
	}

	// 5. ChatAgent + tools
	readTool := fs.NewReadTool()
	writeTool := fs.NewWriteTool()
	editTool := fs.NewEditTool()
	// ponytail: ScriptedChatClient defaults to 2 scripts — smoke test + HTTP turn both get a response
	stubClient := model.ScriptedChatClient("hello from alfred")

	chat := agent.NewChatAgent("alfred.chat", "You are Alfred, a helpful research assistant.", stubClient)
	chat.RegisterTool(readTool)
	chat.RegisterTool(writeTool)
	chat.RegisterTool(editTool)
	log.Printf("chat agent ready: id=%s tools=%d", chat.ID(), len(chat.Tools()))

	// 6. Wire bridge
	bridge := runtime.NewAgentBridge(chat)
	_ = bridge

	// 7. Runtime backed by hybrid store
	rt := runtime.NewAlfredRuntimeWithHybrid(cfg.BearerToken, hybrid)
	rt.SetModelClient(stubClient)
	rt.RegisterTool(readTool)
	rt.RegisterTool(writeTool)
	rt.RegisterTool(editTool)
	log.Printf("runtime ready")

	// 8. Web UI (htmx shell)
	webHandlers := web.NewHandlers()
	webHandlers.Threads = rt.ThreadStore()
	webHandlers.Thread = rt.ThreadStore()
	webHandlers.Create = rt.ThreadStore()
	webHandlers.Session = rt.SessionStore()
	webHandlers.Starter = rt
	web.Register(rt.Mux(), webHandlers, "web/static")
	log.Printf("web ui ready: / and /design")

	// 8b. Capability IPC routes.
	capability.RegisterCapabilityRoutes(rt.Mux(), broker, discovery)
	log.Printf("capability routes ready: GET /v1/capabilities")

	// 9. Smoke-test the agent
	if err := smokeTest(chat); err != nil {
		log.Printf("smoke test failed: %v", err)
	} else {
		log.Printf("smoke test passed")
	}

	log.Printf("ready (Ctrl-C to exit)")
	fmt.Println("READY")

	// 10. Start HTTP server in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := rt.Start(ctx, cfg.Port); err != nil {
			log.Printf("http server: %v", err)
		}
	}()

	// 11. Wait for SIGINT / SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	log.Printf("received signal %s; shutting down", sig)
	cancel()
	time.Sleep(100 * time.Millisecond)
}

func smokeTest(a *agent.ChatAgent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	thread := &contract.Thread{ID: "smoke-thr", Title: "smoke", Status: contract.ThreadStatusIdle}
	turn, err := a.StartTurn(ctx, thread, contract.UserInput{Text: "hello"})
	if err != nil {
		return err
	}
	if turn.Status != contract.TurnStatusCompleted {
		return fmt.Errorf("turn status: %s", turn.Status)
	}
	return nil
}

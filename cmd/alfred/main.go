// Command alfred is the Phase 1 hello-world orchestrator.
//
// Boot order:
//  1. parse flags (--port, --version)
//  2. initialize the capability broker
//  3. register the EchoServer MCP worker (with a stub echo tool)
//  4. register ReadTool with the ChatAgent
//  5. start the agent loop on a stub model client
//  6. log "ready" with the listening URL
//  7. wait for SIGINT/SIGTERM and shut down cleanly with exit code 0
//
// Phase 1 does not expose an HTTP server (that lands in Phase 2). The
// binary stays alive so the smoke test can verify signal handling.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alfred/alfred/internal/agent"
	"github.com/alfred/alfred/internal/capability"
	"github.com/alfred/alfred/internal/contract"
	"github.com/alfred/alfred/internal/fs"
	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/model"
)

const version = "0.1.0-phase1"

func main() {
	port := flag.Int("port", 8899, "port placeholder (HTTP server lands in Phase 2)")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("alfred %s\n", version)
		return
	}

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Printf("alfred %s starting on port %d (placeholder)", version, *port)

	// 1. Capability broker
	broker := capability.NewBroker()

	// 2. Hello-world MCP worker (EchoServer).
	echoServer := worker.NewEchoServer()
	log.Printf("mcp worker ready: id=%s tools=%d", echoServer.ID(), len(echoServer.Tools()))

	// 3. Capability: announce the echo worker.
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

	// 4. ReadTool + ChatAgent.
	readTool := fs.NewReadTool()
	stubClient := model.ScriptedChatClient("hello from alfred")

	chat := agent.NewChatAgent("alfred.chat", "You are Alfred, a helpful research assistant.", stubClient)
	chat.RegisterTool(readTool)
	log.Printf("chat agent ready: id=%s tools=%d", chat.ID(), len(chat.Tools()))

	// 5. Smoke-test the agent end-to-end so the boot log is meaningful.
	if err := smokeTest(chat); err != nil {
		log.Printf("smoke test failed: %v", err)
	} else {
		log.Printf("smoke test passed: agent produced a turn end-to-end")
	}

	log.Printf("ready (Ctrl-C to exit)")
	fmt.Println("READY")

	// 6. Wait for SIGINT / SIGTERM, then shut down cleanly.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	log.Printf("received signal %s; shutting down", sig)
}

// smokeTest runs one canned turn through the agent so the boot log proves
// the wiring is correct end-to-end.
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

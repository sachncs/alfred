package worker_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alfred/alfred/internal/mcp"
	"github.com/alfred/alfred/internal/mcp/worker"
)

func newEchoServer() *worker.EchoServer {
	return worker.NewEchoServer()
}

// sendAndRecv sends a JSON-RPC message to the server's transport and
// reads the next response. Uses shared mutexes on both buffers to keep
// the test goroutine race-free with the server goroutine.
func sendAndRecv(t *testing.T, in *lockingBuffer, lockedOut *lockingBuffer, msg string) mcp.JSONRPCResponse {
	t.Helper()
	if _, err := in.WriteString(msg + "\n"); err != nil {
		t.Fatalf("write to in: %v", err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for {
		if time.Now().After(deadline) {
			lockedOut.mu.Lock()
			s := lockedOut.b.String()
			lockedOut.mu.Unlock()
			t.Fatalf("timeout waiting for response to %s; have %q", msg, s)
		}
		lockedOut.mu.Lock()
		s := lockedOut.b.String()
		if strings.Contains(s, "\n") {
			lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
			last := lines[len(lines)-1]
			var resp mcp.JSONRPCResponse
			if err := json.Unmarshal([]byte(last), &resp); err == nil {
				// Reset so future reads see only new lines.
				lockedOut.b.Reset()
				lockedOut.mu.Unlock()
				return resp
			}
		}
		lockedOut.mu.Unlock()
		time.Sleep(5 * time.Millisecond)
	}
}

// lockingBuffer wraps a *bytes.Buffer with a mutex so the test
// goroutine and the server goroutine don't race under -race.
// Implements io.Reader so it can be used as a stdio source too.
type lockingBuffer struct {
	mu *sync.Mutex
	b  *bytes.Buffer
}

func (l *lockingBuffer) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}

func (l *lockingBuffer) WriteString(s string) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.WriteString(s)
}

func (l *lockingBuffer) Read(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Read(p)
}

func (l *lockingBuffer) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.String()
}

func TestEchoServerInitialize(t *testing.T) {
	t.Parallel()
	var inMu sync.Mutex
	in := &bytes.Buffer{}
	lockedIn := &lockingBuffer{mu: &inMu, b: in}
	var outMu sync.Mutex
	out := &bytes.Buffer{}
	lockedOut := &lockingBuffer{mu: &outMu, b: out}
	tr := mcp.NewStdioTransport(lockedIn, lockedOut)
	srv := newEchoServer()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = srv.Start(context.Background(), tr)
	}()

	resp := sendAndRecv(t, lockedIn, lockedOut, `{"jsonrpc":"2.0","id":1,"method":"initialize"}`)
	res, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", resp.Result)
	}
	if res["serverInfo"].(map[string]any)["name"] != "echo-worker" {
		t.Fatalf("serverInfo wrong: %v", res)
	}
	if res["protocolVersion"] != "2024-11-05" {
		t.Fatalf("protocolVersion wrong: %v", res)
	}

	_ = tr.Close()
	wg.Wait()
}

func TestEchoServerToolsList(t *testing.T) {
	t.Parallel()
	var inMu sync.Mutex
	in := &bytes.Buffer{}
	lockedIn := &lockingBuffer{mu: &inMu, b: in}
	var outMu sync.Mutex
	out := &bytes.Buffer{}
	lockedOut := &lockingBuffer{mu: &outMu, b: out}
	tr := mcp.NewStdioTransport(lockedIn, lockedOut)
	srv := newEchoServer()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = srv.Start(context.Background(), tr)
	}()

	resp := sendAndRecv(t, lockedIn, lockedOut, `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`)
	res, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result")
	}
	tools, ok := res["tools"].([]any)
	if !ok || len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %v", res["tools"])
	}
	t0 := tools[0].(map[string]any)
	if t0["name"] != "echo" {
		t.Fatalf("tool name: %v", t0)
	}

	_ = tr.Close()
	wg.Wait()
}

func TestEchoServerToolsCallSuccess(t *testing.T) {
	t.Parallel()
	var inMu sync.Mutex
	in := &bytes.Buffer{}
	lockedIn := &lockingBuffer{mu: &inMu, b: in}
	var outMu sync.Mutex
	out := &bytes.Buffer{}
	lockedOut := &lockingBuffer{mu: &outMu, b: out}
	tr := mcp.NewStdioTransport(lockedIn, lockedOut)
	srv := newEchoServer()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = srv.Start(context.Background(), tr)
	}()

	resp := sendAndRecv(t, lockedIn, lockedOut, `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"echo","arguments":{"message":"hello"}}}`)
	res, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T (error=%v)", resp.Result, resp.Error)
	}
	content, _ := res["content"].([]any)
	if len(content) != 1 {
		t.Fatalf("expected 1 content item, got %v", res)
	}
	item := content[0].(map[string]any)
	text, _ := item["text"].(string)
	if !strings.Contains(text, "hello") {
		t.Fatalf("expected echo result, got %q", text)
	}
	if isErr, _ := res["isError"].(bool); isErr {
		t.Fatalf("expected not error, got isError=true")
	}

	_ = tr.Close()
	wg.Wait()
}

func TestEchoServerToolsCallMissingArgument(t *testing.T) {
	t.Parallel()
	var inMu sync.Mutex
	in := &bytes.Buffer{}
	lockedIn := &lockingBuffer{mu: &inMu, b: in}
	var outMu sync.Mutex
	out := &bytes.Buffer{}
	lockedOut := &lockingBuffer{mu: &outMu, b: out}
	tr := mcp.NewStdioTransport(lockedIn, lockedOut)
	srv := newEchoServer()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = srv.Start(context.Background(), tr)
	}()

	resp := sendAndRecv(t, lockedIn, lockedOut, `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"echo","arguments":{}}}`)
	res, _ := resp.Result.(map[string]any)
	if res == nil {
		t.Fatalf("expected map result")
	}
	if isErr, _ := res["isError"].(bool); !isErr {
		t.Fatalf("expected isError=true for missing argument")
	}

	_ = tr.Close()
	wg.Wait()
}

func TestEchoServerToolsCallUnknownTool(t *testing.T) {
	t.Parallel()
	var inMu sync.Mutex
	in := &bytes.Buffer{}
	lockedIn := &lockingBuffer{mu: &inMu, b: in}
	var outMu sync.Mutex
	out := &bytes.Buffer{}
	lockedOut := &lockingBuffer{mu: &outMu, b: out}
	tr := mcp.NewStdioTransport(lockedIn, lockedOut)
	srv := newEchoServer()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = srv.Start(context.Background(), tr)
	}()

	resp := sendAndRecv(t, lockedIn, lockedOut, `{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"ghost","arguments":{}}}`)
	if resp.Error == nil {
		t.Fatalf("expected error response, got %+v", resp)
	}
	if resp.Error.Code != mcp.CodeMethodNotFound {
		t.Fatalf("expected CodeMethodNotFound, got %d", resp.Error.Code)
	}

	_ = tr.Close()
	wg.Wait()
}

func TestEchoServerShutdown(t *testing.T) {
	t.Parallel()
	var inMu sync.Mutex
	in := &bytes.Buffer{}
	lockedIn := &lockingBuffer{mu: &inMu, b: in}
	tr := mcp.NewStdioTransport(lockedIn, &bytes.Buffer{})
	srv := newEchoServer()

	done := make(chan error, 1)
	go func() {
		done <- srv.Start(context.Background(), tr)
	}()

	if _, err := lockedIn.WriteString(`{"jsonrpc":"2.0","id":6,"method":"shutdown"}` + "\n"); err != nil {
		t.Fatalf("write to lockedIn: %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("shutdown returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("shutdown did not return in time")
	}
}

func TestEchoServerPing(t *testing.T) {
	t.Parallel()
	var inMu sync.Mutex
	in := &bytes.Buffer{}
	lockedIn := &lockingBuffer{mu: &inMu, b: in}
	var outMu sync.Mutex
	out := &bytes.Buffer{}
	lockedOut := &lockingBuffer{mu: &outMu, b: out}
	tr := mcp.NewStdioTransport(lockedIn, lockedOut)
	srv := newEchoServer()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = srv.Start(context.Background(), tr)
	}()

	resp := sendAndRecv(t, lockedIn, lockedOut, `{"jsonrpc":"2.0","id":7,"method":"ping"}`)
	if resp.Error != nil {
		t.Fatalf("ping should not error: %+v", resp)
	}

	_ = tr.Close()
	wg.Wait()
}

func TestEchoServerIDAndTools(t *testing.T) {
	t.Parallel()
	srv := newEchoServer()
	if srv.ID() != "echo-worker" {
		t.Fatalf("id: %s", srv.ID())
	}
	tools := srv.Tools()
	if len(tools) != 1 || tools[0].Name() != "echo" {
		t.Fatalf("tools: %+v", tools)
	}
}

func TestEchoServerUnknownMethod(t *testing.T) {
	t.Parallel()
	var inMu sync.Mutex
	in := &bytes.Buffer{}
	lockedIn := &lockingBuffer{mu: &inMu, b: in}
	var outMu sync.Mutex
	out := &bytes.Buffer{}
	lockedOut := &lockingBuffer{mu: &outMu, b: out}
	tr := mcp.NewStdioTransport(lockedIn, lockedOut)
	srv := newEchoServer()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = srv.Start(context.Background(), tr)
	}()

	resp := sendAndRecv(t, lockedIn, lockedOut, `{"jsonrpc":"2.0","id":8,"method":"frobnicate"}`)
	if resp.Error == nil {
		t.Fatalf("expected error for unknown method")
	}
	if resp.Error.Code != mcp.CodeMethodNotFound {
		t.Fatalf("expected CodeMethodNotFound, got %d", resp.Error.Code)
	}

	_ = tr.Close()
	wg.Wait()
}

func TestEchoServerParseError(t *testing.T) {
	t.Parallel()
	var inMu sync.Mutex
	in := &bytes.Buffer{}
	lockedIn := &lockingBuffer{mu: &inMu, b: in}
	var outMu sync.Mutex
	out := &bytes.Buffer{}
	lockedOut := &lockingBuffer{mu: &outMu, b: out}
	tr := mcp.NewStdioTransport(lockedIn, lockedOut)
	srv := newEchoServer()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = srv.Start(context.Background(), tr)
	}()

	resp := sendAndRecv(t, lockedIn, lockedOut, `not json`)
	if resp.Error == nil {
		t.Fatalf("expected parse error")
	}
	if resp.Error.Code != mcp.CodeParseError {
		t.Fatalf("expected CodeParseError, got %d", resp.Error.Code)
	}

	_ = tr.Close()
	wg.Wait()
}

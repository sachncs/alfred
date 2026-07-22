package mcp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alfred/alfred/internal/mcp"
	"github.com/alfred/alfred/internal/tool"
)

// Compile-time interface conformance.
var _ mcp.Transport = (*mcp.StdioTransport)(nil)
var _ mcp.Server = (*stubServer)(nil)

type stubServer struct {
	id    string
	tools []tool.Tool
}

func (s *stubServer) ID() string         { return s.id }
func (s *stubServer) Tools() []tool.Tool { return s.tools }
func (s *stubServer) Start(ctx context.Context, t mcp.Transport) error {
	// Echo server: respond to any request with the same method and an empty
	// result. Real MCP parsing happens in Phase 5 with the worker impl.
	for {
		raw, err := t.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, mcp.ErrTransportClosed) || errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		var req mcp.JSONRPCRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			resp := mcp.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      json.RawMessage(`null`),
				Error:   &mcp.JSONRPCError{Code: mcp.CodeParseError, Message: err.Error()},
			}
			b, _ := json.Marshal(resp)
			_ = t.WriteMessage(ctx, b)
			continue
		}
		resp := mcp.JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"echo": req.Method}}
		b, _ := json.Marshal(resp)
		if err := t.WriteMessage(ctx, b); err != nil {
			return err
		}
	}
}

func TestStdioTransportRoundTrip(t *testing.T) {
	t.Parallel()
	var in, out bytes.Buffer
	tr := mcp.NewStdioTransport(&in, &out)
	ctx := context.Background()

	in.WriteString(`{"jsonrpc":"2.0","id":1,"method":"ping"}` + "\n")

	raw, err := tr.ReadMessage(ctx)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var req mcp.JSONRPCRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req.Method != "ping" {
		t.Fatalf("method: %s", req.Method)
	}

	resp := mcp.JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"ok": true}}
	b, _ := json.Marshal(resp)
	if err := tr.WriteMessage(ctx, b); err != nil {
		t.Fatalf("write: %v", err)
	}

	if !strings.Contains(out.String(), `"result"`) {
		t.Fatalf("output missing result: %s", out.String())
	}
}

func TestStdioTransportCloseThenRead(t *testing.T) {
	t.Parallel()
	var in, out bytes.Buffer
	tr := mcp.NewStdioTransport(&in, &out)
	_ = tr.Close()
	_, err := tr.ReadMessage(context.Background())
	if !errors.Is(err, mcp.ErrTransportClosed) {
		t.Fatalf("expected ErrTransportClosed, got: %v", err)
	}
}

func TestStdioTransportCloseThenWrite(t *testing.T) {
	t.Parallel()
	var in, out bytes.Buffer
	tr := mcp.NewStdioTransport(&in, &out)
	_ = tr.Close()
	if err := tr.WriteMessage(context.Background(), json.RawMessage(`{}`)); !errors.Is(err, mcp.ErrTransportClosed) {
		t.Fatalf("expected ErrTransportClosed, got: %v", err)
	}
}

func TestStdioTransportReadEOF(t *testing.T) {
	t.Parallel()
	var in, out bytes.Buffer
	tr := mcp.NewStdioTransport(&in, &out)
	// Empty input → scanner returns false → ErrTransportClosed
	_, err := tr.ReadMessage(context.Background())
	if !errors.Is(err, mcp.ErrTransportClosed) {
		t.Fatalf("expected ErrTransportClosed on EOF, got: %v", err)
	}
}

func TestStdioTransportLongLine(t *testing.T) {
	t.Parallel()
	var in, out bytes.Buffer
	tr := mcp.NewStdioTransport(&in, &out)
	long := strings.Repeat("a", 1024*1024) // 1 MiB
	in.WriteString(`{"jsonrpc":"2.0","id":1,"method":"big","data":"` + long + `"}` + "\n")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	raw, err := tr.ReadMessage(ctx)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(raw), long) {
		t.Fatalf("long line truncated")
	}
}

func TestStdioTransportContextCancel(t *testing.T) {
	t.Parallel()
	var in, out bytes.Buffer
	tr := mcp.NewStdioTransport(&in, &out)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := tr.ReadMessage(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}

func TestStubServerHandlesInitialize(t *testing.T) {
	t.Parallel()
	var in, out bytes.Buffer
	var outMu sync.Mutex
	safeOut := &lockingBuffer{mu: &outMu, b: &out}
	tr := mcp.NewStdioTransport(&in, safeOut)
	srv := &stubServer{id: "test"}

	in.WriteString(`{"jsonrpc":"2.0","id":1,"method":"initialize"}` + "\n")
	in.WriteString(`{"jsonrpc":"2.0","id":2,"method":"shutdown"}` + "\n")

	var wg sync.WaitGroup
	wg.Add(1)
	errCh := make(chan error, 1)
	go func() {
		defer wg.Done()
		errCh <- srv.Start(context.Background(), tr)
	}()

	deadline := time.After(2 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatalf("timeout waiting for output")
		default:
		}
		outMu.Lock()
		done := strings.Count(out.String(), "\n") >= 2
		outMu.Unlock()
		if done {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	_ = tr.Close()
	wg.Wait()

	outMu.Lock()
	outStr := out.String()
	outMu.Unlock()
	if !strings.Contains(outStr, `"initialize"`) {
		t.Fatalf("output missing initialize echo: %s", outStr)
	}
	if !strings.Contains(outStr, `"shutdown"`) {
		t.Fatalf("output missing shutdown echo: %s", outStr)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("server exit: %v", err)
	}
}

func TestStubServerHandlesParseError(t *testing.T) {
	t.Parallel()
	var in, out bytes.Buffer
	var outMu sync.Mutex
	safeOut := &lockingBuffer{mu: &outMu, b: &out}
	tr := mcp.NewStdioTransport(&in, safeOut)
	srv := &stubServer{id: "test"}

	in.WriteString(`not valid json` + "\n")

	var wg sync.WaitGroup
	wg.Add(1)
	errCh := make(chan error, 1)
	go func() {
		defer wg.Done()
		errCh <- srv.Start(context.Background(), tr)
	}()

	deadline := time.After(2 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatalf("timeout")
		default:
		}
		outMu.Lock()
		found := strings.Contains(out.String(), `"code":-32700`)
		outMu.Unlock()
		if found {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	_ = tr.Close()
	wg.Wait()
}

// lockingBuffer wraps an io.Writer with a mutex so concurrent writes
// from a goroutine and reads from the test don't race under -race.
type lockingBuffer struct {
	mu *sync.Mutex
	b  *bytes.Buffer
}

func (l *lockingBuffer) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}

package tool_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/alfred/alfred/internal/tool"
)

// stubTool is a minimal Tool implementation used to verify the
// interface can be satisfied and the contract holds.
type stubTool struct {
	name string
	body func(ctx context.Context, input json.RawMessage, tc *tool.Context) (*tool.Result, error)
}

func (s *stubTool) Name() string                   { return s.name }
func (s *stubTool) Description() string            { return "stub for tests" }
func (s *stubTool) Schema() json.RawMessage         { return json.RawMessage(`{"type":"object"}`) }
func (s *stubTool) Execute(ctx context.Context, input json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	return s.body(ctx, input, tc)
}

// Compile-time interface conformance check.
var _ tool.Tool = (*stubTool)(nil)

func TestInterfaceConformance(t *testing.T) {
	var _ tool.Tool = &stubTool{name: "stub"}
}

func TestResultSuccess(t *testing.T) {
	r := tool.Success("ok")
	if !r.OK || r.Content != "ok" {
		t.Fatalf("unexpected result: %+v", r)
	}
}

func TestResultFailure(t *testing.T) {
	r := tool.Failure(errors.New("boom"))
	if r.OK || r.Error == "" {
		t.Fatalf("unexpected result: %+v", r)
	}
}

func TestResultWithMetadata(t *testing.T) {
	r := tool.Success("ok").WithMetadata(map[string]any{"k": "v"})
	if r.Metadata["k"] != "v" {
		t.Fatalf("metadata not set: %+v", r)
	}
}

func TestContextNilSafe(t *testing.T) {
	var c *tool.Context
	if c.Context() == nil {
		t.Fatalf("nil context should fall back to background")
	}
	if got := c.Context(); got == nil {
		t.Fatalf("nil receiver returned nil context")
	}
}

func TestContextCancellationPropagates(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	c2 := tool.NewContext(ctx, "thr", "turn", "call", "/tmp")
	if c2.Context().Err() == nil {
		t.Fatalf("expected cancellation error to propagate")
	}
}

func TestStubToolExecute(t *testing.T) {
	called := false
	var wg sync.WaitGroup
	wg.Add(1)
	st := &stubTool{
		name: "echo",
		body: func(ctx context.Context, input json.RawMessage, tc *tool.Context) (*tool.Result, error) {
			called = true
			if string(input) != `"hi"` {
				t.Errorf("unexpected input: %s", input)
			}
			wg.Done()
			return tool.Success(string(input)), nil
		},
	}
	r, err := st.Execute(context.Background(), json.RawMessage(`"hi"`), tool.NewContext(context.Background(), "t", "u", "c", "/tmp"))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !called {
		t.Fatalf("body not called")
	}
	if !r.OK {
		t.Fatalf("expected OK")
	}
	wg.Wait()
}

func TestJSONSerialization(t *testing.T) {
	r := tool.Success("hello").WithMetadata(map[string]any{"lines": 3})
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(b)
	if !strings.Contains(s, `"ok":true`) {
		t.Fatalf("ok field missing: %s", s)
	}
	if !strings.Contains(s, `"content":"hello"`) {
		t.Fatalf("content field missing: %s", s)
	}
	if !strings.Contains(s, `"lines":3`) {
		t.Fatalf("metadata not serialized: %s", s)
	}
}

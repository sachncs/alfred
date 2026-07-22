package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alfred/alfred/internal/contract"
	"github.com/alfred/alfred/internal/model"
	"github.com/alfred/alfred/internal/store"
)

// slowClient emits a DeltaText chunk after a delay. If the context is cancelled
// before the delay elapses, no chunks are emitted — that's the hook for the
// detached-context test: req.Context() is cancelled when ServeHTTP returns,
// so a goroutine using it never gets to emit. context.Background() survives.
type slowClient struct {
	text  string
	delay time.Duration
}

func (c *slowClient) Stream(ctx context.Context, _ model.Request) (<-chan model.StreamChunk, error) {
	ch := make(chan model.StreamChunk, 2)
	go func() {
		defer close(ch)
		time.Sleep(c.delay)
		if ctx.Err() != nil {
			return
		}
		ch <- model.StreamChunk{DeltaText: c.text}
		if ctx.Err() != nil {
			return
		}
		ch <- model.StreamChunk{Done: true}
	}()
	return ch, nil
}

func newAlfredWithClient(t *testing.T, client model.Client) *AlfredRuntime {
	t.Helper()
	ts, _ := store.NewFileThreadStore(t.TempDir())
	ss, _ := store.NewFileSessionStore(t.TempDir())
	rt := NewAlfredRuntime("", ts, ss)
	rt.SetModelClient(client)
	return rt
}

func TestAlfredRuntimeTurnFlow(t *testing.T) {
	client := &stubClient{responses: []responseSeq{
		{text: "Hello from agent"},
	}}
	rt := newAlfredWithClient(t, client)

	// Create thread
	body, _ := json.Marshal(contract.CreateThreadRequest{Title: "turn-flow"})
	req := httptest.NewRequest("POST", "/v1/threads", bytes.NewReader(body))
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	if w.Code != 201 {
		t.Fatalf("create thread: %d", w.Code)
	}
	var createResp contract.CreateThreadResponse
	_ = json.NewDecoder(w.Body).Decode(&createResp)
	threadID := createResp.Thread.ID

	// Start turn
	turnBody, _ := json.Marshal(contract.StartTurnRequest{
		Input: contract.UserInput{Text: "hello"},
	})
	req = httptest.NewRequest("POST", "/v1/threads/"+string(threadID)+"/turns", bytes.NewReader(turnBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	if w.Code != 202 {
		t.Fatalf("start turn: %d", w.Code)
	}

	// Wait for goroutine to finish
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		events, _ := rt.SessionStore().Read(threadID, 0)
		if len(events) > 0 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Get turn - should have items
	req = httptest.NewRequest("GET", "/v1/threads/"+string(threadID)+"/turns/turn-1", nil)
	w = httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("get turn: %d", w.Code)
	}

	// Verify session store has items
	events, _ := rt.SessionStore().Read(threadID, 0)
	if len(events) == 0 {
		t.Error("expected events in session store after turn execution")
	}
}

func TestAlfredRuntimeTurnFlowNoClient(t *testing.T) {
	ts, _ := store.NewFileThreadStore(t.TempDir())
	ss, _ := store.NewFileSessionStore(t.TempDir())
	rt := NewAlfredRuntime("", ts, ss)
	// no model client set

	body, _ := json.Marshal(contract.CreateThreadRequest{Title: "no-client"})
	req := httptest.NewRequest("POST", "/v1/threads", bytes.NewReader(body))
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	var createResp contract.CreateThreadResponse
	_ = json.NewDecoder(w.Body).Decode(&createResp)
	threadID := createResp.Thread.ID

	turnBody, _ := json.Marshal(contract.StartTurnRequest{
		Input: contract.UserInput{Text: "hello"},
	})
	req = httptest.NewRequest("POST", "/v1/threads/"+string(threadID)+"/turns", bytes.NewReader(turnBody))
	w = httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	// Still returns 202 even if loop will fail (runs async)
	if w.Code != 202 {
		t.Errorf("start turn without client: %d, want 202", w.Code)
	}
}

// TestAlfredRuntimeTurnDetachedContext proves that the goroutine spawned by
// handleStartTurn survives the request context being cancelled.
//
// Mechanism: the slowClient sleeps before emitting its DeltaText chunk and
// checks ctx.Err() between sends. The test cancels the request context after
// ServeHTTP returns (simulating what real net/http does when the response is
// flushed — the request context is cancelled). If executeTurn passed
// req.Context() into TurnLoop, that context is now cancelled — the sleep
// completes, ctx.Err() is non-nil, and no chunks are emitted, so the
// assistant text never lands in the session store.
//
// The fix is executeTurn using context.Background() so the goroutine outlives
// the HTTP request. If this test ever fails (no assistant text in the store),
// someone reintroduced req.Context() into the goroutine.
//
// ponytail: httptest.NewRequest does not cancel its context when ServeHTTP
// returns (no real client to disconnect from), so we cancel manually.
func TestAlfredRuntimeTurnDetachedContext(t *testing.T) {
	client := &slowClient{text: "detached-context-ok", delay: 80 * time.Millisecond}
	rt := newAlfredWithClient(t, client)

	body, _ := json.Marshal(contract.CreateThreadRequest{Title: "detached-ctx"})
	req := httptest.NewRequest("POST", "/v1/threads", bytes.NewReader(body))
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	var createResp contract.CreateThreadResponse
	_ = json.NewDecoder(w.Body).Decode(&createResp)
	threadID := createResp.Thread.ID

	turnBody, _ := json.Marshal(contract.StartTurnRequest{
		Input: contract.UserInput{Text: "hi"},
	})
	req = httptest.NewRequest("POST", "/v1/threads/"+string(threadID)+"/turns", bytes.NewReader(turnBody))
	req.Header.Set("Content-Type", "application/json")
	// ponytail: build a cancellable context so we can simulate real-http cancellation
	ctx, cancel := context.WithCancel(req.Context())
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	if w.Code != 202 {
		t.Fatalf("start turn: %d, want 202", w.Code)
	}
	// Simulate real net/http cancelling the request context when the
	// response is flushed — without this, the goroutine using req.Context()
	// would still see ctx.Err() == nil and accidentally succeed.
	cancel()
	// ServeHTTP has now returned and the request context is cancelled.
	// The goroutine must still be able to complete its work.

	deadline := time.Now().Add(2 * time.Second)
	var events []contract.TurnItem
	for time.Now().Before(deadline) {
		events, _ = rt.SessionStore().Read(threadID, 0)
		if len(events) >= 2 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if len(events) < 2 {
		t.Fatalf("goroutine was killed by request context — got %d items, want >=2 (user + assistant)", len(events))
	}

	foundAssistant := false
	for _, ev := range events {
		if ev.Text != nil && strings.Contains(*ev.Text, "detached-context-ok") {
			foundAssistant = true
			break
		}
	}
	if !foundAssistant {
		t.Errorf("assistant text missing from session store; goroutine died when request context was cancelled")
	}
}

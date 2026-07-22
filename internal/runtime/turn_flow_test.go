package runtime

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alfred/alfred/internal/contract"
	"github.com/alfred/alfred/internal/model"
	"github.com/alfred/alfred/internal/store"
)

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

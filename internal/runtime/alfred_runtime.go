package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/alfred/alfred/internal/contract"
	"github.com/alfred/alfred/internal/model"
	"github.com/alfred/alfred/internal/settings"
	"github.com/alfred/alfred/internal/store"
)

// SSEEvent is a server-sent event.
type SSEEvent struct {
	ID    string `json:"id"`
	Event string `json:"event"`
	Data  any    `json:"data"`
}

// AlfredRuntime is a level-4 runtime that adds SSE streaming and replay buffers.
type AlfredRuntime struct {
	*LocalRuntime
	client     model.Client
	settings   *settings.SettingsStore
	settingsMu sync.RWMutex
	replayMu   sync.RWMutex
	replayBuf  map[contract.ThreadID][]SSEEvent
	subsMu     sync.RWMutex
	subs       map[contract.ThreadID][]chan SSEEvent
}

// NewAlfredRuntime creates an AlfredRuntime.
func NewAlfredRuntime(bearerToken string, ts ThreadStore, ss SessionStore) *AlfredRuntime {
	rt := &AlfredRuntime{
		LocalRuntime: NewLocalRuntime(bearerToken, ts, ss),
		replayBuf:    make(map[contract.ThreadID][]SSEEvent),
		subs:         make(map[contract.ThreadID][]chan SSEEvent),
	}
	rt.parent = rt
	rt.setupSSERoutes()
	return rt
}

// NewAlfredRuntimeWithHybrid creates an AlfredRuntime backed by a hybrid
// (SQLite + JSONL) thread store.
func NewAlfredRuntimeWithHybrid(bearerToken string, hybrid *store.HybridThreadStore) *AlfredRuntime {
	return NewAlfredRuntime(bearerToken, hybrid, hybrid)
}

// SetModelClient sets the model client used for turn execution.
func (r *AlfredRuntime) SetModelClient(c model.Client) { r.client = c }

// SetSettings attaches a settings store to the runtime.
func (r *AlfredRuntime) SetSettings(s *settings.SettingsStore) {
	r.settingsMu.Lock()
	r.settings = s
	r.settingsMu.Unlock()
}

// LoadSettings returns the current persisted settings.
func (r *AlfredRuntime) LoadSettings() (contract.AppSettingsV1, error) {
	r.settingsMu.RLock()
	s := r.settings
	r.settingsMu.RUnlock()
	if s == nil {
		return settings.DefaultAppSettings(), nil
	}
	return s.Load()
}

// StartTurn creates a queued turn, persists it, and executes it async.
// Returns the queued turn immediately; SSE streams the result when done.
func (r *AlfredRuntime) StartTurn(thread *contract.Thread, input contract.UserInput) (*contract.Turn, error) {
	turnID := contract.TurnID(fmt.Sprintf("turn-%d", time.Now().UnixNano()))
	turn := &contract.Turn{
		ID:        turnID,
		ThreadID:  thread.ID,
		Status:    contract.TurnStatusQueued,
		StartedAt: time.Now().UTC(),
		Items:     []contract.TurnItem{},
	}

	if sql := r.sqlStore(); sql != nil {
		if err := sql.InsertTurn(turn); err != nil {
			defaultTurnLog.Printf("web: insert turn row %s: %v", turnID, err)
		}
	}

	go r.executeTurn(context.Background(), r, thread, turnID, input)
	return turn, nil
}

func (r *AlfredRuntime) setupSSERoutes() {
	r.mux.HandleFunc("GET /v1/threads/{id}/events", r.handleSSE)
}

// RunTurn creates a TurnLoop with the runtime's model client and tools,
// and runs the turn to completion. Returns the final turn.
func (r *AlfredRuntime) RunTurn(ctx context.Context, thread *contract.Thread, input contract.UserInput) (*contract.Turn, error) {
	return r.RunTurnWithID(ctx, thread, input, "")
}

// RunTurnWithID is like RunTurn but uses the supplied turn ID instead of
// generating a new one. Used by HTTP routes that pre-allocate the ID.
func (r *AlfredRuntime) RunTurnWithID(ctx context.Context, thread *contract.Thread, input contract.UserInput, turnID contract.TurnID) (*contract.Turn, error) {
	if r.client == nil {
		return nil, errors.New("no model client configured")
	}
	tools := r.Tools()
	loop := NewTurnLoop(r.client, tools, r.ThreadStore(), r.PublishEvent)
	return loop.RunTurnWithID(ctx, thread, input, turnID)
}

func (r *AlfredRuntime) handleSSE(w http.ResponseWriter, req *http.Request) {
	id := contract.ThreadID(req.PathValue("id"))

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	r.replayMu.RLock()
	events := r.replayBuf[id]
	r.replayMu.RUnlock()

	for _, ev := range events {
		writeSSE(w, ev)
		flusher.Flush()
	}

	ch := make(chan SSEEvent, 64)
	r.subsMu.Lock()
	r.subs[id] = append(r.subs[id], ch)
	r.subsMu.Unlock()

	defer func() {
		r.subsMu.Lock()
		subs := r.subs[id]
		for i, s := range subs {
			if s == ch {
				r.subs[id] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
		r.subsMu.Unlock()
		close(ch)
	}()

	for {
		select {
		case <-req.Context().Done():
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			writeSSE(w, ev)
			flusher.Flush()
		}
	}
}

// PublishEvent sends an event to all subscribers of a thread and appends to replay buffer.
func (r *AlfredRuntime) PublishEvent(threadID contract.ThreadID, ev SSEEvent) {
	r.replayMu.Lock()
	r.replayBuf[threadID] = append(r.replayBuf[threadID], ev)
	r.replayMu.Unlock()

	r.subsMu.RLock()
	subs := r.subs[threadID]
	r.subsMu.RUnlock()

	for _, ch := range subs {
		select {
		case ch <- ev:
		default:
			// ponytail: drop if subscriber is slow, don't block publisher
		}
	}
}

func writeSSE(w http.ResponseWriter, ev SSEEvent) {
	if ev.ID != "" {
		_, _ = fmt.Fprintf(w, "id: %s\n", ev.ID)
	}
	if ev.Event != "" {
		_, _ = fmt.Fprintf(w, "event: %s\n", ev.Event)
	}
	data, _ := json.Marshal(ev.Data)
	_, _ = fmt.Fprintf(w, "data: %s\n\n", data)
}

package runtime

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/alfred/alfred/internal/contract"
	internallog "github.com/alfred/alfred/internal/log"
	"github.com/alfred/alfred/internal/store"
)

// defaultTurnLog is the package logger for turn-related events.
var defaultTurnLog = internallog.New()

// sqlTurnStore is implemented by HybridThreadStore.SQLite() — it lets
// handleStartTurn/handleGetTurn write/read turn rows.
type sqlTurnStore interface {
	InsertTurn(t *contract.Turn) error
	UpdateTurn(t *contract.Turn) error
	GetTurn(id string) (*contract.Turn, error)
}

func (r *LocalRuntime) handleStartTurn(w http.ResponseWriter, req *http.Request) {
	id := contract.ThreadID(req.PathValue("id"))
	th, err := r.ThreadStore().Get(id)
	if err != nil {
		RouteError(w, contract.CodeNotFound, err.Error(), 404)
		return
	}

	var in contract.StartTurnRequest
	if err := RouteDecode(req, &in); err != nil {
		RouteError(w, contract.CodeValidation, err.Error(), 400)
		return
	}

	turnID := contract.TurnID(fmt.Sprintf("turn-%d", time.Now().UnixNano()))
	turn := &contract.Turn{
		ID:        turnID,
		ThreadID:  th.ID,
		Status:    contract.TurnStatusQueued,
		StartedAt: time.Now().UTC(),
		Items:     []contract.TurnItem{},
	}

	// Persist the queued turn row immediately so GET can find it.
	if sql := r.sqlStore(); sql != nil {
		if err := sql.InsertTurn(turn); err != nil {
			defaultTurnLog.Printf("insert turn row %s: %v", turnID, err)
			w.Header().Set("X-Alfred-Persistence-Error", err.Error())
		}
	}

	if ar, ok := r.parentAlfred(); ok {
		// ponytail: detached context — turn outlives the HTTP request.
		// Pass only the turn ID (no shared pointer); the goroutine re-fetches
		// the row from SQLite to mutate status/items.
		go r.executeTurn(context.Background(), ar, th, turnID, in.Input)
	}

	RouteJSON(w, contract.StartTurnResponse{Turn: *turn}, 202)
}

// sqlStore returns the SQLite turn-row writer if the runtime has one.
func (r *LocalRuntime) sqlStore() sqlTurnStore {
	if hybrid, ok := r.ThreadStore().(*store.HybridThreadStore); ok {
		return hybrid.SQLite()
	}
	return nil
}

func (r *LocalRuntime) handleGetTurn(w http.ResponseWriter, req *http.Request) {
	turnID := contract.TurnID(req.PathValue("turnId"))
	id := contract.ThreadID(req.PathValue("id"))

	// Look up authoritative turn status from SQLite.
	status := contract.TurnStatusQueued
	var startedAt, endedAt time.Time
	if sql := r.sqlStore(); sql != nil {
		if t, err := sql.GetTurn(string(turnID)); err == nil {
			status = t.Status
			startedAt = t.StartedAt
			endedAt = t.EndedAt
		}
	}

	// Read items from session store, filter to this turn.
	items, err := r.SessionStore().Read(id, 0)
	if err != nil {
		RouteError(w, contract.CodeInternal, err.Error(), 500)
		return
	}
	turnItems := filterItemsByTurn(items, turnID)
	if turnItems == nil {
		turnItems = []contract.TurnItem{}
	}

	RouteJSON(w, contract.Turn{
		ID:        turnID,
		ThreadID:  id,
		Status:    status,
		StartedAt: startedAt,
		EndedAt:   endedAt,
		Items:     turnItems,
	}, 200)
}

// filterItemsByTurn keeps only items whose Metadata["turnId"] matches turnID.
func filterItemsByTurn(items []contract.TurnItem, turnID contract.TurnID) []contract.TurnItem {
	out := make([]contract.TurnItem, 0, len(items))
	for _, it := range items {
		if v, ok := it.Metadata["turnId"].(string); ok && v == string(turnID) {
			out = append(out, it)
		}
	}
	return out
}

// parentAlfred returns the AlfredRuntime if this LocalRuntime is embedded in one.
func (r *LocalRuntime) parentAlfred() (*AlfredRuntime, bool) {
	if ar, ok := r.parent.(*AlfredRuntime); ok {
		return ar, true
	}
	return nil, false
}

// executeTurn runs the turn loop, persists items + updates the turn row, and
// publishes events. Errors are surfaced via SSE; partial results are still
// persisted so resume can pick them up.
//
// The turn row is fetched fresh from SQLite in this goroutine so the only
// shared state between handleStartTurn and executeTurn is the SQLite row.
func (r *LocalRuntime) executeTurn(ctx context.Context, ar *AlfredRuntime, thread *contract.Thread, turnID contract.TurnID, input contract.UserInput) {
	sql := r.sqlStore()
	turn := &contract.Turn{ID: turnID, ThreadID: thread.ID}
	if sql != nil {
		if t, err := sql.GetTurn(string(turnID)); err == nil {
			turn = t
		}
	}

	finalTurn, err := ar.RunTurnWithID(ctx, thread, input, turnID)

	// Merge final items into our turn.
	if finalTurn != nil {
		turn.Items = append(turn.Items, finalTurn.Items...)
		if !finalTurn.StartedAt.IsZero() {
			turn.StartedAt = finalTurn.StartedAt
		}
		if !finalTurn.EndedAt.IsZero() {
			turn.EndedAt = finalTurn.EndedAt
		}
		if finalTurn.Status != "" {
			turn.Status = finalTurn.Status
		}
	}

	// Persist items to session store.
	if len(turn.Items) > 0 {
		_ = r.SessionStore().Append(thread.ID, turn.Items)
	}

	// Update turn row in SQLite.
	if sql != nil {
		if err != nil {
			turn.Status = contract.TurnStatusFailed
			if turn.EndedAt.IsZero() {
				turn.EndedAt = time.Now().UTC()
			}
		}
		if uerr := sql.UpdateTurn(turn); uerr != nil {
			defaultTurnLog.Printf("update turn row %s: %v", turnID, uerr)
			ar.PublishEvent(thread.ID, SSEEvent{Event: "turn.persistence_error", Data: map[string]string{"turnId": string(turnID), "error": uerr.Error()}})
		}
	}

	if err != nil {
		ar.PublishEvent(thread.ID, SSEEvent{Event: "turn.failed", Data: map[string]string{"error": err.Error()}})
	} else {
		ar.PublishEvent(thread.ID, SSEEvent{Event: "turn.completed", Data: turn})
	}
}

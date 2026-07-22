package runtime

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/alfred/alfred/internal/contract"
)

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

	turn := &contract.Turn{
		ID:        contract.TurnID(fmt.Sprintf("turn-%d", time.Now().UnixNano())),
		ThreadID:  th.ID,
		Status:    contract.TurnStatusQueued,
		StartedAt: time.Now().UTC(),
		Items:     []contract.TurnItem{},
	}

	if ar, ok := r.parentAlfred(); ok {
		// ponytail: detached context — turn outlives the HTTP request
		go r.executeTurn(context.Background(), ar, th, in.Input)
	}

	RouteJSON(w, contract.StartTurnResponse{Turn: *turn}, 202)
}

func (r *LocalRuntime) handleGetTurn(w http.ResponseWriter, req *http.Request) {
	turnID := contract.TurnID(req.PathValue("turnId"))
	id := contract.ThreadID(req.PathValue("id"))

	items, err := r.SessionStore().Read(id, 0)
	if err != nil {
		RouteError(w, contract.CodeInternal, err.Error(), 500)
		return
	}

	if len(items) == 0 {
		RouteJSON(w, map[string]any{
			"id":       turnID,
			"threadId": id,
			"status":   contract.TurnStatusQueued,
			"items":    []contract.TurnItem{},
		}, 200)
		return
	}

	RouteJSON(w, contract.Turn{
		ID:       turnID,
		ThreadID: id,
		Status:   contract.TurnStatusCompleted,
		Items:    items,
	}, 200)
}

// parentAlfred returns the AlfredRuntime if this LocalRuntime is embedded in one.
func (r *LocalRuntime) parentAlfred() (*AlfredRuntime, bool) {
	if ar, ok := r.parent.(*AlfredRuntime); ok {
		return ar, true
	}
	return nil, false
}

// executeTurn runs the turn loop and persists items to the session store.
func (r *LocalRuntime) executeTurn(ctx context.Context, ar *AlfredRuntime, thread *contract.Thread, input contract.UserInput) {
	turn, err := ar.RunTurn(ctx, thread, input)
	if turn != nil && len(turn.Items) > 0 {
		_ = r.SessionStore().Append(thread.ID, turn.Items)
	}
	if err != nil {
		ar.PublishEvent(thread.ID, SSEEvent{Event: "turn.failed", Data: map[string]string{"error": err.Error()}})
	}
}

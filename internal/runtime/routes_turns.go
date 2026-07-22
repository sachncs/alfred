package runtime

import (
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

	RouteJSON(w, contract.StartTurnResponse{Turn: *turn}, 202)
}

func (r *LocalRuntime) handleGetTurn(w http.ResponseWriter, req *http.Request) {
	RouteJSON(w, map[string]string{"status": "not_implemented"}, 501)
}

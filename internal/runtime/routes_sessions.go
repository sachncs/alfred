package runtime

import (
	"net/http"

	"github.com/sachncs/alfred/internal/contract"
)

func (r *LocalRuntime) handleGetSession(w http.ResponseWriter, req *http.Request) {
	id := contract.ThreadID(req.PathValue("id"))
	events, err := r.SessionStore().Read(id, 0)
	if err != nil {
		RouteError(w, contract.CodeInternal, err.Error(), 500)
		return
	}
	RouteJSON(w, map[string]any{"events": events}, 200)
}

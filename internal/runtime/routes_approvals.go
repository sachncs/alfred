package runtime

import (
	"net/http"

	"github.com/sachncs/alfred/internal/contract"
)

func (r *LocalRuntime) handleResolveApproval(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	var in struct {
		Approved bool `json:"approved"`
	}
	if err := RouteDecode(req, &in); err != nil {
		RouteError(w, contract.CodeValidation, err.Error(), 400)
		return
	}
	resolved, err := r.ApprovalStore().Resolve(id, in.Approved)
	if err != nil {
		RouteError(w, contract.CodeNotFound, err.Error(), 404)
		return
	}
	RouteJSON(w, resolved, 200)
}

func (r *LocalRuntime) handleListApprovals(w http.ResponseWriter, req *http.Request) {
	RouteJSON(w, map[string]any{"approvals": r.ApprovalStore().List()}, 200)
}

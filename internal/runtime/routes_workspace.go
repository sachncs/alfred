package runtime

import (
	"net/http"
)

func (r *LocalRuntime) handleWorkspaceStatus(w http.ResponseWriter, req *http.Request) {
	RouteJSON(w, map[string]any{
		"status": "ok",
		"tools":  len(r.Tools()),
	}, 200)
}

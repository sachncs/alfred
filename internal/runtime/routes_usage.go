package runtime

import (
	"net/http"
)

func (r *LocalRuntime) handleGetUsage(w http.ResponseWriter, req *http.Request) {
	RouteJSON(w, map[string]any{"tokens": 0, "cost": 0.0}, 200)
}

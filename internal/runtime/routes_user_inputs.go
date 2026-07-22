package runtime

import (
	"net/http"
)

func (r *LocalRuntime) handleResolveUserInput(w http.ResponseWriter, req *http.Request) {
	RouteJSON(w, map[string]string{"status": "ok"}, 200)
}

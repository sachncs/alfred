package runtime

import (
	"net/http"
)

func (r *LocalRuntime) handleListMemory(w http.ResponseWriter, req *http.Request) {
	RouteJSON(w, map[string]any{"entries": []any{}}, 200)
}

func (r *LocalRuntime) handleSetMemory(w http.ResponseWriter, req *http.Request) {
	RouteJSON(w, map[string]string{"status": "ok"}, 200)
}

func (r *LocalRuntime) handleDeleteMemory(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(204)
}

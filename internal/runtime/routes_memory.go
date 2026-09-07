package runtime

import (
	"net/http"

	"github.com/alfred/alfred/internal/contract"
)

func (r *LocalRuntime) handleListMemory(w http.ResponseWriter, req *http.Request) {
	scope := MemoryScope(req.PathValue("scope"))
	switch scope {
	case MemoryScopeUser, MemoryScopeWorkspace, MemoryScopeProject:
		// valid
	default:
		RouteError(w, contract.CodeValidation, "scope must be user, workspace, or project", 400)
		return
	}
	entries, err := r.MemoryStore().List(scope)
	if err != nil {
		RouteError(w, contract.CodeInternal, err.Error(), 500)
		return
	}
	if entries == nil {
		entries = []MemoryEntry{}
	}
	RouteJSON(w, map[string]any{"entries": entries}, 200)
}

func (r *LocalRuntime) handleSetMemory(w http.ResponseWriter, req *http.Request) {
	scope := MemoryScope(req.PathValue("scope"))
	switch scope {
	case MemoryScopeUser, MemoryScopeWorkspace, MemoryScopeProject:
	default:
		RouteError(w, contract.CodeValidation, "scope must be user, workspace, or project", 400)
		return
	}
	var in struct {
		Key   string `json:"key"`
		Value any    `json:"value"`
	}
	if err := RouteDecode(req, &in); err != nil {
		RouteError(w, contract.CodeValidation, err.Error(), 400)
		return
	}
	if in.Key == "" {
		RouteError(w, contract.CodeValidation, "key is required", 400)
		return
	}
	if err := r.MemoryStore().Set(scope, in.Key, in.Value); err != nil {
		RouteError(w, contract.CodeInternal, err.Error(), 500)
		return
	}
	RouteJSON(w, map[string]string{"status": "ok", "scope": string(scope), "key": in.Key}, 200)
}

func (r *LocalRuntime) handleDeleteMemory(w http.ResponseWriter, req *http.Request) {
	scope := MemoryScope(req.PathValue("scope"))
	key := req.PathValue("key")
	switch scope {
	case MemoryScopeUser, MemoryScopeWorkspace, MemoryScopeProject:
	default:
		RouteError(w, contract.CodeValidation, "scope must be user, workspace, or project", 400)
		return
	}
	if err := r.MemoryStore().Delete(scope, key); err != nil {
		RouteError(w, contract.CodeNotFound, err.Error(), 404)
		return
	}
	w.WriteHeader(204)
}

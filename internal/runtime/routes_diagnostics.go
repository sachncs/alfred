package runtime

import (
	"net/http"
)

func (r *LocalRuntime) handleToolDiagnostics(w http.ResponseWriter, req *http.Request) {
	tools := make([]map[string]any, 0, len(r.Tools()))
	for _, t := range r.Tools() {
		tools = append(tools, map[string]any{
			"name":        t.Name(),
			"description": t.Description(),
		})
	}
	RouteJSON(w, map[string]any{"tools": tools}, 200)
}

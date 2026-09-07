package runtime

import (
	"net/http"
)

func (r *LocalRuntime) handleGetUsage(w http.ResponseWriter, req *http.Request) {
	in, out := r.UsageTracker().Totals()
	RouteJSON(w, map[string]any{
		"tokens": map[string]int{"input": in, "output": out},
		"cost":   0.0,
		"recent": r.UsageTracker().Recent(50),
	}, 200)
}

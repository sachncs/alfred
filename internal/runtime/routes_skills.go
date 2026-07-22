package runtime

import (
	"net/http"
)

func (r *LocalRuntime) handleListSkills(w http.ResponseWriter, req *http.Request) {
	RouteJSON(w, map[string]any{"skills": []any{}}, 200)
}

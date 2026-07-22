package runtime

import (
	"encoding/json"
	"net/http"

	"github.com/alfred/alfred/internal/contract"
)

// RouteError writes a JSON error response.
func RouteError(w http.ResponseWriter, code string, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(contract.Errorf(code, msg))
}

// RouteJSON writes a JSON response.
func RouteJSON(w http.ResponseWriter, v any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// RouteDecode decodes a JSON request body into v.
func RouteDecode(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

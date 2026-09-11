package capability

import (
	"encoding/json"
	"net/http"

	"github.com/sachncs/alfred/internal/contract"
)

// HandleCapabilities serves GET /v1/capabilities — lists all registered capabilities.
func HandleCapabilities(broker *Broker, discovery *Discovery) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		caps := discovery.Capabilities(r.Context())
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"capabilities": caps,
			"count":        len(caps),
		})
	}
}

// HandleCapabilityDetail serves GET /v1/capabilities/{id} — describes one capability.
func HandleCapabilityDetail(broker *Broker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := ID(r.PathValue("id"))
		c := broker.Get(id)
		if c == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(404)
			_ = json.NewEncoder(w).Encode(contract.Errorf(contract.CodeNotFound, "capability not found"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":        string(c.ID()),
			"available": c.IsAvailable(r.Context()),
		})
	}
}

// HandleCapabilityDispatch serves POST /v1/capabilities/{id}/dispatch — invokes a capability.
func HandleCapabilityDispatch(broker *Broker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := ID(r.PathValue("id"))
		var input json.RawMessage
		if r.Body != nil {
			defer r.Body.Close() //nolint:errcheck
			_ = json.NewDecoder(r.Body).Decode(&input)
		}
		output, err := broker.Dispatch(r.Context(), id, input)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(500)
			_ = json.NewEncoder(w).Encode(DispatchResult{Error: err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(DispatchResult{Output: output})
	}
}

// RegisterCapabilityRoutes mounts capability routes on the given mux.
func RegisterCapabilityRoutes(mux *http.ServeMux, broker *Broker, discovery *Discovery) {
	mux.HandleFunc("GET /v1/capabilities", HandleCapabilities(broker, discovery))
	mux.HandleFunc("GET /v1/capabilities/{id}", HandleCapabilityDetail(broker))
	mux.HandleFunc("POST /v1/capabilities/{id}/dispatch", HandleCapabilityDispatch(broker))
}

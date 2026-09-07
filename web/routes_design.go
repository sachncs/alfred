package web

import "net/http"

// handleDesign renders the design primitives showcase page.
func (h *Handlers) handleDesign(w http.ResponseWriter, r *http.Request) {
	render(w, h.design, "layout", nil)
}

package web

import "net/http"

func (h *Handlers) handleSettings(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title": "Settings",
	}
	render(w, h.settings, "settings", data)
}

package web

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/sachncs/alfred/internal/contract"
)

// SettingsLoader loads persisted settings.
type SettingsLoader interface {
	Load() (contract.AppSettingsV1, error)
}

// SettingsSaver saves full settings.
type SettingsSaver interface {
	Save(contract.AppSettingsV1) error
}

// SettingsPatcher merges a partial into saved settings.
type SettingsPatcher interface {
	Patch(contract.AppSettingsV1) (contract.AppSettingsV1, error)
}

// HandleGetSettings serves GET /v1/settings — returns the current AppSettingsV1.
func HandleGetSettings(loader SettingsLoader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		settings, err := loader.Load()
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Alfred-Settings-Warning", err.Error())
			w.WriteHeader(500)
			_ = json.NewEncoder(w).Encode(contract.Errorf(contract.CodeInternal, err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(settings)
	}
}

// HandlePostSettings serves POST /v1/settings — replaces settings.
func HandlePostSettings(saver SettingsSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var settings contract.AppSettingsV1
		if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(400)
			_ = json.NewEncoder(w).Encode(contract.Errorf(contract.CodeValidation, "invalid JSON"))
			return
		}
		if err := saver.Save(settings); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			_ = json.NewEncoder(w).Encode(contract.Errorf(contract.CodeInternal, err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "saved"})
	}
}

// HandlePatchSettings serves PATCH /v1/settings — merges partial into saved settings.
func HandlePatchSettings(patcher SettingsPatcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var partial contract.AppSettingsV1
		if err := json.NewDecoder(r.Body).Decode(&partial); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(400)
			_ = json.NewEncoder(w).Encode(contract.Errorf(contract.CodeValidation, "invalid JSON"))
			return
		}
		merged, err := patcher.Patch(partial)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			_ = json.NewEncoder(w).Encode(contract.Errorf(contract.CodeInternal, err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(merged)
	}
}

// RegisterSettingsRoutes mounts settings routes on the mux.
func RegisterSettingsRoutes(mux *http.ServeMux, loader SettingsLoader, saver SettingsSaver, patcher SettingsPatcher) {
	mux.HandleFunc("GET /v1/settings", HandleGetSettings(loader))
	mux.HandleFunc("POST /v1/settings", HandlePostSettings(saver))
	mux.HandleFunc("PATCH /v1/settings", HandlePatchSettings(patcher))
	log.Printf("settings routes ready: GET/POST/PATCH /v1/settings")
}

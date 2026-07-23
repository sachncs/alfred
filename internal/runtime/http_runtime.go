package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alfred/alfred/internal/contract"
	"github.com/alfred/alfred/internal/tool"
)

// HTTPRuntime is a level-2 runtime that serves the HTTP API.
type HTTPRuntime struct {
	state        *RuntimeState
	mux          *http.ServeMux
	tools        []tool.Tool
	bearerToken  string
	threadStore  ThreadStore
	sessionStore SessionStore
	memoryStore  *MemoryStore
	skillsLoader *SkillsLoader
	usageTracker *UsageTracker
	approvals    *ApprovalStore
	attachments  *AttachmentStore
	parent       any // ponytail: outer runtime (e.g. *AlfredRuntime) for context lookup
}

// NewHTTPRuntime creates an HTTPRuntime with the given bearer token.
func NewHTTPRuntime(bearerToken string, ts ThreadStore, ss SessionStore) *HTTPRuntime {
	r := &HTTPRuntime{
		state:        NewRuntimeState(),
		mux:          http.NewServeMux(),
		bearerToken:  bearerToken,
		threadStore:  ts,
		sessionStore: ss,
		memoryStore:  NewMemoryStore(),
		skillsLoader: NewSkillsLoader(),
		usageTracker: NewUsageTracker(),
		approvals:    NewApprovalStore(),
		attachments:  NewAttachmentStore(),
	}
	r.setupRoutes()
	return r
}

// SetMemoryStore replaces the memory store (used by tests; production wires via NewHTTPRuntime).
func (r *HTTPRuntime) SetMemoryStore(m *MemoryStore) { r.memoryStore = m }

// SetSkillsLoader replaces the skills loader.
func (r *HTTPRuntime) SetSkillsLoader(s *SkillsLoader) { r.skillsLoader = s }

// SetUsageTracker replaces the usage tracker.
func (r *HTTPRuntime) SetUsageTracker(u *UsageTracker) { r.usageTracker = u }

// SetApprovalStore replaces the approval store.
func (r *HTTPRuntime) SetApprovalStore(a *ApprovalStore) { r.approvals = a }

// SetAttachmentStore replaces the attachment store.
func (r *HTTPRuntime) SetAttachmentStore(a *AttachmentStore) { r.attachments = a }

// MemoryStore returns the memory store.
func (r *HTTPRuntime) MemoryStore() *MemoryStore { return r.memoryStore }

// SkillsLoader returns the skills loader.
func (r *HTTPRuntime) SkillsLoader() *SkillsLoader { return r.skillsLoader }

// UsageTracker returns the usage tracker.
func (r *HTTPRuntime) UsageTracker() *UsageTracker { return r.usageTracker }

// ApprovalStore returns the approval store.
func (r *HTTPRuntime) ApprovalStore() *ApprovalStore { return r.approvals }

// AttachmentStore returns the attachment store.
func (r *HTTPRuntime) AttachmentStore() *AttachmentStore { return r.attachments }

func (r *HTTPRuntime) setupRoutes() {
	r.mux.HandleFunc("GET /v1/health", r.handleHealth)
}

// handleHealth serves GET /v1/health with a JSON health response.
func (r *HTTPRuntime) handleHealth(w http.ResponseWriter, req *http.Request) {
	resp := contract.HealthResponse{
		Status:    "ok",
		Version:   "0.4.0-phase4",
		Timestamp: time.Now().UTC(),
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// Start begins serving on the given port. Blocks until ctx is cancelled.
func (r *HTTPRuntime) Start(ctx context.Context, port int) error {
	r.state.SetRunning()
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r.Handler(),
	}

	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()

	return srv.ListenAndServe()
}

// Handler returns the http.Handler (with auth middleware applied).
func (r *HTTPRuntime) Handler() http.Handler {
	return BearerAuth(r.bearerToken, r.mux)
}

// Mux returns the underlying ServeMux for external route registration (e.g. web UI).
func (r *HTTPRuntime) Mux() *http.ServeMux { return r.mux }

// ThreadStore returns the thread store.
func (r *HTTPRuntime) ThreadStore() ThreadStore { return r.threadStore }

// SessionStore returns the session store.
func (r *HTTPRuntime) SessionStore() SessionStore { return r.sessionStore }

// RegisterTool registers a tool.
func (r *HTTPRuntime) RegisterTool(t tool.Tool) { r.tools = append(r.tools, t) }

// Tools returns all registered tools.
func (r *HTTPRuntime) Tools() []tool.Tool { return r.tools }

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
	state      *RuntimeState
	mux        *http.ServeMux
	tools      []tool.Tool
	bearerToken string
	threadStore  ThreadStore
	sessionStore SessionStore
}

// NewHTTPRuntime creates an HTTPRuntime with the given bearer token.
func NewHTTPRuntime(bearerToken string, ts ThreadStore, ss SessionStore) *HTTPRuntime {
	r := &HTTPRuntime{
		state:       NewRuntimeState(),
		mux:         http.NewServeMux(),
		bearerToken: bearerToken,
		threadStore:  ts,
		sessionStore: ss,
	}
	r.setupRoutes()
	return r
}

func (r *HTTPRuntime) setupRoutes() {
	r.mux.HandleFunc("GET /health", r.handleHealth)
	r.mux.HandleFunc("GET /healthz", r.handleHealth)
}

func (r *HTTPRuntime) handleHealth(w http.ResponseWriter, req *http.Request) {
	resp := contract.HealthResponse{
		Status:    "ok",
		Version:   "0.2.0-phase2",
		Timestamp: time.Now().UTC(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Start begins serving on the given port. Blocks until ctx is cancelled.
func (r *HTTPRuntime) Start(ctx context.Context, port int) error {
	r.state.SetRunning()
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r.Handler(),
	}

	// Shutdown server when context is cancelled.
	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	return srv.ListenAndServe()
}

// Handler returns the http.Handler (with auth middleware applied).
func (r *HTTPRuntime) Handler() http.Handler {
	return BearerAuth(r.bearerToken, r.mux)
}

// ThreadStore returns the thread store.
func (r *HTTPRuntime) ThreadStore() ThreadStore { return r.threadStore }

// SessionStore returns the session store.
func (r *HTTPRuntime) SessionStore() SessionStore { return r.sessionStore }

// RegisterTool registers a tool.
func (r *HTTPRuntime) RegisterTool(t tool.Tool) { r.tools = append(r.tools, t) }

// Tools returns all registered tools.
func (r *HTTPRuntime) Tools() []tool.Tool { return r.tools }

package runtime

import (
	"github.com/alfred/alfred/internal/tool"
)

// LocalRuntime is a level-3 runtime that wires all /v1/ routes.
type LocalRuntime struct {
	*HTTPRuntime
}

// NewLocalRuntime creates a LocalRuntime with all routes wired.
func NewLocalRuntime(bearerToken string, ts ThreadStore, ss SessionStore) *LocalRuntime {
	rt := &LocalRuntime{
		HTTPRuntime: NewHTTPRuntime(bearerToken, ts, ss),
	}
	rt.setupV1Routes()
	return rt
}

func (r *LocalRuntime) setupV1Routes() {
	mux := r.HTTPRuntime.mux

	// Threads
	mux.HandleFunc("GET /v1/threads", r.handleListThreads)
	mux.HandleFunc("POST /v1/threads", r.handleCreateThread)
	mux.HandleFunc("GET /v1/threads/{id}", r.handleGetThread)
	mux.HandleFunc("DELETE /v1/threads/{id}", r.handleDeleteThread)

	// Turns
	mux.HandleFunc("POST /v1/threads/{id}/turns", r.handleStartTurn)
	mux.HandleFunc("GET /v1/threads/{id}/turns/{turnId}", r.handleGetTurn)

	// Sessions
	mux.HandleFunc("GET /v1/sessions/{id}", r.handleGetSession)

	// Approvals
	mux.HandleFunc("POST /v1/approvals/{id}/resolve", r.handleResolveApproval)

	// User inputs
	mux.HandleFunc("POST /v1/user-inputs/{id}/resolve", r.handleResolveUserInput)

	// Memory
	mux.HandleFunc("GET /v1/memory/{scope}", r.handleListMemory)
	mux.HandleFunc("POST /v1/memory/{scope}", r.handleSetMemory)
	mux.HandleFunc("DELETE /v1/memory/{scope}/{key}", r.handleDeleteMemory)

	// Skills
	mux.HandleFunc("GET /v1/skills", r.handleListSkills)

	// Attachments
	mux.HandleFunc("POST /v1/attachments", r.handleUploadAttachment)

	// Workspace
	mux.HandleFunc("GET /v1/workspace/status", r.handleWorkspaceStatus)

	// Usage
	mux.HandleFunc("GET /v1/usage", r.handleGetUsage)

	// Tool diagnostics
	mux.HandleFunc("GET /v1/tool-diagnostics", r.handleToolDiagnostics)
}

// Ensure LocalRuntime satisfies Runtime at compile time.
var _ Runtime = (*LocalRuntime)(nil)

// RegisterTool registers a tool with the runtime.
func (r *LocalRuntime) RegisterTool(t tool.Tool) {
	r.HTTPRuntime.RegisterTool(t)
}

// Tools returns all registered tools.
func (r *LocalRuntime) Tools() []tool.Tool {
	return r.HTTPRuntime.Tools()
}

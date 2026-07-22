package runtime

import (
	"fmt"
	"net/http"
	"time"

	"github.com/alfred/alfred/internal/contract"
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

func (r *LocalRuntime) handleListThreads(w http.ResponseWriter, req *http.Request) {
	threads, nextToken, err := r.ThreadStore().List(50, req.URL.Query().Get("token"))
	if err != nil {
		RouteError(w, contract.CodeInternal, err.Error(), 500)
		return
	}
	RouteJSON(w, contract.ListThreadsResponse{Threads: threads, NextToken: nextToken}, 200)
}

func (r *LocalRuntime) handleCreateThread(w http.ResponseWriter, req *http.Request) {
	var in contract.CreateThreadRequest
	if err := RouteDecode(req, &in); err != nil {
		RouteError(w, contract.CodeValidation, err.Error(), 400)
		return
	}
	th := &contract.Thread{
		ID:        contract.ThreadID(fmt.Sprintf("thr-%d", time.Now().UnixNano())),
		Title:     in.Title,
		Status:    contract.ThreadStatusIdle,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Metadata:  in.Metadata,
	}
	if err := r.ThreadStore().Create(th); err != nil {
		RouteError(w, contract.CodeInternal, err.Error(), 500)
		return
	}
	RouteJSON(w, contract.CreateThreadResponse{Thread: *th}, 201)
}

func (r *LocalRuntime) handleGetThread(w http.ResponseWriter, req *http.Request) {
	id := contract.ThreadID(req.PathValue("id"))
	th, err := r.ThreadStore().Get(id)
	if err != nil {
		RouteError(w, contract.CodeNotFound, err.Error(), 404)
		return
	}
	RouteJSON(w, th, 200)
}

func (r *LocalRuntime) handleDeleteThread(w http.ResponseWriter, req *http.Request) {
	id := contract.ThreadID(req.PathValue("id"))
	if err := r.ThreadStore().Delete(id); err != nil {
		RouteError(w, contract.CodeInternal, err.Error(), 500)
		return
	}
	w.WriteHeader(204)
}

func (r *LocalRuntime) handleStartTurn(w http.ResponseWriter, req *http.Request) {
	id := contract.ThreadID(req.PathValue("id"))
	th, err := r.ThreadStore().Get(id)
	if err != nil {
		RouteError(w, contract.CodeNotFound, err.Error(), 404)
		return
	}

	var in contract.StartTurnRequest
	if err := RouteDecode(req, &in); err != nil {
		RouteError(w, contract.CodeValidation, err.Error(), 400)
		return
	}

	turn := &contract.Turn{
		ID:        contract.TurnID(fmt.Sprintf("turn-%d", time.Now().UnixNano())),
		ThreadID:  th.ID,
		Status:    contract.TurnStatusQueued,
		StartedAt: time.Now().UTC(),
		Items:     []contract.TurnItem{},
	}

	RouteJSON(w, contract.StartTurnResponse{Turn: *turn}, 202)
}

func (r *LocalRuntime) handleGetTurn(w http.ResponseWriter, req *http.Request) {
	RouteJSON(w, map[string]string{"status": "not_implemented"}, 501)
}

func (r *LocalRuntime) handleGetSession(w http.ResponseWriter, req *http.Request) {
	id := contract.ThreadID(req.PathValue("id"))
	events, err := r.SessionStore().Read(id, 0)
	if err != nil {
		RouteError(w, contract.CodeInternal, err.Error(), 500)
		return
	}
	RouteJSON(w, map[string]any{"events": events}, 200)
}

func (r *LocalRuntime) handleResolveApproval(w http.ResponseWriter, req *http.Request) {
	RouteJSON(w, map[string]string{"status": "ok"}, 200)
}

func (r *LocalRuntime) handleResolveUserInput(w http.ResponseWriter, req *http.Request) {
	RouteJSON(w, map[string]string{"status": "ok"}, 200)
}

func (r *LocalRuntime) handleListMemory(w http.ResponseWriter, req *http.Request) {
	RouteJSON(w, map[string]any{"entries": []any{}}, 200)
}

func (r *LocalRuntime) handleSetMemory(w http.ResponseWriter, req *http.Request) {
	RouteJSON(w, map[string]string{"status": "ok"}, 200)
}

func (r *LocalRuntime) handleDeleteMemory(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(204)
}

func (r *LocalRuntime) handleListSkills(w http.ResponseWriter, req *http.Request) {
	RouteJSON(w, map[string]any{"skills": []any{}}, 200)
}

func (r *LocalRuntime) handleUploadAttachment(w http.ResponseWriter, req *http.Request) {
	RouteJSON(w, map[string]string{"status": "ok"}, 200)
}

func (r *LocalRuntime) handleWorkspaceStatus(w http.ResponseWriter, req *http.Request) {
 RouteJSON(w, map[string]any{
		"status": "ok",
		"tools":  len(r.Tools()),
	}, 200)
}

func (r *LocalRuntime) handleGetUsage(w http.ResponseWriter, req *http.Request) {
	RouteJSON(w, map[string]any{"tokens": 0, "cost": 0.0}, 200)
}

func (r *LocalRuntime) handleToolDiagnostics(w http.ResponseWriter, req *http.Request) {
	tools := make([]map[string]any, 0, len(r.Tools()))
	for _, t := range r.Tools() {
		tools = append(tools, map[string]any{
			"name":        t.Name(),
			"description": t.Description(),
		})
	}
	RouteJSON(w, map[string]any{"tools": tools}, 200)
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

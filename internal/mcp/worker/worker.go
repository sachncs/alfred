// Package worker provides the abstract WorkerServer base class that all
// MCP workers in Phase 5+ inherit from.
package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alfred/alfred/internal/mcp"
	"github.com/alfred/alfred/internal/tool"
)

// WorkerServer is the abstract base for any concrete MCP worker.
// It implements Server by holding a name and a static tool catalog;
// subclasses override catalog construction.
type WorkerServer struct {
	id      string
	tools   []tool.Tool
	handler RequestHandler
}

// RequestHandler is a function that processes a tool call.
// It receives the tool call and a context.Context, and returns the
// tool result.
type RequestHandler func(ctx context.Context, name string, input json.RawMessage, tc *tool.Context) (*tool.Result, error)

// NewWorkerServer constructs a WorkerServer with the given id and tools.
func NewWorkerServer(id string, tools []tool.Tool) *WorkerServer {
	return &WorkerServer{id: id, tools: tools}
}

// WithHandler attaches a default request handler used by tools that
// don't override their own. Subclasses typically override Start.
func (w *WorkerServer) WithHandler(h RequestHandler) *WorkerServer {
	w.handler = h
	return w
}

// ID implements mcp.Server.
func (w *WorkerServer) ID() string { return w.id }

// Tools implements mcp.Server.
func (w *WorkerServer) Tools() []tool.Tool { return w.tools }

// Start implements mcp.Server. The default implementation handles a
// minimal MCP handshake: responds to `initialize`, `tools/list`, and
// `tools/call`. Subclasses can override for richer protocols.
func (w *WorkerServer) Start(ctx context.Context, t mcp.Transport) error {
	for {
		raw, err := t.ReadMessage(ctx)
		if err != nil {
			return nil // EOF / cancel → graceful exit
		}

		var req mcp.JSONRPCRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			resp := mcp.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      json.RawMessage(`null`),
				Error:   &mcp.JSONRPCError{Code: mcp.CodeParseError, Message: err.Error()},
			}
			if err := w.write(ctx, t, resp); err != nil {
				return err
			}
			continue
		}

		var resp mcp.JSONRPCResponse
		switch req.Method {
		case "initialize":
			resp = mcp.JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{
				"protocolVersion": "2024-11-05",
				"serverInfo":       map[string]any{"name": w.id, "version": "0.1.0"},
				"capabilities":     map[string]any{"tools": map[string]any{}},
			}}
		case "tools/list":
			resp = mcp.JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{
				"tools": w.toolList(),
			}}
		case "tools/call":
			resp = w.handleCall(ctx, t, req)
		case "ping":
			resp = mcp.JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{}}
		case "shutdown":
			resp = mcp.JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: nil}
			if err := w.write(ctx, t, resp); err != nil {
				return err
			}
			return nil
		default:
			resp = mcp.JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &mcp.JSONRPCError{
				Code: mcp.CodeMethodNotFound, Message: fmt.Sprintf("unknown method: %s", req.Method),
			}}
		}

		if err := w.write(ctx, t, resp); err != nil {
			return err
		}
	}
}

func (w *WorkerServer) write(ctx context.Context, t mcp.Transport, resp mcp.JSONRPCResponse) error {
	b, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("marshal response: %w", err)
	}
	return t.WriteMessage(ctx, b)
}

func (w *WorkerServer) toolList() []map[string]any {
	out := make([]map[string]any, 0, len(w.tools))
	for _, t := range w.tools {
		entry := map[string]any{
			"name":        t.Name(),
			"description": t.Description(),
			"inputSchema": json.RawMessage(t.Schema()),
		}
		out = append(out, entry)
	}
	return out
}

// handleCall executes a tools/call request by dispatching to the
// registered tool. The default handler passes through to w.handler if
// set, otherwise returns an error.
func (w *WorkerServer) handleCall(ctx context.Context, t mcp.Transport, req mcp.JSONRPCRequest) mcp.JSONRPCResponse {
	var call struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &call); err != nil {
		return mcp.JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &mcp.JSONRPCError{
			Code: mcp.CodeInvalidRequest, Message: err.Error(),
		}}
	}

	toolCtx := tool.NewContext(ctx, "", "", "", "")
	for _, candidate := range w.tools {
		if candidate.Name() == call.Name {
			result, err := candidate.Execute(ctx, call.Arguments, toolCtx)
			if err != nil {
				return mcp.JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &mcp.JSONRPCError{
					Code: mcp.CodeInternalError, Message: err.Error(),
				}}
			}
			if result == nil {
				result = tool.FailureMsg("nil result")
			}
			return mcp.JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{
				"content": []map[string]any{{"type": "text", "text": result.Content}},
				"isError": !result.OK,
			}}
		}
	}
	return mcp.JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &mcp.JSONRPCError{
		Code: mcp.CodeMethodNotFound, Message: fmt.Sprintf("tool not found: %s", call.Name),
	}}
}

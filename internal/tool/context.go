package tool

import (
	"context"
	"sync"
)

// Context carries per-call state for a tool invocation. It is built by
// the runtime and threaded through Execute; tools should treat it as
// read-only.
type Context struct {
	// ThreadID identifies the conversation that triggered the call.
	ThreadID string

	// TurnID identifies the turn within the thread.
	TurnID string

	// ToolCallID identifies this specific invocation.
	ToolCallID string

	// WorkspaceRoot is the absolute path to the active workspace.
	WorkspaceRoot string

	// Approver is consulted when a tool wants to gate a sensitive action.
	// If nil, the tool runs in best-effort mode without approval.
	Approver Approver

	// ctx is the parent context for cancellation.
	ctx context.Context

	once sync.Once
}

// Approver asks for user approval of a tool action.
type Approver interface {
	RequestApproval(ctx context.Context, reason string) (bool, error)
}

// NewContext constructs a Context bound to the given parent context.
func NewContext(parent context.Context, threadID, turnID, toolCallID, workspaceRoot string) *Context {
	return &Context{
		ctx:           parent,
		ThreadID:      threadID,
		TurnID:        turnID,
		ToolCallID:    toolCallID,
		WorkspaceRoot: workspaceRoot,
	}
}

// Context returns the underlying context.Context for cancellation / deadline.
func (c *Context) Context() context.Context {
	if c == nil {
		return context.Background()
	}
	if c.ctx == nil {
		return context.Background()
	}
	return c.ctx
}

// WithApprover returns a copy of the Context with the given Approver.
func (c *Context) WithApprover(a Approver) *Context {
	clone := *c
	clone.Approver = a
	return &clone
}

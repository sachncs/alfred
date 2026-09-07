// Package runtime defines the Runtime interface and its implementations
// (HTTPRuntime, LocalRuntime, AlfredRuntime). The runtime drives the
// turn loop, manages threads, and serves the HTTP/SSE API.
package runtime

import (
	"context"
	"net/http"

	"github.com/alfred/alfred/internal/contract"
	"github.com/alfred/alfred/internal/tool"
)

// Runtime is the contract every runtime implementation must satisfy.
type Runtime interface {
	// Start begins serving on the given port. Blocks until ctx is cancelled.
	Start(ctx context.Context, port int) error

	// Handler returns the http.Handler (for embedding or testing).
	Handler() http.Handler

	// ThreadStore returns the thread store.
	ThreadStore() ThreadStore

	// SessionStore returns the session store.
	SessionStore() SessionStore

	// RegisterTool registers a tool with the runtime.
	RegisterTool(t tool.Tool)

	// Tools returns all registered tools.
	Tools() []tool.Tool
}

// ThreadStore is the persistence layer for threads.
type ThreadStore interface {
	Create(t *contract.Thread) error
	Get(id contract.ThreadID) (*contract.Thread, error)
	List(limit int, token string) ([]contract.Thread, string, error)
	Update(t *contract.Thread) error
	Delete(id contract.ThreadID) error
}

// SessionStore is the persistence layer for turn events.
type SessionStore interface {
	Append(threadID contract.ThreadID, events []contract.TurnItem) error
	Read(threadID contract.ThreadID, offset int) ([]contract.TurnItem, error)
	Prune(threadID contract.ThreadID, maxEvents int) error
}

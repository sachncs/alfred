// Package agent defines the base Agent interface and shared state.
package agent

import (
	"context"

	"github.com/sachncs/alfred/internal/contract"
	"github.com/sachncs/alfred/internal/tool"
)

// Agent is the base interface every concrete agent (ChatAgent,
// ResearchAgent, WorkflowAgent, etc.) must satisfy.
//
// Implementations are responsible for managing the turn loop, dispatching
// tool calls, and producing events. The runtime host (Phase 2) wires
// these into the HTTP server.
type Agent interface {
	// ID returns the stable, machine-readable identifier of the agent
	// (e.g. "alfred", "code_chat", "research"). Used for routing and
	// capability discovery.
	ID() string

	// StartTurn runs a new turn against the given thread. The returned
	// Turn is owned by the caller and reflects the final state.
	StartTurn(ctx context.Context, thread *contract.Thread, input contract.UserInput) (*contract.Turn, error)

	// Interrupt cancels the in-flight turn on the given thread. Returns
	// an error if there is no active turn or the agent is single-shot.
	Interrupt(threadID contract.ThreadID) error

	// Resume rebuilds the agent's view of a thread from storage and
	// returns the most recent turn. Used after a process restart.
	Resume(threadID contract.ThreadID) (*contract.Turn, error)

	// RegisterTool registers a tool with the agent. Tools registered
	// after StartTurn begin will NOT participate in that turn; the
	// runtime re-registers the tool catalog per turn.
	RegisterTool(t tool.Tool)

	// Tools returns the current tool catalog (defensive copy).
	Tools() []tool.Tool
}

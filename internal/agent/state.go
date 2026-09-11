package agent

import (
	"sync"
	"time"

	"github.com/sachncs/alfred/internal/contract"
)

// AgentState is the in-memory state of a running agent instance.
// Persisted state lives in internal/store; this is the runtime view.
type AgentState struct {
	mu sync.Mutex

	// CurrentThreadID is the thread the agent is actively processing.
	// Empty when idle.
	CurrentThreadID contract.ThreadID

	// CurrentTurnID is the turn within the active thread.
	CurrentTurnID contract.TurnID

	// TurnCount is the total number of turns completed since boot.
	TurnCount int

	// LastError is the last error the agent encountered, if any.
	LastError error

	// LastTurnAt is the timestamp of the most recent turn completion.
	LastTurnAt time.Time

	// StartedAt is the timestamp the agent was constructed.
	StartedAt time.Time
}

// NewAgentState constructs a fresh AgentState.
func NewAgentState() *AgentState {
	return &AgentState{StartedAt: time.Now().UTC()}
}

// BeginTurn marks the agent as busy with the given thread/turn.
func (s *AgentState) BeginTurn(threadID contract.ThreadID, turnID contract.TurnID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CurrentThreadID = threadID
	s.CurrentTurnID = turnID
}

// EndTurn marks the agent as idle after the given turn. It records the
// completion timestamp and (optionally) the final error.
func (s *AgentState) EndTurn(turnID contract.TurnID, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CurrentThreadID = ""
	s.CurrentTurnID = ""
	s.TurnCount++
	s.LastError = err
	s.LastTurnAt = time.Now().UTC()
	_ = turnID // reserved for future per-turn accounting
}

// Snapshot returns a defensive copy of the state.
func (s *AgentState) Snapshot() AgentState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return AgentState{
		CurrentThreadID: s.CurrentThreadID,
		CurrentTurnID:   s.CurrentTurnID,
		TurnCount:       s.TurnCount,
		LastError:       s.LastError,
		LastTurnAt:      s.LastTurnAt,
		StartedAt:       s.StartedAt,
	}
}

// IsBusy returns true if the agent is currently processing a turn.
func (s *AgentState) IsBusy() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.CurrentThreadID != ""
}

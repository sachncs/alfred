package runtime

import (
	"sync"
	"time"

	"github.com/alfred/alfred/internal/contract"
)

// RuntimeState tracks the mutable state of a running runtime.
type RuntimeState struct {
	mu sync.RWMutex

	// Status is the current status of the runtime.
	Status string

	// StartedAt is when the runtime was started.
	StartedAt time.Time

	// ActiveThreads is the count of threads with running turns.
	ActiveThreads int

	// TotalTurns is the total number of turns executed.
	TotalTurns int64

	// LastError is the most recent error, if any.
	LastError error
}

// NewRuntimeState creates a new RuntimeState.
func NewRuntimeState() *RuntimeState {
	return &RuntimeState{
		Status: "idle",
	}
}

// SetRunning marks the runtime as running.
func (s *RuntimeState) SetRunning() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status = "running"
	s.StartedAt = time.Now().UTC()
}

// SetStopped marks the runtime as stopped.
func (s *RuntimeState) SetStopped() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status = "stopped"
}

// RecordTurn increments the turn counter.
func (s *RuntimeState) RecordTurn() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalTurns++
}

// SetError records an error.
func (s *RuntimeState) SetError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastError = err
}

// Snapshot returns a defensive copy of the state.
func (s *RuntimeState) Snapshot() RuntimeState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return RuntimeState{
		Status:        s.Status,
		StartedAt:     s.StartedAt,
		ActiveThreads: s.ActiveThreads,
		TotalTurns:    s.TotalTurns,
		LastError:     s.LastError,
	}
}

// ThreadStatus tracks the status of a single thread.
type ThreadStatus struct {
	ID        contract.ThreadID
	Status    contract.ThreadStatus
	StartedAt time.Time
	EndedAt   time.Time
}

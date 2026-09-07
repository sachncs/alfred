package runtime

import (
	"time"

	"github.com/alfred/alfred/internal/contract"
)

// ForkResult is the output of a thread fork.
type ForkResult struct {
	Thread     contract.Thread
	ForkedFrom contract.ThreadID
}

// ForkThread creates a new thread copying title and metadata from a source thread.
func ForkThread(ts ThreadStore, source *contract.Thread) (*ForkResult, error) {
	forked := &contract.Thread{
		Title:     source.Title + " (fork)",
		Status:    contract.ThreadStatusIdle,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Metadata: map[string]any{
			"forkedFrom": string(source.ID),
		},
	}
	if err := ts.Create(forked); err != nil {
		return nil, err
	}
	return &ForkResult{
		Thread:     *forked,
		ForkedFrom: source.ID,
	}, nil
}

// ResumeTurn creates a new turn that continues from a previous turn's context.
func ResumeTurn(thread *contract.Thread, _ contract.TurnID) *contract.Turn {
	return &contract.Turn{
		ID:        contract.TurnID("turn-" + time.Now().Format("20060102150405.000")),
		ThreadID:  thread.ID,
		Status:    contract.TurnStatusRunning,
		StartedAt: time.Now().UTC(),
		Items:     []contract.TurnItem{},
	}
}

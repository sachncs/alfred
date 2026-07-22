package runtime

import (
	"github.com/alfred/alfred/internal/contract"
	"github.com/alfred/alfred/internal/model"
)

// Compactor summarizes older turns to keep the context window under budget.
type Compactor struct {
	client       model.Client
	maxTurns     int
	summaryAfter int
}

// NewCompactor creates a Compactor that summarizes after summaryAfter turns
// and keeps at most maxTurns in context.
func NewCompactor(client model.Client, maxTurns, summaryAfter int) *Compactor {
	if maxTurns <= 0 {
		maxTurns = 40
	}
	if summaryAfter <= 0 {
		summaryAfter = 20
	}
	return &Compactor{client: client, maxTurns: maxTurns, summaryAfter: summaryAfter}
}

// MaybeCompact returns a compacted copy of the thread's turns if over budget.
// Returns the original turns if no compaction is needed.
func (c *Compactor) MaybeCompact(turns []contract.Turn) []contract.Turn {
	if len(turns) <= c.summaryAfter {
		return turns
	}

	// Keep the most recent summaryAfter turns, mark the rest as compaction-eligible.
	cutoff := len(turns) - c.summaryAfter
	compacted := make([]contract.Turn, 0, len(turns))

	// First turn gets a summary marker.
	summary := contract.Turn{
		ID:     contract.TurnID("compact-summary"),
		Status: contract.TurnStatusCompleted,
		Items: []contract.TurnItem{{
			Kind: contract.ItemKindToolResult,
			Text: strPtr("[compacted: older conversation summarized]"),
		}},
	}
	compacted = append(compacted, summary)

	// Append recent turns.
	compacted = append(compacted, turns[cutoff:]...)
	return compacted
}

// ShouldCompact reports whether compaction is needed.
func (c *Compactor) ShouldCompact(turnCount int) bool {
	return turnCount > c.summaryAfter
}

func strPtr(s string) *string { return &s }

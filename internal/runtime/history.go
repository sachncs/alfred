package runtime

import (
	"github.com/alfred/alfred/internal/contract"
)

// HistoryPruner removes old turns that fall outside the retention window.
type HistoryPruner struct {
	maxTurns int
}

// NewHistoryPruner creates a pruner that keeps at most maxTurns.
func NewHistoryPruner(maxTurns int) *HistoryPruner {
	if maxTurns <= 0 {
		maxTurns = 100
	}
	return &HistoryPruner{maxTurns: maxTurns}
}

// Prune removes old turns, keeping the most recent maxTurns.
func (p *HistoryPruner) Prune(turns []contract.Turn) []contract.Turn {
	if len(turns) <= p.maxTurns {
		return turns
	}
	return turns[len(turns)-p.maxTurns:]
}

// ShouldPrune reports whether pruning is needed.
func (p *HistoryPruner) ShouldPrune(turnCount int) bool {
	return turnCount > p.maxTurns
}

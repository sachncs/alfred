package runtime

import (
	"sync"
)

// ToolBudget limits how many times a tool can be called per turn.
type ToolBudget struct {
	mu       sync.Mutex
	limits   map[string]int // tool name → max calls per turn
	counts   map[string]int // tool name → current calls in this turn
	turnID   string
}

// NewToolBudget creates a budget with per-tool limits. 0 = unlimited.
func NewToolBudget(limits map[string]int) *ToolBudget {
	return &ToolBudget{limits: limits, counts: make(map[string]int)}
}

// StartTurn resets counters for a new turn.
func (b *ToolBudget) StartTurn(turnID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.turnID = turnID
	b.counts = make(map[string]int)
}

// Allow reports whether the tool can be called again.
func (b *ToolBudget) Allow(toolName string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	limit, ok := b.limits[toolName]
	if !ok || limit <= 0 {
		return true
	}
	return b.counts[toolName] < limit
}

// Record marks a tool call. Panics if Allow was false (caller bug).
func (b *ToolBudget) Record(toolName string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.counts[toolName]++
}

// Remaining returns how many more calls are allowed for a tool.
func (b *ToolBudget) Remaining(toolName string) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	limit, ok := b.limits[toolName]
	if !ok || limit <= 0 {
		return -1 // unlimited
	}
	r := limit - b.counts[toolName]
	if r < 0 {
		return 0
	}
	return r
}

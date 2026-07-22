package runtime

import (
	"sync"
	"time"
)

// UsageRecord tracks a single token usage event.
type UsageRecord struct {
	Timestamp    time.Time
	Model        string
	InputTokens  int
	OutputTokens int
}

// UsageTracker aggregates usage across turns.
type UsageTracker struct {
	mu       sync.Mutex
	records  []UsageRecord
	totalIn  int
	totalOut int
}

// NewUsageTracker creates a tracker.
func NewUsageTracker() *UsageTracker {
	return &UsageTracker{}
}

// Record adds a usage record.
func (t *UsageTracker) Record(model string, inputTokens, outputTokens int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.records = append(t.records, UsageRecord{
		Timestamp:    time.Now().UTC(),
		Model:        model,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
	})
	t.totalIn += inputTokens
	t.totalOut += outputTokens
}

// Totals returns total input and output tokens.
func (t *UsageTracker) Totals() (int, int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.totalIn, t.totalOut
}

// Recent returns the last n records.
func (t *UsageTracker) Recent(n int) []UsageRecord {
	t.mu.Lock()
	defer t.mu.Unlock()
	if n <= 0 || n > len(t.records) {
		n = len(t.records)
	}
	out := make([]UsageRecord, n)
	copy(out, t.records[len(t.records)-n:])
	return out
}

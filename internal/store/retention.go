package store

import (
	"time"

	"github.com/alfred/alfred/internal/contract"
)

// Retention prunes events older than maxAge or beyond maxPerThread limits.
//
// ponytail: separate "age" and "count" policies. Apply both — events older
// than maxAge go, then trim to maxPerThread most recent.
type Retention struct {
	maxAge       time.Duration
	maxPerThread int
	hybrid       *HybridThreadStore
}

// NewRetention creates a Retention policy.
// maxAge=0 disables age pruning; maxPerThread=0 disables count pruning.
func NewRetention(h *HybridThreadStore, maxAge time.Duration, maxPerThread int) *Retention {
	return &Retention{hybrid: h, maxAge: maxAge, maxPerThread: maxPerThread}
}

// PruneAll applies the policy to every thread in the store.
func (r *Retention) PruneAll() error {
	threads, _, err := r.hybrid.List(1000, "")
	if err != nil {
		return err
	}
	for _, th := range threads {
		if err := r.PruneThread(th.ID); err != nil {
			return err
		}
	}
	return nil
}

// PruneThread prunes a single thread.
func (r *Retention) PruneThread(threadID contract.ThreadID) error {
	events, err := r.hybrid.ReadEvents(threadID, 0)
	if err != nil {
		return err
	}

	cutoff := time.Now().Add(-r.maxAge)
	filtered := events
	if r.maxAge > 0 {
		kept := events[:0]
		for _, e := range events {
			if e.CreatedAt.After(cutoff) {
				kept = append(kept, e)
			}
		}
		filtered = kept
	}

	if r.maxPerThread > 0 && len(filtered) > r.maxPerThread {
		filtered = filtered[len(filtered)-r.maxPerThread:]
	}

	if len(filtered) == len(events) {
		return nil
	}
	return r.hybrid.ReplaceJSONL(threadID, filtered)
}

package runtime

import "sync"

// CacheTelemetry tracks prompt cache hits and misses.
type CacheTelemetry struct {
	mu     sync.Mutex
	hits   map[string]int
	misses map[string]int
}

// NewCacheTelemetry creates a tracker.
func NewCacheTelemetry() *CacheTelemetry {
	return &CacheTelemetry{
		hits:   make(map[string]int),
		misses: make(map[string]int),
	}
}

// RecordHit records a cache hit for the given prefix.
func (ct *CacheTelemetry) RecordHit(prefix string) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.hits[prefix]++
}

// RecordMiss records a cache miss for the given prefix.
func (ct *CacheTelemetry) RecordMiss(prefix string) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.misses[prefix]++
}

// Stats returns hits and misses for the given prefix.
func (ct *CacheTelemetry) Stats(prefix string) (hits, misses int) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	return ct.hits[prefix], ct.misses[prefix]
}

// TotalStats returns aggregate hit/miss counts across all prefixes.
func (ct *CacheTelemetry) TotalStats() (totalHits, totalMisses int) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	for _, v := range ct.hits {
		totalHits += v
	}
	for _, v := range ct.misses {
		totalMisses += v
	}
	return
}

// Reset clears all counters.
func (ct *CacheTelemetry) Reset() {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.hits = make(map[string]int)
	ct.misses = make(map[string]int)
}

package runtime

import (
	"fmt"
	"sync"
)

// MemoryScope identifies a memory storage scope.
type MemoryScope string

const (
	MemoryScopeUser      MemoryScope = "user"
	MemoryScopeWorkspace MemoryScope = "workspace"
	MemoryScopeProject   MemoryScope = "project"
)

// MemoryEntry is a single key-value memory entry.
type MemoryEntry struct {
	Key   string         `json:"key"`
	Value any            `json:"value"`
	Meta  map[string]any `json:"meta,omitempty"`
}

// MemoryStore provides per-scope key-value memory.
// ponytail: in-process map. Wire to SQLite-backed store when persistence matters.
type MemoryStore struct {
	mu      sync.RWMutex
	entries map[MemoryScope]map[string]MemoryEntry
}

// NewMemoryStore creates an empty memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{entries: make(map[MemoryScope]map[string]MemoryEntry)}
}

// Set creates or updates a key in the given scope.
func (s *MemoryStore) Set(scope MemoryScope, key string, value any) error {
	if key == "" {
		return fmt.Errorf("empty memory key")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.entries[scope]; !ok {
		s.entries[scope] = make(map[string]MemoryEntry)
	}
	s.entries[scope][key] = MemoryEntry{Key: key, Value: value}
	return nil
}

// Get retrieves a key from the given scope.
func (s *MemoryStore) Get(scope MemoryScope, key string) (*MemoryEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	scopeEntries, ok := s.entries[scope]
	if !ok {
		return nil, fmt.Errorf("scope %q has no entries", scope)
	}
	e, ok := scopeEntries[key]
	if !ok {
		return nil, fmt.Errorf("key %q not found in scope %s", scope, scope)
	}
	return &e, nil
}

// Delete removes a key from the given scope.
func (s *MemoryStore) Delete(scope MemoryScope, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	scopeEntries, ok := s.entries[scope]
	if !ok {
		return nil
	}
	if _, ok := scopeEntries[key]; !ok {
		return fmt.Errorf("key %q not found in scope %s", scope, scope)
	}
	delete(scopeEntries, key)
	return nil
}

// List returns all entries in the given scope.
func (s *MemoryStore) List(scope MemoryScope) ([]MemoryEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	scopeEntries, ok := s.entries[scope]
	if !ok {
		return []MemoryEntry{}, nil
	}
	out := make([]MemoryEntry, 0, len(scopeEntries))
	for _, e := range scopeEntries {
		out = append(out, e)
	}
	return out, nil
}

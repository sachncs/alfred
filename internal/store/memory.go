package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

// MemoryStore provides per-scope key-value memory persistence.
type MemoryStore struct {
	mu  sync.RWMutex
	dir string
}

// NewMemoryStore creates a memory store rooted at dir.
func NewMemoryStore(dir string) (*MemoryStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create memory store dir: %w", err)
	}
	return &MemoryStore{dir: dir}, nil
}

func (s *MemoryStore) scopePath(scope MemoryScope) string {
	return filepath.Join(s.dir, string(scope)+".json")
}

func (s *MemoryStore) load(scope MemoryScope) (map[string]MemoryEntry, error) {
	p := s.scopePath(scope)
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]MemoryEntry{}, nil
		}
		return nil, err
	}
	var entries map[string]MemoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func (s *MemoryStore) save(scope MemoryScope, entries map[string]MemoryEntry) error {
	return AtomicWrite(s.scopePath(scope), mustMarshal(entries))
}

// Set creates or updates a key in the given scope.
func (s *MemoryStore) Set(scope MemoryScope, key string, value any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := s.load(scope)
	if err != nil {
		return err
	}
	entries[key] = MemoryEntry{Key: key, Value: value}
	return s.save(scope, entries)
}

// Get retrieves a key from the given scope.
func (s *MemoryStore) Get(scope MemoryScope, key string) (*MemoryEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := s.load(scope)
	if err != nil {
		return nil, err
	}
	e, ok := entries[key]
	if !ok {
		return nil, fmt.Errorf("key %q not found in scope %s", key, scope)
	}
	return &e, nil
}

// Delete removes a key from the given scope.
func (s *MemoryStore) Delete(scope MemoryScope, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := s.load(scope)
	if err != nil {
		return err
	}
	delete(entries, key)
	return s.save(scope, entries)
}

// List returns all entries in the given scope.
func (s *MemoryStore) List(scope MemoryScope) ([]MemoryEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := s.load(scope)
	if err != nil {
		return nil, err
	}
	out := make([]MemoryEntry, 0, len(entries))
	for _, e := range entries {
		out = append(out, e)
	}
	return out, nil
}

func mustMarshal(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

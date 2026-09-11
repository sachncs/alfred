// Package store provides persistence implementations for threads,
// sessions, and memory.
package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/sachncs/alfred/internal/contract"
)

// FileThreadStore persists threads as individual JSON files in a directory.
type FileThreadStore struct {
	mu  sync.RWMutex
	dir string
}

// NewFileThreadStore creates a store rooted at dir, creating it if needed.
func NewFileThreadStore(dir string) (*FileThreadStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create thread store dir: %w", err)
	}
	return &FileThreadStore{dir: dir}, nil
}

func (s *FileThreadStore) path(id contract.ThreadID) string {
	return filepath.Join(s.dir, string(id)+".json")
}

// Create persists a new thread. Fails if it already exists.
func (s *FileThreadStore) Create(t *contract.Thread) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.path(t.ID)
	if _, err := os.Stat(p); err == nil {
		return fmt.Errorf("thread %s already exists", t.ID)
	}
	return atomicWriteJSON(p, t)
}

// Get reads a thread by ID.
func (s *FileThreadStore) Get(id contract.ThreadID) (*contract.Thread, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p := s.path(id)
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("thread %s not found", id)
		}
		return nil, err
	}
	var t contract.Thread
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// List returns all threads, sorted by CreatedAt descending.
func (s *FileThreadStore) List(limit int, token string) ([]contract.Thread, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, "", err
	}

	var threads []contract.Thread
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.dir, e.Name()))
		if err != nil {
			continue
		}
		var t contract.Thread
		if err := json.Unmarshal(data, &t); err != nil {
			continue
		}
		threads = append(threads, t)
	}

	sort.Slice(threads, func(i, j int) bool {
		return threads[i].CreatedAt.After(threads[j].CreatedAt)
	})

	// Simple pagination: skip past token, take limit.
	start := 0
	if token != "" {
		for i, t := range threads {
			if string(t.ID) == token {
				start = i + 1
				break
			}
		}
	}
	if start > len(threads) {
		start = len(threads)
	}
	threads = threads[start:]

	nextToken := ""
	if limit > 0 && len(threads) > limit {
		nextToken = string(threads[limit-1].ID)
		threads = threads[:limit]
	}

	return threads, nextToken, nil
}

// Update overwrites an existing thread.
func (s *FileThreadStore) Update(t *contract.Thread) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.path(t.ID)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return fmt.Errorf("thread %s not found", t.ID)
	}
	return atomicWriteJSON(p, t)
}

// Delete removes a thread file.
func (s *FileThreadStore) Delete(id contract.ThreadID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.path(id)
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// atomicWriteJSON writes v as JSON to path atomically (tmp + rename).
func atomicWriteJSON(path string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

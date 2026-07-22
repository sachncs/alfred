package store

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/alfred/alfred/internal/contract"
)

// FileSessionStore persists turn events as JSONL (one JSON object per line)
// in per-thread files. Append-only for writes, full scan for reads.
type FileSessionStore struct {
	mu  sync.RWMutex
	dir string
}

// NewFileSessionStore creates a session store rooted at dir.
func NewFileSessionStore(dir string) (*FileSessionStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create session store dir: %w", err)
	}
	return &FileSessionStore{dir: dir}, nil
}

func (s *FileSessionStore) path(threadID contract.ThreadID) string {
	return filepath.Join(s.dir, string(threadID)+".jsonl")
}

// Append appends events to the thread's JSONL file.
func (s *FileSessionStore) Append(threadID contract.ThreadID, events []contract.TurnItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	p := s.path(threadID)
	f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	enc := json.NewEncoder(f)
	for _, ev := range events {
		if err := enc.Encode(ev); err != nil {
			return err
		}
	}
	return nil
}

// Read returns events for the thread starting from offset (0-indexed line).
func (s *FileSessionStore) Read(threadID contract.ThreadID, offset int) ([]contract.TurnItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p := s.path(threadID)
	f, err := os.Open(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var events []contract.TurnItem
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024) // 1 MiB max line
	idx := 0
	for scanner.Scan() {
		if idx < offset {
			idx++
			continue
		}
		var ev contract.TurnItem
		if err := json.Unmarshal(scanner.Bytes(), &ev); err != nil {
			continue // skip malformed lines
		}
		events = append(events, ev)
		idx++
	}
	return events, scanner.Err()
}

// Prune keeps only the last maxEvents events for the thread.
func (s *FileSessionStore) Prune(threadID contract.ThreadID, maxEvents int) error {
	events, err := s.Read(threadID, 0)
	if err != nil {
		return err
	}
	if len(events) <= maxEvents {
		return nil
	}
	events = events[len(events)-maxEvents:]

	s.mu.Lock()
	defer s.mu.Unlock()

	p := s.path(threadID)
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	enc := json.NewEncoder(f)
	for _, ev := range events {
		if err := enc.Encode(ev); err != nil {
			return err
		}
	}
	return nil
}

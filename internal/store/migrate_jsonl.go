package store

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/alfred/alfred/internal/contract"
)

// MigrateFromJSONL imports legacy JSONL files from srcDir into the hybrid store.
//
// Layout expected:
//   - threads.jsonl: one contract.Thread per line
//   - <threadID>.jsonl: one contract.TurnItem per line (Phase 2 layout)
//
// Idempotent: a thread already in SQLite is skipped.
// Threads are imported first, then events — the events table has a FK to threads.
func MigrateFromJSONL(hybrid *HybridThreadStore, srcDir string) error {
	existing := map[string]bool{}
	if list, _, err := hybrid.List(10000, ""); err == nil {
		for _, t := range list {
			existing[string(t.ID)] = true
		}
	}

	threadsFile := filepath.Join(srcDir, "threads.jsonl")
	if _, err := os.Stat(threadsFile); err == nil {
		if err := migrateThreads(hybrid, threadsFile, existing); err != nil {
			return err
		}
		// ponytail: refresh existing set so events for newly-imported threads are appended
		if list, _, err := hybrid.List(10000, ""); err == nil {
			for _, t := range list {
				existing[string(t.ID)] = true
			}
		}
	}

	// Walk all <id>.jsonl files (skipping threads.jsonl).
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if filepath.Ext(name) != ".jsonl" || name == "threads.jsonl" {
			continue
		}
		threadID := contract.ThreadID(name[:len(name)-len(".jsonl")])
		if err := migrateEvents(hybrid, threadID, filepath.Join(srcDir, name)); err != nil {
			return err
		}
	}
	return nil
}

func migrateThreads(hybrid *HybridThreadStore, path string, skip map[string]bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	for scanner.Scan() {
		var t contract.Thread
		if err := json.Unmarshal(scanner.Bytes(), &t); err != nil {
			continue
		}
		if skip[string(t.ID)] {
			continue
		}
		if err := hybrid.Create(&t); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func migrateEvents(hybrid *HybridThreadStore, threadID contract.ThreadID, path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer func() { _ = f.Close() }()

	var items []contract.TurnItem
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	for scanner.Scan() {
		var it contract.TurnItem
		if err := json.Unmarshal(scanner.Bytes(), &it); err != nil {
			continue
		}
		if it.CreatedAt.IsZero() {
			it.CreatedAt = time.Now().UTC()
		}
		items = append(items, it)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	// ponytail: skip events for threads that don't exist — they're orphan files
	if _, err := hybrid.Get(threadID); err != nil {
		return nil
	}
	return hybrid.AppendEvents(threadID, items)
}

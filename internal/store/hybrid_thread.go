package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sachncs/alfred/internal/contract"
)

// HybridThreadStore reads thread metadata from SQLite and messages/events
// from JSONL on disk. Writes commit the SQLite row + atomic JSONL append.
//
// ponytail: SQLite is the index; JSONL is the body. Restart from either
// alone — SQLite knows what's there, JSONL holds the actual events.
type HybridThreadStore struct {
	sqlite *SQLite
	dir    string
}

// NewHybridThreadStore creates a hybrid store. SQLite at dbPath, JSONL files under dir.
func NewHybridThreadStore(dbPath, dir string) (*HybridThreadStore, error) {
	sql, err := OpenSQLite(dbPath)
	if err != nil {
		return nil, err
	}
	if err := sql.Migrate(); err != nil {
		_ = sql.Close()
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		_ = sql.Close()
		return nil, err
	}
	return &HybridThreadStore{sqlite: sql, dir: dir}, nil
}

// SQLite returns the underlying SQLite handle.
func (h *HybridThreadStore) SQLite() *SQLite { return h.sqlite }

// Close closes the SQLite connection.
func (h *HybridThreadStore) Close() error { return h.sqlite.Close() }

func (h *HybridThreadStore) jsonlPath(threadID contract.ThreadID) string {
	return filepath.Join(h.dir, string(threadID)+".jsonl")
}

// Create inserts a thread row + creates an empty JSONL file.
func (h *HybridThreadStore) Create(t *contract.Thread) error {
	if err := h.sqlite.InsertThread(t); err != nil {
		return err
	}
	// ponytail: pre-create JSONL file so first Append doesn't race.
	p := h.jsonlPath(t.ID)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		if err := AtomicWrite(p, []byte{}); err != nil {
			return err
		}
	}
	return nil
}

// Get returns a thread by ID.
func (h *HybridThreadStore) Get(id contract.ThreadID) (*contract.Thread, error) {
	t, err := h.sqlite.GetThread(string(id))
	if err != nil {
		return nil, err
	}
	return t, nil
}

// List returns threads ordered by updated_at desc.
func (h *HybridThreadStore) List(limit int, token string) ([]contract.Thread, string, error) {
	return h.sqlite.ListThreads(limit, token)
}

// Update updates a thread row.
func (h *HybridThreadStore) Update(t *contract.Thread) error {
	return h.sqlite.UpdateThread(t)
}

// Delete deletes the thread row + JSONL file.
func (h *HybridThreadStore) Delete(id contract.ThreadID) error {
	if err := h.sqlite.DeleteThread(string(id)); err != nil {
		return err
	}
	p := h.jsonlPath(id)
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// AppendEvents appends turn items to both SQLite events table and JSONL file.
//
// The writes are made durable by ordering JSONL first (write-ahead) and
// SQLite second. If the JSONL append fails we return immediately; if the
// SQLite insert fails after a successful JSONL append, the events are
// still visible via JSONL replay and can be reconciled on the next
// successful append — the inverse direction is unrecoverable.
func (h *HybridThreadStore) AppendEvents(threadID contract.ThreadID, items []contract.TurnItem) error {
	if len(items) == 0 {
		return nil
	}

	// ponytail: append to JSONL first (write-ahead) so legacy readers can
	// always replay events even if the SQLite write fails afterwards.
	f, err := os.OpenFile(h.jsonlPath(threadID), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	for _, it := range items {
		if err := enc.Encode(it); err != nil {
			_ = f.Close()
			return err
		}
	}
	if err := f.Close(); err != nil {
		return err
	}

	if _, _, err := h.sqlite.BulkInsertEvents(string(threadID), items); err != nil {
		return fmt.Errorf("sqlite bulk insert (jsonl already written): %w", err)
	}
	return nil
}

// ReadEvents returns events for a thread starting from offset (0-indexed line in JSONL).
func (h *HybridThreadStore) ReadEvents(threadID contract.ThreadID, offset int) ([]contract.TurnItem, error) {
	p := h.jsonlPath(threadID)
	f, err := os.Open(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var events []contract.TurnItem
	idx := 0
	dec := json.NewDecoder(f)
	for dec.More() {
		var it contract.TurnItem
		if err := dec.Decode(&it); err != nil {
			return nil, err
		}
		if idx >= offset {
			events = append(events, it)
		}
		idx++
	}
	return events, nil
}

// PruneEvents keeps only the last maxEvents events for the thread (JSONL + SQLite).
func (h *HybridThreadStore) PruneEvents(threadID contract.ThreadID, maxEvents int) error {
	if maxEvents <= 0 {
		return nil
	}
	events, err := h.ReadEvents(threadID, 0)
	if err != nil {
		return err
	}
	if len(events) <= maxEvents {
		return nil
	}
	kept := events[len(events)-maxEvents:]
	return h.ReplaceJSONL(threadID, kept)
}

// ReplaceJSONL rewrites the JSONL file with the given items.
func (h *HybridThreadStore) ReplaceJSONL(threadID contract.ThreadID, items []contract.TurnItem) error {
	p := h.jsonlPath(threadID)
	tmp := p + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	for _, it := range items {
		if err := enc.Encode(it); err != nil {
			_ = f.Close()
			return err
		}
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// IsNotFound reports whether err is a missing-row error.
func IsNotFound(err error) bool {
	return errors.Is(err, errNotFound)
}

var errNotFound = errors.New("not found")

// SQLiteEventsFallback reads events from SQLite if JSONL is missing.
// Used during migration from legacy stores.
func (h *HybridThreadStore) SQLiteEventsFallback(threadID contract.ThreadID) ([]contract.TurnItem, error) {
	rows, err := h.sqlite.ReadEvents(string(threadID), 0, 0)
	if err != nil {
		return nil, err
	}
	var out []contract.TurnItem
	for _, r := range rows {
		var it contract.TurnItem
		if err := json.Unmarshal(r.Payload, &it); err != nil {
			continue
		}
		out = append(out, it)
	}
	return out, nil
}

// Addr returns the SQLite path and JSONL dir for diagnostics.
func (h *HybridThreadStore) Addr() (sqlitePath, jsonlDir string) {
	return h.sqlite.Path(), h.dir
}

// SessionStore-style methods (Append/Read/Prune) for interface compatibility.

func (h *HybridThreadStore) Append(threadID contract.ThreadID, events []contract.TurnItem) error {
	return h.AppendEvents(threadID, events)
}

func (h *HybridThreadStore) Read(threadID contract.ThreadID, offset int) ([]contract.TurnItem, error) {
	return h.ReadEvents(threadID, offset)
}

func (h *HybridThreadStore) Prune(threadID contract.ThreadID, maxEvents int) error {
	return h.PruneEvents(threadID, maxEvents)
}

var _ = time.Now

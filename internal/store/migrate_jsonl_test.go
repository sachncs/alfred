package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sachncs/alfred/internal/contract"
)

func TestMigrateFromJSONL(t *testing.T) {
	src := t.TempDir()

	// threads.jsonl with 1 thread.
	th1 := &contract.Thread{ID: "thr1", Title: "first", Status: contract.ThreadStatusIdle, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	writeJSONL(t, filepath.Join(src, "threads.jsonl"), th1)

	// events JSONL for thr1 + orphan file thr2.
	writeJSONL(t, filepath.Join(src, "thr1.jsonl"),
		contract.TurnItem{Kind: contract.ItemKindUserMessage, CreatedAt: time.Now().UTC()},
	)
	writeJSONL(t, filepath.Join(src, "thr2.jsonl"),
		contract.TurnItem{Kind: contract.ItemKindUserMessage, CreatedAt: time.Now().UTC()},
		contract.TurnItem{Kind: contract.ItemKindAssistantText, CreatedAt: time.Now().UTC()},
	)

	// Run migration.
	dst := t.TempDir()
	hybrid, err := NewHybridThreadStore(filepath.Join(dst, "test.db"), filepath.Join(dst, "events"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = hybrid.Close() }()

	if err := MigrateFromJSONL(hybrid, src); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// thr1: imported normally, not a placeholder.
	got, _ := hybrid.Get("thr1")
	if got == nil || got.Title != "first" {
		t.Errorf("thr1 not migrated: %+v", got)
	}
	if got != nil && got.Metadata["recovered"] != nil {
		t.Errorf("thr1 should not be marked recovered, got %v", got.Metadata)
	}
	events1, _ := hybrid.ReadEvents("thr1", 0)
	if len(events1) != 1 {
		t.Errorf("thr1 events = %d, want 1", len(events1))
	}

	// thr2: orphan — placeholder thread created, events imported, recovered flag set.
	got2, _ := hybrid.Get("thr2")
	if got2 == nil {
		t.Fatal("thr2 placeholder not created")
	}
	if got2.Title != "(recovered from orphan events)" {
		t.Errorf("placeholder title = %q", got2.Title)
	}
	if got2.Metadata["recovered"] != true {
		t.Errorf("placeholder should have recovered=true, got %v", got2.Metadata)
	}
	events2, _ := hybrid.ReadEvents("thr2", 0)
	if len(events2) != 2 {
		t.Errorf("thr2 events = %d, want 2 (orphan events should not be dropped)", len(events2))
	}

	// Idempotent: run again, no duplicates.
	if err := MigrateFromJSONL(hybrid, src); err != nil {
		t.Fatal(err)
	}
	list, _, _ := hybrid.List(100, "")
	if len(list) != 2 {
		t.Errorf("after re-run: list = %d, want 2", len(list))
	}
	events2After, _ := hybrid.ReadEvents("thr2", 0)
	if len(events2After) != 2 {
		t.Errorf("after re-run: thr2 events = %d, want 2 (idempotent)", len(events2After))
	}
}

func writeJSONL(t *testing.T, path string, items ...any) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	enc := json.NewEncoder(f)
	for _, it := range items {
		if err := enc.Encode(it); err != nil {
			t.Fatal(err)
		}
	}
}

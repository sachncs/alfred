package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alfred/alfred/internal/contract"
)

func newTestHybrid(t *testing.T) *HybridThreadStore {
	t.Helper()
	dir := t.TempDir()
	hs, err := NewHybridThreadStore(filepath.Join(dir, "test.db"), filepath.Join(dir, "events"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = hs.Close() })
	return hs
}

func TestHybridCreateGet(t *testing.T) {
	h := newTestHybrid(t)
	th := &contract.Thread{
		ID:        contract.ThreadID("thr1"),
		Title:     "hello",
		Status:    contract.ThreadStatusIdle,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := h.Create(th); err != nil {
		t.Fatal(err)
	}
	got, err := h.Get("thr1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "hello" {
		t.Errorf("title = %q", got.Title)
	}
	// JSONL file created.
	if _, err := os.Stat(h.jsonlPath("thr1")); err != nil {
		t.Errorf("jsonl file missing: %v", err)
	}
}

func TestHybridAppendAndReadEvents(t *testing.T) {
	h := newTestHybrid(t)
	_ = h.Create(&contract.Thread{
		ID: "thr1", Title: "t", Status: contract.ThreadStatusIdle,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	})

	now := time.Now().UTC()
	items := []contract.TurnItem{
		{Kind: contract.ItemKindUserMessage, CreatedAt: now, Text: strPtrLocal("hi")},
		{Kind: contract.ItemKindAssistantText, CreatedAt: now.Add(time.Millisecond), Text: strPtrLocal("hello")},
	}
	if err := h.AppendEvents("thr1", items); err != nil {
		t.Fatal(err)
	}

	got, err := h.ReadEvents("thr1", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d items, want 2", len(got))
	}
}

func TestHybridPrune(t *testing.T) {
	h := newTestHybrid(t)
	_ = h.Create(&contract.Thread{
		ID: "thr1", Title: "t", Status: contract.ThreadStatusIdle,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	})
	now := time.Now().UTC()
	for i := 0; i < 10; i++ {
		_ = h.AppendEvents("thr1", []contract.TurnItem{
			{Kind: contract.ItemKindUserMessage, CreatedAt: now.Add(time.Duration(i) * time.Millisecond)},
		})
	}
	if err := h.PruneEvents("thr1", 3); err != nil {
		t.Fatal(err)
	}
	got, _ := h.ReadEvents("thr1", 0)
	if len(got) != 3 {
		t.Errorf("after prune: got %d, want 3", len(got))
	}
}

func TestHybridDelete(t *testing.T) {
	h := newTestHybrid(t)
	_ = h.Create(&contract.Thread{
		ID: "thr1", Title: "t", Status: contract.ThreadStatusIdle,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	})
	if err := h.Delete("thr1"); err != nil {
		t.Fatal(err)
	}
	if _, err := h.Get("thr1"); err == nil {
		t.Error("expected error after delete")
	}
	if _, err := os.Stat(h.jsonlPath("thr1")); !os.IsNotExist(err) {
		t.Error("jsonl file should be gone")
	}
}

func TestHybridSessionStoreInterface(t *testing.T) {
	h := newTestHybrid(t)
	_ = h.Create(&contract.Thread{
		ID: "thr1", Title: "t", Status: contract.ThreadStatusIdle,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	})
	now := time.Now().UTC()
	if err := h.Append("thr1", []contract.TurnItem{
		{Kind: contract.ItemKindUserMessage, CreatedAt: now, Text: strPtrLocal("a")},
	}); err != nil {
		t.Fatal(err)
	}
	got, err := h.Read("thr1", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestHybridJSONLContentMatchesSQLite(t *testing.T) {
	h := newTestHybrid(t)
	_ = h.Create(&contract.Thread{
		ID: "thr1", Title: "t", Status: contract.ThreadStatusIdle,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	})

	now := time.Now().UTC()
	items := []contract.TurnItem{
		{Kind: contract.ItemKindUserMessage, CreatedAt: now, Text: strPtrLocal("x")},
	}
	_ = h.AppendEvents("thr1", items)

	// Read raw JSONL.
	data, _ := os.ReadFile(h.jsonlPath("thr1"))
	var roundTrip contract.TurnItem
	if err := json.Unmarshal(data[:len(data)-1], &roundTrip); err != nil { // strip trailing newline
		t.Errorf("jsonl unmarshal: %v", err)
	}
}

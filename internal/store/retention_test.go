package store

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/alfred/alfred/internal/contract"
)

func TestRetentionAgePrune(t *testing.T) {
	h := newTestHybrid(t)
	_ = h.Create(&contract.Thread{
		ID: "thr1", Title: "t", Status: contract.ThreadStatusIdle,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	})

	old := time.Now().Add(-2 * time.Hour)
	recent := time.Now().Add(-time.Minute)
	if err := h.AppendEvents("thr1", []contract.TurnItem{
		{Kind: contract.ItemKindUserMessage, CreatedAt: old},
		{Kind: contract.ItemKindUserMessage, CreatedAt: recent},
	}); err != nil {
		t.Fatal(err)
	}

	r := NewRetention(h, time.Hour, 0)
	if err := r.PruneThread("thr1"); err != nil {
		t.Fatal(err)
	}
	got, _ := h.ReadEvents("thr1", 0)
	if len(got) != 1 {
		t.Errorf("after age prune: got %d, want 1", len(got))
	}
}

func TestRetentionCountPrune(t *testing.T) {
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

	r := NewRetention(h, 0, 3)
	if err := r.PruneThread("thr1"); err != nil {
		t.Fatal(err)
	}
	got, _ := h.ReadEvents("thr1", 0)
	if len(got) != 3 {
		t.Errorf("after count prune: got %d, want 3", len(got))
	}
}

func TestRetentionNoopWhenWithinLimits(t *testing.T) {
	h := newTestHybrid(t)
	_ = h.Create(&contract.Thread{
		ID: "thr1", Title: "t", Status: contract.ThreadStatusIdle,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	})
	_ = h.AppendEvents("thr1", []contract.TurnItem{
		{Kind: contract.ItemKindUserMessage, CreatedAt: time.Now().UTC()},
	})
	r := NewRetention(h, time.Hour, 100)
	if err := r.PruneThread("thr1"); err != nil {
		t.Fatal(err)
	}
	got, _ := h.ReadEvents("thr1", 0)
	if len(got) != 1 {
		t.Errorf("noop prune changed event count: %d", len(got))
	}
	_ = json.Marshal
}

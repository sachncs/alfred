package store

import (
	"testing"

	"github.com/alfred/alfred/internal/contract"
)

func TestFileSessionStoreAppendAndRead(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewFileSessionStore(dir)

	events := []contract.TurnItem{
		{ID: "e1", Kind: contract.ItemKindUserMessage, Text: strPtr("hello")},
		{ID: "e2", Kind: contract.ItemKindAssistantText, Text: strPtr("world")},
	}
	if err := s.Append("thr1", events); err != nil {
		t.Fatal(err)
	}

	got, err := s.Read("thr1", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("read: got %d events, want 2", len(got))
	}
	if got[0].ID != "e1" {
		t.Errorf("event 0 ID = %q, want e1", got[0].ID)
	}
	if got[1].ID != "e2" {
		t.Errorf("event 1 ID = %q, want e2", got[1].ID)
	}
}

func TestFileSessionStoreAppendMultipleTimes(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewFileSessionStore(dir)

	s.Append("thr1", []contract.TurnItem{{ID: "e1"}})
	s.Append("thr1", []contract.TurnItem{{ID: "e2"}})
	s.Append("thr1", []contract.TurnItem{{ID: "e3"}})

	got, _ := s.Read("thr1", 0)
	if len(got) != 3 {
		t.Fatalf("read: got %d events, want 3", len(got))
	}
}

func TestFileSessionStoreReadWithOffset(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewFileSessionStore(dir)

	s.Append("thr1", []contract.TurnItem{
		{ID: "e1"}, {ID: "e2"}, {ID: "e3"},
	})

	got, _ := s.Read("thr1", 1)
	if len(got) != 2 {
		t.Fatalf("read offset 1: got %d events, want 2", len(got))
	}
	if got[0].ID != "e2" {
		t.Errorf("first event after offset 1 = %q, want e2", got[0].ID)
	}
}

func TestFileSessionStoreReadNonexistent(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewFileSessionStore(dir)

	got, err := s.Read("nope", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %d events", len(got))
	}
}

func TestFileSessionStorePrune(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewFileSessionStore(dir)

	for i := 0; i < 10; i++ {
		s.Append("thr1", []contract.TurnItem{{ID: contract.ItemID("e" + string(rune('0'+i)))}})
	}

	if err := s.Prune("thr1", 3); err != nil {
		t.Fatal(err)
	}

	got, _ := s.Read("thr1", 0)
	if len(got) != 3 {
		t.Fatalf("after prune: got %d events, want 3", len(got))
	}
	if got[0].ID != "e7" {
		t.Errorf("first event after prune = %q, want e7", got[0].ID)
	}
}

func TestFileSessionStorePruneNoop(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewFileSessionStore(dir)

	s.Append("thr1", []contract.TurnItem{{ID: "e1"}})
	if err := s.Prune("thr1", 10); err != nil {
		t.Fatal(err)
	}
	got, _ := s.Read("thr1", 0)
	if len(got) != 1 {
		t.Errorf("prune with high max should be noop, got %d", len(got))
	}
}

func strPtr(s string) *string { return &s }

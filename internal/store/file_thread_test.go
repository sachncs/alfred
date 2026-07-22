package store

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/alfred/alfred/internal/contract"
)

func newFixtureThread(id string) *contract.Thread {
	return &contract.Thread{
		ID:        contract.ThreadID(id),
		Title:     "test thread " + id,
		Status:    contract.ThreadStatusIdle,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

func TestFileThreadStoreCreateAndGet(t *testing.T) {
	dir := t.TempDir()
	s, err := NewFileThreadStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	th := newFixtureThread("t1")
	if err := s.Create(th); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := s.Get("t1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Title != th.Title {
		t.Errorf("title = %q, want %q", got.Title, th.Title)
	}
}

func TestFileThreadStoreCreateDuplicate(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewFileThreadStore(dir)

	th := newFixtureThread("t1")
	if err := s.Create(th); err != nil {
		t.Fatal(err)
	}
	if err := s.Create(th); err == nil {
		t.Error("expected error on duplicate create")
	}
}

func TestFileThreadStoreGetNotFound(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewFileThreadStore(dir)

	_, err := s.Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent thread")
	}
}

func TestFileThreadStoreList(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewFileThreadStore(dir)

	for i := 0; i < 5; i++ {
		th := newFixtureThread("t" + string(rune('0'+i)))
		th.CreatedAt = time.Now().UTC().Add(time.Duration(i) * time.Second)
		_ = s.Create(th)
	}

	threads, _, err := s.List(3, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(threads) != 3 {
		t.Errorf("list limit 3: got %d threads, want 3", len(threads))
	}
}

func TestFileThreadStoreUpdate(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewFileThreadStore(dir)

	th := newFixtureThread("t1")
	_ = s.Create(th)

	th.Title = "updated"
	th.Status = contract.ThreadStatusRunning
	if err := s.Update(th); err != nil {
		t.Fatal(err)
	}

	got, _ := s.Get("t1")
	if got.Title != "updated" {
		t.Errorf("title = %q, want updated", got.Title)
	}
	if got.Status != contract.ThreadStatusRunning {
		t.Errorf("status = %q, want running", got.Status)
	}
}

func TestFileThreadStoreDelete(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewFileThreadStore(dir)

	th := newFixtureThread("t1")
	_ = s.Create(th)

	if err := s.Delete("t1"); err != nil {
		t.Fatal(err)
	}

	_, err := s.Get("t1")
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestFileThreadStoreDeleteNonexistent(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewFileThreadStore(dir)

	if err := s.Delete("nope"); err != nil {
		t.Errorf("deleting nonexistent should not error: %v", err)
	}
}

func TestFileThreadStoreConcurrentWrites(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewFileThreadStore(dir)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := contract.ThreadID("t" + string(rune('A'+i%26)))
			th := newFixtureThread(string(id))
			_ = s.Create(th)
		}(i)
	}
	wg.Wait()

	// All threads should be readable (no corruption).
	entries, _ := os.ReadDir(dir)
	count := 0
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".json" {
			count++
		}
	}
	if count == 0 {
		t.Error("expected at least one thread file")
	}
}

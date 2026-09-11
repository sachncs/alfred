package store

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/sachncs/alfred/internal/contract"
)

func newTestSQLite(t *testing.T) *SQLite {
	t.Helper()
	db, err := OpenSQLite(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Migrate(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestThreadRowCRUD(t *testing.T) {
	db := newTestSQLite(t)

	th := &contract.Thread{
		ID:        contract.ThreadID("t1"),
		Title:     "hello",
		Status:    contract.ThreadStatusIdle,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Metadata:  map[string]any{"k": "v"},
	}
	if err := db.InsertThread(th); err != nil {
		t.Fatalf("insert: %v", err)
	}

	got, err := db.GetThread("t1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Title != "hello" {
		t.Errorf("title = %q", got.Title)
	}
	if got.Metadata["k"] != "v" {
		t.Errorf("metadata = %v", got.Metadata)
	}

	th.Title = "updated"
	th.UpdatedAt = time.Now().UTC()
	if err := db.UpdateThread(th); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, _ = db.GetThread("t1")
	if got.Title != "updated" {
		t.Errorf("after update: title = %q", got.Title)
	}

	if err := db.DeleteThread("t1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := db.GetThread("t1"); err == nil {
		t.Error("expected error after delete")
	}
}

func TestThreadList(t *testing.T) {
	db := newTestSQLite(t)
	for i := 0; i < 5; i++ {
		_ = db.InsertThread(&contract.Thread{
			ID:        contract.ThreadID(string(rune('a' + i))),
			Title:     "t",
			Status:    contract.ThreadStatusIdle,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC().Add(time.Duration(i) * time.Second),
		})
	}
	threads, token, err := db.ListThreads(3, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(threads) != 3 {
		t.Errorf("list limit 3: got %d", len(threads))
	}
	if token == "" {
		t.Error("expected next token")
	}
}

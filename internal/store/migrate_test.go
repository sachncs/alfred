package store

import (
	"path/filepath"
	"testing"
)

func TestMigrate(t *testing.T) {
	dir := t.TempDir()
	db, err := OpenSQLite(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	// schema_version row exists.
	var v int
	if err := db.QueryRow("SELECT version FROM schema_version").Scan(&v); err != nil {
		t.Fatalf("schema_version query: %v", err)
	}
	if v != 1 {
		t.Errorf("version = %d, want 1", v)
	}

	// All expected tables exist.
	wantTables := []string{"threads", "sessions", "turns", "events", "usage", "settings"}
	for _, name := range wantTables {
		var n int
		err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", name).Scan(&n)
		if err != nil {
			t.Errorf("query %s: %v", name, err)
		}
		if n != 1 {
			t.Errorf("table %s missing", name)
		}
	}
}

func TestMigrateIdempotent(t *testing.T) {
	dir := t.TempDir()
	db, err := OpenSQLite(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	for i := 0; i < 3; i++ {
		if err := db.Migrate(); err != nil {
			t.Fatalf("Migrate iteration %d: %v", i, err)
		}
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_version").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("schema_version rows = %d, want 1 (idempotent)", count)
	}
}

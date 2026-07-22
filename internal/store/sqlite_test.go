package store

import (
	"path/filepath"
	"testing"
)

func TestOpenSQLite(t *testing.T) {
	dir := t.TempDir()
	db, err := OpenSQLite(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	defer func() { _ = db.Close() }()

	var v int
	if err := db.QueryRow("SELECT 1").Scan(&v); err != nil {
		t.Fatalf("query: %v", err)
	}
	if v != 1 {
		t.Errorf("got %d, want 1", v)
	}
}

func TestSQLiteExec(t *testing.T) {
	dir := t.TempDir()
	db, err := OpenSQLite(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	if _, err := db.Exec("CREATE TABLE t (x INTEGER)"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO t VALUES (42)"); err != nil {
		t.Fatal(err)
	}
	var x int
	if err := db.QueryRow("SELECT x FROM t").Scan(&x); err != nil {
		t.Fatal(err)
	}
	if x != 42 {
		t.Errorf("got %d, want 42", x)
	}
}

func TestSQLiteClose(t *testing.T) {
	dir := t.TempDir()
	db, err := OpenSQLite(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Errorf("close: %v", err)
	}
}

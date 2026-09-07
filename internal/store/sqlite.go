package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

// SQLite is a connection manager for a single SQLite database file.
//
// ponytail: single connection, no pool — SQLite serialises writes anyway, and
// modernc.org/sqlite is in-process so a pool buys nothing. Swap for sqlx +
// connection pool if throughput becomes a bottleneck.
type SQLite struct {
	mu   sync.Mutex
	db   *sql.DB
	path string
}

// OpenSQLite opens (or creates) a SQLite database at the given path.
func OpenSQLite(path string) (*SQLite, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("mkdir: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	// ponytail: SQLite wants single-writer; cap to 1 conn.
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("wal: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return nil, fmt.Errorf("foreign_keys: %w", err)
	}
	return &SQLite{db: db, path: path}, nil
}

// Path returns the database file path.
func (s *SQLite) Path() string { return s.path }

// DB returns the underlying *sql.DB. Callers should not close it.
func (s *SQLite) DB() *sql.DB { return s.db }

// Close closes the database.
func (s *SQLite) Close() error { return s.db.Close() }

// Exec runs a query that doesn't return rows.
func (s *SQLite) Exec(query string, args ...any) (sql.Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Exec(query, args...)
}

// Query runs a query that returns rows.
func (s *SQLite) Query(query string, args ...any) (*sql.Rows, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Query(query, args...)
}

// QueryRow runs a query that returns a single row.
func (s *SQLite) QueryRow(query string, args ...any) *sql.Row {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.QueryRow(query, args...)
}

// Tx runs fn inside a transaction.
func (s *SQLite) Tx(fn func(*sql.Tx) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

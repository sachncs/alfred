package store

import (
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strings"
	"time"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// migrations are applied in lexical order by filename.
type migration struct {
	version int
	name    string
	sql     string
}

func loadMigrations() ([]migration, error) {
	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return nil, err
	}
	var out []migration
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		body, err := migrationFS.ReadFile("migrations/" + e.Name())
		if err != nil {
			return nil, err
		}
		var version int
		if _, err := fmt.Sscanf(e.Name(), "%04d_", &version); err != nil {
			return nil, fmt.Errorf("bad migration filename %q: %w", e.Name(), err)
		}
		out = append(out, migration{version: version, name: e.Name(), sql: string(body)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].version < out[j].version })
	return out, nil
}

// Migrate applies all pending migrations to the database.
func (s *SQLite) Migrate() error {
	if _, err := s.Exec(`CREATE TABLE IF NOT EXISTS schema_version (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		return fmt.Errorf("create schema_version: %w", err)
	}

	migs, err := loadMigrations()
	if err != nil {
		return err
	}

	applied, err := s.appliedVersions()
	if err != nil {
		return err
	}

	for _, m := range migs {
		if applied[m.version] {
			continue
		}
		if err := s.applyMigration(m); err != nil {
			return fmt.Errorf("migration %s: %w", m.name, err)
		}
	}
	return nil
}

func (s *SQLite) appliedVersions() (map[int]bool, error) {
	rows, err := s.Query("SELECT version FROM schema_version")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make(map[int]bool)
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out[v] = true
	}
	return out, rows.Err()
}

func (s *SQLite) applyMigration(m migration) error {
	return s.Tx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(m.sql); err != nil {
			return err
		}
		_, err := tx.Exec("INSERT INTO schema_version (version, applied_at) VALUES (?, ?)",
			m.version, time.Now().UTC().Format(time.RFC3339))
		return err
	})
}

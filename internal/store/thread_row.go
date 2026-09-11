package store

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/sachncs/alfred/internal/contract"
)

// ThreadRow mirrors the threads table.
type ThreadRow struct {
	ID        string
	Title     string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
	Metadata  map[string]any
}

// threadRowFromContract converts a contract.Thread to a ThreadRow.
func threadRowFromContract(t *contract.Thread) ThreadRow {
	meta := t.Metadata
	if meta == nil {
		meta = map[string]any{}
	}
	return ThreadRow{
		ID:        string(t.ID),
		Title:     t.Title,
		Status:    string(t.Status),
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
		Metadata:  meta,
	}
}

// contractFromThreadRow converts a ThreadRow back to a contract.Thread.
func contractFromThreadRow(r ThreadRow) contract.Thread {
	return contract.Thread{
		ID:        contract.ThreadID(r.ID),
		Title:     r.Title,
		Status:    contract.ThreadStatus(r.Status),
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
		Metadata:  r.Metadata,
	}
}

// InsertThread inserts a thread row.
func (s *SQLite) InsertThread(t *contract.Thread) error {
	row := threadRowFromContract(t)
	metaJSON, err := json.Marshal(row.Metadata)
	if err != nil {
		return err
	}
	_, err = s.Exec(`INSERT INTO threads (id, title, status, created_at, updated_at, metadata_json)
		VALUES (?, ?, ?, ?, ?, ?)`,
		row.ID, row.Title, row.Status,
		row.CreatedAt.UTC().Format(time.RFC3339),
		row.UpdatedAt.UTC().Format(time.RFC3339),
		string(metaJSON))
	return err
}

// GetThread fetches a thread row by ID.
func (s *SQLite) GetThread(id string) (*contract.Thread, error) {
	row := s.QueryRow(`SELECT title, status, created_at, updated_at, metadata_json
		FROM threads WHERE id = ?`, id)
	var r ThreadRow
	r.ID = id
	var created, updated, metaJSON string
	if err := row.Scan(&r.Title, &r.Status, &created, &updated, &metaJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	r.CreatedAt, _ = time.Parse(time.RFC3339, created)
	r.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
	if metaJSON != "" {
		_ = json.Unmarshal([]byte(metaJSON), &r.Metadata)
	}
	t := contractFromThreadRow(r)
	return &t, nil
}

// UpdateThread updates a thread row.
func (s *SQLite) UpdateThread(t *contract.Thread) error {
	row := threadRowFromContract(t)
	metaJSON, _ := json.Marshal(row.Metadata)
	_, err := s.Exec(`UPDATE threads SET title=?, status=?, updated_at=?, metadata_json=? WHERE id=?`,
		row.Title, row.Status,
		row.UpdatedAt.UTC().Format(time.RFC3339),
		string(metaJSON), row.ID)
	return err
}

// DeleteThread deletes a thread row.
func (s *SQLite) DeleteThread(id string) error {
	_, err := s.Exec("DELETE FROM threads WHERE id=?", id)
	return err
}

// ListThreads returns threads ordered by updated_at desc, paginated by limit+token.
func (s *SQLite) ListThreads(limit int, token string) ([]contract.Thread, string, error) {
	if limit <= 0 {
		limit = 50
	}
	var beforeTime time.Time
	if token != "" {
		t, err := time.Parse(time.RFC3339, token)
		if err == nil {
			beforeTime = t
		}
	}
	q := `SELECT id, title, status, created_at, updated_at, metadata_json FROM threads`
	args := []any{}
	if !beforeTime.IsZero() {
		q += ` WHERE updated_at < ?`
		args = append(args, beforeTime.UTC().Format(time.RFC3339))
	}
	q += ` ORDER BY updated_at DESC LIMIT ?`
	args = append(args, limit+1)

	rows, err := s.Query(q, args...)
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = rows.Close() }()

	var out []contract.Thread
	for rows.Next() {
		var r ThreadRow
		var created, updated, metaJSON string
		if err := rows.Scan(&r.ID, &r.Title, &r.Status, &created, &updated, &metaJSON); err != nil {
			return nil, "", err
		}
		r.CreatedAt, _ = time.Parse(time.RFC3339, created)
		r.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		if metaJSON != "" {
			_ = json.Unmarshal([]byte(metaJSON), &r.Metadata)
		}
		out = append(out, contractFromThreadRow(r))
	}
	var nextToken string
	if len(out) > limit {
		nextToken = out[limit-1].UpdatedAt.Format(time.RFC3339)
		out = out[:limit]
	}
	return out, nextToken, rows.Err()
}

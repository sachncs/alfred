package store

import (
	"database/sql"
	"time"
)

// SessionRow mirrors the sessions table.
type SessionRow struct {
	ID        string
	ThreadID  string
	StartedAt time.Time
	EndedAt   sql.NullTime
}

// InsertSession inserts a session row.
func (s *SQLite) InsertSession(sess SessionRow) error {
	var ended interface{}
	if sess.EndedAt.Valid {
		ended = sess.EndedAt.Time.UTC().Format(time.RFC3339)
	}
	_, err := s.Exec(`INSERT INTO sessions (id, thread_id, started_at, ended_at) VALUES (?, ?, ?, ?)`,
		sess.ID, sess.ThreadID,
		sess.StartedAt.UTC().Format(time.RFC3339),
		ended)
	return err
}

// GetSession returns a session row by ID.
func (s *SQLite) GetSession(id string) (*SessionRow, error) {
	row := s.QueryRow(`SELECT id, thread_id, started_at, ended_at FROM sessions WHERE id=?`, id)
	var sess SessionRow
	var started string
	var ended sql.NullString
	if err := row.Scan(&sess.ID, &sess.ThreadID, &started, &ended); err != nil {
		return nil, err
	}
	sess.StartedAt, _ = time.Parse(time.RFC3339, started)
	if ended.Valid {
		t, _ := time.Parse(time.RFC3339, ended.String)
		sess.EndedAt = sql.NullTime{Time: t, Valid: true}
	}
	return &sess, nil
}

// EndSession marks a session as ended.
func (s *SQLite) EndSession(id string, endedAt time.Time) error {
	_, err := s.Exec(`UPDATE sessions SET ended_at=? WHERE id=?`,
		endedAt.UTC().Format(time.RFC3339), id)
	return err
}

// ListSessions returns sessions for a thread ordered by started_at desc.
func (s *SQLite) ListSessions(threadID string) ([]SessionRow, error) {
	rows, err := s.Query(`SELECT id, thread_id, started_at, ended_at FROM sessions
		WHERE thread_id=? ORDER BY started_at DESC`, threadID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []SessionRow
	for rows.Next() {
		var sess SessionRow
		var started string
		var ended sql.NullString
		if err := rows.Scan(&sess.ID, &sess.ThreadID, &started, &ended); err != nil {
			return nil, err
		}
		sess.StartedAt, _ = time.Parse(time.RFC3339, started)
		if ended.Valid {
			t, _ := time.Parse(time.RFC3339, ended.String)
			sess.EndedAt = sql.NullTime{Time: t, Valid: true}
		}
		out = append(out, sess)
	}
	return out, rows.Err()
}

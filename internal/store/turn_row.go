package store

import (
	"database/sql"
	"time"

	"github.com/sachncs/alfred/internal/contract"
)

// TurnRow mirrors the turns table.
type TurnRow struct {
	ID        string
	ThreadID  string
	SessionID sql.NullString
	Status    string
	StartedAt time.Time
	EndedAt   sql.NullTime
}

func turnRowFromContract(t *contract.Turn) TurnRow {
	r := TurnRow{
		ID:        string(t.ID),
		ThreadID:  string(t.ThreadID),
		Status:    string(t.Status),
		StartedAt: t.StartedAt,
	}
	if !t.EndedAt.IsZero() {
		r.EndedAt = sql.NullTime{Time: t.EndedAt, Valid: true}
	}
	return r
}

func contractFromTurnRow(r TurnRow) contract.Turn {
	t := contract.Turn{
		ID:        contract.TurnID(r.ID),
		ThreadID:  contract.ThreadID(r.ThreadID),
		Status:    contract.TurnStatus(r.Status),
		StartedAt: r.StartedAt,
	}
	if r.EndedAt.Valid {
		t.EndedAt = r.EndedAt.Time
	}
	return t
}

// InsertTurn inserts a turn row.
func (s *SQLite) InsertTurn(t *contract.Turn) error {
	row := turnRowFromContract(t)
	var sessionID, ended interface{}
	if row.SessionID.Valid {
		sessionID = row.SessionID.String
	}
	if row.EndedAt.Valid {
		ended = row.EndedAt.Time.UTC().Format(time.RFC3339)
	}
	_, err := s.Exec(`INSERT INTO turns (id, thread_id, session_id, status, started_at, ended_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		row.ID, row.ThreadID, sessionID, row.Status,
		row.StartedAt.UTC().Format(time.RFC3339), ended)
	return err
}

// UpdateTurn updates a turn row's status and ended_at.
func (s *SQLite) UpdateTurn(t *contract.Turn) error {
	row := turnRowFromContract(t)
	var ended interface{}
	if row.EndedAt.Valid {
		ended = row.EndedAt.Time.UTC().Format(time.RFC3339)
	}
	_, err := s.Exec(`UPDATE turns SET status=?, ended_at=? WHERE id=?`,
		row.Status, ended, row.ID)
	return err
}

// GetTurn returns a turn row by ID.
func (s *SQLite) GetTurn(id string) (*contract.Turn, error) {
	row := s.QueryRow(`SELECT id, thread_id, status, started_at, ended_at
		FROM turns WHERE id=?`, id)
	var r TurnRow
	var started, ended sql.NullString
	if err := row.Scan(&r.ID, &r.ThreadID, &r.Status, &started, &ended); err != nil {
		return nil, err
	}
	r.StartedAt, _ = time.Parse(time.RFC3339, started.String)
	if ended.Valid {
		t, _ := time.Parse(time.RFC3339, ended.String)
		r.EndedAt = sql.NullTime{Time: t, Valid: true}
	}
	t := contractFromTurnRow(r)
	return &t, nil
}

// ListTurns returns turns for a thread ordered by started_at desc.
func (s *SQLite) ListTurns(threadID string) ([]contract.Turn, error) {
	rows, err := s.Query(`SELECT id, thread_id, status, started_at, ended_at
		FROM turns WHERE thread_id=? ORDER BY started_at DESC`, threadID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []contract.Turn
	for rows.Next() {
		var r TurnRow
		var started, ended sql.NullString
		if err := rows.Scan(&r.ID, &r.ThreadID, &r.Status, &started, &ended); err != nil {
			return nil, err
		}
		r.StartedAt, _ = time.Parse(time.RFC3339, started.String)
		if ended.Valid {
			t, _ := time.Parse(time.RFC3339, ended.String)
			r.EndedAt = sql.NullTime{Time: t, Valid: true}
		}
		out = append(out, contractFromTurnRow(r))
	}
	return out, rows.Err()
}

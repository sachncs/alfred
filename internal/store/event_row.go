package store

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/sachncs/alfred/internal/contract"
)

// EventRow mirrors the events table.
type EventRow struct {
	Seq       int64
	ThreadID  string
	TurnID    string
	Kind      string
	Payload   []byte
	CreatedAt time.Time
}

// InsertEvent inserts a single event and returns its seq.
func (s *SQLite) InsertEvent(threadID, turnID, kind string, payload []byte, createdAt time.Time) (int64, error) {
	res, err := s.Exec(`INSERT INTO events (thread_id, turn_id, kind, payload_json, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		threadID, turnID, kind, string(payload),
		createdAt.UTC().Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// BulkInsertEvents inserts many events atomically. Returns the first and last seq assigned.
func (s *SQLite) BulkInsertEvents(threadID string, items []contract.TurnItem) (firstSeq, lastSeq int64, err error) {
	if len(items) == 0 {
		return 0, 0, nil
	}
	err = s.Tx(func(tx *sql.Tx) error {
		for _, it := range items {
			payload, mErr := json.Marshal(it)
			if mErr != nil {
				return mErr
			}
			turnID := ""
			if it.Metadata != nil {
				if v, ok := it.Metadata["turnId"].(string); ok {
					turnID = v
				}
			}
			res, eErr := tx.Exec(`INSERT INTO events (thread_id, turn_id, kind, payload_json, created_at)
				VALUES (?, ?, ?, ?, ?)`,
				threadID, turnID, string(it.Kind), string(payload),
				it.CreatedAt.UTC().Format(time.RFC3339))
			if eErr != nil {
				return eErr
			}
			id, _ := res.LastInsertId()
			if firstSeq == 0 {
				firstSeq = id
			}
			lastSeq = id
		}
		return nil
	})
	return firstSeq, lastSeq, err
}

// ReadEvents returns events for a thread in seq range [fromSeq, toSeq]. Pass 0 for fromSeq/toSeq for open bounds.
func (s *SQLite) ReadEvents(threadID string, fromSeq, toSeq int64) ([]EventRow, error) {
	q := `SELECT seq, thread_id, COALESCE(turn_id, ''), kind, payload_json, created_at
		FROM events WHERE thread_id=?`
	args := []any{threadID}
	if fromSeq > 0 {
		q += ` AND seq >= ?`
		args = append(args, fromSeq)
	}
	if toSeq > 0 {
		q += ` AND seq <= ?`
		args = append(args, toSeq)
	}
	q += ` ORDER BY seq ASC`

	rows, err := s.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []EventRow
	for rows.Next() {
		var r EventRow
		var created string
		if err := rows.Scan(&r.Seq, &r.ThreadID, &r.TurnID, &r.Kind, &r.Payload, &created); err != nil {
			return nil, err
		}
		r.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, r)
	}
	return out, rows.Err()
}

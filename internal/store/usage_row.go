package store

import (
	"time"
)

// UsageRow mirrors the usage table.
type UsageRow struct {
	ID           int64
	ThreadID     string
	Model        string
	InputTokens  int
	OutputTokens int
	Day          string
	CreatedAt    time.Time
}

// RecordUsage inserts a usage record.
func (s *SQLite) RecordUsage(threadID, model string, in, out int, day string, at time.Time) error {
	_, err := s.Exec(`INSERT INTO usage (thread_id, model, input_tokens, output_tokens, day, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		threadID, model, in, out, day, at.UTC().Format(time.RFC3339))
	return err
}

// UsageAggregate is a summary row.
type UsageAggregate struct {
	TotalIn  int
	TotalOut int
	Count    int
}

// AggregateUsagePerThread sums tokens for a single thread.
func (s *SQLite) AggregateUsagePerThread(threadID string) (UsageAggregate, error) {
	var agg UsageAggregate
	err := s.QueryRow(`SELECT COALESCE(SUM(input_tokens),0), COALESCE(SUM(output_tokens),0), COUNT(*)
		FROM usage WHERE thread_id=?`, threadID).Scan(&agg.TotalIn, &agg.TotalOut, &agg.Count)
	return agg, err
}

// AggregateUsagePerDay returns per-day totals.
func (s *SQLite) AggregateUsagePerDay() (map[string]UsageAggregate, error) {
	rows, err := s.Query(`SELECT day, COALESCE(SUM(input_tokens),0), COALESCE(SUM(output_tokens),0), COUNT(*)
		FROM usage GROUP BY day ORDER BY day`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := make(map[string]UsageAggregate)
	for rows.Next() {
		var day string
		var agg UsageAggregate
		if err := rows.Scan(&day, &agg.TotalIn, &agg.TotalOut, &agg.Count); err != nil {
			return nil, err
		}
		out[day] = agg
	}
	return out, rows.Err()
}

// AggregateUsagePerModel returns per-model totals.
func (s *SQLite) AggregateUsagePerModel() (map[string]UsageAggregate, error) {
	rows, err := s.Query(`SELECT model, COALESCE(SUM(input_tokens),0), COALESCE(SUM(output_tokens),0), COUNT(*)
		FROM usage GROUP BY model ORDER BY model`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := make(map[string]UsageAggregate)
	for rows.Next() {
		var model string
		var agg UsageAggregate
		if err := rows.Scan(&model, &agg.TotalIn, &agg.TotalOut, &agg.Count); err != nil {
			return nil, err
		}
		out[model] = agg
	}
	return out, rows.Err()
}

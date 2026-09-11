package store

import (
	"database/sql"
	"testing"
	"time"

	"github.com/sachncs/alfred/internal/contract"
)

func TestSessionRowCRUD(t *testing.T) {
	db := newTestSQLite(t)
	// Need a parent thread for FK.
	_ = db.InsertThread(&contract.Thread{
		ID: "thr1", Title: "t", Status: contract.ThreadStatusIdle,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	})

	sess := SessionRow{
		ID:        "sess1",
		ThreadID:  "thr1",
		StartedAt: time.Now().UTC(),
	}
	if err := db.InsertSession(sess); err != nil {
		t.Fatalf("insert: %v", err)
	}
	got, err := db.GetSession("sess1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ThreadID != "thr1" {
		t.Errorf("thread_id = %q", got.ThreadID)
	}
	if got.EndedAt.Valid {
		t.Error("expected ended_at to be null")
	}

	if err := db.EndSession("sess1", time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	got, _ = db.GetSession("sess1")
	if !got.EndedAt.Valid {
		t.Error("expected ended_at to be set")
	}

	list, err := db.ListSessions("thr1")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Errorf("list = %d, want 1", len(list))
	}
}

func TestTurnRowCRUD(t *testing.T) {
	db := newTestSQLite(t)
	_ = db.InsertThread(&contract.Thread{
		ID: "thr1", Title: "t", Status: contract.ThreadStatusIdle,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	})

	turn := &contract.Turn{
		ID:        "t1",
		ThreadID:  "thr1",
		Status:    contract.TurnStatusRunning,
		StartedAt: time.Now().UTC(),
	}
	if err := db.InsertTurn(turn); err != nil {
		t.Fatalf("insert: %v", err)
	}
	got, err := db.GetTurn("t1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != contract.TurnStatusRunning {
		t.Errorf("status = %s", got.Status)
	}

	turn.Status = contract.TurnStatusCompleted
	turn.EndedAt = time.Now().UTC()
	if err := db.UpdateTurn(turn); err != nil {
		t.Fatal(err)
	}
	got, _ = db.GetTurn("t1")
	if got.Status != contract.TurnStatusCompleted {
		t.Errorf("after update: status = %s", got.Status)
	}

	list, err := db.ListTurns("thr1")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Errorf("list = %d, want 1", len(list))
	}
}

func TestEventRowBulkInsert(t *testing.T) {
	db := newTestSQLite(t)
	_ = db.InsertThread(&contract.Thread{
		ID: "thr1", Title: "t", Status: contract.ThreadStatusIdle,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	})

	now := time.Now().UTC()
	items := []contract.TurnItem{
		{Kind: contract.ItemKindUserMessage, CreatedAt: now, Text: strPtrLocal("hi")},
		{Kind: contract.ItemKindAssistantText, CreatedAt: now.Add(time.Millisecond), Text: strPtrLocal("hello")},
		{Kind: contract.ItemKindToolResult, CreatedAt: now.Add(2 * time.Millisecond)},
	}

	first, last, err := db.BulkInsertEvents("thr1", items)
	if err != nil {
		t.Fatalf("bulk insert: %v", err)
	}
	if first == 0 || last == 0 || last < first {
		t.Errorf("seq range invalid: %d..%d", first, last)
	}

	events, err := db.ReadEvents("thr1", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 3 {
		t.Errorf("read all: got %d, want 3", len(events))
	}

	// Range query.
	events, err = db.ReadEvents("thr1", first+1, last-1)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Errorf("read range: got %d, want 1", len(events))
	}
}

func TestUsageRowAggregates(t *testing.T) {
	db := newTestSQLite(t)
	now := time.Now().UTC()
	day := now.Format("2006-01-02")
	if err := db.RecordUsage("thr1", "gpt-4", 100, 50, day, now); err != nil {
		t.Fatal(err)
	}
	if err := db.RecordUsage("thr1", "gpt-4", 200, 100, day, now); err != nil {
		t.Fatal(err)
	}
	if err := db.RecordUsage("thr1", "claude", 300, 150, day, now); err != nil {
		t.Fatal(err)
	}

	agg, err := db.AggregateUsagePerThread("thr1")
	if err != nil {
		t.Fatal(err)
	}
	if agg.TotalIn != 600 || agg.TotalOut != 300 || agg.Count != 3 {
		t.Errorf("per-thread: %+v", agg)
	}

	byDay, err := db.AggregateUsagePerDay()
	if err != nil {
		t.Fatal(err)
	}
	if len(byDay) != 1 {
		t.Errorf("per-day: got %d", len(byDay))
	}

	byModel, err := db.AggregateUsagePerModel()
	if err != nil {
		t.Fatal(err)
	}
	if len(byModel) != 2 {
		t.Errorf("per-model: got %d", len(byModel))
	}
}

func strPtrLocal(s string) *string { return &s }

var _ = sql.ErrNoRows

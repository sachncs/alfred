package schedule

import (
	"context"
	"encoding/json"
	"testing"

	alfredtool "github.com/alfred/alfred/internal/tool"
)

func TestScheduleCreateTool(t *testing.T) {
	store := newTaskStore()
	tl := &scheduleCreateTool{store: store}

	t.Run("missing name", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"schedule":"daily"}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})

	t.Run("missing schedule", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"name":"test"}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})

	t.Run("valid create", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"name":"backup","schedule":"daily","kind":"cron"}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})

	t.Run("defaults kind to once", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"name":"test","schedule":"now"}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`not json`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})
}

func TestScheduleRunTool(t *testing.T) {
	store := newTaskStore()
	tl := &scheduleRunTool{store: store}

	t.Run("missing taskId", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})

	t.Run("unknown task", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"taskId":"unknown"}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})

	t.Run("valid run", func(t *testing.T) {
		task := store.create("test", "now", "once")
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"taskId":"`+task.ID+`"}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`not json`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})
}

func TestScheduleAuditTool(t *testing.T) {
	store := newTaskStore()
	tl := &scheduleAuditTool{store: store}

	t.Run("audit all", func(t *testing.T) {
		store.create("task1", "daily", "cron")
		res, err := tl.Execute(context.Background(), nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})

	t.Run("audit specific task", func(t *testing.T) {
		task := store.create("task2", "weekly", "cron")
		store.addRun(task.ID, "completed", "ok")
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"taskId":"`+task.ID+`"}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})
}

func TestTaskStore(t *testing.T) {
	store := newTaskStore()

	t.Run("create and get", func(t *testing.T) {
		task := store.create("test", "daily", "cron")
		got, ok := store.get(task.ID)
		if !ok || got.Name != "test" {
			t.Fatal("expected to find task")
		}
	})

	t.Run("list", func(t *testing.T) {
		tasks := store.list()
		if len(tasks) == 0 {
			t.Fatal("expected tasks")
		}
	})

	t.Run("add run", func(t *testing.T) {
		task := store.create("run-test", "now", "once")
		run := store.addRun(task.ID, "completed", "output")
		if run.Status != "completed" {
			t.Fatal("expected completed run")
		}
		runs := store.getRuns(task.ID)
		if len(runs) == 0 {
			t.Fatal("expected runs")
		}
	})
}

func TestToolInterface(t *testing.T) {
	var _ alfredtool.Tool = &scheduleCreateTool{}
	var _ alfredtool.Tool = &scheduleRunTool{}
	var _ alfredtool.Tool = &scheduleAuditTool{}
}

package multiagent

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	alfredtool "github.com/sachncs/alfred/internal/tool"
)

func TestDelegateTaskTool(t *testing.T) {
	store := newTaskStore()
	tl := &delegateTaskTool{store: store}

	t.Run("missing task", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure for empty task")
		}
	})

	t.Run("delegate with default agent", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"task":"do something"}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})

	t.Run("delegate with specific agent", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"task":"analyze","agentType":"analyst"}`), nil)
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
			t.Fatal("expected failure for invalid json")
		}
	})
}

func TestTaskStatusTool(t *testing.T) {
	store := newTaskStore()
	tl := &taskStatusTool{store: store}

	t.Run("missing task", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure for empty task")
		}
	})

	t.Run("unknown task", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"taskId":"unknown"}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure for unknown task")
		}
	})

	t.Run("known task", func(t *testing.T) {
		child := store.create("test task", "general")
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"taskId":"`+child.ID+`"}`), nil)
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
		child := store.create("task1", "general")
		if child.Status != StatusQueued {
			t.Fatal("expected queued status")
		}
		got, ok := store.get(child.ID)
		if !ok || got.Task != "task1" {
			t.Fatal("expected to find task")
		}
	})

	t.Run("update", func(t *testing.T) {
		child := store.create("task2", "general")
		ok := store.update(child.ID, StatusComplete, "done", "")
		if !ok {
			t.Fatal("expected update to succeed")
		}
		got, _ := store.get(child.ID)
		if got.Status != StatusComplete || got.Result != "done" {
			t.Fatal("expected updated status")
		}
	})

	t.Run("update nonexistent", func(t *testing.T) {
		ok := store.update("nope", StatusComplete, "", "")
		if ok {
			t.Fatal("expected update to fail")
		}
	})

	t.Run("list", func(t *testing.T) {
		store.create("a", "general")
		store.create("b", "general")
		list := store.list()
		if len(list) < 2 {
			t.Fatal("expected at least 2 tasks")
		}
	})
}

func TestGoroutineExecution(t *testing.T) {
	store := newTaskStore()
	tl := &delegateTaskTool{store: store}

	child := store.create("fast task", "general")
	tl.runTask(child)

	// Give goroutine a moment
	time.Sleep(10 * time.Millisecond)

	got, ok := store.get(child.ID)
	if !ok {
		t.Fatal("expected task to exist")
	}
	if got.Status != StatusComplete {
		t.Fatalf("expected completed, got %s", got.Status)
	}
	if got.Result == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestToolInterface(t *testing.T) {
	var _ alfredtool.Tool = &delegateTaskTool{}
	var _ alfredtool.Tool = &taskStatusTool{}
}

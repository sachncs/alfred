package feedbackgateway

import (
	"context"
	"encoding/json"
	"testing"

	alfredtool "github.com/alfred/alfred/internal/tool"
)

func TestFeedbackSubmitTool(t *testing.T) {
	store := newIssueStore()
	tl := &feedbackSubmitTool{store: store}

	t.Run("missing title", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"body":"test"}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})

	t.Run("missing body", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"title":"test"}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})

	t.Run("valid submit", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"title":"Bug report","body":"Something is wrong","labels":["bug"]}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})

	t.Run("idempotent submit", func(t *testing.T) {
		res1, _ := tl.Execute(context.Background(), json.RawMessage(`{"title":"Dupe","body":"Same content"}`), nil)
		res2, _ := tl.Execute(context.Background(), json.RawMessage(`{"title":"Dupe","body":"Same content"}`), nil)
		if !res1.OK || !res2.OK {
			t.Fatal("expected both to succeed")
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

func TestFeedbackStatusTool(t *testing.T) {
	store := newIssueStore()
	tl := &feedbackStatusTool{store: store}

	t.Run("missing issueId", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})

	t.Run("unknown issue", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"issueId":"unknown"}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})

	t.Run("known issue", func(t *testing.T) {
		issue := store.submit("test", "body", nil)
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"issueId":"`+issue.ID+`"}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})
}

func TestIssueStore(t *testing.T) {
	store := newIssueStore()

	t.Run("submit and get", func(t *testing.T) {
		issue := store.submit("title", "body", []string{"bug"})
		if issue.Status != StatusCreated {
			t.Fatal("expected created status")
		}
		got, ok := store.get(issue.ID)
		if !ok || got.Title != "title" {
			t.Fatal("expected to find issue")
		}
	})

	t.Run("idempotency", func(t *testing.T) {
		i1 := store.submit("dup", "same", nil)
		i2 := store.submit("dup", "same", nil)
		if i2.Status != StatusDuplicate {
			t.Fatal("expected duplicate status")
		}
		if i1.ID != i2.ID {
			t.Fatal("expected same ID for duplicate")
		}
	})

	t.Run("get missing", func(t *testing.T) {
		_, ok := store.get("nope")
		if ok {
			t.Fatal("expected not found")
		}
	})

	t.Run("find by key", func(t *testing.T) {
		issue := store.submit("findme", "content", nil)
		got, ok := store.findByKey(issue.IdempotencyKey)
		if !ok || got.ID != issue.ID {
			t.Fatal("expected to find by key")
		}
	})
}

func TestIdempotencyKey(t *testing.T) {
	k1 := idempotencyKey("title", "body")
	k2 := idempotencyKey("title", "body")
	k3 := idempotencyKey("title", "different")
	if k1 != k2 {
		t.Fatal("expected same key for same input")
	}
	if k1 == k3 {
		t.Fatal("expected different key for different input")
	}
}

func TestToolInterface(t *testing.T) {
	var _ alfredtool.Tool = &feedbackSubmitTool{}
	var _ alfredtool.Tool = &feedbackStatusTool{}
}

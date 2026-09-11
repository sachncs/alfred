package paperradar

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sachncs/alfred/internal/search/providers"
	"github.com/sachncs/alfred/internal/tool"
)

func TestPaperSearchTool(t *testing.T) {
	svc := providers.NewResearchService()
	store := newPaperStore()
	tool := &paperSearchTool{svc: svc, store: store}

	t.Run("missing query", func(t *testing.T) {
		res, err := tool.Execute(context.Background(), json.RawMessage(`{}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure for empty query")
		}
	})

	t.Run("valid query returns result", func(t *testing.T) {
		res, err := tool.Execute(context.Background(), json.RawMessage(`{"query":"CRISPR","maxResults":3}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		res, err := tool.Execute(context.Background(), json.RawMessage(`not json`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure for invalid json")
		}
	})
}

func TestPaperDigestTool(t *testing.T) {
	store := newPaperStore()
	store.add(&Paper{ID: "2301.00001", Title: "Test Paper", Authors: []string{"Alice"}, Snippet: "A test abstract."})
	tool := &paperDigestTool{store: store}

	t.Run("unknown paper", func(t *testing.T) {
		res, err := tool.Execute(context.Background(), json.RawMessage(`{"paperId":"unknown"}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure for unknown paper")
		}
	})

	t.Run("known paper", func(t *testing.T) {
		res, err := tool.Execute(context.Background(), json.RawMessage(`{"paperId":"2301.00001"}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})
}

func TestPaperProfileTool(t *testing.T) {
	store := newPaperStore()
	store.add(&Paper{ID: "2301.00001", Title: "Test"})
	tool := &paperProfileTool{store: store}

	t.Run("unknown paper", func(t *testing.T) {
		res, err := tool.Execute(context.Background(), json.RawMessage(`{"paperId":"unknown"}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure for unknown paper")
		}
	})

	t.Run("create profile", func(t *testing.T) {
		res, err := tool.Execute(context.Background(), json.RawMessage(`{"paperId":"2301.00001","focus":"methods"}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})
}

func TestPaperStore(t *testing.T) {
	store := newPaperStore()

	t.Run("add and get", func(t *testing.T) {
		p := &Paper{ID: "p1", Title: "Paper One"}
		store.add(p)
		got, ok := store.get("p1")
		if !ok || got.Title != "Paper One" {
			t.Fatal("expected to find paper")
		}
	})

	t.Run("get missing", func(t *testing.T) {
		_, ok := store.get("missing")
		if ok {
			t.Fatal("expected not found")
		}
	})

	t.Run("profile upsert", func(t *testing.T) {
		prof := &ReadingProfile{PaperID: "p1", Focus: "intro", Status: "active"}
		store.upsertProfile(prof)
		got, ok := store.getProfile("p1")
		if !ok || got.Focus != "intro" {
			t.Fatal("expected profile")
		}
	})
}

func TestPaperID(t *testing.T) {
	id := paperID("https://arxiv.org/abs/2301.00001")
	if id != "2301.00001" {
		t.Fatalf("expected 2301.00001, got %s", id)
	}

	id2 := paperID("nourl")
	if id2 != "nourl" {
		t.Fatalf("expected nourl, got %s", id2)
	}
}

func TestGenerateDigest(t *testing.T) {
	p := &Paper{
		Title:   "Test Paper",
		Authors: []string{"Alice", "Bob"},
		Source:  "arxiv",
		Date:    "2024-01-01",
		Snippet: "Abstract text here.",
	}
	d := generateDigest(p)
	if d == "" {
		t.Fatal("expected non-empty digest")
	}
	if !containsStr(d, "Test Paper") {
		t.Fatal("digest should contain title")
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsSubstr(s, sub))
}

func containsSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestToolInterface(t *testing.T) {
	var _ tool.Tool = &paperSearchTool{}
	var _ tool.Tool = &paperDigestTool{}
	var _ tool.Tool = &paperProfileTool{}
}

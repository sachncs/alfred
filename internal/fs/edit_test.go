package fs

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sachncs/alfred/internal/tool"
)

func TestEditToolBasic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	_ = os.WriteFile(path, []byte("hello world"), 0o644)

	et := NewEditTool()
	tc := tool.NewContext(context.Background(), "", "", "", dir)

	input, _ := json.Marshal(EditInput{Path: "test.txt", Old: "world", New: "Go"})
	res, err := et.Execute(context.Background(), input, tc)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK {
		t.Fatalf("edit failed: %s", res.Error)
	}

	data, _ := os.ReadFile(path)
	if string(data) != "hello Go" {
		t.Errorf("content = %q, want 'hello Go'", data)
	}
}

func TestEditToolDiff(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	_ = os.WriteFile(path, []byte("line1\nline2\nline3"), 0o644)

	et := NewEditTool()
	tc := tool.NewContext(context.Background(), "", "", "", dir)

	input, _ := json.Marshal(EditInput{Path: "test.txt", Old: "line2", New: "LINE2"})
	res, _ := et.Execute(context.Background(), input, tc)

	var out EditOutput
	_ = json.Unmarshal(res.Structured, &out)
	if out.Diff == "" {
		t.Error("expected non-empty diff")
	}
}

func TestEditToolNotFound(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	_ = os.WriteFile(path, []byte("hello"), 0o644)

	et := NewEditTool()
	tc := tool.NewContext(context.Background(), "", "", "", dir)

	input, _ := json.Marshal(EditInput{Path: "test.txt", Old: "nonexistent", New: "x"})
	res, _ := et.Execute(context.Background(), input, tc)
	if res.OK {
		t.Error("expected failure for text not found")
	}
}

func TestEditToolMultipleMatches(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	_ = os.WriteFile(path, []byte("aaa bbb aaa"), 0o644)

	et := NewEditTool()
	tc := tool.NewContext(context.Background(), "", "", "", dir)

	input, _ := json.Marshal(EditInput{Path: "test.txt", Old: "aaa", New: "xxx"})
	res, _ := et.Execute(context.Background(), input, tc)
	if res.OK {
		t.Error("expected failure for multiple matches")
	}
}

func TestEditToolFuzzyMatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	// Content has extra spaces.
	_ = os.WriteFile(path, []byte("hello   world"), 0o644)

	et := NewEditTool()
	tc := tool.NewContext(context.Background(), "", "", "", dir)

	// Fuzzy search with normalized whitespace should match.
	input, _ := json.Marshal(EditInput{Path: "test.txt", Old: "hello world", New: "hi Go", Fuzzy: true})
	res, err := et.Execute(context.Background(), input, tc)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK {
		t.Fatalf("fuzzy edit failed: %s", res.Error)
	}
}

func TestEditToolPathEscape(t *testing.T) {
	dir := t.TempDir()
	et := NewEditTool()
	tc := tool.NewContext(context.Background(), "", "", "", dir)

	input, _ := json.Marshal(EditInput{Path: "../escape.txt", Old: "a", New: "b"})
	res, _ := et.Execute(context.Background(), input, tc)
	if res.OK {
		t.Error("expected failure for path escape")
	}
}

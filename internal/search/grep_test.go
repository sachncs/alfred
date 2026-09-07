package search

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/alfred/alfred/internal/tool"
)

func TestGrepToolBasic(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello world\nfoo bar\nhello go"), 0o644)

	gt := NewGrepTool()
	tc := tool.NewContext(context.Background(), "", "", "", dir)

	input, _ := json.Marshal(GrepInput{Pattern: "hello"})
	res, err := gt.Execute(context.Background(), input, tc)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK {
		t.Fatalf("grep failed: %s", res.Error)
	}

	var out GrepOutput
	_ = json.Unmarshal(res.Structured, &out)
	if out.Count != 2 {
		t.Errorf("count = %d, want 2", out.Count)
	}
}

func TestGrepToolGlobFilter(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "a.go"), []byte("package main"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "b.txt"), []byte("package main"), 0o644)

	gt := NewGrepTool()
	tc := tool.NewContext(context.Background(), "", "", "", dir)

	input, _ := json.Marshal(GrepInput{Pattern: "package", Glob: "*.go"})
	res, _ := gt.Execute(context.Background(), input, tc)

	var out GrepOutput
	_ = json.Unmarshal(res.Structured, &out)
	if out.Count != 1 {
		t.Errorf("count = %d, want 1", out.Count)
	}
	if out.Matches[0].File != filepath.Join(dir, "a.go") {
		t.Errorf("matched file = %s, want a.go", out.Matches[0].File)
	}
}

func TestGrepToolNoMatches(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0o644)

	gt := NewGrepTool()
	tc := tool.NewContext(context.Background(), "", "", "", dir)

	input, _ := json.Marshal(GrepInput{Pattern: "nonexistent"})
	res, _ := gt.Execute(context.Background(), input, tc)

	var out GrepOutput
	_ = json.Unmarshal(res.Structured, &out)
	if out.Count != 0 {
		t.Errorf("count = %d, want 0", out.Count)
	}
}

func TestGrepToolInvalidRegex(t *testing.T) {
	gt := NewGrepTool()
	input, _ := json.Marshal(GrepInput{Pattern: "[invalid"})
	res, _ := gt.Execute(context.Background(), input, nil)
	if res.OK {
		t.Error("expected failure for invalid regex")
	}
}

func TestGrepToolEmptyPattern(t *testing.T) {
	gt := NewGrepTool()
	input, _ := json.Marshal(GrepInput{Pattern: ""})
	res, _ := gt.Execute(context.Background(), input, nil)
	if res.OK {
		t.Error("expected failure for empty pattern")
	}
}

func TestGrepToolSkipsGit(t *testing.T) {
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, ".git"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, ".git", "config"), []byte("match me"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "a.txt"), []byte("match me"), 0o644)

	gt := NewGrepTool()
	tc := tool.NewContext(context.Background(), "", "", "", dir)

	input, _ := json.Marshal(GrepInput{Pattern: "match me"})
	res, _ := gt.Execute(context.Background(), input, tc)

	var out GrepOutput
	_ = json.Unmarshal(res.Structured, &out)
	if out.Count != 1 {
		t.Errorf("count = %d, want 1 (should skip .git)", out.Count)
	}
}

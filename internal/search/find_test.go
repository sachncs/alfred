package search

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/alfred/alfred/internal/tool"
)

func TestFindToolBasic(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "a.go"), []byte("package main"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "b.txt"), []byte("hello"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "c.go"), []byte("package main"), 0o644)

	ft := NewFindTool()
	tc := tool.NewContext(context.Background(), "", "", "", dir)

	input, _ := json.Marshal(FindInput{Pattern: "*.go"})
	res, err := ft.Execute(context.Background(), input, tc)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK {
		t.Fatalf("find failed: %s", res.Error)
	}

	var out FindOutput
	_ = json.Unmarshal(res.Structured, &out)
	if out.Count != 2 {
		t.Errorf("count = %d, want 2", out.Count)
	}
}

func TestFindToolSkipsGit(t *testing.T) {
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, ".git"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, ".git", "config"), []byte("data"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "a.txt"), []byte("data"), 0o644)

	ft := NewFindTool()
	tc := tool.NewContext(context.Background(), "", "", "", dir)

	input, _ := json.Marshal(FindInput{Pattern: "*"})
	res, _ := ft.Execute(context.Background(), input, tc)

	var out FindOutput
	_ = json.Unmarshal(res.Structured, &out)
	if out.Count != 1 {
		t.Errorf("count = %d, want 1 (skip .git)", out.Count)
	}
}

func TestFindToolNoMatches(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "a.txt"), []byte("data"), 0o644)

	ft := NewFindTool()
	tc := tool.NewContext(context.Background(), "", "", "", dir)

	input, _ := json.Marshal(FindInput{Pattern: "*.go"})
	res, _ := ft.Execute(context.Background(), input, tc)

	var out FindOutput
	_ = json.Unmarshal(res.Structured, &out)
	if out.Count != 0 {
		t.Errorf("count = %d, want 0", out.Count)
	}
}

func TestFindToolEmptyPattern(t *testing.T) {
	ft := NewFindTool()
	input, _ := json.Marshal(FindInput{Pattern: ""})
	res, _ := ft.Execute(context.Background(), input, nil)
	if res.OK {
		t.Error("expected failure for empty pattern")
	}
}

package fs

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/alfred/alfred/internal/tool"
)

func TestLsToolBasic(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0o644)
	_ = os.MkdirAll(filepath.Join(dir, "sub"), 0o755)

	lt := NewLsTool()
	tc := tool.NewContext(context.Background(), "", "", "", dir)

	input, _ := json.Marshal(LsInput{Path: "."})
	res, err := lt.Execute(context.Background(), input, tc)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK {
		t.Fatalf("ls failed: %s", res.Error)
	}

	var out LsOutput
	_ = json.Unmarshal(res.Structured, &out)
	if out.Count != 2 {
		t.Errorf("count = %d, want 2", out.Count)
	}
}

func TestLsToolSubDir(t *testing.T) {
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, "sub"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "sub", "file.txt"), []byte("data"), 0o644)

	lt := NewLsTool()
	tc := tool.NewContext(context.Background(), "", "", "", dir)

	input, _ := json.Marshal(LsInput{Path: "sub"})
	res, _ := lt.Execute(context.Background(), input, tc)

	var out LsOutput
	_ = json.Unmarshal(res.Structured, &out)
	if out.Count != 1 {
		t.Errorf("count = %d, want 1", out.Count)
	}
	if out.Entries[0].Name != "file.txt" {
		t.Errorf("name = %q, want file.txt", out.Entries[0].Name)
	}
}

func TestLsToolMetadata(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "data.txt"), []byte("hello world"), 0o644)

	lt := NewLsTool()
	tc := tool.NewContext(context.Background(), "", "", "", dir)

	input, _ := json.Marshal(LsInput{Path: "."})
	res, _ := lt.Execute(context.Background(), input, tc)

	var out LsOutput
	_ = json.Unmarshal(res.Structured, &out)
	e := out.Entries[0]
	if e.Size != 11 {
		t.Errorf("size = %d, want 11", e.Size)
	}
	if e.IsDir {
		t.Error("expected isDir=false for file")
	}
	if e.ModTime == 0 {
		t.Error("expected non-zero modtime")
	}
}

func TestLsToolPathEscape(t *testing.T) {
	dir := t.TempDir()
	lt := NewLsTool()
	tc := tool.NewContext(context.Background(), "", "", "", dir)

	input, _ := json.Marshal(LsInput{Path: "../escape"})
	res, _ := lt.Execute(context.Background(), input, tc)
	if res.OK {
		t.Error("expected failure for path escape")
	}
}

func TestLsToolEmptyPath(t *testing.T) {
	lt := NewLsTool()
	input, _ := json.Marshal(LsInput{Path: ""})
	res, _ := lt.Execute(context.Background(), input, nil)
	if res.OK {
		t.Error("expected failure for empty path")
	}
}

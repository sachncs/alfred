package fs

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/alfred/alfred/internal/tool"
)

func TestWriteToolBasic(t *testing.T) {
	dir := t.TempDir()
	wt := NewWriteTool()
	tc := tool.NewContext(context.Background(), "", "", "", dir)

	input, _ := json.Marshal(WriteInput{Path: "hello.txt", Content: "hello world"})
	res, err := wt.Execute(context.Background(), input, tc)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK {
		t.Fatalf("write failed: %s", res.Error)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if string(data) != "hello world" {
		t.Errorf("content = %q, want hello world", data)
	}
}

func TestWriteToolCreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	wt := NewWriteTool()
	tc := tool.NewContext(context.Background(), "", "", "", dir)

	input, _ := json.Marshal(WriteInput{Path: "a/b/c/deep.txt", Content: "nested"})
	res, err := wt.Execute(context.Background(), input, tc)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK {
		t.Fatalf("write failed: %s", res.Error)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "a/b/c/deep.txt"))
	if string(data) != "nested" {
		t.Errorf("content = %q, want nested", data)
	}
}

func TestWriteToolOverwrite(t *testing.T) {
	dir := t.TempDir()
	wt := NewWriteTool()
	tc := tool.NewContext(context.Background(), "", "", "", dir)

	input, _ := json.Marshal(WriteInput{Path: "file.txt", Content: "first"})
	wt.Execute(context.Background(), input, tc)

	input, _ = json.Marshal(WriteInput{Path: "file.txt", Content: "second"})
	wt.Execute(context.Background(), input, tc)

	data, _ := os.ReadFile(filepath.Join(dir, "file.txt"))
	if string(data) != "second" {
		t.Errorf("content = %q, want second", data)
	}
}

func TestWriteToolBOMStripped(t *testing.T) {
	dir := t.TempDir()
	wt := NewWriteTool()
	tc := tool.NewContext(context.Background(), "", "", "", dir)

	input, _ := json.Marshal(WriteInput{Path: "bom.txt", Content: "\xEF\xBB\xBFwith bom"})
	wt.Execute(context.Background(), input, tc)

	data, _ := os.ReadFile(filepath.Join(dir, "bom.txt"))
	if hasBOM(data) {
		t.Error("BOM should be stripped")
	}
	if string(data) != "with bom" {
		t.Errorf("content = %q, want 'with bom'", data)
	}
}

func TestWriteToolCRLFNormalized(t *testing.T) {
	dir := t.TempDir()
	wt := NewWriteTool()
	tc := tool.NewContext(context.Background(), "", "", "", dir)

	input, _ := json.Marshal(WriteInput{Path: "crlf.txt", Content: "line1\r\nline2\r\n"})
	wt.Execute(context.Background(), input, tc)

	data, _ := os.ReadFile(filepath.Join(dir, "crlf.txt"))
	if string(data) != "line1\nline2\n" {
		t.Errorf("content = %q, want LF line endings", data)
	}
}

func TestWriteToolPathEscape(t *testing.T) {
	dir := t.TempDir()
	wt := NewWriteTool()
	tc := tool.NewContext(context.Background(), "", "", "", dir)

	input, _ := json.Marshal(WriteInput{Path: "../escape.txt", Content: "bad"})
	res, err := wt.Execute(context.Background(), input, tc)
	if err != nil {
		t.Fatal(err)
	}
	if res.OK {
		t.Error("expected failure for path escaping workspace")
	}
}

func TestWriteToolEmptyPath(t *testing.T) {
	wt := NewWriteTool()
	input, _ := json.Marshal(WriteInput{Path: "", Content: "data"})
	res, err := wt.Execute(context.Background(), input, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.OK {
		t.Error("expected failure for empty path")
	}
}

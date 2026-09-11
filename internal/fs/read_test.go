package fs

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sachncs/alfred/internal/tool"
)

// helper: write a fixture and return its absolute path.
func writeFixture(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return p
}

// helper: build a fresh Context.
func newCtx(_ string) context.Context {
	return context.Background()
}

func TestResolvePathAbsolute(t *testing.T) {
	t.Parallel()
	r := &FileSystemTool{}
	dir := t.TempDir()
	got, err := r.ResolvePath(dir, "/")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !filepath.IsAbs(got) {
		t.Fatalf("expected absolute, got %q", got)
	}
}

func TestResolvePathRelativeJoinsWorkspace(t *testing.T) {
	t.Parallel()
	r := &FileSystemTool{}
	dir := t.TempDir()
	got, err := r.ResolvePath("foo/bar.txt", dir)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	want := filepath.Clean(filepath.Join(dir, "foo/bar.txt"))
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestResolvePathBlocksEscape(t *testing.T) {
	t.Parallel()
	r := &FileSystemTool{}
	dir := t.TempDir()
	if _, err := r.ResolvePath("../etc/passwd", dir); err == nil {
		t.Fatalf("expected escape to be blocked")
	}
}

func TestResolvePathEmptyWorkspace(t *testing.T) {
	t.Parallel()
	r := &FileSystemTool{}
	if _, err := r.ResolvePath("a", ""); err == nil {
		t.Fatalf("expected error on empty workspace")
	}
}

func TestResolvePathEmptyInput(t *testing.T) {
	t.Parallel()
	r := &FileSystemTool{}
	if _, err := r.ResolvePath("", "/tmp"); err == nil {
		t.Fatalf("expected error on empty path")
	}
}

func TestReadToolReadsText(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFixture(t, dir, "hello.txt", "hello\nworld\n")
	rt := NewReadTool()
	in, _ := json.Marshal(ReadInput{Path: "hello.txt"})
	tc := tool.NewContext(newCtx(dir), "thr-1", "turn-1", "call-1", dir)
	r, err := rt.Execute(newCtx(dir), in, tc)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !r.OK {
		t.Fatalf("expected OK, got %+v", r)
	}
	if !strings.Contains(r.Content, "hello") {
		t.Fatalf("content missing: %q", r.Content)
	}
	var out ReadOutput
	if err := json.Unmarshal(r.Structured, &out); err != nil {
		t.Fatalf("structured: %v", err)
	}
	if out.Path != "hello.txt" {
		t.Fatalf("path mismatch: %q", out.Path)
	}
	if out.MIMEType != "text/plain" {
		t.Fatalf("mime: %q", out.MIMEType)
	}
	if out.Lines != 2 {
		t.Fatalf("lines: %d", out.Lines)
	}
}

func TestReadToolDetectsMIMEByExtension(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFixture(t, dir, "main.go", "package main\n")
	writeFixture(t, dir, "data.json", `{"a":1}`)
	writeFixture(t, dir, "page.html", "<html>")

	tests := []struct {
		path string
		mime string
	}{
		{"main.go", "text/x-go"},
		{"data.json", "application/json"},
		{"page.html", "text/html"},
	}
	rt := NewReadTool()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.path, func(t *testing.T) {
			t.Parallel()
			in, _ := json.Marshal(ReadInput{Path: tc.path})
			ctx := tool.NewContext(newCtx(dir), "thr", "turn", "call", dir)
			r, err := rt.Execute(newCtx(dir), in, ctx)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			var out ReadOutput
			_ = json.Unmarshal(r.Structured, &out)
			if out.MIMEType != tc.mime {
				t.Fatalf("got %q want %q", out.MIMEType, tc.mime)
			}
		})
	}
}

func TestReadToolBinaryFallback(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	binary := []byte{0x00, 0x01, 0x02, 0x00, 0x03}
	if err := os.WriteFile(filepath.Join(dir, "blob.bin"), binary, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	rt := NewReadTool()
	in, _ := json.Marshal(ReadInput{Path: "blob.bin"})
	ctx := tool.NewContext(newCtx(dir), "thr", "turn", "call", dir)
	r, err := rt.Execute(newCtx(dir), in, ctx)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !r.OK {
		t.Fatalf("expected OK")
	}
	var out ReadOutput
	_ = json.Unmarshal(r.Structured, &out)
	if out.MIMEType != "application/octet-stream" {
		t.Fatalf("mime: %q", out.MIMEType)
	}
}

func TestReadToolTruncatesToMaxBytes(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFixture(t, dir, "big.txt", strings.Repeat("a", 1000))
	rt := NewReadTool()
	in, _ := json.Marshal(ReadInput{Path: "big.txt", MaxBytes: 50})
	ctx := tool.NewContext(newCtx(dir), "thr", "turn", "call", dir)
	r, err := rt.Execute(newCtx(dir), in, ctx)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(r.Content) > 50 {
		t.Fatalf("content not truncated: %d bytes", len(r.Content))
	}
	var out ReadOutput
	_ = json.Unmarshal(r.Structured, &out)
	if !out.Truncated {
		t.Fatalf("expected truncated=true")
	}
}

func TestReadToolRejectsDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	rt := NewReadTool()
	in, _ := json.Marshal(ReadInput{Path: "subdir"})
	ctx := tool.NewContext(newCtx(dir), "thr", "turn", "call", dir)
	r, err := rt.Execute(newCtx(dir), in, ctx)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if r.OK {
		t.Fatalf("expected failure for directory, got OK")
	}
}

func TestReadToolFailsOnMissing(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	rt := NewReadTool()
	in, _ := json.Marshal(ReadInput{Path: "nope.txt"})
	ctx := tool.NewContext(newCtx(dir), "thr", "turn", "call", dir)
	r, err := rt.Execute(newCtx(dir), in, ctx)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if r.OK {
		t.Fatalf("expected failure, got OK")
	}
	if r.Error == "" {
		t.Fatalf("expected error message")
	}
}

func TestReadToolRespectsCancellation(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFixture(t, dir, "f.txt", "x")
	rt := NewReadTool()
	in, _ := json.Marshal(ReadInput{Path: "f.txt"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	tcCtx := tool.NewContext(ctx, "thr", "turn", "call", dir)
	r, err := rt.Execute(ctx, in, tcCtx)
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got: %v", err)
	}
	// Either returned a failure Result OR returned an error — both acceptable.
	if r != nil && r.OK {
		t.Fatalf("expected failure after cancel, got OK")
	}
}

func TestReadToolEmptyPathRejected(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	rt := NewReadTool()
	in, _ := json.Marshal(ReadInput{Path: ""})
	ctx := tool.NewContext(newCtx(dir), "thr", "turn", "call", dir)
	r, err := rt.Execute(newCtx(dir), in, ctx)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if r.OK {
		t.Fatalf("expected failure on empty path")
	}
}

func TestReadToolInvalidJSON(t *testing.T) {
	t.Parallel()
	rt := NewReadTool()
	ctx := tool.NewContext(newCtx("/tmp"), "thr", "turn", "call", "/tmp")
	r, err := rt.Execute(newCtx("/tmp"), json.RawMessage(`{"path":`), ctx)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if r.OK {
		t.Fatalf("expected failure on invalid JSON")
	}
}

func TestReadToolNameAndSchema(t *testing.T) {
	t.Parallel()
	rt := NewReadTool()
	if rt.Name() != "fs_read" {
		t.Fatalf("name: %q", rt.Name())
	}
	if rt.Description() == "" {
		t.Fatalf("description empty")
	}
	if len(rt.Schema()) == 0 {
		t.Fatalf("schema empty")
	}
	// Schema must mention "path" since it's required.
	if !strings.Contains(string(rt.Schema()), `"path"`) {
		t.Fatalf("schema missing path property: %s", rt.Schema())
	}
}

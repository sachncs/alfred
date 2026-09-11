package fs

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sachncs/alfred/internal/tool"
)

func TestApplyPatchBasic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	_ = os.WriteFile(path, []byte("hello world"), 0o644)

	apt := NewApplyPatchTool()
	tc := tool.NewContext(context.Background(), "", "", "", dir)

	patch := `--- a/test.txt
+++ b/test.txt
@@ -1 +1 @@
-hello world
+hello Go
`
	input, _ := json.Marshal(ApplyPatchInput{Path: "test.txt", Patch: patch})
	res, err := apt.Execute(context.Background(), input, tc)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK {
		t.Fatalf("apply patch failed: %s", res.Error)
	}

	data, _ := os.ReadFile(path)
	if string(data) != "hello Go" {
		t.Errorf("content = %q, want 'hello Go'", data)
	}
}

func TestApplyPatchMultipleHunks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	_ = os.WriteFile(path, []byte("aaa\nbbb\nccc"), 0o644)

	apt := NewApplyPatchTool()
	tc := tool.NewContext(context.Background(), "", "", "", dir)

	// Two separate hunks in one patch: replace "aaa" then replace "ccc".
	patch := `--- a/test.txt
+++ b/test.txt
@@ -1 +1 @@
-aaa
+AAA
--- a/test.txt
+++ b/test.txt
@@ -3 +3 @@
-ccc
+CCC
`
	input, _ := json.Marshal(ApplyPatchInput{Path: "test.txt", Patch: patch})
	res, err := apt.Execute(context.Background(), input, tc)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK {
		t.Fatalf("apply patch failed: %s", res.Error)
	}

	data, _ := os.ReadFile(path)
	expected := "AAA\nbbb\nCCC"
	if string(data) != expected {
		t.Errorf("content = %q, want %q", data, expected)
	}
}

func TestApplyPatchDryRun(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	_ = os.WriteFile(path, []byte("original"), 0o644)

	apt := NewApplyPatchTool()
	tc := tool.NewContext(context.Background(), "", "", "", dir)

	patch := `--- a/test.txt
+++ b/test.txt
@@ -1 +1 @@
-original
+modified
`
	input, _ := json.Marshal(ApplyPatchInput{Path: "test.txt", Patch: patch, DryRun: true})
	res, err := apt.Execute(context.Background(), input, tc)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK {
		t.Fatalf("dry run failed: %s", res.Error)
	}

	data, _ := os.ReadFile(path)
	if string(data) != "original" {
		t.Errorf("dry run should not modify file, got %q", data)
	}
}

func TestApplyPatchHunkNotFound(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	_ = os.WriteFile(path, []byte("hello"), 0o644)

	apt := NewApplyPatchTool()
	tc := tool.NewContext(context.Background(), "", "", "", dir)

	patch := `--- a/test.txt
+++ b/test.txt
@@ -1 +1 @@
-nonexistent text
+replaced
`
	input, _ := json.Marshal(ApplyPatchInput{Path: "test.txt", Patch: patch})
	res, _ := apt.Execute(context.Background(), input, tc)
	if res.OK {
		t.Error("expected failure for hunk not found")
	}
}

func TestApplyPatchPathEscape(t *testing.T) {
	dir := t.TempDir()
	apt := NewApplyPatchTool()
	tc := tool.NewContext(context.Background(), "", "", "", dir)

	input, _ := json.Marshal(ApplyPatchInput{Path: "../escape.txt", Patch: "--- a\n+++ b\n"})
	res, _ := apt.Execute(context.Background(), input, tc)
	if res.OK {
		t.Error("expected failure for path escape")
	}
}

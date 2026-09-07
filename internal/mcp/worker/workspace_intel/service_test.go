package workspaceintel

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	alfredtool "github.com/alfred/alfred/internal/tool"
)

func TestWorkspaceListTool(t *testing.T) {
	tl := &workspaceListTool{}

	t.Run("list current dir", func(t *testing.T) {
		tc := &alfredtool.Context{WorkspaceRoot: t.TempDir()}
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"path":"."}`), tc)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})

	t.Run("empty path defaults to dot", func(t *testing.T) {
		tc := &alfredtool.Context{WorkspaceRoot: t.TempDir()}
		res, err := tl.Execute(context.Background(), json.RawMessage(`{}`), tc)
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

func TestWorkspaceReadTool(t *testing.T) {
	tl := &workspaceReadTool{}

	t.Run("read existing file", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "test.txt")
		if err := os.WriteFile(f, []byte("hello world"), 0644); err != nil {
			t.Fatal(err)
		}
		tc := &alfredtool.Context{WorkspaceRoot: dir}
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"path":"test.txt"}`), tc)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})

	t.Run("missing path", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure for missing path")
		}
	})

	t.Run("read nonexistent file", func(t *testing.T) {
		tc := &alfredtool.Context{WorkspaceRoot: t.TempDir()}
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"path":"nope.txt"}`), tc)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure for nonexistent file")
		}
	})
}

func TestWorkspaceVisualInspectTool(t *testing.T) {
	tl := &workspaceVisualInspectTool{}

	t.Run("inspect file", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "data.csv")
		if err := os.WriteFile(f, []byte("a,b,c\n1,2,3"), 0644); err != nil {
			t.Fatal(err)
		}
		tc := &alfredtool.Context{WorkspaceRoot: dir}
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"path":"data.csv"}`), tc)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})

	t.Run("inspect directory", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, "sub"), 0755); err != nil {
			t.Fatal(err)
		}
		tc := &alfredtool.Context{WorkspaceRoot: dir}
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"path":"."}`), tc)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})

	t.Run("missing path", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure for missing path")
		}
	})

	t.Run("nonexistent path", func(t *testing.T) {
		tc := &alfredtool.Context{WorkspaceRoot: t.TempDir()}
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"path":"nope"}`), tc)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure for nonexistent path")
		}
	})
}

func TestToolInterface(t *testing.T) {
	var _ alfredtool.Tool = &workspaceListTool{}
	var _ alfredtool.Tool = &workspaceReadTool{}
	var _ alfredtool.Tool = &workspaceVisualInspectTool{}
}

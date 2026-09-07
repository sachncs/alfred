package workspaceomics

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	alfredtool "github.com/alfred/alfred/internal/tool"
)

func TestOmicsPreviewTool(t *testing.T) {
	tl := &omicsPreviewTool{}

	t.Run("missing path", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		tc := &alfredtool.Context{WorkspaceRoot: t.TempDir()}
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"path":"nope.mtx"}`), tc)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})

	t.Run("parse Matrix Market", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "data.mtx")
		content := `%%MatrixMarket matrix coordinate real general
% comment line
3 3 4
1 1 1.5
2 2 2.5
3 3 3.5
1 2 0.5`
		if err := os.WriteFile(f, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		tc := &alfredtool.Context{WorkspaceRoot: dir}
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"path":"data.mtx"}`), tc)
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
			t.Fatal("expected failure")
		}
	})
}

func TestOmicsSelectDatasetTool(t *testing.T) {
	tl := &omicsSelectDatasetTool{}

	t.Run("missing fields", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})

	t.Run("valid select", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"path":"data.mtx","datasetId":"ds1"}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})
}

func TestParseMatrixMarket(t *testing.T) {
	t.Run("full file", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "test.mtx")
		content := `%%MatrixMarket matrix coordinate real general
% A small test matrix
4 4 5
1 1 1.0
2 2 2.0
3 3 3.0
4 4 4.0
1 2 0.5`
		if err := os.WriteFile(f, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		data, err := parseMatrixMarket(f)
		if err != nil {
			t.Fatal(err)
		}
		if data.Rows != 4 || data.Cols != 4 {
			t.Fatalf("expected 4x4, got %dx%d", data.Rows, data.Cols)
		}
		if data.NonZero != 5 {
			t.Fatalf("expected 5 non-zeros, got %d", data.NonZero)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "empty.mtx")
		if err := os.WriteFile(f, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
		data, err := parseMatrixMarket(f)
		if err != nil {
			t.Fatal(err)
		}
		if data.Rows != 0 {
			t.Fatalf("expected 0 rows, got %d", data.Rows)
		}
	})
}

func TestToolInterface(t *testing.T) {
	var _ alfredtool.Tool = &omicsPreviewTool{}
	var _ alfredtool.Tool = &omicsSelectDatasetTool{}
}

package workspacetabular

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	alfredtool "github.com/sachncs/alfred/internal/tool"
)

func TestTabularPreviewTool(t *testing.T) {
	tl := &tabularPreviewTool{}

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
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"path":"nope.csv"}`), tc)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})

	t.Run("parse CSV", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "data.csv")
		if err := os.WriteFile(f, []byte("name,age,score\nAlice,30,95.5\nBob,25,88.0\nCharlie,35,72.3"), 0644); err != nil {
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

	t.Run("parse TSV", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "data.tsv")
		if err := os.WriteFile(f, []byte("col1\tcol2\tcol3\n1\thello\t3.14\n2\tworld\t2.71"), 0644); err != nil {
			t.Fatal(err)
		}
		tc := &alfredtool.Context{WorkspaceRoot: dir}
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"path":"data.tsv"}`), tc)
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

func TestDetectDelimiter(t *testing.T) {
	tests := []struct {
		input string
		want  rune
	}{
		{"a,b,c\n1,2,3", ','},
		{"a\tb\tc\n1\t2\t3", '\t'},
	}
	for _, tt := range tests {
		got := detectDelimiter([]byte(tt.input))
		if got != tt.want {
			t.Errorf("detectDelimiter(%q) = %c, want %c", tt.input, got, tt.want)
		}
	}
}

func TestParseCSV(t *testing.T) {
	t.Run("basic CSV", func(t *testing.T) {
		data := "name,age\nAlice,30\nBob,25"
		preview, cols, total, err := parseCSV(strings.NewReader(data), ',', 10)
		if err != nil {
			t.Fatal(err)
		}
		if total != 2 {
			t.Fatalf("expected 2 rows, got %d", total)
		}
		if len(cols) != 2 {
			t.Fatalf("expected 2 columns, got %d", len(cols))
		}
		if cols[0].Name != "name" || cols[1].Name != "age" {
			t.Fatalf("unexpected column names: %v", cols)
		}
		if len(preview) != 2 {
			t.Fatalf("expected 2 preview rows, got %d", len(preview))
		}
	})

	t.Run("empty CSV", func(t *testing.T) {
		data := "col1,col2\n"
		_, _, total, err := parseCSV(strings.NewReader(data), ',', 10)
		if err != nil {
			t.Fatal(err)
		}
		if total != 0 {
			t.Fatalf("expected 0 rows, got %d", total)
		}
	})

	t.Run("maxRows limit", func(t *testing.T) {
		data := "a,b\n1,2\n3,4\n5,6\n7,8"
		preview, _, total, err := parseCSV(strings.NewReader(data), ',', 2)
		if err != nil {
			t.Fatal(err)
		}
		if total != 4 {
			t.Fatalf("expected 4 total rows, got %d", total)
		}
		if len(preview) != 2 {
			t.Fatalf("expected 2 preview rows, got %d", len(preview))
		}
	})
}

func TestAnalyzeColumn(t *testing.T) {
	t.Run("integer column", func(t *testing.T) {
		col := analyzeColumn("age", []string{"30", "25", "35"})
		if col.Type != "integer" {
			t.Fatalf("expected integer, got %s", col.Type)
		}
		if col.NonEmpty != 3 {
			t.Fatalf("expected 3 non-empty, got %d", col.NonEmpty)
		}
	})

	t.Run("number column", func(t *testing.T) {
		col := analyzeColumn("score", []string{"3.14", "2.71", "1.41"})
		if col.Type != "number" {
			t.Fatalf("expected number, got %s", col.Type)
		}
	})

	t.Run("string column", func(t *testing.T) {
		col := analyzeColumn("name", []string{"Alice", "Bob", ""})
		if col.Type != "string" {
			t.Fatalf("expected string, got %s", col.Type)
		}
		if col.NonEmpty != 2 {
			t.Fatalf("expected 2 non-empty, got %d", col.NonEmpty)
		}
	})

	t.Run("empty column", func(t *testing.T) {
		col := analyzeColumn("empty", []string{})
		if col.NonEmpty != 0 {
			t.Fatalf("expected 0 non-empty, got %d", col.NonEmpty)
		}
	})
}

func TestToolInterface(t *testing.T) {
	var _ alfredtool.Tool = &tabularPreviewTool{}
}

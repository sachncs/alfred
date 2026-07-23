package workspacesequence

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	alfredtool "github.com/alfred/alfred/internal/tool"
)

func TestSequencePreviewTool(t *testing.T) {
	tl := &sequencePreviewTool{}

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
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"path":"nope.fa"}`), tc)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})

	t.Run("parse FASTA", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "seq.fa")
		if err := os.WriteFile(f, []byte(">seq1 test sequence\nACGTACGT\n>seq2 another\nTTTTCCCC"), 0644); err != nil {
			t.Fatal(err)
		}
		tc := &alfredtool.Context{WorkspaceRoot: dir}
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"path":"seq.fa"}`), tc)
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

func TestParseFasta(t *testing.T) {
	t.Run("single sequence", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "test.fa")
		if err := os.WriteFile(f, []byte(">header\nACGTACGT"), 0644); err != nil {
			t.Fatal(err)
		}
		seqs, err := parseFasta(f)
		if err != nil {
			t.Fatal(err)
		}
		if len(seqs) != 1 {
			t.Fatalf("expected 1 sequence, got %d", len(seqs))
		}
		if seqs[0].ID != "header" {
			t.Fatalf("expected ID 'header', got %q", seqs[0].ID)
		}
		if seqs[0].Length != 8 {
			t.Fatalf("expected length 8, got %d", seqs[0].Length)
		}
		if seqs[0].Type != "DNA" {
			t.Fatalf("expected DNA, got %s", seqs[0].Type)
		}
	})

	t.Run("multiple sequences", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "multi.fa")
		if err := os.WriteFile(f, []byte(">s1 desc1\nACGT\n>s2\nCCCC"), 0644); err != nil {
			t.Fatal(err)
		}
		seqs, err := parseFasta(f)
		if err != nil {
			t.Fatal(err)
		}
		if len(seqs) != 2 {
			t.Fatalf("expected 2 sequences, got %d", len(seqs))
		}
		if seqs[0].Description != "desc1" {
			t.Fatalf("expected desc1, got %q", seqs[0].Description)
		}
	})

	t.Run("RNA sequence", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "rna.fa")
		if err := os.WriteFile(f, []byte(">rna\nACGUACGU"), 0644); err != nil {
			t.Fatal(err)
		}
		seqs, err := parseFasta(f)
		if err != nil {
			t.Fatal(err)
		}
		if seqs[0].Type != "RNA" {
			t.Fatalf("expected RNA, got %s", seqs[0].Type)
		}
	})

	t.Run("protein sequence", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "prot.fa")
		if err := os.WriteFile(f, []byte(">prot\nMNLKIHGFEDCBA"), 0644); err != nil {
			t.Fatal(err)
		}
		seqs, err := parseFasta(f)
		if err != nil {
			t.Fatal(err)
		}
		if seqs[0].Type != "protein" {
			t.Fatalf("expected protein, got %s", seqs[0].Type)
		}
	})
}

func TestDetectType(t *testing.T) {
	tests := []struct {
		seq  string
		want string
	}{
		{"ACGT", "DNA"},
		{"ACGU", "RNA"},
		{"MNLKIHGFEDCBA", "protein"},
		{"", "unknown"},
		{"ACGTACGT", "DNA"},
	}
	for _, tt := range tests {
		got := detectType(tt.seq)
		if got != tt.want {
			t.Errorf("detectType(%q) = %s, want %s", tt.seq, got, tt.want)
		}
	}
}

func TestGCContent(t *testing.T) {
	tests := []struct {
		seq  string
		want float64
	}{
		{"ACGT", 50.0},
		{"AAAA", 0.0},
		{"CCCC", 100.0},
		{"", 0.0},
	}
	for _, tt := range tests {
		got := gcContent(tt.seq)
		if got != tt.want {
			t.Errorf("gcContent(%q) = %v, want %v", tt.seq, got, tt.want)
		}
	}
}

func TestToolInterface(t *testing.T) {
	var _ alfredtool.Tool = &sequencePreviewTool{}
}

package workspacemolecular

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	alfredtool "github.com/sachncs/alfred/internal/tool"
)

func TestMolecularPreviewTool(t *testing.T) {
	tl := &molecularPreviewTool{}

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
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"path":"nope.pdb"}`), tc)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})

	t.Run("parse PDB", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "test.pdb")
		content := `TITLE     TEST STRUCTURE
EXPDTA    X-RAY DIFFRACTION
REMARK   2 RESOLUTION.   2.50 ANGSTROMS.
ATOM      1  N   ALA A   1       1.000   2.000   3.000  1.00 10.00           N
ATOM      2  CA  ALA A   1       2.000   3.000   4.000  1.00 10.00           C
ATOM      3  C   ALA A   1       3.000   4.000   5.000  1.00 10.00           C
END`
		if err := os.WriteFile(f, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		tc := &alfredtool.Context{WorkspaceRoot: dir}
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"path":"test.pdb"}`), tc)
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

func TestMolecularUpdateWorkbenchTool(t *testing.T) {
	tl := &molecularUpdateWorkbenchTool{}

	t.Run("missing path", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})

	t.Run("valid update", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"path":"test.pdb","mutations":[]}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})
}

func TestParsePDB(t *testing.T) {
	t.Run("full PDB file", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "full.pdb")
		content := `TITLE     PROTEIN STRUCTURE
EXPDTA    NMR
ATOM      1  N   ALA A   1       1.000   2.000   3.000  1.00 10.00           N
ATOM      2  CA  ALA A   1       2.000   3.000   4.000  1.00 10.00           C
ATOM      3  N   GLY B   1      10.000  11.000  12.000  1.00 10.00           N
END`
		if err := os.WriteFile(f, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		info, err := parsePDB(f)
		if err != nil {
			t.Fatal(err)
		}
		if info.TotalAtoms != 3 {
			t.Fatalf("expected 3 atoms, got %d", info.TotalAtoms)
		}
		if info.TotalResidues != 2 {
			t.Fatalf("expected 2 residues, got %d", info.TotalResidues)
		}
		if len(info.Chains) != 2 {
			t.Fatalf("expected 2 chains, got %d", len(info.Chains))
		}
	})

	t.Run("empty file", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "empty.pdb")
		if err := os.WriteFile(f, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
		info, err := parsePDB(f)
		if err != nil {
			t.Fatal(err)
		}
		if info.TotalAtoms != 0 {
			t.Fatalf("expected 0 atoms, got %d", info.TotalAtoms)
		}
	})
}

func TestParseCoord(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"  1.000", 1.0},
		{" -2.500", -2.5},
		{"100.000", 100.0},
	}
	for _, tt := range tests {
		got := parseCoord(tt.input)
		if got != tt.want {
			t.Errorf("parseCoord(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestToolInterface(t *testing.T) {
	var _ alfredtool.Tool = &molecularPreviewTool{}
	var _ alfredtool.Tool = &molecularUpdateWorkbenchTool{}
}

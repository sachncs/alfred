package workspacespectra

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	alfredtool "github.com/alfred/alfred/internal/tool"
)

func TestSpectraPreviewTool(t *testing.T) {
	tl := &spectraPreviewTool{}

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
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"path":"nope.jdx"}`), tc)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})

	t.Run("parse JCAMP-DX", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "spec.jdx")
		content := `##TITLE=Test Spectrum
##XUNITS=WAVENUMBERS
##YUNITS=ABSORBANCE
##NPOINTS=5
4000 0.1
3500 0.2
3000 0.5
2500 0.3
2000 0.1`
		if err := os.WriteFile(f, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		tc := &alfredtool.Context{WorkspaceRoot: dir}
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"path":"spec.jdx"}`), tc)
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

func TestParseJCAMPDX(t *testing.T) {
	t.Run("full JCAMP-DX file", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "test.jdx")
		content := `##TITLE=IR Spectrum
##XUNITS=WAVENUMBERS
##YUNITS=TRANSMITTANCE
##NPOINTS=4
1000 80.5
1500 60.2
2000 40.0
2500 20.1`
		if err := os.WriteFile(f, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		data, err := parseJCAMPDX(f)
		if err != nil {
			t.Fatal(err)
		}
		if data.Points != 4 {
			t.Fatalf("expected 4 points, got %d", data.Points)
		}
		if data.Metadata["TITLE"] != "IR Spectrum" {
			t.Fatalf("expected TITLE 'IR Spectrum', got %q", data.Metadata["TITLE"])
		}
		if data.XRange[0] != 1000 || data.XRange[1] != 2500 {
			t.Fatalf("unexpected XRange: %v", data.XRange)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "empty.jdx")
		if err := os.WriteFile(f, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
		data, err := parseJCAMPDX(f)
		if err != nil {
			t.Fatal(err)
		}
		if data.Points != 0 {
			t.Fatalf("expected 0 points, got %d", data.Points)
		}
	})
}

func TestParseXYLine(t *testing.T) {
	tests := []struct {
		line  string
		wantX float64
		wantY float64
		nil   bool
	}{
		{"1000 0.5", 1000, 0.5, false},
		{"2.5,3.14", 2.5, 3.14, false},
		{"", 0, 0, true},
		{"not a number", 0, 0, true},
	}
	for _, tt := range tests {
		got := parseXYLine(tt.line)
		if tt.nil {
			if got != nil {
				t.Errorf("parseXYLine(%q) = %v, want nil", tt.line, got)
			}
			continue
		}
		if got == nil {
			t.Errorf("parseXYLine(%q) = nil, want [%v, %v]", tt.line, tt.wantX, tt.wantY)
			continue
		}
		if got[0] != tt.wantX || got[1] != tt.wantY {
			t.Errorf("parseXYLine(%q) = [%v, %v], want [%v, %v]", tt.line, got[0], got[1], tt.wantX, tt.wantY)
		}
	}
}

func TestToolInterface(t *testing.T) {
	var _ alfredtool.Tool = &spectraPreviewTool{}
}

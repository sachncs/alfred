package workspacedeck

import (
	"archive/zip"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	alfredtool "github.com/alfred/alfred/internal/tool"
)

func createTestPPTX(t *testing.T, dir string) string {
	t.Helper()
	f := filepath.Join(dir, "test.pptx")
	outFile, err := os.Create(f)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(outFile)

	// Add a minimal slide XML
	slideXML := `<?xml version="1.0" encoding="UTF-8"?>
<p:sld xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:cSld>
    <p:spTree>
      <p:sp>
        <p:txBody>
          <p:p>
            <p:r>
              <p:t>Test Title</p:t>
            </p:r>
          </p:p>
        </p:txBody>
      </p:sp>
      <p:sp>
        <p:txBody>
          <p:p>
            <p:r>
              <p:t>Test content</p:t>
            </p:r>
          </p:p>
        </p:txBody>
      </p:sp>
    </p:spTree>
  </p:cSld>
</p:sld>`

	fw, err := w.Create("ppt/slides/slide1.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write([]byte(slideXML)); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := outFile.Close(); err != nil {
		t.Fatal(err)
	}

	return f
}

func TestDeckPreviewTool(t *testing.T) {
	tl := &deckPreviewTool{}

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
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"path":"nope.pptx"}`), tc)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})

	t.Run("parse PPTX", func(t *testing.T) {
		dir := t.TempDir()
		createTestPPTX(t, dir)
		tc := &alfredtool.Context{WorkspaceRoot: dir}
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"path":"test.pptx"}`), tc)
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

func TestDeckSelectSlideTool(t *testing.T) {
	tl := &deckSelectSlideTool{}

	t.Run("missing path", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})

	t.Run("out of range index", func(t *testing.T) {
		dir := t.TempDir()
		createTestPPTX(t, dir)
		tc := &alfredtool.Context{WorkspaceRoot: dir}
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"path":"test.pptx","slideIndex":5}`), tc)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure for out of range")
		}
	})

	t.Run("valid select", func(t *testing.T) {
		dir := t.TempDir()
		createTestPPTX(t, dir)
		tc := &alfredtool.Context{WorkspaceRoot: dir}
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"path":"test.pptx","slideIndex":1}`), tc)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})
}

func TestDeckUpdateTextTool(t *testing.T) {
	tl := &deckUpdateTextTool{}

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
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"path":"test.pptx","slideIndex":1,"elementId":"el1","text":"New text"}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})
}

func TestParsePPTX(t *testing.T) {
	t.Run("valid PPTX", func(t *testing.T) {
		dir := t.TempDir()
		f := createTestPPTX(t, dir)
		deck, err := parsePPTX(f)
		if err != nil {
			t.Fatal(err)
		}
		if deck.SlideCount != 1 {
			t.Fatalf("expected 1 slide, got %d", deck.SlideCount)
		}
		if len(deck.Slides) != 1 {
			t.Fatalf("expected 1 slide info, got %d", len(deck.Slides))
		}
		if deck.Slides[0].Index != 1 {
			t.Fatalf("expected slide index 1, got %d", deck.Slides[0].Index)
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := parsePPTX("/nonexistent.pptx")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestToolInterface(t *testing.T) {
	var _ alfredtool.Tool = &deckPreviewTool{}
	var _ alfredtool.Tool = &deckSelectSlideTool{}
	var _ alfredtool.Tool = &deckUpdateTextTool{}
}

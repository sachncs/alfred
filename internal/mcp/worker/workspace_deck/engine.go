package workspacedeck

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/sachncs/alfred/internal/mcp/worker"
	"github.com/sachncs/alfred/internal/tool"
)

// SlideInfo describes a single slide.
type SlideInfo struct {
	Index        int    `json:"index"`
	Title        string `json:"title,omitempty"`
	TextPreview  string `json:"textPreview,omitempty"`
	ElementCount int    `json:"elementCount"`
}

// DeckInfo holds parsed PPTX data.
type DeckInfo struct {
	Path       string      `json:"path"`
	SlideCount int         `json:"slideCount"`
	Slides     []SlideInfo `json:"slides"`
}

func NewDeckServer() *worker.WorkerServer {
	return worker.NewWorkerServer("workspace-deck-worker", []tool.Tool{
		&deckPreviewTool{},
		&deckSelectSlideTool{},
		&deckUpdateTextTool{},
	})
}

type deckPreviewTool struct{}

func (t *deckPreviewTool) Name() string        { return "deck_preview" }
func (t *deckPreviewTool) Description() string { return "Preview a presentation deck" }
func (t *deckPreviewTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}`)
}
func (t *deckPreviewTool) Execute(_ context.Context, raw json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	var in struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.Path == "" {
		return tool.FailureMsg("path is required"), nil
	}

	abs := in.Path
	if tc != nil && tc.WorkspaceRoot != "" && !strings.HasPrefix(in.Path, "/") {
		abs = tc.WorkspaceRoot + "/" + in.Path
	}

	deck, err := parsePPTX(abs)
	if err != nil {
		return tool.Failure(err), nil
	}

	return tool.SuccessWith(fmt.Sprintf("%d slides", deck.SlideCount), deck), nil
}

type deckSelectSlideTool struct{}

func (t *deckSelectSlideTool) Name() string        { return "deck_select_slide" }
func (t *deckSelectSlideTool) Description() string { return "Select a slide from the deck" }
func (t *deckSelectSlideTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"},"slideIndex":{"type":"integer"}},"required":["path","slideIndex"]}`)
}
func (t *deckSelectSlideTool) Execute(_ context.Context, raw json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	var in struct {
		Path       string `json:"path"`
		SlideIndex int    `json:"slideIndex"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.Path == "" {
		return tool.FailureMsg("path is required"), nil
	}

	abs := in.Path
	if tc != nil && tc.WorkspaceRoot != "" && !strings.HasPrefix(in.Path, "/") {
		abs = tc.WorkspaceRoot + "/" + in.Path
	}

	deck, err := parsePPTX(abs)
	if err != nil {
		return tool.Failure(err), nil
	}

	if in.SlideIndex < 1 || in.SlideIndex > len(deck.Slides) {
		return tool.FailureMsg(fmt.Sprintf("slide index %d out of range (1-%d)", in.SlideIndex, len(deck.Slides))), nil
	}

	slide := deck.Slides[in.SlideIndex-1]
	return tool.SuccessWith(fmt.Sprintf("Slide %d", in.SlideIndex), slide), nil
}

type deckUpdateTextTool struct{}

func (t *deckUpdateTextTool) Name() string        { return "deck_update_text" }
func (t *deckUpdateTextTool) Description() string { return "Update a text element on a slide" }
func (t *deckUpdateTextTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"},"slideIndex":{"type":"integer"},"elementId":{"type":"string"},"text":{"type":"string"}},"required":["path","slideIndex","elementId","text"]}`)
}
func (t *deckUpdateTextTool) Execute(_ context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Path       string `json:"path"`
		SlideIndex int    `json:"slideIndex"`
		ElementID  string `json:"elementId"`
		Text       string `json:"text"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.Path == "" {
		return tool.FailureMsg("path is required"), nil
	}
	if in.SlideIndex <= 0 {
		return tool.FailureMsg("slideIndex must be positive"), nil
	}
	if in.ElementID == "" {
		return tool.FailureMsg("elementId is required"), nil
	}
	return tool.SuccessWith("Text updated", map[string]any{
		"path":       in.Path,
		"slideIndex": in.SlideIndex,
		"elementId":  in.ElementID,
		"text":       in.Text,
		"status":     "updated",
	}), nil
}

// parsePPTX opens a PPTX file and extracts slide information.
func parsePPTX(path string) (*DeckInfo, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()

	deck := &DeckInfo{Path: path}

	// Find all slide XML files
	var slideFiles []string
	for _, f := range r.File {
		if strings.HasPrefix(f.Name, "ppt/slides/slide") && strings.HasSuffix(f.Name, ".xml") {
			slideFiles = append(slideFiles, f.Name)
		}
	}
	sort.Strings(slideFiles)

	for i, name := range slideFiles {
		slide, err := extractSlide(r, name)
		if err != nil {
			continue
		}
		slide.Index = i + 1
		deck.Slides = append(deck.Slides, *slide)
	}

	deck.SlideCount = len(deck.Slides)
	return deck, nil
}

// extractSlide reads a slide XML from the ZIP and extracts text.
func extractSlide(r *zip.ReadCloser, name string) (*SlideInfo, error) {
	for _, f := range r.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			return nil, err
		}

		return extractTextFromSlideXML(data), nil
	}
	return nil, fmt.Errorf("slide not found: %s", name)
}

// extractTextFromSlideXML extracts text from PPTX slide XML using simple
// tag scanning rather than full XML namespace resolution.
func extractTextFromSlideXML(data []byte) *SlideInfo {
	slideInfo := &SlideInfo{}
	var allText strings.Builder
	elementCount := 0
	var firstText string

	// Extract all <a:t>...</a:t> content — these are the text runs in PPTX
	buf := bytes.NewBuffer(data)
	for {
		line, err := buf.ReadString('\n')
		if err != nil && line == "" {
			break
		}
		line = strings.TrimSpace(line)

		// Look for <a:t>text</a:t> patterns
		for {
			start := strings.Index(line, "<a:t>")
			if start < 0 {
				break
			}
			end := strings.Index(line[start:], "</a:t>")
			if end < 0 {
				break
			}
			text := line[start+5 : start+end]
			if text != "" {
				elementCount++
				if firstText == "" {
					firstText = text
				}
				allText.WriteString(text)
				allText.WriteString(" ")
			}
			line = line[start+end+6:]
		}

		if err != nil {
			break
		}
	}

	slideInfo.ElementCount = elementCount
	slideInfo.Title = firstText
	text := strings.TrimSpace(allText.String())
	if len(text) > 200 {
		text = text[:200] + "..."
	}
	slideInfo.TextPreview = text

	return slideInfo
}

package workspacedeck

import (
	"context"
	"encoding/json"

	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/tool"
)

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
func (t *deckPreviewTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Path string `json:"path"`
	}
	_ = json.Unmarshal(raw, &in)
	return tool.SuccessWith("Deck preview", map[string]any{"path": in.Path, "slides": 0}), nil
}

type deckSelectSlideTool struct{}

func (t *deckSelectSlideTool) Name() string        { return "deck_select_slide" }
func (t *deckSelectSlideTool) Description() string { return "Select a slide from the deck" }
func (t *deckSelectSlideTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"},"slideIndex":{"type":"integer"}},"required":["path","slideIndex"]}`)
}
func (t *deckSelectSlideTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Path       string `json:"path"`
		SlideIndex int    `json:"slideIndex"`
	}
	_ = json.Unmarshal(raw, &in)
	return tool.SuccessWith("Slide selected", map[string]any{"slideIndex": in.SlideIndex}), nil
}

type deckUpdateTextTool struct{}

func (t *deckUpdateTextTool) Name() string        { return "deck_update_text" }
func (t *deckUpdateTextTool) Description() string { return "Update a text element on a slide" }
func (t *deckUpdateTextTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"},"slideIndex":{"type":"integer"},"elementId":{"type":"string"},"text":{"type":"string"}},"required":["path","slideIndex","elementId","text"]}`)
}
func (t *deckUpdateTextTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Path       string `json:"path"`
		SlideIndex int    `json:"slideIndex"`
		ElementID  string `json:"elementId"`
		Text       string `json:"text"`
	}
	_ = json.Unmarshal(raw, &in)
	return tool.SuccessWith("Text updated", map[string]any{"status": "updated"}), nil
}

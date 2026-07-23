package workspacespectra

import (
	"context"
	"encoding/json"

	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/tool"
)

func NewSpectraServer() *worker.WorkerServer {
	return worker.NewWorkerServer("workspace-spectra-worker", []tool.Tool{
		&spectraPreviewTool{},
	})
}

type spectraPreviewTool struct{}

func (t *spectraPreviewTool) Name() string        { return "spectra_preview" }
func (t *spectraPreviewTool) Description() string { return "Preview spectra data (JCAMP-DX)" }
func (t *spectraPreviewTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}`)
}
func (t *spectraPreviewTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Path string `json:"path"`
	}
	_ = json.Unmarshal(raw, &in)
	return tool.SuccessWith("Spectra preview", map[string]any{"path": in.Path, "points": 0}), nil
}

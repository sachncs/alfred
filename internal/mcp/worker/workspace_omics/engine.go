package workspaceomics

import (
	"context"
	"encoding/json"

	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/tool"
)

func NewOmicsServer() *worker.WorkerServer {
	return worker.NewWorkerServer("workspace-omics-worker", []tool.Tool{
		&omicsPreviewTool{},
		&omicsSelectDatasetTool{},
	})
}

type omicsPreviewTool struct{}

func (t *omicsPreviewTool) Name() string        { return "omics_preview" }
func (t *omicsPreviewTool) Description() string { return "Preview omics data (Matrix Market)" }
func (t *omicsPreviewTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}`)
}
func (t *omicsPreviewTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Path string `json:"path"`
	}
	_ = json.Unmarshal(raw, &in)
	return tool.SuccessWith("Omics preview", map[string]any{"path": in.Path, "rows": 0, "columns": 0}), nil
}

type omicsSelectDatasetTool struct{}

func (t *omicsSelectDatasetTool) Name() string        { return "omics_select_dataset" }
func (t *omicsSelectDatasetTool) Description() string { return "Select a dataset from omics data" }
func (t *omicsSelectDatasetTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"},"datasetId":{"type":"string"}},"required":["path","datasetId"]}`)
}
func (t *omicsSelectDatasetTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Path      string `json:"path"`
		DatasetID string `json:"datasetId"`
	}
	_ = json.Unmarshal(raw, &in)
	return tool.SuccessWith("Dataset selected", map[string]any{"datasetId": in.DatasetID}), nil
}

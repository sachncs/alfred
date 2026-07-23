package workspacetabular

import (
	"context"
	"encoding/json"

	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/tool"
)

func NewTabularServer() *worker.WorkerServer {
	return worker.NewWorkerServer("workspace-tabular-worker", []tool.Tool{
		&tabularPreviewTool{},
	})
}

type tabularPreviewTool struct{}

func (t *tabularPreviewTool) Name() string        { return "tabular_preview" }
func (t *tabularPreviewTool) Description() string { return "Preview tabular data (CSV/TSV/Parquet)" }
func (t *tabularPreviewTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"},"maxRows":{"type":"integer"}},"required":["path"]}`)
}
func (t *tabularPreviewTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Path    string `json:"path"`
		MaxRows int    `json:"maxRows"`
	}
	_ = json.Unmarshal(raw, &in)
	return tool.SuccessWith("Tabular preview", map[string]any{"path": in.Path, "rows": 0, "columns": []string{}}), nil
}

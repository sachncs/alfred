package workspacesequence

import (
	"context"
	"encoding/json"

	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/tool"
)

func NewSequenceServer() *worker.WorkerServer {
	return worker.NewWorkerServer("workspace-sequence-worker", []tool.Tool{
		&sequencePreviewTool{},
	})
}

type sequencePreviewTool struct{}

func (t *sequencePreviewTool) Name() string { return "sequence_preview" }
func (t *sequencePreviewTool) Description() string {
	return "Preview a sequence record (FASTA/GenBank)"
}
func (t *sequencePreviewTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}`)
}
func (t *sequencePreviewTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Path string `json:"path"`
	}
	_ = json.Unmarshal(raw, &in)
	return tool.SuccessWith("Sequence preview", map[string]any{"path": in.Path, "length": 0}), nil
}

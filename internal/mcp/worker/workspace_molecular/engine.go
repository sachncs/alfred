package workspacemolecular

import (
	"context"
	"encoding/json"

	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/tool"
)

func NewMolecularServer() *worker.WorkerServer {
	return worker.NewWorkerServer("workspace-molecular-worker", []tool.Tool{
		&molecularPreviewTool{},
		&molecularUpdateWorkbenchTool{},
	})
}

type molecularPreviewTool struct{}

func (t *molecularPreviewTool) Name() string        { return "molecular_preview" }
func (t *molecularPreviewTool) Description() string { return "Preview a PDB molecular structure" }
func (t *molecularPreviewTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}`)
}
func (t *molecularPreviewTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Path string `json:"path"`
	}
	_ = json.Unmarshal(raw, &in)
	return tool.SuccessWith("Molecular preview", map[string]any{"path": in.Path, "chains": 0, "residues": 0}), nil
}

type molecularUpdateWorkbenchTool struct{}

func (t *molecularUpdateWorkbenchTool) Name() string        { return "molecular_update_workbench" }
func (t *molecularUpdateWorkbenchTool) Description() string { return "Update the molecular workbench" }
func (t *molecularUpdateWorkbenchTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"},"mutations":{"type":"array","items":{"type":"object"}}},"required":["path"]}`)
}
func (t *molecularUpdateWorkbenchTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Path      string `json:"path"`
		Mutations any    `json:"mutations"`
	}
	_ = json.Unmarshal(raw, &in)
	return tool.SuccessWith("Workbench updated", map[string]any{"status": "updated"}), nil
}

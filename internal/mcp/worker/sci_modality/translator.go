package scimodality

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/tool"
)

// NewSciModalityServer creates the sci-modality MCP worker.
func NewSciModalityServer() *worker.WorkerServer {
	return worker.NewWorkerServer("sci-modality-worker", []tool.Tool{
		&translateModalityTool{},
	})
}

type translateModalityTool struct{}

func (t *translateModalityTool) Name() string { return "translate_modality" }
func (t *translateModalityTool) Description() string {
	return "Translate a scientific modality to text evidence"
}
func (t *translateModalityTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"modality":{"type":"string"},"input":{"type":"string"}},"required":["modality","input"]}`)
}
func (t *translateModalityTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Modality string `json:"modality"`
		Input    string `json:"input"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	return tool.SuccessWith("Modality translated", map[string]any{
		"modality": in.Modality,
		"evidence": "Stub evidence text for " + in.Modality,
	}), nil
}

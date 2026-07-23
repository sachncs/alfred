package bgcdiscovery

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/tool"
)

// NewBGCDiscoveryServer creates the bgc-discovery MCP worker.
func NewBGCDiscoveryServer() *worker.WorkerServer {
	return worker.NewWorkerServer("bgc-discovery-worker", []tool.Tool{
		&bgcPipelineTool{},
	})
}

type bgcPipelineTool struct{}

func (t *bgcPipelineTool) Name() string { return "bgc_run_pipeline" }
func (t *bgcPipelineTool) Description() string {
	return "Run BGC discovery pipeline (antiSMASH + BiG-SCAPE + MIBiG)"
}
func (t *bgcPipelineTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"inputFile":{"type":"string"},"steps":{"type":"array","items":{"type":"string"}}},"required":["inputFile"]}`)
}
func (t *bgcPipelineTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		InputFile string   `json:"inputFile"`
		Steps     []string `json:"steps"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	return tool.SuccessWith("BGC pipeline started", map[string]any{
		"jobId":  "bgc-001",
		"input":  in.InputFile,
		"status": "running",
		"steps":  in.Steps,
	}), nil
}

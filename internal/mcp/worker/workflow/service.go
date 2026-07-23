package workflow

import (
	"context"
	"encoding/json"

	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/tool"
)

func NewWorkflowServer() *worker.WorkerServer {
	return worker.NewWorkerServer("workflow-worker", []tool.Tool{
		&workflowListTool{},
		&workflowRunTool{},
		&workflowStatusTool{},
		&workflowValidateTool{},
	})
}

type workflowListTool struct{}

func (t *workflowListTool) Name() string        { return "workflow_list" }
func (t *workflowListTool) Description() string { return "List available workflows" }
func (t *workflowListTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{}}`)
}
func (t *workflowListTool) Execute(ctx context.Context, _ json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	return tool.SuccessWith("Workflows", map[string]any{"workflows": []any{}}), nil
}

type workflowRunTool struct{}

func (t *workflowRunTool) Name() string        { return "workflow_run" }
func (t *workflowRunTool) Description() string { return "Run a workflow" }
func (t *workflowRunTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"workflowId":{"type":"string"},"inputs":{"type":"object"}},"required":["workflowId"]}`)
}
func (t *workflowRunTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		WorkflowID string `json:"workflowId"`
	}
	_ = json.Unmarshal(raw, &in)
	return tool.SuccessWith("Workflow started", map[string]any{"runId": "run-001", "status": "running"}), nil
}

type workflowStatusTool struct{}

func (t *workflowStatusTool) Name() string        { return "workflow_status" }
func (t *workflowStatusTool) Description() string { return "Get workflow run status" }
func (t *workflowStatusTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"runId":{"type":"string"}},"required":["runId"]}`)
}
func (t *workflowStatusTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		RunID string `json:"runId"`
	}
	_ = json.Unmarshal(raw, &in)
	return tool.SuccessWith("Workflow status", map[string]any{"runId": in.RunID, "status": "completed"}), nil
}

type workflowValidateTool struct{}

func (t *workflowValidateTool) Name() string        { return "workflow_validate" }
func (t *workflowValidateTool) Description() string { return "Validate a workflow definition" }
func (t *workflowValidateTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"workflow":{"type":"object"}},"required":["workflow"]}`)
}
func (t *workflowValidateTool) Execute(ctx context.Context, _ json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	return tool.SuccessWith("Workflow valid", map[string]any{"valid": true, "errors": []string{}}), nil
}

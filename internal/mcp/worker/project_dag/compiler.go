package projectdag

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/tool"
)

// NewProjectDAGServer creates the project-dag MCP worker.
func NewProjectDAGServer() *worker.WorkerServer {
	return worker.NewWorkerServer("project-dag-worker", []tool.Tool{
		&projectUpdateTool{},
		&projectGraphTool{},
		&projectReviewTool{},
	})
}

type projectUpdateTool struct{}

func (t *projectUpdateTool) Name() string { return "project_update" }
func (t *projectUpdateTool) Description() string {
	return "Compile thread evidence into a project graph"
}
func (t *projectUpdateTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"threadId":{"type":"string"}},"required":["threadId"]}`)
}
func (t *projectUpdateTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		ThreadID string `json:"threadId"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	return tool.SuccessWith("Project graph updated", map[string]any{
		"projectId": "proj-001",
		"nodes":     0,
		"edges":     0,
	}), nil
}

type projectGraphTool struct{}

func (t *projectGraphTool) Name() string        { return "project_graph" }
func (t *projectGraphTool) Description() string { return "Return the current project graph" }
func (t *projectGraphTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"projectId":{"type":"string"}},"required":["projectId"]}`)
}
func (t *projectGraphTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		ProjectID string `json:"projectId"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	return tool.SuccessWith("Project graph", map[string]any{
		"projectId": in.ProjectID,
		"nodes":     []any{},
		"edges":     []any{},
	}), nil
}

type projectReviewTool struct{}

func (t *projectReviewTool) Name() string        { return "project_review" }
func (t *projectReviewTool) Description() string { return "Request human review of a project node" }
func (t *projectReviewTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"nodeId":{"type":"string"},"reason":{"type":"string"}},"required":["nodeId"]}`)
}
func (t *projectReviewTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		NodeID string `json:"nodeId"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	return tool.SuccessWith("Review requested", map[string]any{
		"nodeId": in.NodeID,
		"status": "pending_review",
	}), nil
}

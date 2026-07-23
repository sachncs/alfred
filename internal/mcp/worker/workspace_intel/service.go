package workspaceintel

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/tool"
)

func NewWorkspaceIntelServer() *worker.WorkerServer {
	return worker.NewWorkerServer("workspace-intel-worker", []tool.Tool{
		&workspaceListTool{},
		&workspaceReadTool{},
		&workspaceVisualInspectTool{},
	})
}

type workspaceListTool struct{}

func (t *workspaceListTool) Name() string        { return "workspace_list" }
func (t *workspaceListTool) Description() string { return "List files in the workspace" }
func (t *workspaceListTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"}}}`)
}
func (t *workspaceListTool) Execute(ctx context.Context, _ json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	return tool.SuccessWith("Workspace listing", map[string]any{
		"entries": []map[string]any{},
	}), nil
}

type workspaceReadTool struct{}

func (t *workspaceReadTool) Name() string        { return "workspace_read" }
func (t *workspaceReadTool) Description() string { return "Read a file from the workspace" }
func (t *workspaceReadTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}`)
}
func (t *workspaceReadTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	return tool.SuccessWith("File content", map[string]any{
		"path":    in.Path,
		"content": "stub content",
	}), nil
}

type workspaceVisualInspectTool struct{}

func (t *workspaceVisualInspectTool) Name() string { return "workspace_visual_inspect" }
func (t *workspaceVisualInspectTool) Description() string {
	return "Visual inspection via model router"
}
func (t *workspaceVisualInspectTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}`)
}
func (t *workspaceVisualInspectTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	return tool.SuccessWith("Visual inspection", map[string]any{
		"path":        in.Path,
		"description": "Stub visual inspection",
	}), nil
}

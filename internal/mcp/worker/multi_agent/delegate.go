package multiagent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/tool"
)

// NewMultiAgentServer creates the multi-agent MCP worker.
func NewMultiAgentServer() *worker.WorkerServer {
	return worker.NewWorkerServer("multi-agent-worker", []tool.Tool{
		&delegateTaskTool{},
	})
}

type delegateTaskTool struct{}

func (t *delegateTaskTool) Name() string { return "delegate_task" }
func (t *delegateTaskTool) Description() string {
	return "Delegate a task to a sub-agent for parallel execution"
}
func (t *delegateTaskTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"task":{"type":"string"},"agentType":{"type":"string"}},"required":["task"]}`)
}
func (t *delegateTaskTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Task      string `json:"task"`
		AgentType string `json:"agentType"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.Task == "" {
		return tool.FailureMsg("task is required"), nil
	}
	return tool.SuccessWith("Task delegated", map[string]any{
		"taskId":    "delegate-001",
		"task":      in.Task,
		"agentType": in.AgentType,
		"status":    "queued",
	}), nil
}

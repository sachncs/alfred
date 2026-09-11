package plan_gateway

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sachncs/alfred/internal/mcp/worker"
	"github.com/sachncs/alfred/internal/tool"
)

// NewPlanGatewayWorker creates a plan-gateway MCP worker.
// ponytail: proxies to OpenCode for coding-plan generation.
func NewPlanGatewayWorker() *worker.WorkerServer {
	gw := &planGateway{}
	planTool := &planTool{gw: gw}
	return worker.NewWorkerServer("plan-gateway", []tool.Tool{planTool}).WithHandler(
		func(ctx context.Context, name string, input json.RawMessage, tc *tool.Context) (*tool.Result, error) {
			if name == "generate_plan" {
				return planTool.Execute(ctx, input, tc)
			}
			return tool.FailureMsg(fmt.Sprintf("unknown tool: %s", name)), nil
		},
	)
}

type planGateway struct{}

type planTool struct {
	gw *planGateway
}

func (p *planTool) Name() string { return "generate_plan" }

func (p *planTool) Description() string {
	return "Generate an implementation plan for a coding task. Returns a structured plan with steps, dependencies, and acceptance criteria."
}

func (p *planTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"task": {"type": "string", "description": "Description of the task to plan"},
			"context": {"type": "string", "description": "Additional context about the codebase"}
		},
		"required": ["task"]
	}`)
}

func (p *planTool) Execute(_ context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Task    string `json:"task"`
		Context string `json:"context"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.Task == "" {
		return tool.FailureMsg("task is required"), nil
	}

	// ponytail: generate a basic plan structure — full OpenCode integration in Phase 8
	plan := map[string]any{
		"task": in.Task,
		"steps": []map[string]string{
			{"action": "analyze", "description": "Analyze requirements and existing code"},
			{"action": "design", "description": "Design the implementation approach"},
			{"action": "implement", "description": "Implement the core changes"},
			{"action": "test", "description": "Write and run tests"},
			{"action": "verify", "description": "Verify build, vet, and lint pass"},
		},
		"acceptance": "All tests pass, build succeeds, no lint warnings",
	}
	b, _ := json.Marshal(plan)
	return tool.SuccessWith(fmt.Sprintf("Plan generated for: %s", in.Task), json.RawMessage(b)), nil
}

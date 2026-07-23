package schedule

import (
	"context"
	"encoding/json"

	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/tool"
)

func NewScheduleServer() *worker.WorkerServer {
	return worker.NewWorkerServer("schedule-worker", []tool.Tool{
		&scheduleCreateTool{},
		&scheduleRunTool{},
		&scheduleAuditTool{},
	})
}

type scheduleCreateTool struct{}

func (t *scheduleCreateTool) Name() string        { return "schedule_create" }
func (t *scheduleCreateTool) Description() string { return "Create a scheduled task" }
func (t *scheduleCreateTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"},"schedule":{"type":"string"},"kind":{"type":"string"}},"required":["name","schedule"]}`)
}
func (t *scheduleCreateTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Name     string `json:"name"`
		Schedule string `json:"schedule"`
		Kind     string `json:"kind"`
	}
	_ = json.Unmarshal(raw, &in)
	return tool.SuccessWith("Task created", map[string]any{"taskId": "sched-001", "name": in.Name}), nil
}

type scheduleRunTool struct{}

func (t *scheduleRunTool) Name() string        { return "schedule_run" }
func (t *scheduleRunTool) Description() string { return "Run a scheduled task immediately" }
func (t *scheduleRunTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"taskId":{"type":"string"}},"required":["taskId"]}`)
}
func (t *scheduleRunTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		TaskID string `json:"taskId"`
	}
	_ = json.Unmarshal(raw, &in)
	return tool.SuccessWith("Task triggered", map[string]any{"taskId": in.TaskID, "runId": "run-001"}), nil
}

type scheduleAuditTool struct{}

func (t *scheduleAuditTool) Name() string        { return "schedule_audit" }
func (t *scheduleAuditTool) Description() string { return "Audit task run history" }
func (t *scheduleAuditTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"taskId":{"type":"string"}}}`)
}
func (t *scheduleAuditTool) Execute(ctx context.Context, _ json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	return tool.SuccessWith("Audit complete", map[string]any{"runs": []any{}}), nil
}

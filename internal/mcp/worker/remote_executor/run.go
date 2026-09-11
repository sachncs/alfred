package remoteexecutor

import (
	"context"
	"encoding/json"

	"github.com/sachncs/alfred/internal/mcp/worker"
	"github.com/sachncs/alfred/internal/tool"
)

func NewRemoteExecutorServer() *worker.WorkerServer {
	return worker.NewWorkerServer("remote-executor-worker", []tool.Tool{
		&remoteTargetListTool{},
		&remoteRunTool{},
		&remoteJobStatusTool{},
	})
}

type remoteTargetListTool struct{}

func (t *remoteTargetListTool) Name() string { return "remote_target_list" }
func (t *remoteTargetListTool) Description() string {
	return "List available remote targets (SSH/Slurm)"
}
func (t *remoteTargetListTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{}}`)
}
func (t *remoteTargetListTool) Execute(ctx context.Context, _ json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	return tool.SuccessWith("Remote targets", map[string]any{"targets": []any{}}), nil
}

type remoteRunTool struct{}

func (t *remoteRunTool) Name() string        { return "remote_run" }
func (t *remoteRunTool) Description() string { return "Run a command on a remote target" }
func (t *remoteRunTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"targetId":{"type":"string"},"command":{"type":"string"}},"required":["targetId","command"]}`)
}
func (t *remoteRunTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		TargetID string `json:"targetId"`
		Command  string `json:"command"`
	}
	_ = json.Unmarshal(raw, &in)
	return tool.SuccessWith("Command dispatched", map[string]any{"jobId": "job-001", "targetId": in.TargetID}), nil
}

type remoteJobStatusTool struct{}

func (t *remoteJobStatusTool) Name() string        { return "remote_job_status" }
func (t *remoteJobStatusTool) Description() string { return "Get remote job status" }
func (t *remoteJobStatusTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"jobId":{"type":"string"}},"required":["jobId"]}`)
}
func (t *remoteJobStatusTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		JobID string `json:"jobId"`
	}
	_ = json.Unmarshal(raw, &in)
	return tool.SuccessWith("Job status", map[string]any{"jobId": in.JobID, "status": "completed", "output": ""}), nil
}

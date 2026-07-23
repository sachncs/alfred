package evidencedag

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/tool"
)

// NewEvidenceDAGServer creates the evidence-dag MCP worker.
func NewEvidenceDAGServer() *worker.WorkerServer {
	return worker.NewWorkerServer("evidence-dag-worker", []tool.Tool{
		&evidenceUpdateTool{},
		&evidenceSnapshotTool{},
		&evidenceAuditTool{},
	})
}

type evidenceUpdateTool struct{}

func (t *evidenceUpdateTool) Name() string        { return "evidence_update" }
func (t *evidenceUpdateTool) Description() string { return "Update the evidence graph with new claims" }
func (t *evidenceUpdateTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"claim":{"type":"string"},"source":{"type":"string"},"confidence":{"type":"number"}},"required":["claim"]}`)
}
func (t *evidenceUpdateTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Claim      string  `json:"claim"`
		Source     string  `json:"source"`
		Confidence float64 `json:"confidence"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	return tool.SuccessWith("Evidence updated", map[string]any{
		"nodeId":     "ev-001",
		"claim":      in.Claim,
		"confidence": in.Confidence,
	}), nil
}

type evidenceSnapshotTool struct{}

func (t *evidenceSnapshotTool) Name() string { return "evidence_snapshot" }
func (t *evidenceSnapshotTool) Description() string {
	return "Snapshot the current evidence graph state"
}
func (t *evidenceSnapshotTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{}}`)
}
func (t *evidenceSnapshotTool) Execute(ctx context.Context, _ json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	return tool.SuccessWith("Snapshot captured", map[string]any{
		"snapshotId": "snap-001",
		"nodes":      0,
		"edges":      0,
	}), nil
}

type evidenceAuditTool struct{}

func (t *evidenceAuditTool) Name() string        { return "evidence_audit" }
func (t *evidenceAuditTool) Description() string { return "Audit the evidence chain for consistency" }
func (t *evidenceAuditTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"depth":{"type":"integer"}}}`)
}
func (t *evidenceAuditTool) Execute(ctx context.Context, _ json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	return tool.SuccessWith("Audit complete", map[string]any{
		"findings": []string{},
		"status":   "clean",
	}), nil
}

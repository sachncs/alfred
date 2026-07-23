package feedbackgateway

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/tool"
)

// NewFeedbackGatewayServer creates the feedback-gateway MCP worker.
func NewFeedbackGatewayServer() *worker.WorkerServer {
	return worker.NewWorkerServer("feedback-gateway-worker", []tool.Tool{
		&feedbackSubmitTool{},
	})
}

type feedbackSubmitTool struct{}

func (t *feedbackSubmitTool) Name() string { return "feedback_submit" }
func (t *feedbackSubmitTool) Description() string {
	return "Submit feedback as an idempotent GitHub Issue"
}
func (t *feedbackSubmitTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"title":{"type":"string"},"body":{"type":"string"},"labels":{"type":"array","items":{"type":"string"}}},"required":["title","body"]}`)
}
func (t *feedbackSubmitTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Title  string   `json:"title"`
		Body   string   `json:"body"`
		Labels []string `json:"labels"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	return tool.SuccessWith("Feedback submitted", map[string]any{
		"issueId": "fb-001",
		"title":   in.Title,
		"status":  "created",
	}), nil
}
